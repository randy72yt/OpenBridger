# OpenBridger 动态定价控制 — 测试用例设计

分支：`test/openbridger-qa`　最后更新：2026-09-16　状态：**待评审**

## 1. 范围与依据

| 项目 | 内容 |
| --- | --- |
| 被测对象 | 动态定价控制模块（OpenBridger 自研，upstream main 不存在） |
| 代码范围 | `service/pricing_control.go`、`model/pricing_control.go`、`model/model_pricing_config.go`、`controller/pricing_control.go`、`router/api-router.go`（`/api/pricing-control/*`）、`web/src/features/system-settings/models/pricing-control-*` |
| 设计依据 | `docs/pricing-control-plane.md` |
| 验收依据 | `docs/openbridger/production-acceptance.md` 的 BILL-003、BILL-004、BILL-005 |
| 本轮不做 | 全量回归、公开页面与文档站验收、支付与邮件（项目档案记为未配置） |

### 1.1 与既有资产的关系

项目已有 `production-acceptance.md`（30 步人工脚本 + BILL-003/004/005），但其执行总表全部为「未执行」。本文件不重复那份文档，**只补充它缺失的部分**：

- 那份是人工 UI 操作脚本，判断标准是页面表现；本文件给出**可自动化执行的服务层用例**和**精确数值预期**。
- 那份没有覆盖定价公式的边界与异常输入。
- 那份没有覆盖三数据库矩阵的真实执行方式。

两者并行：本文件的用例可先在服务层跑，通过后再执行人工脚本做 UI 确认。

## 2. 被测计算模型（来自代码，非设计文档）

已通过阅读 `buildPriceProposal` 与现有测试断言双向验证：

```text
fallback      = FallbackProbabilityBPS / 10000
overhead      = 1 + OverheadBPS / 10000

加权成本 = 主渠道成本 × (1 - fallback) + 备用渠道成本 × fallback   // 无备用渠道时 = 主渠道成本
压力成本 = 备用渠道成本                                          // 无备用渠道时 = 主渠道成本

含运营加权成本 = 加权成本 × overhead
含运营压力成本 = 压力成本 × overhead

建议售价 = 含运营加权成本 / (1 - TargetMarginBPS/10000)

// 仅当 ServiceTier == "stable" 且存在备用渠道
建议售价 = max(建议售价, 含运营压力成本 / (1 - MinimumMarginBPS/10000))

最终售价 = round(建议售价, 1e-8)
预期毛利 = min(1 - 含运营加权输入成本/输入售价, 1 - 含运营加权输出成本/输出售价) × 10000
           // 仅当 输入售价 > 0 且 输出售价 > 0；否则直接取 TargetMarginBPS
```

已用现有测试反算通过：`primary=1/5, backup=2/10, fallback=1000bps, overhead=500bps, target=3500bps` → 加权 `1.155/5.775`，建议价 `1.77692308/8.88461538`；`stable` 档压力价 `2.625/13.125`，毛利 `5600bps`。公式理解无误。

## 3. 评审前必读：4 处代码与设计不符

这几条不是用例，是我在读代码时发现的**待确认项**，需要你先定性，因为它们决定对应用例的「正确结果」怎么写。

### 风险 R1 — 零成本渠道产出 0 售价，却显示 35% 毛利

代码事实：
- `validPriceNumber(0) == true`，即 `InputCost=0` 的报价**可以合法导入**。
- default 档：`建议售价 = 0 / (1-0.35) = 0`。
- 毛利回算有守卫 `if inputPrice > 0 && outputPrice > 0`，为假时 `margin = TargetMarginBPS`。

结果：`ProposedInputPrice = 0`，而 `ExpectedMarginBPS = 3500`。**售价为 0 却报告毛利 35%，数据自相矛盾**，审核台无法据此判断风险。

设计文档原文：「免费或异常低价渠道不能形成零成本售价。经济线路可使用免费渠道并设置额度、并发和停服规则；稳定线路必须按可用的付费备用渠道计算保护价。」

stable 档同样守不住：备用渠道成本为 0 时，压力成本为 0，`max(0, 0) = 0`。

**需要你确认**：0 售价是否允许（文档同时说了「经济线路可使用免费渠道」）。但无论结论如何，「0 售价 + 35% 毛利」这个组合应当修正。

### 风险 R2 — 过期或版本变化的方案仍可被批准

项目档案 DEC-004（2026-09-09）：「上线前审核台展示当前价、建议价、涨跌幅和压力毛利；**过期或价格版本变化的方案禁止审批**。」

代码事实：`ApprovePricingProposal` 只校验 `status = pending`，**不检查报价是否过期，也不检查 `PricingChanged`**。`ListPricingProposalReviews` 计算了 `PricingChanged` 和过期时间，但那是给前端展示用的，服务端批准入口没有复用。

