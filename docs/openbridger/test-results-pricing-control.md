# OpenBridger 动态定价控制 — 测试执行结果

分支：`test/openbridger-qa`　执行时间：2026-09-16　状态：**已执行，含 4 项失败**

本报告只记录实际执行结果与客观事实。是否修复、如何修复不在本文件范围内。

## 1. 执行概要

| 项目 | 数量 |
| --- | --- |
| 已执行用例 | 43 |
| 通过 | 39 |
| 失败 | 4 |
| 未执行（阻断） | 6 |
| 跳过后补测 | 0 |

失败 4 项均为**代码与既有约定不符**，不是断言写法问题（断言依据与修正过程见第 5 节）。

## 2. 执行环境

| 项目 | 实际值 |
| --- | --- |
| Go | go1.26.1 darwin/amd64 |
| SQLite | in-memory（`service` 包 TestMain） |
| MySQL | 8.0.46（docker，端口 3399） |
| PostgreSQL | 17.11（docker，端口 5499） |
| 上游 API Key | 无（本地 `one-api.db` 渠道数为 0） |

## 3. 结果明细

### 3.1 PC-CALC 计算正确性 — 7 条全部通过

| ID | 场景 | 实际结果 | 结论 |
| --- | --- | --- | --- |
| CALC-001 | default 无备用 | 成本 1.05/5.25，售价 1.61538462/8.07692308，毛利 3500 | 通过 |
| CALC-002 | default 有备用 | 成本 1.155/5.775，售价 1.77692308/8.88461538，毛利 3500 | 通过 |
| CALC-003 | stable 压力保护价 | 售价 2.625/13.125，毛利 5600 | 通过 |
| CALC-004 | stable 无备用 | 退化为 1.61538462/8.07692308 | 通过 |
| CALC-005 | 压力价抬高后毛利回算 | 毛利 5600，不等于目标毛利 3500 | 通过 |
| CALC-006 | overhead=0 | 成本 1.1/5.5，售价 1.69230769/8.46153846 | 通过 |
| CALC-007 | 发布表达式与建议价一致 | 表达式 `tier("base", p * 1.77692308 + c * 8.88461538)` | 通过 |

### 3.2 PC-EDGE 边界与异常输入 — 14 通过 / 2 失败

| ID | 场景 | 实际结果 | 结论 |
| --- | --- | --- | --- |
| EDGE-001 | 零成本渠道 | 售价 0.00000000/0.00000000，毛利 3500 | **失败** |
| EDGE-002 | stable 备用零成本 | 售价 1.45384615/7.26923077（主渠道成本为正，保护价未退化） | 通过 |
| EDGE-003 | 极低非零成本 | 售价 0.00000162/0.00000808，未下溢 | 通过 |
| EDGE-004 | fallback=0 | default 1.53846154/7.69230769；stable 仍取备用压力价 2.5/12.5 | 通过 |
| EDGE-005 | fallback=10000 | 售价 3.07692308/15.38461538（等于备用渠道成本） | 通过 |
| EDGE-006 | target=9999 | 售价 10000.00000000/50000.00000001，无溢出 | 通过 |
| EDGE-007 | 备用远贵于主 | 售价取保护价 62.5/125，压力毛利 2000 | 通过 |
| EDGE-008 | 舍入一致性 | 建议价 1.42857143，表达式 `p * 1.42857143`，逐位一致 | 通过 |
| EDGE-016 | 计费单位 | 100 万 token 时 rawCost 10661538.46，换算后 10.66153846 | 通过 |
| EDGE-009 | 缓存读取成本 | CacheReadCost=0.5 时售价 1.53846154/7.69230769，与缓存成本 0 时相同 | 通过（行为记录） |
| EDGE-010 | 非法成本值 | -1/NaN/±Inf 全部被拒，入库 0 条 | 通过 |
| EDGE-011 | 币种校验 | CNY 拒绝、空拒绝、小写 usd 接受 | 通过 |
| EDGE-012 | 成功率越界 | -1 与 10001 均被拒 | 通过 |
| EDGE-013 | 批量部分非法 | 返回错误，但已入库 1 条 | **失败** |
| EDGE-014 | 成本无变化 | 涨跌 0，PricingChanged=false | 通过 |
| EDGE-015 | 缺字段 | 空 model、channel_id=0、空批次均被拒 | 通过 |

