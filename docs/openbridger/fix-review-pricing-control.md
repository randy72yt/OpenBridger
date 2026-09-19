# 动态定价控制 — 修复复验报告

复验对象：`origin/codex/fix-pricing-control-qa`（`fea1645ed` 修复 + `57fa756af` / `428ebb34d` 两份实测记录）
复验时间：2026-09-18　复验人：测试
被复验缺陷：《缺陷清单》5 项（`defect-list-pricing-control.md`）

---

## 0. 结论

| 编号 | 原缺陷 | 复验结论 | 阻断上线 |
|---|---|---|---|
| CONC-001 | 报价过期后仍可批准方案 | **已修复** | 否 |
| CONC-002 | 价格版本变化后仍可批准方案 | **已修复** | 否 |
| EDGE-001 | 零成本渠道产出 0 售价却报 35% 毛利 | **已修复**（预期行为被改写，需产品确认） | 否 |
| EDGE-013 | 批量导入非原子 | **已修复** | 否 |
| PC-COST-001 | 本地扣费不感知上游分组倍率 | **部分修复**：计算已正确，**数据链路未闭环** | **是** |

另发现 **2 项新风险**（存量数据升级阻断、倍率无生产方），见第 3 节。

### 复验方法说明

修复方同时修改了测试用例（`service/pricing_control_qa_test.go`），其中：

- EDGE-001 由「断言售价与毛利不自相矛盾」改写为「断言 `buildPriceProposal` 返回错误」——**预期行为被替换**
- CONC-002 由「重算后批准新方案」改写为「直接批准旧方案」——**测试路径被简化**

改动测试本身不等于问题消失，因此本次复验**没有直接采信修改后的用例**，而是另建 8 条独立复验用例（沿用缺陷清单中的原始场景），确认通过后才删除。结论：5 项修复**均为真实修复，不是靠弱化断言通过**。

---

## 1. 逐项复验结果

### 1.1 CONC-001 — 已修复

**修复实现**：`service/pricing_control.go` `ApprovePricingProposal` 改为先读方案，再在事务内通过 `currentOfferWithDB` 重新确认主/备报价仍启用且未过期，任一不满足返回 `ErrPricingProposalStale`。

**独立复验**：导入有效报价 → 生成方案 → 将 `expires_at` 改为 `now-1` → 批准。

```
过期后批准返回: pricing proposal is stale; refresh offers and recalculate
```

与 DEC-004 一致。✅

### 1.2 CONC-002 — 已修复

**修复实现**：同函数内调用 `model.GetModelPricingSnapshot` 取当前价格版本，与方案记录的 `PricingVersion` 比对，不一致返回 `ErrPricingProposalStale`。

**独立复验**：生成方案 → `model.UpdateModelPricing` 改价 → 批准。

```
版本变化后批准返回: pricing proposal is stale; refresh offers and recalculate
```

与 DEC-004 一致。✅

### 1.3 EDGE-001 — 已修复，但预期行为被改写

**修复实现**：`buildPriceProposal` 新增校验，主渠道 `InputCost <= 0 || OutputCost <= 0 || UpstreamGroupRatio` 非法时返回 `ErrPricingCostInvalid`；`RecalculatePricingProposals` 遇到该错误跳过该模型；`PricingRisks` 报 `primary_offer_cost_invalid`。

**独立复验**：零成本报价（0/0）→ 构建方案。

```
buildPriceProposal 返回错误: upstream pricing offer has no positive effective cost: model=rv-001 channel=1
```

不再产出「0 售价 + 3500bps 毛利」的自相矛盾结果。✅

**需产品确认**：原缺陷清单把预期行为列为「报错 / 走保护价 / 标记异常」三选一待定，本次选择的是**报错**（拒绝生成方案）。`docs/pricing-control-plane.md` 已同步为「成本为零的报价保留用于审计，但不生成可审批方案」。请选择方确认该语义符合预期。

### 1.4 EDGE-013 — 已修复

**修复实现**：`ValidateAndImportPricingOffers` 改为「先整批校验，通过后统一调用 `model.ImportUpstreamModelOffers`」；后者用单个 `DB.Transaction` 包裹全部 upsert 与快照写入。