结果：管理员拿到一个基于已过期成本生成的旧方案，服务端仍可批准并发布。这直接命中 production-acceptance 的 BILL-004「过期报价：阻止发布并要求刷新」。

### 风险 R4 — 批量导入非原子，非法项会留下部分入库

`ValidateAndImportPricingOffers` 的循环体内，校验通过后**立即调用 `model.UpsertUpstreamModelOffer` 逐条落库**；遇到非法项时 `return` 退出。因此一批报价中若第 2 条非法：第 1 条已写入、第 3 条未处理，形成**部分成功**且没有回滚。

对照 PC-EDGE-011 的币种规范化逻辑（`usd` → `USD`），实际业务中大小写/币种不统一的批量导入很容易踩到。这属于数据一致性问题，可能让渠道成本停在半更新状态。

### 风险 R3 — 缓存读取成本未参与定价

`UpstreamModelOffer.CacheReadCost` 在导入时被校验，但 `buildPriceProposal` 只使用 `InputCost` 和 `OutputCost`；发布的计费表达式也只有 `p`（输入）和 `c`（输出）两项。

若上游对缓存读取的收费不可忽略，缓存命中的调用毛利会低于页面显示值。**需要确认**：首发模型的缓存计价是否走 `p` 或 `c`，若是则可关闭此项。

## 4. 测试用例

优先级沿用项目约定：P0 = 阻断上线，P1 = 重要，P2 = 建议。
标记 `[存疑]` 的用例，正确结果取决于第 3 节你的定性。

### 4.1 计算正确性（PC-CALC）

| ID | 优先级 | 场景 | 输入 | 正确结果 |
| --- | --- | --- | --- | --- |
| PC-CALC-001 | P0 | default 档，无备用渠道 | 主 1/5，overhead 500，target 3500 | 成本 1.05/5.25；售价 1.61538462/8.07692308；毛利 3500 |
| PC-CALC-002 | P0 | default 档，有备用渠道 | 主 1/5，备 2/10，fallback 1000，overhead 500，target 3500 | 成本 1.155/5.775；售价 1.77692308/8.88461538；毛利 3500 |
| PC-CALC-003 | P0 | stable 档压力保护价生效 | 同上，tier=stable，min 2000 | 售价取 max(1.77692308, 2.1/0.8)=2.625；输出 13.125；毛利 5600 |
| PC-CALC-004 | P0 | stable 档但无备用渠道 | 主 1/5，tier=stable | 保护价不生效，退化为 default 档结果 1.61538462/8.07692308 |
| PC-CALC-005 | P0 | 压力价抬高售价后毛利回算 | 同 PC-CALC-003 | `ExpectedMarginBPS` = 5600，不等于 `TargetMarginBPS`；审核台显示的预期毛利必须跟着变 |
| PC-CALC-006 | P1 | overhead = 0 | 主 1/5，备 2/10，fallback 1000，overhead 0，target 3500 | 成本 1.1/5.5；售价 1.69230769/8.46153846 |
| PC-CALC-007 | P0 | 手工复算与页面一致（BILL-003 步骤 4） | 任取一组真实成本 | 页面建议价与按第 2 节公式手算结果在显示精度内一致；若触发保护价/舍入，页面必须标明触发的规则 |

### 4.2 边界与异常输入（PC-EDGE）

