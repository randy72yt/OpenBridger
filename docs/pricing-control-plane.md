# OpenBridger 动态定价控制模块设计

## 目标

OpenBridger 自动发现 DDD、CodeGo 等上游渠道提供的模型，记录各渠道的采购成本和可用性，并根据主备路由、失败重试、支付成本和目标毛利生成公开售价。模型目录、采购成本和用户售价分别管理，避免把上游成本隐藏在一组无法追溯的固定倍率中。

## 部署决定

- 功能保留在 OpenBridger 同一仓库，复用现有配置、鉴权、审计、系统任务和计费表达式。
- 本地和单节点部署由主进程运行定价任务。
- 生产环境可使用同一份程序启动独立 Worker，主节点负责 API 和管理后台，Worker 负责模型、成本及健康数据同步。
- Worker 不直接维护另一套业务规则，也不同时读写主库和日志库。它通过内部服务层执行任务；若以后部署为独立服务，则通过带服务身份认证的内部 API 提交采集结果。

## 数据库边界

### 主业务库

以下数据写入现有主业务库，并继续支持 SQLite、MySQL 和 PostgreSQL：

1. `upstream_model_offers`
   - 渠道 ID、上游模型名称和映射后的公开模型名称。
   - 输入、输出、缓存、图片、音频及任务类采购单价。
   - 币种、单位、倍率、来源类型、来源地址和采集时间。
   - 有效状态、数据版本和价格过期时间。

2. `upstream_cost_snapshots`
   - 每次发现的采购价格快照。
   - 原始价格、标准化人民币成本、数据摘要和变更幅度。
   - 仅追加，不覆盖历史记录。

3. `model_route_policies`
   - 公开模型、服务等级、主渠道和备用渠道。
   - 最大重试次数、切换条件、允许的最大备用成本和停服规则。

4. `model_price_policies`
   - 目标毛利、最低毛利、支付与基础设施成本率。
   - 价格缓冲、最小调整幅度、自动发布范围和人工审核阈值。

5. `model_price_proposals`
   - 当前成本、综合成本、建议售价、现行售价和预计毛利。
   - 状态为待审核、已批准、已拒绝、已发布或已失效。
   - 保存计算输入、规则版本、审核人与发布时间。

6. `pricing_sync_runs`
   - 同步任务状态、渠道结果、发现模型数、价格变化数和错误摘要。

现有模型计费表达式仍是用户结算的唯一依据。批准发布后，定价模块通过现有模型定价服务更新表达式，并保留并发版本检查和管理审计。

### 日志库

日志库不保存定价主数据，只保留：

- 用户实际调用、Token 用量、最终选择的渠道和结算金额。
- 请求发生时冻结的计费表达式与规则版本。
- 实际重试次数、失败渠道和最终渠道。

定价分析需要历史用量时，由主应用的统计服务读取日志库并生成聚合结果，再把必要的成功率、重试率和用量统计写回主库。定价 Worker 不直接修改日志库。

## 数据采集

每个上游实现一个适配器：

```text
DiscoverModels(ctx, channel) -> 模型目录
FetchOffers(ctx, channel)    -> 采购价格与单位
FetchHealth(ctx, channel)    -> 可选的余额和渠道状态
```

数据来源按以下优先级使用：

1. 上游正式价格 API。
2. 上游提供的 JSON 或配置接口。
3. 管理员上传的结构化价格文件。
4. 需要登录态的页面采集，仅作为临时适配器。

任何采集结果必须保存来源、采集时间和内容摘要。上游接口只返回模型名称时，仅更新模型目录，不推断或覆盖采购价格。新发现但没有有效成本的模型保持隐藏状态，不能进入付费路由。

## 主备路由成本

经济线路和稳定线路分别计算：

```text
经济线路成本 = 主渠道成本 + 基础设施成本

稳定线路预期成本 =
  主渠道成本 × 主渠道成功率
  + 备用渠道成本 × 切换概率
  + 平均重试成本
  + 基础设施成本
  + 支付手续费分摊
```

稳定线路同时计算压力成本：主渠道完全不可用时，所有请求进入备用渠道所产生的成本。如果现行售价低于压力成本保护线，系统不得继续无限制切换到备用渠道。

免费或异常低价渠道不能形成零成本售价。经济线路可使用免费渠道并设置额度、并发和停服规则；稳定线路必须按可用的付费备用渠道计算保护价。

## 售价规则

```text
建议售价 = 综合成本 / (1 - 目标毛利率)

保护售价 = max(
  建议售价,
  主渠道成本 × 主渠道保护倍率,
  备用压力成本 × 备用保护倍率
)
```

