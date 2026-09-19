# 按次计费模型（生图 / MJ 类）纳入动态定价 —— 改造清单

| 项 | 值 |
|---|---|
| 分支 | `codex/fix-pricing-control-qa` |
| 提出依据 | `docs/openbridger/defect-list-pricing-control.md`、`docs/openbridger/fix-review-pricing-control.md`（第 6 节） |
| 当前状态 | **未实现**，本文为实现规格 |
| 影响的代码区域 | `model/pricing_control.go`、`service/pricing_control.go`、`service/pricing_offer_sync.go`、`web/src/features/system-settings/models/*` |
| 不涉及 | 计费内核（`relay/helper/price.go` 及其 per-call 分支）、`ratio_sync` |

---

## 1. 背景

### 1.1 QA 发现

`ddd054896` 引入上游报价同步后，40 个上游模型中有 **3 个被跳过**。原因是报价表的成本字段单位为「USD / 百万 Token」，而这 3 个是 `quota_type=1` 的按次计费模型：

| 模型 | model_price | 计费方式 | enable_groups | group_ratio |
|---|---|---|---|---|
| `gemini-3-pro-image` | 0.12 | USD / 张 | media | 1 |
| `gemini-3-pro-image-4k` | 0.22 | USD / 张 | media | 1 |
| `gpt-image-2` | 1.00 | USD / 张 | media, sub-openai | media = 1 |

跳过发生在 `service/pricing_offer_sync.go:273`：

```go
func normalizedUpstreamTokenCosts(item upstreamPriceItem) (float64, float64, float64, error) {
	if item.QuotaType == 1 {
		return 0, 0, 0, errors.New("fixed-price model is not token-priced")
	}
	...
```

### 1.2 为什么必须改

1. 生图是高频核心品类，且单价高（`$1.00/次`），定价错了直接亏钱；
2. 模型仍在迭代（`gpt-image-2` → `2.5`），靠人工维护一个持续增加、单价昂贵的品类不可持续；
3. **数据已经全部到位，只是一行 `return` 把它挡住了** —— 上游 `/api/pricing` 返回体里的 `model_price` 字段就是准确单价，`upstreamPriceItem` 结构体（`service/pricing_offer_sync.go:49-51`）已经把 `QuotaType`、`ModelPrice` 都解析进来了。

---

## 2. 可行性依据（已验证）

改造不需要碰计费内核，三条链路本来就是通的：

| 结论 | 证据 |
|---|---|
| **发布通道存在**：`ModelPrice` 可直接写入 | `model/model_pricing_config.go:50` 的 `modelPricingOptionKeys` 白名单已包含 `"ModelPrice"`；`UpdateModelPricing`（同文件 :269）接受 `PricingValues{"ModelPrice": X}` |
| **不需要设 `quota_type`**：写了 `ModelPrice` 就自动变按次 | `model/pricing.go:381-389`：<br>`modelPrice, findPrice := ratio_setting.GetModelPrice(model, false)`<br>`if findPrice { pricing.ModelPrice = modelPrice; pricing.QuotaType = 1 } else { ...ratio 路径... pricing.QuotaType = 0 }` |
| **切换回 token 也自动**：不提供 `ModelPrice` 就会被删掉 | `model/model_pricing_config.go:296-301`：`UpdateModelPricing` 对白名单内每个 key 先 `delete` 再按需写回，因此只发 `ModelPrice` 会清掉 `billing_expr` / `model_ratio`；只发 `billing_expr` 会清掉 `ModelPrice` |
| **按次扣费路径已在用** | `relay/relay_task.go:292`、`relay/mjproxy_handler.go:206/519` 走 `helper.ModelPriceHelperPerCall` |

**结论：本次改造全部落在动态定价模块内部 + UI，计费侧零改动。**

---

## 3. 目标与非目标

**目标**