### 3.3 PC-STATE 状态机 — 9 条全部通过

| ID | 场景 | 结论 |
| --- | --- | --- |
| STATE-001 | pending→approved→published | 通过（approved_by=7，时间戳正常） |
| STATE-002 | 重复发布 | 通过（第二次 ErrPricingProposalState） |
| STATE-003 | 未批准直接发布 | 通过 |
| STATE-004 | rejected 后批准/发布 | 通过 |
| STATE-005 | stable 档发布 | 通过（错误：`only the default service tier can publish a global model price`） |
| STATE-006 | 价格锁定 | 通过（不生成方案） |
| STATE-007 | 策略禁用 | 通过（不生成方案） |
| STATE-008 | 连续重算 3 次 | 通过（pending 仍为 1 条） |
| STATE-009 | 发布后定价模式 | 通过（billing_mode=tiered_expr） |

### 3.4 PC-CONC 过期、并发与风险 — 6 通过 / 2 失败

| ID | 场景 | 实际结果 | 结论 |
| --- | --- | --- | --- |
| CONC-001 | 报价过期后批准 | 批准成功（未拦截） | **失败** |
| CONC-002 | 价格版本变化后批准 | 批准成功（未拦截） | **失败** |
| CONC-003 | 并发发布 | 成功数 1，失败方返回 `model pricing changed; reload before saving` | 通过 |
| CONC-004 | 空批次不覆盖成本 | 有效成本保持 1/5 | 通过 |
| CONC-005 | 主渠道过期 | 不生成方案，风险码 `primary_offer_missing_or_expired` | 通过 |
| CONC-006 | 备用渠道过期 | 不生成方案，风险码 `backup_offer_missing_or_expired` | 通过 |
| CONC-007 | disabled 报价 | 等同缺失，不生成方案 | 通过 |
| CONC-008 | 同模型同渠道 | 取 collected_at 最新的报价 | 通过 |

### 3.5 PC-DB 三数据库矩阵 — 全部通过

| 数据库 | 版本 | 建议售价 | 毛利 | 快照行数 | 发布 | 结论 |
| --- | --- | --- | --- | --- | --- | --- |
| SQLite | in-memory | 3.28125000/16.40625000 | 5342 | 2 | 成功 | 通过 |
| MySQL | 8.0.46 | 3.28125000/16.40625000 | 5342 | 2 | 成功 | 通过 |
| PostgreSQL | 17.11 | 3.28125000/16.40625000 | 5342 | 2 | 成功 | 通过 |

三种数据库的售价与毛利完全一致（差值 0）。数值已手工复算：加权成本 1.46025832，压力成本 2.625，保护价 `2.625/(1-0.2)=3.28125`，建议价 `1.46025832/0.65=2.2466…`，取 max 后为 3.28125。

### 3.5.1 原有矩阵测试 `TestPricingControlDatabaseMatrix` 的执行情况

该测试是仓库自带的三库矩阵用例，此前因未设置 `TEST_MYSQL_DSN`/`TEST_POSTGRES_DSN` 一直处于 skip 状态，**从未真实执行过**。本次执行结果：

| 执行条件 | 结果 |
| --- | --- |
| 干净的空库（新建 `qaclean`） | MySQL 通过、PostgreSQL 通过 |
| 已被其他测试写入过数据的库 | MySQL 失败、PostgreSQL 失败 |

失败时的表现为「应生成 1 个方案，实际生成 3 个」。`RecalculatePricingProposals` 遍历全部启用策略，因此库中残留的其他策略会被一并重算。经核查，失败时库中存在 3 条策略，其中 2 条为其他测试写入后未清理的残留。