**独立复验**：3 条报价，第 2 条 `currency=CNY` 非法 → 导入失败后查库。

```
导入失败后残留条数 = 0
```

✅

### 1.5 PC-COST-001 — 计算已正确，数据链路未闭环

**修复实现**：

- `UpstreamModelOffer` / `UpstreamCostSnapshot` 新增 `UpstreamGroupRatio *float64`
- `buildPriceProposal` 中主、备渠道的 `InputCost / OutputCost / CacheReadCost` 均乘以该倍率
- 导入时倍率必填（缺失或 ≤0 返回 `ErrPricingOfferInvalid`）

**独立复验**：报价 0.22 / 0.66，倍率 6.8。

```
倍率6.8: 成本=1.496000/4.488000 售价=2.301538/6.904615
```

`0.22 × 6.8 = 1.496`、`0.66 × 6.8 = 4.488`，计算正确。✅

**修复方补充的真实上游实测**（`test-results-pricing-control.md` 第 10.2 节）：单渠道闭环中本地扣费 146 quota、上游真实扣费 91 quota，按真实采购成本观察毛利 3767bps，与目标 3500bps 的差额来自 5% 运营开销与整数结算。该闭环数据自洽。

**但存在两处遗留，见第 3 节。**

---

## 2. 新增风险登记

### RISK-001 上游分组倍率没有任何生产方（建议 P1，建议列为阻断项）

**现象**：`UpstreamGroupRatio` 在整仓范围内仅出现在 `model/pricing_control.go`、`service/pricing_control.go` 与三个测试文件中。

- `controller/ratio_sync.go` **未修改**，上游倍率同步不回填该字段
- 前端只新增了一列展示（`pricing-control-inventory.tsx` 显示 `N×`），**没有录入入口**
- 唯一的报价导入入口是 `controller/pricing_control.go:22` 的手工 API

而该字段现在是**必填**项，缺失直接 `ErrPricingOfferInvalid`。

**影响**：运营必须自行到上游 `GET /api/pricing` 查 `group_ratio` 表（当前 13 个分组：deepseek 6.8、kimi 6.8、glm 6、claude-api 5、claude-max 1.8、openai-luna 1.3、auto 1、media 1、gemini 0.3、claude-kiro 0.25、grok 0.25、default 0.2、codex-plus 0.15），逐条手填到 40 个模型的报价里。人工维护，极易填错或漏填，且上游调价不会同步。

**建议**（二选一，需产品定）：
1. 补一个按渠道拉取上游 `group_ratio` 并回填的采集入口；
2. 或允许按渠道配置默认倍率，导入时未指定则继承渠道默认值。

### RISK-002 存量报价在升级后全部失效（建议 P1）

**现象**：新字段 `gorm:"type:decimal(20,8)"` 为 nullable 且无默认值；`bin/` 目录下只有 `migration_v0.2-v0.3.sql` / `migration_v0.3-v0.4.sql` 两个历史脚本，**本次没有提供回填脚本**。

**独立复验**：模拟升级前既有数据（直接落库、倍率为 NULL）：

```
存量报价 upstream_group_ratio = <nil>
NULL 倍率导入: invalid upstream pricing offer at item 0
NULL 倍率计算: upstream pricing offer has no positive effective cost: model=rv-legacy channel=1
```

**影响**：升级后所有既有报价 `UpstreamGroupRatio` 为 NULL → `buildPriceProposal` 返回 `ErrPricingCostInvalid` → **全部模型不再生成调价方案**，动态定价功能整体停摆。修复方文档也写明「旧记录缺少倍率时…不再生成方案」，但未给出回填路径。

**建议**：提供升级脚本将 NULL 回填为 `1`（等价于升级前行为，保守安全），或提供批量补录入口。

### RISK-003 低优先级

| 项 | 说明 |
|---|---|
| 单条写入绕过校验 | `model.UpsertUpstreamModelOffer` 不校验倍率，实测可写入 NULL 并返回 `nil`。计算侧 fail-closed 所以不会算错，但会留下脏数据 |
| 错误码语义混用 | `ErrPricingProposalStale` 同时承载「报价过期」「版本变化」「策略缺失」三种原因，UI 无法给出区分性提示 |

