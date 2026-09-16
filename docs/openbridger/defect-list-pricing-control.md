# OpenBridger 动态定价控制 — 缺陷清单

分支：`test/openbridger-qa`　整理时间：2026-09-16　缺陷总数：**5**

本文档只做缺陷登记与定级依据。**不含修复方案**；是否修复、如何修复、由谁修复由研发与产品决定。

配套文档：
- 用例设计：`docs/openbridger/test-cases-pricing-control.md`
- 执行结果：`docs/openbridger/test-results-pricing-control.md`（含第 9 节真实成本实测）

---

## 0. 缺陷总览

| 编号 | 标题 | 严重程度 | 优先级 | 阻断上线 | 类型 |
|---|---|---|---|---|---|
| PC-COST-001 | 本地扣费不感知上游分组倍率，毛利双向失真（最高 45 倍） | S1 | P0 | **是** | 计费/财务 |
| CONC-001 | 报价过期后仍可批准方案（违反 DEC-004） | S2 | P1 | 是 | 流程管控 |
| CONC-002 | 价格版本变化后仍可批准方案（违反 DEC-004） | S2 | P1 | 是 | 流程管控 |
| EDGE-001 | 零成本渠道产出 0 售价，同时报告 35% 毛利 | S3 | P1 | 否 | 数据自洽 |
| EDGE-013 | 批量导入遇非法项时留下部分入库，非原子 | S3 | P2 | 否 | 数据一致性 |

严重程度口径：S1=财务或安全后果不可逆；S2=违反已明确约定的设计约束；S3=结果自相矛盾或数据不一致，可通过人工流程规避。

---

## 1. PC-COST-001

**标题**：本地扣费不感知上游分组倍率，动态定价毛利双向失真
**严重程度**：S1　**优先级**：P0　**阻断上线**：是
**类型**：计费 / 财务
**状态**：新建　**发现方式**：真实上游端到端实测

### 环境

| 项 | 值 |
|---|---|
| 本地实例 | OpenBridger（new-api 分支），SQLite 隔离库，端口 3211 |
| 上游 | `https://api.dddai.dev/v1`（其本身也是 new-api 实例） |
| 上游 key 分组 | `auto`（自动分组） |
| 本地分组倍率 | 1（默认） |
| 计费口径 | 本地与上游均已同步同一份 `billing_expr` / `model_ratio` |

### 前置条件

1. 已完成上游倍率同步（`POST /api/ratio_sync/fetch` + `PATCH /api/option/model_pricing`），本地倍率与上游一致；
2. 本地分组倍率为默认 1；
3. 上游 key 位于 `auto` 组。

### 复现步骤

1. 记录上游 `GET /v1/dashboard/billing/usage` 的 `total_usage`（换算：`quota = total_usage × 5000`）；
2. 记录本地 `/api/user/self` 的 `used_quota`；
3. 经本地实例调用 `deepseek-v4-flash`，prompt 85 / completion 19；
4. 重复步骤 1、2 取差值；
5. 换 `gemini-3.1-flash-lite`，prompt 2201 / completion 1033 再测一轮。

### 实际结果

| 模型 | 上游分组倍率 | 本地扣费 | 上游真实扣费 | 偏差 |
|---|---|---|---|---|
| deepseek-v4-flash | 6.8（deepseek 组） | 16 quota（$0.000032） | 106 quota（$0.000212） | **少收 6.8 倍** |
| deepseek-v4-flash（第二次） | 6.8 | 32 quota | 220 quota | 少收 6.9 倍 |
| gemini-3.1-flash-lite | 0.3（gemini 组） | 1325 quota（推算） | 398 quota | **多收 3.3 倍** |

### 期望结果

本地扣费应能反映真实上游成本，毛利报表与真实毛利方向一致。

### 依据

统一公式（两种计费口径均已实测验证，误差 < 0.5%）：

```
上游真实扣费 = 定价基准 × 该模型所属分组的 group_ratio
  定价基准 = (prompt + completion × completion_ratio) × model_ratio   # ratio 模式
          或 billing_expr 结果 / 1e6 × QuotaPerUnit                    # tiered_expr 模式
```

验算：

- deepseek：`(85×0.22 + 19×0.66) / 1e6 × 500000 = 15.62`；`15.62 × 6.8 = 106.2` → 实测 106 ✅
- gemini：`(2201 + 1033×3) × 0.25 = 1325`；`1325 × 0.3 = 397.5` → 实测 398 ✅
- `auto` 组 key 调 deepseek 按 6.8 计、调 gemini 按 0.3 计，说明 `auto` 按目标模型所属分组结算，不是按 `auto` 自身的 1

上游 `group_ratio` 全表（取自 `GET /api/pricing`）：