**客观结论**：该用例对数据库清洁度存在隐含依赖，在未隔离的环境中会产生非产品原因的失败。在干净库上，MySQL 8.0.46 与 PostgreSQL 17.11 均通过。

### 3.6 未执行（阻断）

| ID | 场景 | 阻断原因 |
| --- | --- | --- |
| PC-PERM-001 | 非 Root 调用 9 个定价接口 | 需要运行中的服务与有效会话 |
| PC-PERM-002 | 审计日志内容与脱敏 | 同上 |
| PC-E2E-001 | 真实调用与扣费核对 | 需要上游 API Key；本地 `one-api.db` 渠道数为 0 |
| PC-E2E-002 | 发布后历史日志不变 | 同上 |
| PC-E2E-003 | 主备切换不重复扣费 | 同上，且需要第二个上游 |
| PC-E2E-004 | 发布后审核台刷新 | 同上 |

## 4. 失败项详情

### 4.1 EDGE-001 — 零成本渠道产出 0 售价，同时报告 35% 毛利

**依据**：`docs/pricing-control-plane.md`「免费或异常低价渠道不能形成零成本售价。」

**复现输入**：主渠道成本 0/0，default 档，TargetMarginBPS=3500，MinimumMarginBPS=2000，OverheadBPS=500。

**实际输出**：

```text
成本=0.00000000/0.00000000  压力成本=0.00000000/0.00000000
建议售价=0.00000000/0.00000000  预期毛利=3500bps
```

**代码位置**：`service/pricing_control.go`

- 第 270-271 行：`inputPrice = inputCost / targetDenominator`，成本为 0 时售价为 0。
- 第 279-283 行：毛利回算有守卫 `if inputPrice > 0 && outputPrice > 0`，条件不成立时 `margin = policy.TargetMarginBPS`。

**客观描述**：售价与毛利两个字段互相矛盾。`validPriceNumber`（第 397 行）允许 0 成本通过校验。

**边界说明**：EDGE-002（stable 档、备用渠道零成本）**通过**，因为主渠道成本为正时加权成本仍为正；只有当主渠道本身成本为 0 时才会出现 0 售价。

### 4.2 EDGE-013 — 批量导入部分非法时留下部分入库

**复现输入**：3 条报价，第 1 条合法，第 2 条 `currency=CNY`，第 3 条合法。

**实际输出**：函数返回错误，但数据库中已写入 **1 条**（第 1 条），第 3 条未处理。

**代码位置**：`service/pricing_control.go` 第 161-186 行，`for` 循环内校验通过后立即调用 `model.UpsertUpstreamModelOffer`，遇到非法项 `return`，无事务包裹整批。

**客观描述**：导入失败后数据处于部分更新状态，且调用方只能拿到「某一条非法」的错误，无法得知已写入了哪些。

### 4.3 CONC-001 — 报价过期后仍可批准方案

**依据**：`docs/openbridger/project-profile.md` DEC-004「过期或价格版本变化的方案禁止审批」；`production-acceptance.md` BILL-004「过期报价：阻止发布并要求刷新，旧价格不变」。

**复现步骤**：导入有效报价 → 生成 pending 方案 → 将 `expires_at` 改为 `now-1` → 调用 `ApprovePricingProposal`。

**实际输出**：返回 `nil`（批准成功）。

**代码位置**：`service/pricing_control.go` 第 316-330 行，仅校验 `status = pending`，未检查报价有效期。

### 4.4 CONC-002 — 价格版本变化后仍可批准方案

**依据**：同 DEC-004。

**复现步骤**：生成 pending 方案（记录 `PricingVersion`）→ 通过 `model.UpdateModelPricing` 修改该模型价格使版本变化 → 调用 `ApprovePricingProposal`。

**实际输出**：返回 `nil`（批准成功）。

**代码位置**：同上。版本变化标记 `PricingChanged` 在 `ListPricingProposalReviews` 第 112 行已计算，但服务端批准入口未使用该字段。

## 5. 关于断言的两处修正（说明失败项为何只剩 4 条）

