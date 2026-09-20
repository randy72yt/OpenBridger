# OpenBridger 项目控制中心

> 本目录是 OpenBridger 从本地开发到生产运营的长期记录入口。功能、配置、上线决定和验收证据以这里为索引，不依赖聊天记录。

## 文档入口

| 文档 | 用途 |
| --- | --- |
| [项目档案](./project-profile.md) | 保存产品定位、系统组成、已确认决定、当前状态、外部依赖和待确认事项 |
| [生产验收](./production-acceptance.md) | 上线前逐项执行的测试用例、通过标准和证据记录 |
| [上线计划](./launch-plan.md) | 按依赖关系推进上线准备，并记录每个阶段的完成状态 |

第一个版本的发布前必做项与可延期项已整理在[上线计划的 Release 1 待办总览](./launch-plan.md#release-1-待办总览)。发布放行仍以[生产验收](./production-acceptance.md)中所有适用 P0 用例的实际证据为准。

## 动态定价控制面（pricing control）测试与缺陷资产

| 文档 | 用途 | 当前状态 |
| --- | --- | --- |
| [测试用例](./test-cases-pricing-control.md) | 40+ 条用例设计与 4 处风险（R1–R4） | 已评审，已执行 |
| [执行结果](./test-results-pricing-control.md) | 43 条服务层用例、三库矩阵、真实上游实测记录 | 已完成；第 9 节为「本地扣费 vs 上游真实成本」实测 |
| [缺陷清单](./defect-list-pricing-control.md) | 5 项缺陷的完整定义、复现步骤、依据与定级 | **全部已关闭**，作为回归时的原始验收标准 |
| [修复复验报告](./fix-review-pricing-control.md) | 三轮复验结论。**第 0.1 节「最终状态」表为唯一权威结论** | 已完成；无阻断缺陷，可合入 |
| [按次计费模型改造规格](./spec-per-request-pricing.md) | 生图等 `quota_type=1` 模型纳入动态定价的实现方案 | 待排期实施 |

### 阅读顺序

1. 先看 `fix-review-pricing-control.md` **第 0.1 节最终状态表**（12 项，含关闭/遗留状态），确认当前进度；第 1–5 节属于历史记录，仅用于溯源。
2. 再按下列「待办事项 → 阅读位置」定位具体工作的依据。
3. `defect-list-pricing-control.md` 只在做回归时查阅，用于核对原始缺陷定义与验收标准。

### 待办事项与阅读位置

事项记录在项目事项系统，标签统一为 `pricing-control`。

| 事项编号 | 事项 | 优先级 | 开工前读 |
| --- | --- | --- | --- |
| r6UjdF | 配置 `PRICING_CONTROL_ENABLED=true`，否则报价 2 小时后全过期、动态定价停摆 | urgent | `fix-review` 6.9 ①（含三种部署方式的配置位置与验证 SQL） |
| r2VsH5 | 生图等按次计费模型纳入动态定价（gpt-image-2/2.5、gemini-3-pro-image） | high | `spec-per-request-pricing.md` 全文，重点第 2 节可行性依据、第 6/7 节改动清单、第 8 节测试验收 |
| rpOYB6 | 生产验收：核对本地扣费与上游账单增量，确认倍率同步生效 | high | `test-results` 第 9 节（真实成本测量方法）＋ `fix-review` 6.2（倍率匹配规则） |
| rhU308 | 同步结果未展示 warnings，被跳过的模型运营不可见（本分支已修复，待 CI 确认） | medium | `fix-review` 6.3、6.6 RISK-005 |
| r008SV | EDGE-001 零成本渠道语义追认（报错 / 保护价 / 标记异常） | medium | `fix-review` 1.3 ＋ `defect-list` EDGE-001 |
| ryIHwN | 三库矩阵缺 DSN 静默跳过；补存量 NULL 倍率回归用例（本分支已修复并本地三库通过，待 CI 确认） | medium | `fix-review` 3.1、6.5、6.9 ③（含三库基准值） |
| rftiXH | 主备切换与故障转移验证（需第二个上游渠道） | medium | `fix-review` 3.2；前提：`RetryTimes` 默认 0、禁用渠道须走 `POST /api/channel/:id/status` |
| rx3c81 | `ErrPricingProposalStale` 错误码语义混用 | low | `fix-review` 2. RISK-003 |

### 代码位置

上述资产与修复代码位于分支 **`codex/fix-pricing-control-qa`**（以当前分支 HEAD 为准，**未合并到 main**）。合入前请以 `fix-review-pricing-control.md` 第 0.1 节结论为准，遗留项按[上线计划](./launch-plan.md#release-1-待办总览)区分是否阻断首版；`r6UjdF` 为上线必做项。

### 复用价值高的两条经验

- **测真实成本不需要上游管理后台**：若上游是 new-api，用同一把 key 调 `GET /v1/dashboard/billing/usage`，`quota = total_usage × 5000`，取调用前后差值即为上游真实扣费（详见 `test-results` 第 9 节）。
- **「CI 绿」不等于「跑过三库」**：`TestQAPricingDatabaseMatrix` 在缺 DSN 时静默跳过且不计失败，查看 CI 结果必须确认日志里真的出现 `mysql` 与 `postgres` 两个子测试的 PASS 行。

## 维护规则

1. 产品、架构、运营或合规决定确认后，更新“项目档案”的决定记录。
2. 开发项只有在代码完成且相关验证通过后才能标记为“已完成”。
3. 生产配置完成后，只记录配置项名称、负责人和验证结果，不在文档中保存密码、Token、API Key 或商户密钥。
4. 每个生产验收用例必须记录执行日期、环境、执行人、结果和证据位置。
5. 验收失败时记录实际结果和问题编号；修复后重新执行原用例，不通过口头确认关闭。
6. 上线计划按照阶段顺序推进。存在阻断项时，不开始依赖该项的生产发布。

## 状态定义

| 状态 | 含义 |
| --- | --- |
| 待确认 | 仍需要产品或运营决定 |
| 待实施 | 决定已明确，尚未开始开发或配置 |
| 进行中 | 正在开发、配置或验证 |
| 待验收 | 实施完成，等待规定测试 |
| 已完成 | 验收通过并已记录证据 |
| 阻断 | 因外部资料、账号、审批或故障暂时无法继续 |

## 当前发布目标

第一阶段采用邀请制生产试运行，优先验证文本模型 API、主备渠道、计费、余额、日志和管理流程。在线支付是否随第一阶段开放，由上线计划中的支付决策门决定。

生产入口规划：

- 主站与控制台：`https://openbridger.com`
- API Base URL：`https://openbridger.com/v1`
- 文档站：`https://docs.openbridger.com`

上述地址在 DNS、TLS 和反向代理完成前属于规划值。