```
deepseek 6.8   kimi 6.8   glm 6   claude-api 5   claude-max 1.8
openai-luna 1.3   auto 1   media 1   gemini 0.3   claude-kiro 0.25
grok 0.25   default 0.2   codex-plus 0.15   free 0
```

最高 6.8 / 最低非零 0.15 = **约 45 倍**。

根因（代码层，待研发确认）：`ratio_sync` 同步了 `model_ratio / completion_ratio / cache_ratio / create_cache_ratio / model_price / billing_mode / billing_expr`，**未同步 `group_ratio`**；本地分组倍率与上游分组倍率相互独立。

### 影响范围

- 高价组（deepseek / kimi 6.8、glm 6、claude-api 5）：实际亏损，报表仍显示正毛利；
- 低价组（gemini 0.3、claude-kiro 0.25、grok 0.25、default 0.2、codex-plus 0.15）：实际毛利远高于报表，售价虚高；
- `pricing_control` 的全部毛利分析、成本预警、定价建议均建立在失真的成本基线上。

### 定级理由

财务后果不可逆且**双向**（既可能少收也可能多收），单模型偏差最高 45 倍，无法通过人工复核覆盖（报表本身是错的，人工看报表看不出问题）。

### 附件 / 关联

- 关联文档：`test-results-pricing-control.md` 第 9 节（9.1–9.7）
- 关联用例：PC-E2E（真实上游端到端）

---

## 2. CONC-001

**标题**：报价过期后仍可批准方案，违反 DEC-004
**严重程度**：S2　**优先级**：P1　**阻断上线**：是
**类型**：流程管控
**状态**：新建　**发现方式**：服务层自动化测试

### 环境

Go 1.26.1 / SQLite（`service` 包 TestMain，内存库）／MySQL 8.0.46／PostgreSQL 17.11 三库均复现。

### 前置条件

已导入一条有效上游报价，并据此生成 pending 状态方案。

### 复现步骤

1. 导入有效报价 → 生成 pending 方案；
2. 将该报价的 `expires_at` 改为 `now - 1`（已过期）；
3. 调用 `ApprovePricingProposal`。

### 实际结果

返回 `nil`，**批准成功**。

### 期望结果

批准被拒绝，返回明确错误，并要求刷新报价。依据：
- `docs/openbridger/project-profile.md` DEC-004「过期或价格版本变化的方案禁止审批」
- `production-acceptance.md` BILL-004「过期报价：阻止发布并要求刷新，旧价格不变」

### 依据

`service/pricing_control.go` 第 316-330 行：`ApprovePricingProposal` 仅校验 `status = pending`，未检查关联报价的有效期。

### 影响范围

过期价格可被批准并发布上线，用户按失效成本定价。测试 `STATE-*` 系列未覆盖"过期"这一条件，属于审批入口的校验缺口。

### 定级理由

违反需求文档已明确写死的设计约定（DEC-004），不是边界情况而是明确规定未实现。

---

## 3. CONC-002

**标题**：价格版本变化后仍可批准方案，违反 DEC-004
**严重程度**：S2　**优先级**：P1　**阻断上线**：是
**类型**：流程管控
**状态**：新建　**发现方式**：服务层自动化测试

### 环境

同 CONC-001。

### 前置条件

已生成 pending 方案，方案中记录了当时的 `PricingVersion`。

### 复现步骤

1. 生成 pending 方案（记录 `PricingVersion`）；
2. 通过 `model.UpdateModelPricing` 修改该模型价格，使价格版本发生变化；
3. 调用 `ApprovePricingProposal`。

### 实际结果

返回 `nil`，**批准成功**。

### 期望结果

批准被拒绝并提示需重新生成方案。依据同 DEC-004。

### 依据

`service/pricing_control.go` 第 316-330 行，审批入口未使用版本变化标记。而 `PricingChanged` 字段已在 `ListPricingProposalReviews` 第 112 行计算出来——**列表页能看到"价格已变化"，但审批接口不拦**。

### 影响范围

方案基于旧价格快照生成，批准后按旧成本发布，与当前实际成本脱节。与 CONC-001 是同一处代码缺口的两个触发条件，建议合并修复。

### 定级理由

同 CONC-001；且 UI 已提示、后端未拦截，属前后端校验不一致，用户会误以为已被拦住。

---

## 4. EDGE-001

**标题**：零成本渠道产出 0 售价，同时报告 35% 毛利
**严重程度**：S3　**优先级**：P1　**阻断上线**：否
**类型**：数据自洽
**状态**：新建　**发现方式**：服务层自动化测试

### 环境

同 CONC-001。

### 前置条件

主渠道成本为 0/0，策略：`default` 档、`TargetMarginBPS=3500`、`MinimumMarginBPS=2000`、`OverheadBPS=500`。