为保证结果真实，以下两条最初也报失败，经核实是我的断言写错，已修正，不计入失败：

- **CONC-003 并发发布**：我最初断言「最终价格不应是 99」，实际并发顺序不确定，CAS 正确的那次是 99 的方案。实质结论（成功数 = 1、CAS 生效）成立，已改为断言「最终价格必须是两个候选之一」。
- **EDGE-016 计费单位**：我最初直接用 `RunExpr` 返回值对比建议价，实际 `RunExpr` 返回 `p × 系数`，`/1_000_000` 换算在 `relay/helper/price.go` 第 347 行由调用方完成。实际值 `10661538.46 = (1.77692308+8.88461538)×10⁶`，换算后 `10.66153846` 与建议价之和一致，单位链条正确。

另外 EDGE-006 在 target=9999 时输出 `50000.00000001` 而非 `50000`，为分母接近 1e-4 时的浮点误差（1e-8 量级），非缺陷，断言容差已按 1e-7 处理。

## 6. 复现方式

```bash
# 服务层全部用例（SQLite）
go test ./service/ -run 'TestQAPricing' -v

# 三数据库矩阵
docker run -d --name qa-mysql -e MYSQL_ROOT_PASSWORD=qapass -e MYSQL_DATABASE=qaprice -p 3399:3306 mysql:8.0
docker run -d --name qa-pg -e POSTGRES_PASSWORD=qapass -e POSTGRES_DB=qaprice -p 5499:5432 postgres:17

TEST_MYSQL_DSN="root:qapass@tcp(127.0.0.1:3399)/qaprice?charset=utf8mb4&parseTime=True&loc=Local" \
TEST_POSTGRES_DSN="host=127.0.0.1 port=5499 user=postgres password=qapass dbname=qaprice sslmode=disable" \
  go test ./service/ -run 'TestQAPricingDatabaseMatrix' -v
```

**新增测试文件**：

- `service/pricing_control_qa_test.go` — PC-CALC / PC-EDGE / PC-STATE / PC-CONC
- `service/pricing_control_db_test.go` — PC-DB 三数据库一致性

两个文件只新增测试，未修改任何业务代码。

## 7. 真实上游端到端（PC-E2E）执行结果

### 7.1 环境

- 隔离实例：独立 sqlite `/tmp/qa-e2e/ob.db`，端口 3200/3201/3202。项目根 `one-api.db` 未被触碰（修改时间仍为 09-08 22:11）。
- 上游：中转站 `https://api.dddai.dev`（OpenAI 兼容）。该中转站自身即 new-api，错误响应 `type: new_api_error`。
- 渠道：id=1 `qa-upstream`（有效 key）；id=2 `qa-bad-primary`（key 故意无效，priority=100）。
- 模型：`deepseek-v4-flash`。

### 7.2 结果总览

| 用例 | 结果 | 实测 |
|---|---|---|
| E2E-001 真实非流式调用 | 通过 | 200，usage `prompt=85 / completion=21` |
| E2E-002 真实流式调用 | 通过 | 200，扣费与非流式**完全一致** |
| E2E-003 无效模型不扣费 | 通过 | 503 `model_not_found`，无 log、无扣减 |
| E2E-004 主渠道 401 时故障切换 | **未切换** | 上游 401 被透传，未重试备用渠道 |
| E2E-005 扣费金额核对 | **不符** | 实扣 3975，按上游倍率应为约 111 |

### 7.3 E2E-004 详情：上游 401 不触发切换

日志实证：

```
[ERR] channel error (channel #2, status code: 401): Invalid token
[ERR] relay error: Invalid token
```

请求被路由到 priority=100 的坏渠道，上游返回 401，OpenBridger **直接把上游错误原文（含 "Invalid token" 字样）透传给客户端，未切换到可用的 channel #1**。

排查提示：这个 401 极易被误判为"OpenBridger 自身令牌无效"。区分依据是日志中 `channel error (channel #2, ...)` 前缀——那是上游返回的状态，不是网关自身的鉴权失败。