- 生图 / 按次计费模型可自动同步上游报价（含分组倍率）
- 可像 token 模型一样：生成调价方案 → 人工审批 → 发布到自己的销售价
- 发布结果写入 `ModelPrice`（USD / 次），由系统接管计费
- 三库（SQLite / MySQL / PostgreSQL）均验证通过

**非目标**

- 不改变按次计费的扣费公式
- 不为同一模型的多种分辨率 / 尺寸档位建立多档价（见第 10 节待确认项 10.1）
- 不改 `ratio_sync`（它同步的是上游原始倍率，与本模块互不干涉）

---

## 4. 数据口径定义

| 口径 | 字段值 | InputCost 含义 | OutputCost | CacheReadCost | 发布目标 | 售价单位 |
|---|---|---|---|---|---|---|
| token | `cost_basis = "token"`（现有行为，默认） | USD / 百万 input token | USD / 百万 output token | USD / 百万 cache read token | `billing_setting.billing_mode=tiered_expr`<br>`billing_setting.billing_expr=tier("base", p*IP + c*OP)` | USD / 百万 token |
| request | `cost_basis = "request"`（新增） | **USD / 次** | 恒为 `0` | 恒为 `0` | `ModelPrice = X` | USD / 次 |

成本换算（两种口径通用）：

```
真实成本 = 上游单价 × UpstreamGroupRatio × (1 + OverheadBPS/10000)
建议售价 = 真实成本 / (1 - TargetMarginBPS/10000)
```

生图场景示例（`group_ratio.media = 1`）：

| 模型 | 上游单价 | 成本（×1） | 目标毛利 35% 时售价 |
|---|---|---|---|
| `gemini-3-pro-image` | 0.12 | 0.12 | 0.1846 |
| `gemini-3-pro-image-4k` | 0.22 | 0.22 | 0.3385 |
| `gpt-image-2` | 1.00 | 1.00 | 1.5385 |

---

## 5. 数据模型改动

### 5.1 推荐方案 A：新增 `cost_basis`，复用现有成本列

| 表 | 改动 |
|---|---|
| `upstream_model_offers` | 新增 `cost_basis varchar(16) not null default 'token'`（`token` \| `request`），加普通索引 |
| `upstream_cost_snapshots` | 同上（快照若不记录口径，历史回溯会丢失语义） |
| `model_price_proposals` | 新增 `cost_basis varchar(16) not null default 'token'`（发布时需据此决定写 `ModelPrice` 还是 `billing_expr`） |

取值约束在应用层校验，不用 DB enum（保持三库兼容）。

**采用理由**：只加一列，迁移面最小，需要同时维护的列更少，三库风险最低。
**代价**：`request` 口径下 `input_cost` 的语义变为「每次」，需靠 `cost_basis` 一起读，代码里必须显式分支，不能只看数值。

### 5.2 备选方案 B：新增 `request_cost` 独立列

- 优点：语义严格，`input_cost` 永远只表示 token
- 缺点：两列并存容易出现「两个都有值」的脏数据，需要额外校验；UI 取值逻辑也要双分支

**不建议**。除非产品认为 A 的语义复用会造成长期维护风险。

### 5.3 迁移要求

- `AutoMigrate` 必须覆盖两张表的两个\Models（`model/main.go` 现有注册方式）；
- 新列 **必须 nullable 或带默认值**，否则存量数据升级后立即报错 —— 这正是缺陷清单中 RISK-002 的教训；
- 参考 `docs/pricing-control-plane.md` 的迁移验证要求，三库都要跑一遍（命令见第 8.3 节）；
- **无需回填脚本**：`ddd054896` 之后，跑一次报价同步即 upsert 覆盖（冲突键 `(channel_id, upstream_model, public_model)`）。

---

## 6. 后端改动清单

### 6.1 `model/pricing_control.go`