首发默认规则：

- 目标调用毛利率：35%。
- 最低允许毛利率：20%。
- 成本变化低于 5%：不调整公开售价。
- 建议涨幅不超过 5% 且毛利仍安全：进入下一次定时发布窗口。
- 建议涨幅 5% 至 15%：等待管理员审核。
- 成本上涨超过 15% 或预计毛利低于 10%：暂停自动扩量，触发告警并等待审核。
- 成本下降：每天集中生成一次降价建议，避免频繁变价。
- 成本超过 24 小时未更新：标记过期并告警。
- 成本超过 72 小时未更新：禁止自动上线新模型。

价格发布只影响之后创建的请求。已经开始的请求继续使用其冻结的计费快照完成结算。

## 任务流程

复用现有 `SystemTask`、任务锁、租约续期和多节点抢占机制，新增任务类型：

1. `upstream_model_discovery`
2. `upstream_offer_sync`
3. `upstream_health_rollup`
4. `model_price_recalculate`
5. `model_price_publish`

推荐频率：

- 模型目录：每 6 小时。
- 上游采购价格：每 1 小时。
- 成功率与重试率聚合：每 15 分钟。
- 建议售价重算：成本或健康数据变化后触发，并至少每小时校验一次。
- 自动发布窗口：每天一次；风险保护任务即时执行。

任务使用现有数据库锁保证多节点只能有一个执行者。失败采用有限重试，不用失败结果覆盖最后一次有效成本。

## 管理后台

新增“定价控制”页面，展示：

- 每个公开模型在 DDD、CodeGo 等渠道的最新成本和更新时间。
- 主渠道、备用渠道、成功率、重试率和压力成本。
- 现行售价、建议售价、预计毛利和最低毛利。
- 价格来源及最近同步任务结果。
- 待审核变更、差异说明和一键发布。
- 成本过期、价格倒挂、免费线路失效及备用渠道过贵告警。

管理员可以锁定某个模型的售价。锁定后系统继续更新成本和毛利告警，但不会自动修改公开价格。

## 内部接口

- `POST /api/pricing/sync/models`
- `POST /api/pricing/sync/offers`
- `GET /api/pricing/offers`
- `GET /api/pricing/proposals`
- `POST /api/pricing/proposals/:id/approve`
- `POST /api/pricing/proposals/:id/reject`
- `POST /api/pricing/proposals/:id/publish`
- `GET /api/pricing/risk`

管理接口继续使用现有管理员鉴权和审计。若 Worker 独立部署，采集提交接口使用单独的服务身份、请求签名、时间戳和防重放标识，不向 Worker 提供管理员 Session 或数据库超级用户凭据。

## 实施顺序

1. 建立渠道报价、成本快照、路由政策和价格建议数据结构。
2. 接入现有渠道模型发现任务，未定价模型默认隐藏。
3. 实现 DDD 和 CodeGo 报价适配器，并保留表格导入作为回退方式。
4. 实现综合成本、压力成本和建议售价计算。
5. 完成管理后台成本与毛利视图，以及人工审核发布。
6. 增加有限范围自动发布和风险保护。
7. 完成 SQLite、MySQL、PostgreSQL 的新建与升级迁移验证后再进入生产环境。

## 当前实现状态

第一阶段已集成到 OpenBridger：上游标准化成本与历史快照、模型定价策略、主备渠道加权成本、备用渠道压力成本、调价方案、风险检查、人工审批以及向现有计费表达式发布价格。管理员入口位于“系统设置 → 模型设置 → 动态定价控制”。

当前接口统一使用 `/api/pricing-control` 前缀。开启 `PRICING_CONTROL_ENABLED=true` 后，系统会按 `PRICING_OFFER_SYNC_INTERVAL_MINUTES`（默认 60 分钟，最低 15 分钟）从兼容 New API `/api/pricing` 的已配置主、备渠道同步报价和实际分组倍率，再按 `PRICING_RECALCULATE_INTERVAL_MINUTES` 定时重算。管理员也可以在管理页手动同步，或通过 `POST /api/pricing-control/offers/import` 导入标准化的美元/百万 Token 报价。无法确认有效分组倍率、仅提供固定单价或不属于渠道模型范围的报价会被跳过，不会按倍率 1 猜测成本。

最小导入示例：