### 7.4 E2E-005 详情：未知模型按 37.5 兜底计费（已修正根因定位）

> **2026-09-16 修正**：本节初稿将根因归为"new-api 无法获取上游价格"，**该判断错误**。经追加验证，new-api 原生支持从上游同步模型与价格（见 7.5.1）。37.5 兜底的真实触发条件是：**未执行倍率同步**，本地倍率表中无此模型。以下为实测事实与修正后的结论。

**实测事实（未同步状态下）**

- 实际扣费 3975。日志 `other` 字段记录：`model_ratio: 37.5`、`completion_ratio: 1`、`group_ratio: 1`、`billing_source: wallet`。
- 实扣核算：`(85 + 21 × 1) × 37.5 = 3975` ✓
- 兜底位置：`setting/ratio_setting/model_ratio.go:760`，`GetModelRatioOrPrice` 在模型名未命中本地倍率表时 `return 37.5, false, false`（第三个返回值 `exist=false`）。

**修正后的根因**

不是"产品不支持获取上游价格"，而是**测试环境从未执行过倍率同步**。补充实测（`POST /api/ratio_sync/fetch`，channel_id=1）结果：

- 请求成功，返回 **39 个模型的定价值**（`test_results: [{name: 'qa-upstream(1)', status: 'success'}]`）。
- 本地 `current` 全部为空 `{}`，即本地倍率表确实没有这些模型。
- 上游同时下发了 `billing_mode` 与 `billing_expr`：`deepseek-v4-flash` 使用的是**分档计费表达式**（峰谷分时），并非固定 ratio：

  ```
  weekday("UTC") >= 1 && weekday("UTC") <= 5 && ((hour("UTC") >= 1 && hour("UTC") < 4) || (hour("UTC") >= 6 && hour("UTC") < 10))
    ? tier("peak", p * 0.44 + cr * 0.014 + c * 1.32)
    : tier("off_peak", p * 0.22 + cr * 0.007 + c * 0.66)
  ```

- 初稿所称"上游给 `model_ratio = 0.75`"不适用于该模型——`billing_mode = tiered_expr` 时走表达式计费，ratio 字段不参与。

**结论调整**

"差 35.8 倍"是由未同步状态产生的观测值，**不构成产品缺陷的直接证据**。仍然成立且值得关注的事实是：

1. 未命中本地倍率表时静默按 37.5（约 $75/1M tokens，属最贵档）计费，无告警、无日志标记。
2. `ratio_sync.go:677-691` 存在专门的可信度校验：上游返回 `model_ratio=37.5 且 completion_ratio=1.0` 组合时标记 `confidence=false`。这说明该组合在代码内已被认定为无意义信号，但本地兜底路径仍会产生同样的值。
3. 倍率同步为**手动触发**（fetch 仅返回差异，需人工确认后应用），无定时自动同步；模型列表侧则支持自动同步（`UpstreamModelUpdateAutoSyncEnabled`）。

**待补测**：同步应用上游倍率后重新调用，核对扣费金额是否符合表达式预期。本次未执行（需重启实例刷新倍率缓存，且每次真实调用产生上游费用）。

### 7.5 上游能力实测

| 探测项 | 结果 |
|---|---|
| `GET /v1/models` | 分组调整后返回 40 个模型（调整前为空数组） |
| `GET /api/pricing` | 可用，40 条；**匿名可访问**（ratio_sync 拉取时不带认证头，实测仍返回 200 / 17741 字节） |
| 流式 usage | **默认每个 chunk 都带 usage**，无需 `stream_options` |
| 模型名改写 | `deepseek-v4-flash` → 上游返回 `deepseek-v4-flash-ga-260731` |

#### 7.5.1 new-api 上游同步能力（追加验证，修正此前错误判断）

初次报告中"new-api 只取模型 ID 不取价格""37/40 条只有 ratio 因而不可用"的说法**均不成立**。代码与实测如下：

**价格同步** —— `controller/ratio_sync.go`，路由 `POST /api/ratio_sync/fetch`（`RootAuth` 保护），默认端点常量 `defaultEndpoint = "/api/pricing"`（`ratio_sync.go:33`）。支持四种上游格式：