| # | 位置 | 改动 |
|---|---|---|
| 1.1 | `UpstreamModelOffer`（:23-42） | 新增 `CostBasis string` 字段，`gorm:"type:varchar(16);not null;default:token"` |
| 1.2 | `UpstreamCostSnapshot`（:44-56） | 新增同名字段；`upsertUpstreamModelOffer`（:195）创建快照时一并写入 |
| 1.3 | `ModelPriceProposal`（:75-95） | 新增 `CostBasis string` 字段 |
| 1.4 | `validateUpstreamModelOfferWrite`（:158） | 增加 `cost_basis` 白名单校验（`token` / `request`）；**request 口径下不再要求 `OutputCost > 0`**，但要求 `InputCost > 0` 且 `CacheReadCost == 0` |

### 6.2 `service/pricing_offer_sync.go`

| # | 位置 | 改动 |
|---|---|---|
| 2.1 | `normalizedUpstreamTokenCosts`（:273） | `quota_type == 1` 不再返回错误。改为返回 `requestCost = item.ModelPrice`（正值校验保留），并将 output / cache 成本置为 0 |
| 2.2 | `fetchPricingOffersForChannel`（~:200-240） | 构造 offer 时按 item.QuotaType 设置 `CostBasis`：`1 → "request"`，否则 `"token"` |
| 2.3 | 同上 | `SourceType` 保持 `new-api-pricing`；`ExpiresAt = now + 2h` 逻辑不变 |

### 6.3 `service/pricing_control.go`

| # | 位置 | 现状 → 改法 |
|---|---|---|
| 3.1 | `ValidateAndImportPricingOffers`（:158-190） | 现状校验 `InputCost / OutputCost / CacheReadCost` 三项 `validPriceNumber`。改为按 `cost_basis` 分支：request 型只校验 `InputCost` 有限且 `> 0` |
| 3.2 | `buildPriceProposal`（:241-328） | ① 主备 offer 的 `cost_basis` **必须一致**，不一致返回 `ErrPricingCostInvalid`（避免混合口径算出无意义加权平均）<br>② 沿用 :253-277 的倍率乘法与主备加权（`fallbackProbabilityBPS`）、`overhead` 逻辑<br>③ request 型：单一成本 `requestCost`，`inputPrice = requestCost / targetDenominator`，`outputPrice = 0`，跳过 :294-296 的 `ServiceTier == "stable"` 压力成本取 max 时的 output 部分<br>④ `margin` 计算由双值改单值<br>⑤ 返回结构体写入 `CostBasis` |
| 3.3 | `PublishPricingProposal`（:405-431） | 现状固定写 `tiered_expr` + 表达式。改为：<br>`if proposal.CostBasis == "request"` → `Pricing: model.PricingValues{"ModelPrice": X}`<br>`else` → 维持现有表达式逻辑 |
| 3.4 | `PricingRisks`（:433-465） | request 型只判 `InputCost > 0`，不再因 `OutputCost == 0` 误报 `primary_offer_cost_invalid` |
| 3.5 | `minimumPriceMarginBPS`（:132） | request 型跳过 output 部分的毛利计算 |

### 6.4 `controller/pricing_control.go`

| # | 位置 | 改动 |
|---|---|---|
| 4.1 | `ImportPricingOffers`（手工导入入口） | 接受 `cost_basis` 字段；request 型不强制要求 output_cost / cache_read_cost |

### 6.5 Router / 定时任务

无改动。`/api/pricing-control/offers/sync` 与定时任务 `pricing_offer_sync` 直接复用新逻辑。

---

## 7. 前端改动清单

| # | 文件 | 改动 |
|---|---|---|
| 7.1 | `web/src/features/system-settings/models/pricing-control-types.ts` | `UpstreamModelOffer`、`ModelPriceProposal` 类型补 `cost_basis` |
| 7.2 | `pricing-control-inventory.tsx` | 表头 / 文案：request 型显示 `t('Request cost')` 单列而非 `t('Input cost')` + `t('Output cost')` 两列；数值后缀按口径区分 |
| 7.3 | `pricing-proposal-review.tsx` | 顶部说明 `t('Prices and costs are USD per million tokens.')` 需按口径切换；request 型时改为类似 `t('Prices and costs are USD per request.')`；提案表格 Input/Output 双行改单行 |
| 7.4 | `pricing-control-policy-form.tsx` | 无需改（毛利、overhead、fallback 对两种口径通用） |
| 7.5 | `web/src/i18n/locales/{en,zh}.json` | 新增文案 key（同步 run `bun run i18n:sync`） |
| 7.6 | `pricing-control-section.tsx` | 同步结果 toast：`ddd054896` 目前只提示成功/失败，**不展示 `channels[].warnings`**。建议补一个可展开的 warnings 列表（对应 RISK-005） |