```json
[
  {
    "channel_id": 1,
    "upstream_model": "provider-model-name",
    "public_model": "openbridger-model-name",
    "input_cost": 1.5,
    "output_cost": 7.5,
    "cache_read_cost": 0.15,
    "upstream_group_ratio": 1.0,
    "currency": "USD",
    "source_type": "manual",
    "source_version": "2026-09-08",
    "success_rate_bps": 9890,
    "enabled": true
  }
]
```

`upstream_group_ratio` 是上游账号或自动分组实际结算倍率。动态定价会先将报价乘以该倍率，再计算采购成本和售价。每条导入报价都必须显式提供大于零的倍率；如果报价已经是最终采购价，则填写 `1`。这样可以避免未知倍率被默认为无折扣价格而造成错误毛利。

未提供采集时间时使用导入时间，未提供有效期时默认 24 小时后过期。输入或输出成本为零的报价会保留用于审计，但不会生成可审批的调价方案，并会进入风险列表。

DDD、CodeGo 的自动采集适配器仍需根据它们实际提供的价格接口、账号折扣、充值汇率和模型名称映射实现。适配器完成前不把其公开零售价直接当成 OpenBridger 采购成本，避免产生错误毛利。日志库成功率聚合和模型自动隐藏仍保留在后续阶段。

售价默认仍由管理员批准后发布。主节点另设 `PRICING_AUTO_PUBLISH=true` 且 `PRICING_AUTO_PUBLISH_OPERATOR_ID` 为启用中的 root 用户时，只有 `pricing_offer_sync` 会在成功导入并重算后自动发布。模型策略的 `auto_publish_change_bps` 为 0 时保持人工审批；大于 0 时，仅当 default 档、报价有效、涨跌幅不超过该带宽、预期毛利不低于 `minimum_margin_bps`、建议售价不低于压力成本（`below_cost` 硬底线，即使最低毛利设为 0 也不会低于成本发布）、且现行表达式为空或已是 `tier("base", p * … + c * …)` 且没有缓存/音频/图像倍率时才写入售价。

自动发布前还会做成本多次校验（`PRICING_AUTO_PUBLISH_VERIFY_SAMPLES`，默认 3，设 1 关闭）：主、备渠道的当前报价必须与最近 N 次同步采样的成本快照一致（容差 `PRICING_AUTO_PUBLISH_VERIFY_TOLERANCE_BPS`，默认 5%），否则跳过（`cost_unverified`），避免单次上游读数异常驱动无人值守调价。成本经多次采样确认一致的降价可以自动发布；未启用或未通过校验的降价、超带宽、复杂表达式和首次没有现售价的模型继续留在待审列表。`pricing_recalculate` 不发布售价。

### 2026-11 整改后的状态（与上文的差异以此节为准）

| 主题 | 现状 |
|---|---|
| 报价有效期 | `PRICING_OFFER_TTL_MINUTES`（默认 120，最低 15）。**必须 ≥ 同步间隔**；`TTL < INTERVAL` 时首轮同步记 warning |
| 分组倍率 | 解析上游 `model_group_ratio[model][group]`，存在时优先于顶层 `group_ratio[group]`；一个模型可服务多组时由 `PRICING_GROUP_RATIO_POLICY`（`key_group` / `max` / `min`）决定；`max`/`min` 在两层倍率间取边界，避免低估成本。`upstream_model_offers.upstream_group` 记录取值来源组 |
| 采集端点 | 渠道 `setting.pricing_endpoint` 可覆盖，默认 `/api/pricing`；不再硬编码，与管理员「同步端」口径统一 |
| 来源优先级 | `new-api-pricing` > `manual`；低优先级不得覆盖未过期的高优先级行，被保留的数量计入 `kept` |
| 同步结果留痕 | `pricing_sync_runs` 记录每次同步的 trigger / status / imported / kept / skipped / created_proposals / 渠道明细，页面刷新不丢 |
| 手动 vs 定时 | 两条入口共用 `RunPricingOfferSync` → 采集 → 重算 →（`PRICING_AUTO_PUBLISH=true` 时）自动发布 → 落库 |
| 售价写入方 | 只有「批准 → 发布」。`PATCH /api/option/model_pricing` 在启用动态定价后拒绝批量写（403），单模型需 `override.enabled=true` + 理由；`PublishPricingProposal` 不受影响 |
| 策略层级 | `model_price_policies.scope`：`global`（`public_model='*'`，只提供毛利/成本默认值，不生成方案、不进风险列表）+ `model`（覆盖项留 0 表示继承全局，只影响该模型） |
| 清单 | `GET /api/pricing-control/offers` 支持 `q` / `source_type` / `channel_id` / `include_expired` / `limit` / `offset`，默认只返回 `enabled=1 且未过期`；页面直接展示「实际采购成本 = 基础单价 × 分组倍率」 |
| 方案列表 | `GET /api/pricing-control/proposals` 分页（`limit`/`offset`），返回 `{items,total}`，不再硬截断 500 条 |
| 风险列表 | 每条带 `channel_id` / `channel_role` / `last_collected_at` / `expires_at` / `upstream_group` / `suggested_action`，页面给「重新同步」「查看报价」入口 |
| 数据保留 | `pricing_retention` 任务：成本快照保留 `PRICING_SNAPSHOT_RETENTION_DAYS`（30）天，过期报价保留 `PRICING_EXPIRED_OFFER_RETENTION_DAYS`（7）天。快照仍为追加式，因为自动发布的多重校验依赖最近 N 条样本 |
| 同步资料 | 新建模型默认只写 `description/icon/tags/vendor/endpoints`，`status` 默认 1；`status` 与 `name_rule` 必须显式勾选才写，避免批量同步反向隐藏在售模型 |