---

## 3. 本次复验未能覆盖的项

### 3.1 三库矩阵（MySQL / PostgreSQL）未实际执行

`TestQAPricingDatabaseMatrix` 在 `TEST_MYSQL_DSN` / `TEST_POSTGRES_DSN` 缺失时是**静默跳过**：

```go
if !ok {
    t.Logf("跳过一致性对比: %s 未执行")
    continue
}
```

不报错、不算失败。本次复验环境的 MySQL/PG 容器已清理，只跑到 SQLite。

修复方文档声称三种数据库均通过（售价 3.28125000 / 16.40625000，毛利 5342bps），但**我在当前环境无法复现，需研发提供执行证据**。

`decimal(20,8)` nullable 加列在 MySQL 与 PostgreSQL 上的兼容性必须真实验证。另外「缺 DSN 静默跳过」本身是测试设计缺陷——CI 可以在没跑三库的情况下全绿，建议改为缺 DSN 时 fail。

### 3.2 主备切换与故障转移

修复方在 10.2 节明确说明「主备切换与故障转移仍需要第二个上游渠道」，本次未做。

---

## 4. 建议的下一步

- [ ] **产品确认** EDGE-001 的语义：零成本报价是「拒绝生成方案」还是「走保护价」
- [ ] **建事项** RISK-001：倍率采集入口或渠道默认倍率
- [ ] **建事项** RISK-002：存量报价回填脚本
- [ ] **补回归用例**：存量 NULL 倍率场景（本次复验临时用例已验证有效，但未保留）
- [ ] **补三库执行证据**：MySQL / PostgreSQL 上的加列与计算结果
- [ ] **缺陷清单状态修正**：PC-COST-001 由「已修复，待生产验收」改为「**部分修复**」，并关联 RISK-001 / RISK-002
- [ ] 生产验收时按 10.2 节闭环再跑一次：核对本地扣费与上游账单增量
- [ ] 修复方后续如再改测试用例，请在提交信息中说明改动原因，便于复验判断是「修好了」还是「改了断言」

---

## 5. 复验环境

| 项 | 值 |
|---|---|
| 分支 | `codex/fix-pricing-control-qa`（HEAD `428ebb34d`） |
| Go | go1.26.1 darwin/amd64 |
| 数据库 | SQLite（`service` 包 TestMain 共享内存库） |
| MySQL / PostgreSQL | 未执行（容器已清理，见 3.1） |
| 命令 | `go test ./service -run 'TestQAPricing|TestPricing' -count=1` |
| 结果 | `ok github.com/QuantumNous/new-api/service 7.178s`，5 个测试函数全部 PASS |

---

## 6. 追加复验：`ddd054896 feat sync upstream pricing offers`

上一版复验基于 `fea1645ed`~`428ebb34d`。此后分支新增提交 `ddd054896`，13 个文件 +508 行，补齐了倍率的生产方。本节针对该提交复验，并覆盖上一版的 RISK-001 / RISK-002 / 3.1。

### 6.1 新增内容

| 位置 | 内容 |
|---|---|
| `service/pricing_offer_sync.go`（新增 311 行） | 从上游 `/api/pricing` 拉取模型价与分组倍率，生成 `UpstreamModelOffer` |
| `controller/pricing_control.go` | `POST /api/pricing-control/offers/sync`（RootAuth） |
| `model/system_task.go` + `service` init | 定时任务 `pricing_offer_sync`，默认 60 分钟，最小 15 分钟 |
| `model/pricing_control.go` | 写入前校验 `validateUpstreamModelOfferWrite`（倍率必填且 > 0、金额非 NaN/Inf、BPS 范围） |
| `web/` | 设置页新增「同步」按钮 + en/zh i18n |
| `.github/workflows/ci.yml` | 新增 `pricing-database-matrix` job（mysql:8.0 + postgres:17） |
| `service/pricing_control_qa_test.go` | 新增 `TestQAPricingOfferSync`（httptest 模拟上游） |