---

## 8. 测试清单

### 8.1 单元测试（`service/`）

沿用 `service/pricing_control_qa_test.go` 既有风格（该文件的辅助函数 `qaOffer`、`truncate` 等可直接复用）：

| 用例 | 期望 |
|---|---|
| 同步侧：`quota_type=1` 的模型不再被跳过 | `warnings` 中不再出现 `xxx: fixed-price model is not token-priced`；产出 offer 的 `CostBasis == "request"`，`InputCost == model_price` |
| 导入校验：request 型 offer 允许 `output_cost=0` | 不返回 `ErrPricingOfferInvalid` |
| 导入校验：`cost_basis` 非法值 | 返回 `ErrPricingOfferInvalid` |
| 方案生成：request 型，无备 | `ProposedInputPrice == cost/(1-margin)`，`ProposedOutputPrice == 0`，`ExpectedMarginBPS` 与目标一致 |
| 方案生成：request 型，主 `request` + 备 `token` | 返回 `ErrPricingCostInvalid`（口径不一致） |
| 方案生成：request 型成本为 0 | 返回 `ErrPricingCostInvalid`（防回归 EDGE-001） |
| 发布：request 型提案发布 | `model_pricing` 中该模型 `ModelPrice == 售价`，且 `billing_expr` 被清除；随后 `GetModelPrice(name)` 命中、`QuotaType == 1` |
| 发布：token 型提案发布（回归） | 仍写 `billing_expr`，`ModelPrice` 被清除 |
| 风险清单：request 型合法报价 | 不出现 `primary_offer_cost_invalid` |
| 过期 / 版本变化后批准（回归 CONC-001/002） | 仍返回 `ErrPricingProposalStale` |

### 8.2 端到端（本地隔离实例）

```
起两个同模型渠道（或一个渠道 + 备用），`POST /api/pricing-control/offers/sync`
→ GET /api/pricing-control/offers 确认生图模型已入库且 cost_basis=request
→ POST /api/pricing-control/policies 为生图模型建策略（目标毛利 3500）
→ POST /api/pricing-control/recalculate，查看生成的方案单价
→ 批准 + 发布 → 实际调用一次，核对扣费 == ModelPrice
```

### 8.3 三库矩阵（必做）

```bash
docker run -d --name qa-mysql -e MYSQL_ROOT_PASSWORD=qapass -e MYSQL_DATABASE=qaprice -p 3399:3306 mysql:8.0
docker run -d --name qa-pg   -e POSTGRES_PASSWORD=qapass -e POSTGRES_DB=qaprice -p 5499:5432 postgres:17

GOWORK=off \
TEST_MYSQL_DSN="root:qapass@tcp(127.0.0.1:3399)/qaprice?charset=utf8mb4&parseTime=True&loc=Local" \
TEST_POSTGRES_DSN="host=127.0.0.1 port=5499 user=postgres password=qapass dbname=qaprice sslmode=disable" \
go test ./service -run '^TestQAPricingDatabaseMatrix$' -count=1 -v
```

要求：`sqlite` / `mysql` / `postgres` 三个子测试均 PASS 且售价、毛利数值一致。（本机已验证现有代码三库通过：`3.28125000 / 16.40625000`，毛利 `5342bps`。）

**注意**：`TestQAPricingDatabaseMatrix` 在缺 DSN 时子测试是 `t.Skip`、外层对比是 `t.Logf`，job 仍显示绿色。因此「测试通过」必须以日志里真的出现三个子测试的 PASS 行为准。