| ID | 优先级 | 场景 | 输入 | 正确结果 |
| --- | --- | --- | --- | --- |
| PC-EDGE-001 | P0 `[存疑]` | 免费渠道零成本 | 主 0/0，default 档，target 3500 | R1。至少不得出现「售价 0 且毛利 3500」。若允许 0 售价，毛利应为 0 或该方案被标记为风险/禁止发布 |
| PC-EDGE-002 | P0 `[存疑]` | stable 档备用渠道零成本 | 主 1/5，备 0/0，tier=stable，min 2000 | R1。保护价退化为 0 时不得直接发布，应告警 |
| PC-EDGE-003 | P1 | 极低非零成本 | 主 0.000001/0.000005，target 3500 | 售价按公式正常产出，不因下溢变成 0；舍入后仍 > 0 |
| PC-EDGE-004 | P0 | fallback = 0 | 备 2/10，fallback 0，tier=default | 加权成本 = 主渠道成本；stable 档下压力成本仍取备用渠道成本 |
| PC-EDGE-005 | P0 | fallback = 10000 | 主 1/5，备 2/10，fallback 10000 | 加权成本 = 备用渠道成本 2/10；售价按备渠道计算 |
| PC-EDGE-006 | P1 | 极端目标毛利 | target 9999，min 0 | 校验通过（9999 < 10000）；售价 = 成本 × 10000，不溢出为 Inf/NaN |
| PC-EDGE-007 | P0 | 价格倒挂：备用远贵于主 | 主 1/5，备 50/100，fallback 5000，overhead 0，tier=stable，min 2000，target 3500 | 加权成本 25.5/52.5，建议价 39.23076923/80.76923077，压力价 62.5/125 → 售价取保护价 **62.5/125**；`WorstMarginBPS` = 2000（等于 min margin）；审核台须显示压力毛利 |
| PC-EDGE-008 | P0 | 舍入一致性 | 使售价出现 9 位小数的成本组合 | `ProposedInputPrice` 保留 8 位；发布后的表达式 `tier("base", p * X + c * Y)` 中的 X/Y 与建议价完全一致，无二次舍入 |
| PC-EDGE-009 | P1 `[存疑]` | 缓存读取成本 | CacheReadCost > 0 | R3。确认缓存 token 的计价归属；若独立计价则必须在售价中体现 |
| PC-EDGE-010 | P0 | 非法成本值 | 成本为 -1 / NaN / Inf | 导入被拒，返回 `ErrPricingOfferInvalid`，且不写入任何一条记录（事务性） |
| PC-EDGE-011 | P0 | 非 USD 币种 | currency = CNY / cny / "" | 被拒。注意 `usd` 小写应被规范化为 USD 并接受（现有测试已覆盖） |
| PC-EDGE-012 | P1 | 成功率越界 | success_rate_bps = -1 或 10001 | 被拒 |
| PC-EDGE-013 | P0 `[存疑]` | 部分数据非法 | 3 条报价中第 2 条非法 | R4。期望整批拒绝；当前实现会留下「第 1 条已入库」的部分成功状态，本用例在修复前预期为**失败** |
| PC-EDGE-014 | P1 | 成本无变化 | 同成本重复导入并重算 | `InputPriceChangeBPS` = 0，`OutputPriceChangeBPS` = 0，`PricingChanged` = false |
| PC-EDGE-015 | P1 | 空/缺字段导入 | 缺 public_model 或 channel_id=0 | 被拒，错误信息指明是第几条（现有实现已含 `at item N`） |

### 4.3 状态机（PC-STATE）

| ID | 优先级 | 场景 | 正确结果 |
| --- | --- | --- | --- |
| PC-STATE-001 | P0 | pending → approved → published | 状态依次变化；`approved_by`/`approved_at`/`published_at` 有值 |
| PC-STATE-002 | P0 | 重复发布同一方案 | 第二次返回 `ErrPricingProposalState`，价格不二次变更 |
| PC-STATE-003 | P0 | 未批准直接发布 | 被拒，价格不变 |
| PC-STATE-004 | P1 | rejected 后再批准/发布 | 均被拒 |
| PC-STATE-005 | P0 | stable 档方案发布 | 被拒，错误信息明确说明仅 default 档可发布全局价格；**前端须给出可读提示，不能只报内部错误** |
| PC-STATE-006 | P0 | 价格锁定（PriceLocked=true） | 重算时跳过，不生成新方案 |
| PC-STATE-007 | P1 | 策略禁用（Enabled=false） | 重算时跳过 |
| PC-STATE-008 | P0 | 连续两次重算 | 旧 pending 方案被替换，不累积重复方案（`ReplacePendingModelPriceProposal`） |
| PC-STATE-009 | P1 | 发布后模型定价模式 | `billing_setting.billing_mode` = `tiered_expr`，表达式含建议价；版本号递增 |

### 4.4 过期、并发与风险（PC-CONC）

| ID | 优先级 | 场景 | 正确结果 |
| --- | --- | --- | --- |
| PC-CONC-001 | P0 `[存疑]` | 报价过期后批准旧方案 | R2。按 DEC-004 应禁止审批并要求刷新；当前服务端未拦截，若确认是缺陷则本用例在修复前为**失败** |
| PC-CONC-002 | P0 `[存疑]` | 价格版本变化后批准（PricingChanged=true） | R2。同上应禁止 |
| PC-CONC-003 | P0 | 并发发布 | 两个请求基于同一 `PricingVersion`，先到成功，后到因版本 CAS 失败（`UpdateModelPricing` 的 `ExpectedVersion` 校验）；**余额/价格只变更一次** |
| PC-CONC-004 | P0 | 采集失败不覆盖有效成本 | 导入失败或空批次时，保留上一次有效成本，不得写 0 成本（BILL-004「采集失败」） |
| PC-CONC-005 | P0 | 主渠道报价过期 | 重算跳过该模型（不生成方案）；`PricingRisks()` 返回 `primary_offer_missing_or_expired` |
| PC-CONC-006 | P0 | 备用渠道报价过期 | stable 档重算跳过；风险列表返回 `backup_offer_missing_or_expired` |
| PC-CONC-007 | P0 | 报价 disabled | `enabled=false` 的报价不参与成本计算，等同于缺失 |
| PC-CONC-008 | P1 | 同模型多渠道同名最新优先 | 同一 (public_model, channel_id) 多条报价时取 `collected_at` 最新的一条 |