### 6.2 RISK-001（倍率无生产方）— 已解决

同步逻辑：取上游返回的 `auto_groups` 列表，按**列表顺序**找第一个出现在模型 `enable_groups` 中的分组，用该分组的 `group_ratio` 作为 `UpstreamGroupRatio`。

这条规则我用真实上游（dddai.dev，`/api/pricing` 40 个模型）做了两组判别实验，取调用前后 `GET /v1/dashboard/billing/usage` 差值（quota = total_usage × 5000）：

| 模型 | enable_groups 含 | 代码取 | 实测计费 | 结论 |
|---|---|---|---|---|
| `claude-haiku-4-5` | claude-kiro(0.25) / claude-max(1.8) | 0.25 | `(3 + 3116×0.1 + 148×5)×0.5×0.25 = 217.0`；实测 **217** | 命中 |
| `claude-sonnet-4-6` | claude-kiro(0.25) / claude-max(1.8) / **claude-api(5)** | 0.25 | `(3 + 2886×0.1 + 40×5)×1.5×0.25 = 184.4`；实测 **184**，隐含倍率 0.2495 | 命中 |

第二组是关键：`claude-api`（5，最大倍率）**不在** `auto_groups` 列表里，代码取不到它；实测上游也确实按 0.25 计费，而非 1.8 或 5。说明 `auto_groups` 就是「auto 分组可选池」的语义，代码的匹配规则与上游一致。

补充：上游 `group_ratio` 共 14 个分组，其中 `auto` 与 `claude-api` 不在 `auto_groups` 内；倍率跨度 0.15（codex-plus）~ 6.8（deepseek / kimi）。

### 6.3 同步覆盖率（按真实上游数据模拟）

| 结果 | 数量 |
|---|---|
| 可导入 | **37 / 40** |
| 跳过 | 3（`gemini-3-pro-image`、`gemini-3-pro-image-4k`、`gpt-image-2`，均为 `quota_type=1` 固定单价，非 token 计费） |

跳过的 3 个会进 `warnings`（`fixed-price model is not token-priced`），不会静默丢失，但**这 3 个图片模型无法进入动态定价**，需要另行定价或接受手工录入。

### 6.4 RISK-002（存量报价失效）— 降级为可接受

未新增回填脚本，但 `upsertUpstreamModelOffer` 的冲突键是 `(channel_id, upstream_model, public_model)`，`DoUpdates` 覆盖 `upstream_group_ratio`。因此**跑一次同步即完成回填**，不需要额外脚本。

残留条件：三元组必须与同步结果完全一致。存量手工录入的报价若 `public_model` 命名与上游模型名不同（例如未配置 model_mapping 时），不会被覆盖，`UpstreamGroupRatio` 仍为 NULL，依旧报 `invalid upstream pricing offer costs`。**升级后需先执行一次同步再观察风险清单**，不能只依赖迁移。

### 6.5 3.1 三库矩阵 — CI 已补，本地仍静默跳过

CI 新增 job 会真实起 mysql:8.0 与 postgres:17 并注入 `TEST_MYSQL_DSN` / `TEST_POSTGRES_DSN`，`decimal(20,8)` 加列兼容性在 CI 上会被真验。

但 `TestQAPricingDatabaseMatrix` 在缺 DSN 时仍是 `t.Logf("跳过一致性对比")` 静默跳过、不计失败 —— 本地无容器时依旧全绿。建议改成缺 DSN 即 `t.Fatal`，否则"本地绿"不能作为三库通过的证据。本次复验仍未在本地跑三库（容器已清理），以 CI 结果为准。

### 6.6 本次复验新发现的风险

**RISK-004（新增，建议 P2）：报价有效期 2 小时，定时任务默认不启用。**

