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