### 4.5 权限与审计（PC-PERM）

本轮范围外，但定价接口全部由 `middleware.RootAuth()` 保护且**当前没有任何接口级测试**，建议至少执行前两条。

| ID | 优先级 | 场景 | 正确结果 |
| --- | --- | --- | --- |
| PC-PERM-001 | P0 | 非 Root 用户调用 `/api/pricing-control/*` 全部 9 个接口 | 返回 401/403，数据未修改 |
| PC-PERM-002 | P0 | Root 执行导入/批准/发布 | 审计日志有记录，含操作人与对象；日志**不含**上游 Key、Session、密码 |

### 4.6 三数据库矩阵（PC-DB）

现有 `TestPricingControlDatabaseMatrix` 因缺少 DSN 被跳过（已实测确认）。本轮用 docker 起真实库注入 `TEST_MYSQL_DSN` / `TEST_POSTGRES_DSN` 后执行，满足 AGENTS.md 的三库验证硬要求。

| ID | 优先级 | 场景 | 正确结果 |
| --- | --- | --- | --- |
| PC-DB-001 | P0 | SQLite 全量定价流程 | 导入→重算→批准→发布→快照校验全部通过 |
| PC-DB-002 | P0 | MySQL 8.0 同上 | 同上；`decimal(20,8)` 精度无损 |
| PC-DB-003 | P0 | PostgreSQL 17 同上 | 同上 |
| PC-DB-004 | P0 | 三库售价一致性 | 同一组输入在三种库上产出的建议价**完全相同**（对比到 1e-8） |
| PC-DB-005 | P1 | 唯一索引冲突 | 同 (channel_id, upstream_model, public_model) 重复导入 → upsert 更新而非报错或建重复行 |

### 4.7 真实上游端到端（PC-E2E）

你有可用上游 Key，以下用例在真实链路上验证「定价 → 计费」闭环。

| ID | 优先级 | 场景 | 正确结果 |
| --- | --- | --- | --- |
| PC-E2E-001 | P0 | 导入真实采购成本 → 生成 → 批准 → 发布 → 真实调用 | 新调用按新价格计费；`调用前余额 - 调用后余额 = 日志费用`（允许显示精度的最小舍入差） |
| PC-E2E-002 | P0 | 发布后历史日志 | 发布前已产生的日志费用**不被改写**，价格版本可区分 |
| PC-E2E-003 | P0 | 主备切换 | 切换后用户无需改配置；每次调用只结算一次，无双重扣费 |
| PC-E2E-004 | P1 | 发布后刷新审核台 | 当前价 = 刚发布的价格，`PricingChanged` = false，涨跌 = 0 |

## 5. 执行方式

| 层级 | 落地位置 | 命令 |
| --- | --- | --- |
| 服务层（PC-CALC / EDGE / STATE / CONC） | 扩充 `service/pricing_control_test.go`，复用现有 `TestMain` 的 SQLite in-memory 与 `truncate` | `go test ./service/ -run TestPricingControl -v` |
| 三数据库（PC-DB） | 复用 `TestPricingControlDatabaseMatrix`，docker 起库后注入 DSN | `docker compose -f docker-compose.dev.yml up -d` 后带 DSN 跑测试 |
| 接口与权限（PC-PERM） | 新增 router 层 gin 测试，或脚本调用带/不带 Root Cookie | 待定 |
| 端到端（PC-E2E） | 启动本地服务，脚本 + 手工核对 | `go run .` 或现有 `.local-tests/start.sh` |

不修改任何业务代码，只新增测试文件与用例文档。发现的缺陷单独记录，不在本分支直接修。

## 6. 交付物

1. 本分支新增/扩充的测试用例代码。
2. 一份执行结果记录：每条用例的结论（通过 / 失败 / 阻断），失败项附最小复现与代码位置。
3. 缺陷清单，按 production-acceptance 的判定规则标注 P0/P1，并回写到项目档案的风险登记。

## 7. 待你确认

1. **第 3 节 R1～R4 如何定性**——决定 PC-EDGE-001/002/013、PC-CONC-001/002 写「通过」还是「失败」。
2. **0 售价是否可接受**（R1）；若可接受，毛利该如何显示。
3. **R2（过期方案可批准）和 R4（导入非原子）是否本轮修复**。若只测不修，我把它们记为 P0 失败项；若要修，需要另开修复分支，本分支保持只加测试。
4. 是否需要我同步补齐 PC-PERM 两条权限用例（本轮范围外，但风险高）。