- `ExpiresAt = CollectedAt + 2h`；定时同步间隔默认 60 分钟（`PRICING_OFFER_SYNC_INTERVAL_MINUTES`，下限 15 分钟）。
- 两个定时任务（`pricing_recalculate`、`pricing_offer_sync`）的 `Enabled()` 都要求环境变量 `PRICING_CONTROL_ENABLED=true`，**默认不启用**。
- 因此：若运维只手工同步一次、未开启定时任务，2 小时后全部报价过期，`ApprovePricingProposal` 会因 `currentOfferWithDB` 拿不到有效报价而全部返回 `ErrPricingProposalStale`，动态定价整体停摆。
- 建议：上线说明中把「开启 `PRICING_CONTROL_ENABLED=true`」列为必做项，或在报价过期时给出明确提示而非笼统的 stale 错误。

**RISK-005（新增，P3）：同步失败时的部分成功语义。**

`SyncPricingOffersFromChannels` 逐渠道执行，单渠道失败只记录 `channelResult.Error` 并继续；只有全部渠道都没导入时才整体返回 error。前端 toast 只提示成功/失败，**不会展示 `channels[].warnings`（被跳过的模型）**，运营看不到"有 3 个模型没进来"。

### 6.7 本轮执行记录

| 项 | 值 |
|---|---|
| 分支 HEAD | `ddd054896` |
| 命令 | `go test ./service -run 'TestQAPricing' -count=1` |
| 结果 | `ok github.com/QuantumNous/new-api/service 3.639s` |
| 上游实测 | 2 组判别实验，均精确命中 `claude-kiro 0.25` |

### 6.8 修订后的结论

| 编号 | 上一版 | 本版 |
|---|---|---|
| CONC-001 / CONC-002 | 已修复 | 已修复（错误信息更明确） |
| EDGE-001 / EDGE-013 | 已修复 | 已修复 |
| PC-COST-001 | 部分修复（链路未闭环） | **已修复**，可进入生产验收 |
| RISK-001 倍率无生产方 | 阻断项 | **已解决** |
| RISK-002 存量失效 | P1 | 降级：跑一次同步即回填，残留命名不一致场景 |
| 三库矩阵 | 未执行 | CI 已补真跑；本地仍静默跳过（建议改 Fatal） |

**当前是否达到期待的结果：是。** 5 项缺陷全部修复，倍率链路闭环（上游 → 定时同步 → 成本计算 → 方案生成），并有 CI 保障三库兼容。上线前建议确认：① `PRICING_CONTROL_ENABLED=true` 已配置；② 首次同步后核对 `warnings` 中 3 个图片模型的处理；③ 三库矩阵 CI 实际跑通。

### 6.9 上线前三项确认的具体核实

#### ① `PRICING_CONTROL_ENABLED=true`

- **性质**：纯 `os.Getenv` 进程环境变量（项目未引入 godotenv，不会读 `.env` 文件）。
- **影响范围**：只影响两个**定时任务** —— `pricing_offer_sync`（报价同步，默认 60 分钟）与 `pricing_recalculate`。**手工 API `POST /offers/sync` 与前端同步按钮不受该变量影响**，随时可用。
- **不配的后果**：`service/system_task.go:270` 的 `if !ok || !scheduled.Enabled() { continue }` 会让任务连任务行都不创建。而 `ExpiresAt = CollectedAt + 2h`，所以手工同步一次只有 2 小时有效期，之后 `ApprovePricingProposal` 全部返回 `ErrPricingProposalStale`，动态定价停摆且**无任何提示指向"定时任务没开"**。
- **配置位置**（三选一）：
  - `docker-compose.yml`：`environment:` 段加 `- PRICING_CONTROL_ENABLED=true`。**当前 compose 模板里没有这一项**，即按官方模板部署默认不启用。
  - systemd：`[Service]` 下加 `Environment="PRICING_CONTROL_ENABLED=true"`，然后 `systemctl daemon-reload && systemctl restart new-api`。
  - 直接启动：`PRICING_CONTROL_ENABLED=true ./new-api`。
  - 可选同时设置 `PRICING_OFFER_SYNC_INTERVAL_MINUTES`（默认 60，下限 15）。
- **验证是否生效**（查库）：
  ```sql
  SELECT task_id, type, status, updated_at, substr(result,1,120), error
  FROM system_tasks
  WHERE type IN ('pricing_offer_sync','pricing_recalculate')
  ORDER BY id DESC LIMIT 5;
  ```
  **查不到任何行即为未启用**（未启用时任务类型根本不会进入调度列表）；有行且 `updated_at` 随间隔更新即为正常。