### 8.4 前端

`cd web && bun run test`（现有 CI 已包含该 job）。

---

## 9. 风险与副作用

| 风险 | 说明 | 处置 |
|---|---|---|
| 发布后模型计费口径翻转 | 某模型原本是 token 计费，被误配成 request 策略后发布 → 直接变成 `quota_type=1`，扣费方式彻底改变 | 发布前在审批页显著标注口径；request 型仅允许人工确认发布，不得走 `AutoPublishChangeBPS` 自动发布（**建议对 request 型强制禁用自动发布**） |
| 混合口径的主备渠道 | 主 request 备 token，加权平均无意义 | 已在 3.2 处理为 `ErrPricingCostInvalid` |
| 存量 token 报价受影响 | `cost_basis` 默认 `token`，现有数据行为不变 | 靠默认值兜底，无需回填 |
| 解锁既有缺陷 | `ExpiresAt = now + 2h`，若 `PRICING_CONTROL_ENABLED=true` 未配置，2 小时后依然存在全部停摆问题 | 与本改造无关但必须一并处理，见 6.9 节 ① |
| 图片模型的 `SuccessRateBPS` 恒为 10000 | 同步侧硬编码 10000，未反映真实成功率 | 不在本次范围，登记为观察项 |

---

## 10. 待产品确认

### 10.1 多档价 / 分辨率

`gpt-image-2` 存在不同分辨率档位；上游 `/api/pricing` 当前只返回一个 `model_price`。需确认：

- 各档位是同一模型名不同参数，还是不同模型名？
- 若同模型多档位，动态定价该如何表达（多 public_model？还是不支持该模型自动定价）？

### 10.2 自动发布是否放开

建议：request 型**不支持自动发布**（`AutoPublishChangeBPS` 一律按 0 处理），原因是口径翻转的代价远大于 token 模型。需产品确认。

### 10.3 售价精度

`gpt-image-2` 售价 `1.5385`，`roundPricingValue` 保留 8 位小数，`ModelPrice` 是 float64。是否需要按货币习惯收敛（例如保留 4 位）？

---

## 11. 工作量估算与排期建议

| 阶段 | 内容 | 估算 |
|---|---|---|
| P0 后端核心 | 5.1 + 6.1 + 6.2 + 6.3（含 8.1 单测） | 1 人日 |
| P1 前端 | 7.1–7.5 | 0.5 人日 |
| P2 三库 + E2E | 8.2 + 8.3 | 0.5 人日 |
| P3 UI 细节（warnings 展示，对应 RISK-005） | 7.6 | 0.5 人日（可选） |

**建议顺序**：P0 → P2 → P1 → P3。P0/P2 完成即可灰度验证，P1 只是展示层。

---

## 12. 过渡期方案（P0 上线前）

不是「换算成 USD/百万 Token」（那个确实没有业务含义），而是**直接填写 `ModelPrice`（USD / 次）** —— 这个数值是准确的，不存在折算失真：

```
gpt-image-2 示例：
  上游单价 1.00  ×  group_ratio.media 1  =  成本 1.00
  目标毛利 35%  →  售价 = 1.00 / (1 - 0.35) = 1.5385  USD / 次
```

填入位置：系统设置 → 模型设置 → 模型定价 → 该模型的 `ModelPrice`（或直接调 `/api/pricing-control/offers/import` 手工导入 request 型报价，待 P0 完成后该接口即支持）。

**唯一缺点**：上游调价不会自动同步，需人工跟进。这正是 P0 要解决的问题。

**过渡期务必注意**：

1. 不要给这 3 个模型建动态定价策略 —— 建了之后风险清单会出现 `primary_offer_cost_invalid`，且永远生成不出方案（不报错、不崩溃，容易被误判为系统故障）；
2. 新模型（如 `gpt-image-2.5`）上线后，先确认上游的 `enable_groups` 与 `group_ratio`（`media` 当前为 1，变了成本就变了）。