### 复现步骤

1. 设置主渠道成本 0/0；
2. 生成定价方案并读取建议售价与预期毛利。

### 实际结果

```
成本=0.00000000/0.00000000  压力成本=0.00000000/0.00000000
建议售价=0.00000000/0.00000000  预期毛利=3500bps
```

售价为 0，却报告 35% 毛利——**两个字段互相矛盾**。

### 期望结果

零成本不应产出零售价；若确实无法定价，毛利不应回落到目标值，应标记异常。依据：`docs/pricing-control-plane.md`「免费或异常低价渠道不能形成零成本售价。」

### 依据

- `service/pricing_control.go` 第 270-271 行：`inputPrice = inputCost / targetDenominator`，成本为 0 时售价为 0；
- 第 279-283 行：毛利回算有守卫 `if inputPrice > 0 && outputPrice > 0`，条件不成立时直接 `margin = policy.TargetMarginBPS`；
- 第 397 行 `validPriceNumber` 允许 0 成本通过校验。

### 影响范围

仅当**主渠道本身**成本为 0 时触发。EDGE-002（stable 档、备用渠道零成本）通过，因为加权成本仍为正。

风险在于：0 售价若被发布，等于该模型免费；而报表显示 35% 毛利，运营侧不会察觉。

### 定级理由

可通过"人工复核售价为 0 的方案"规避；但一旦发布即为免费模型，故优先级仍给 P1。

---

## 5. EDGE-013

**标题**：批量导入遇非法项时留下部分入库，非原子
**严重程度**：S3　**优先级**：P2　**阻断上线**：否
**类型**：数据一致性
**状态**：新建　**发现方式**：服务层自动化测试

### 环境

同 CONC-001。

### 前置条件

准备 3 条报价：第 1 条合法、第 2 条 `currency=CNY`（非法）、第 3 条合法。

### 复现步骤

1. 调用批量导入接口，一次性提交上述 3 条；
2. 观察返回值与数据库实际写入情况。

### 实际结果

函数返回错误，但数据库中**已写入 1 条**（第 1 条），第 3 条未处理。

### 期望结果

整批拒绝，或返回成功列表 + 失败明细，由调用方决定。当前状态下调用方无法得知"到底写进去了哪几条"。

### 依据

`service/pricing_control.go` 第 161-186 行：`for` 循环内校验通过后立即 `model.UpsertUpstreamModelOffer`，遇到非法项直接 `return`，整批无事务包裹。

### 影响范围

导入失败后数据处于部分更新状态，需要人工查库才能确认已写入范围。成本基线被部分覆盖后，后续生成的方案基于不完整数据。

### 定级理由

有明确报错、不会静默成功，且可通过"导入后核对条数"规避，故 S3/P2。

---

## 6. 复现方式

```bash
cd /Users/yuntao/code/OpenBridger
git checkout test/openbridger-qa

# CONC-001 / CONC-002 / EDGE-001 / EDGE-013（三库均复现）
go test ./service -run 'TestQAPricingState|TestQAPricingConcurrency|TestQAPricingEdge' -v

# MySQL / PostgreSQL 矩阵（需先注入 DSN）
TEST_MYSQL_DSN=... TEST_POSTGRES_DSN=... \
  go test ./service -run 'TestQAPricing' -v
```

PC-COST-001 需真实上游 key 与本地实例，复现步骤见本文档第 1 节与 `test-results-pricing-control.md` 第 9 节。

---

## 7. 待办

- [ ] PC-COST-001：确认是否需要把上游 `group_ratio` 纳入同步范围，或改为按模型维护"真实成本倍率"——**此项定级依赖该决策**
- [ ] CONC-001 / CONC-002：确认是补审批入口校验，还是调整 DEC-004 约定
- [ ] EDGE-001：确认零成本渠道的预期行为（报错 / 走保护价 / 标记异常）
- [ ] EDGE-013：确认期望语义（整批原子 vs 部分成功 + 明细返回）
- [ ] 5 项缺陷是否建档为研发事项、分配负责人、关联原需求

---

## 8. 更正记录

- 初版曾将"key 在 `auto` 组却按 6.8 计费"列为待用户确认的矛盾。经补充实测（`gemini-3.1-flash-lite` 按 gemini 组 0.3 计费，实测 398 vs 预测 397.5）确认：`auto` 为自动分组，按目标模型所属分组结算，**不存在矛盾，用户说法准确**。已更正 `test-results-pricing-control.md` 第 9.6 节。
- 初版曾表述 PC-COST-001 为"少收"，补充 gemini 实测后确认为**双向偏差**（高价组少收、低价组多收），已更正为第 9.7 节及本清单第 1 节。