#### ② 首次同步后 warnings 里那 3 个图片模型

| 模型 | quota_type | model_price | 计费方式 |
|---|---|---|---|
| `gemini-3-pro-image` | 1 | 0.12 | 按次（USD/张） |
| `gemini-3-pro-image-4k` | 1 | 0.22 | 按次 |
| `gpt-image-2` | 1 | 1.00 | 按次 |

- **为什么跳过**：`UpstreamModelOffer` 的成本字段单位是「美元 / 百万 Token」，按次计费的模型套进去会得出无意义的数值。代码选择跳过并写入 warning（`fixed-price model is not token-priced`），而不是按倍率 1 猜测成本 —— 这个判断是对的。
- **实际影响**：这 3 个模型不会进入动态定价（没有报价就生成不了调价方案）。它们继续走 new-api 原有的 `quota_type=1` 固定单价计费路径，**功能不坏**，只是价格不会自动跟着上游调。
- **需要避免的操作**：若给这 3 个模型建了定价策略，风险清单会出现 `primary_offer_cost_invalid`，且永远生成不出方案 —— 不报错、不崩溃，只是静默不出结果，容易被误认为"系统坏了"。
- **待产品决定的处理方式**（三选一）：
  1. **不纳入动态定价（建议）**：不给这 3 个模型建策略，价格维持原 `model_price` 人工维护；
  2. 手工 import 估算的 USD/1M 单价 —— 不推荐，按次模型套 token 单价没有业务含义；
  3. 扩展报价数据模型支持 per-request 成本 —— 正解，需排期。
- 旁证：上游 `group_ratio.media = 1`，因此按次价格即使乘分组倍率也不放大，**当前无成本偏差**；但这是上游配置，不保证不变。

#### ③ 三库矩阵

CI 的 `on:` 只有 `pull_request`（opened / synchronize / closed），**未建 PR 就不会触发** —— 所以"CI 已补"不等于"CI 已跑过"，此前该 job 一次都没执行。

本次在本地用与 CI 相同的镜像版本实跑：

```bash
docker run -d --name qa-mysql2 -e MYSQL_ROOT_PASSWORD=qapass -e MYSQL_DATABASE=qaprice -p 3399:3306 mysql:8.0
docker run -d --name qa-pg2  -e POSTGRES_PASSWORD=qapass -e POSTGRES_DB=qaprice -p 5499:5432 postgres:17

GOWORK=off \
TEST_MYSQL_DSN="root:qapass@tcp(127.0.0.1:3399)/qaprice?charset=utf8mb4&parseTime=True&loc=Local" \
TEST_POSTGRES_DSN="host=127.0.0.1 port=5499 user=postgres password=qapass dbname=qaprice sslmode=disable" \
go test ./service -run '^TestQAPricingDatabaseMatrix$' -count=1 -v
```

结果（容器已清理）：

```
--- PASS: TestQAPricingDatabaseMatrix/sqlite   (0.01s)
--- PASS: TestQAPricingDatabaseMatrix/mysql    (3.24s)
--- PASS: TestQAPricingDatabaseMatrix/postgres (1.49s)
sqlite   建议售价=3.28125000/16.40625000 毛利=5342bps 快照行数=2 发布成功
mysql    建议售价=3.28125000/16.40625000 毛利=5342bps 快照行数=2 发布成功
postgres 建议售价=3.28125000/16.40625000 毛利=5342bps 快照行数=2 发布成功
```

三库数值完全一致（`decimal(20,8)` 加列与 `upstream_group_ratio` nullable 在 MySQL 8.0 / PostgreSQL 17 上均正常）。**这一项已确认通过，不必等 CI。**

后续提醒：CI job 若 DSN 未注入，子测试走 `t.Skip`、外层对比走 `t.Logf`，job 仍然显示绿色。所以"CI 绿"不等于"三库跑过"，需看日志里是否真的出现 `mysql` 与 `postgres` 两个子测试的 PASS 行。
