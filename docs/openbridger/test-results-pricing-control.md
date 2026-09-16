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