| 类型 | 上游端点 | 处理方式 |
|---|---|---|
| type1 | `/api/ratio_config` | `data` 为 map，直接采用 |
| type2 | `/api/pricing` | `data` 为 `[]Pricing` 列表，转换为统一 map |
| type3 | OpenRouter `/v1/models` | per-token 价格换算为 ratio，`ratio = price × 1000 × USD` |
| type4 | models.dev `/api.json` | USD/1M 换算为 ratio，`ratio = input × USD / 1000` |

同步字段（`pricingSyncFields`，`ratio_sync.go:64-75`）覆盖 `model_ratio`、`completion_ratio`、`cache_ratio`、`create_cache_ratio`、`image_ratio`、`audio_ratio`、`audio_completion_ratio`、`model_price`，以及 `billing_mode` + `billing_expr`。**分档计费表达式可随同步下发**，不止是固定倍率。

另内置两个预设来源（`GetSyncableChannels`）：官方倍率预设 `basellm.github.io`、models.dev 价格预设。

`ratio_sync.go:677-691` 设有可信度校验：上游返回 `model_ratio=37.5 且 completion_ratio=1.0` 时将该条目标记 `confidence=false`，UI 可据此提示。

**模型同步** —— `controller/channel_upstream_update.go`，`fetchChannelUpstreamModelIDs` 拉取 `/v1/models`。支持自动同步：`UpstreamModelUpdateAutoSyncEnabled` 开启后，后台任务 `runChannelUpstreamModelUpdateTaskOnce`（`:688`）按 `getUpstreamModelUpdateMinCheckIntervalSeconds` 节流自动追加新模型（`:534`）。手动入口：`/api/channel/upstream_models/detect|apply`（`:855`、`:911`）。

**实测结果（对本次中转站渠道）**

```
POST /api/ratio_sync/fetch {"channel_ids":[1],"timeout":20}
→ HTTP 200, test_results: [{name: 'qa-upstream(1)', status: 'success'}]
→ differences: 39 个模型, prices: 39 条
→ confidence 统计: 可信 114 条 / 不可信 0 条
```

样例（本地 `current` 均为空 `{}`）：

- `claude-haiku-4-5`：`model_ratio 0.5`、`completion_ratio 5`、`cache_ratio 0.1`、`create_cache_ratio 2`
- `claude-sonnet-5`：`model_ratio 1`、`completion_ratio 5`、`cache_ratio 0.1`、`create_cache_ratio 1.25`
- `deepseek-v4-flash`：`billing_mode = tiered_expr` + 峰谷分时表达式（见 7.4）
- `gemini-2.5-flash-lite`：`billing_mode = tiered_expr` + `tier("standard", p*0.1 + img*0.1 + ai*0.3 + c*0.4 + cr*0.01 + cc*0.0833333333)`

**计价单位换算**：`USD = 500`（`model_ratio.go:15`，`$0.002 = 1`），即 `1 ratio = $0.002/1K tokens = $2/1M tokens`。由此 ratio 与绝对价格的换算链路是通的，此前"需知道上游额度单价才能换算"的判断不成立。

**仍然成立的断点**：上述能力作用于 new-api 原生的 ratio/表达式计价体系；OpenBridger 自研的定价控制平面（`UpstreamModelOffer`）要求输入绝对成本 USD/1M，其唯一写入路径是手工导入 `ValidateAndImportPricingOffers`。**上游价格 → pricing_control offers 这一段在代码上没有自动通道**，这才是此前"拿不到价格"说法中唯一成立的部分。

### 7.6 未完成项

- 主备切换的完整矩阵（同级 priority 重试、5xx 重试、超时重试、重复扣费验证）。
- 阻塞原因：坏渠道一经建入，其渠道缓存/ability 未随禁用操作刷新，后续请求持续命中它；同时每次真实调用都产生上游费用。改用本地 mock 上游可完整覆盖且不产生费用。