### 2026-11 第二阶段：成本倍率改为实测（自校准）

第一阶段的倍率解析仍然建立在「上游会公开真实结算倍率」这个假设上，而用真实 key 对账后这个假设不成立：

- 带 key 与匿名请求 `/api/pricing` 响应**逐字节相同**，`auto_groups` 拿不到「这把 key 属于哪组」，`key_group` 实际等于「按顺序猜第一个组」。
- `model_group_ratio[m][g] / group_ratio[g]` 在 750 个条目里只有 `1/6.8` 和 `1.0` 两种取值，对应上游 `usd_exchange_rate=6.8`，是**币种换算**而非模型级定价。
- 真实流量走的是公开表里没有的组（`claude-kiro`、`openai-luna`、私有 `openai-ent-13481` 等），且同一响应里各模型实际倍率跨 0.02–1.0。
- 结果：`max` 策略把 claude 系成本高估 27–49 倍，`gemini-3.6/3.7/3.8` 与模型级一致（1.0 倍），`grok`/`gemini-pro` 反而差了 6.8 倍。

因此新增 `upstream_cost_calibrations` 与 `pricing_calibration` 任务，用上游自己的消费日志做端到端对账：

```
因子 = (quota ÷ quota_per_unit) ÷ ((prompt−cache)×in + cache×cache_price + completion×out) ÷ 1e6
```

分子是上游实收，分母按我们 offer 表同一套单价计算，因此 `基础成本 × 因子 ≡ 上游账单`，与上游内部用倍率、分档表达式还是按次计费无关。

| 主题 | 现状 |
|---|---|
| 生效顺序 | 校准因子（新鲜且已生效）> `upstream_group_ratio`（兜底）。有校准即可定价，即使上游根本没公布该模型的分组倍率 |
| 失效回退 | 超过 `PRICING_CALIBRATION_STALE_HOURS`（96）未更新的校准自动失效，回退分组倍率；未生效的行（样本不足 / 离散度高）不参与定价 |
| 兜底策略默认值 | `PRICING_GROUP_RATIO_POLICY` 默认改为 `median`（多组取中位数）。`max` 不再作为默认：实测它是把成本抬高一个量级的直接原因 |
| 数据来源 | `GET /api/log/token?type=2`，路径可由 `PRICING_CALIBRATION_LOG_PATH` 覆盖；逐条含 `model`/`quota`/`prompt_tokens`/`completion_tokens`/`cache_read_tokens`/`group` |
| 置信判定 | 样本数 ≥ `PRICING_CALIBRATION_MIN_SAMPLES`（5）且离散度 MAD/中位数 ≤ `PRICING_CALIBRATION_MAX_DISPERSION_PCT`（15%）才 `applied=1`；中位数抗单条异常 |
| 人工兜底 | `PUT /api/pricing-control/calibrations` 可钉住手工实测值（低流量模型），自动任务不再覆盖；填 0 解除 |
| 可观测 | 页面「成本校准」：覆盖率、策略高估倍数、最近校准时间、实测结算分组；`GET /api/pricing-control/calibrations` |
| 风险项 | `*_cost_ratio_deviation`：校准与策略差 ≥ 3 倍时告警（说明策略口径已不可信）。未校准不进风险列表，改为覆盖率数字，避免上百条同义告警 |
| 保留策略 | `PRICING_CALIBRATION_RETENTION_DAYS`（30）天；人工钉住的行不删 |
