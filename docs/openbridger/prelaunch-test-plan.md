# 上线前测试计划

最后更新：2026-09-22
前置：域名已解析、HTTPS 已签发、CF 已配 Full(strict) 与 Bypass、管理员双因子已开、SMTP 通、Turnstile 开。

运营策略（用户 2026-09-22 确定）：

- **注册常开**，不走邀请制。
- **在线支付首版不开放**，试用额度由管理员手工发放。
- 顺序：先跑通无上游的基础流程 → 再挂上游渠道 → 再做上线前总验收。

回归脚本：`scripts/prod-acceptance.sh`（每次发版后跑，退出码非 0 表示有 FAIL）。

---

## 阶段 1：无上游的基础流程（进行中）

目标：证明「注册 → 登录 → 令牌 → 调用」的最小闭环可用，且所有该封的口子都封死。**不需要上游 Key。**

### 1.1 已自动验证 ✅ 17 项全通过

执行 `./scripts/prod-acceptance.sh`，结果 17 PASS / 0 FAIL / 2 SKIP / 4 INFO。

| 组 | 覆盖内容 | 结果 |
| --- | --- | --- |
| A 公开面 | 状态接口、根路径返回前端；`server_address=https://openbridger.com` | PASS |
| B 权限边界 | `/api/user/self`、`/api/token/`、`/api/channel/`、`/v1/models`、`/v1/chat/completions`、`/api/user/pay` 未认证全部 401 | PASS |
| D 支付面 | `POST /api/user/pay`、支付回调 ingress 被拒 | PASS |
| E CF 层 | OpenAI SDK / httpx / axios / Postman / okhttp / Go 六种 UA 全部 200；`/v1/models` 的 `cf-cache-status=DYNAMIC`（不缓存） | PASS |

### 1.2 需要人工做（需浏览器 + 凭证）

| # | 用例 | 通过标准 |
| --- | --- | --- |
| 1 | 注册新账号并收验证码 | 邮件收到；激活后能登录 |
| 2 | 用错误验证码注册 | 被拒，不产生用户记录 |
| 3 | 重复邮箱 / 重复用户名 | 被拒，提示明确 |
| 4 | 用户名 > 20 字符 | 被拒（`model/user.go:81` 的 `validate:"max=20"`） |
| 5 | 登录后开 TOTP | 出现二维码，验证后可正常登录第二次 |
| 6 | 创建 / 只读 / 过期 / 无限额令牌各一枚 | 四类行为符合预期 |
| 7 | 用令牌调 `/v1/models` | 200（当前列表为空，因为没有上架模型） |
| 8 | 管理员给测试账号发放额度 | 控制台看到余额变化 |
| 9 | 额度耗尽后再调 `/v1/chat/completions` | 被拒并给出余额不足的明确提示 |

### 1.3 阻塞项

- **缺一枚有额度的测试令牌** → `OB_API_KEY=sk-xxx ./scripts/prod-acceptance.sh` 才能跑 C 组（认证后闭环）。

### 1.4 本阶段暴露的问题

| # | 问题 | 影响 | 处理 |
| --- | --- | --- | --- |
| N1 | `GET /api/verification` **无需登录**即可发送验证码，限流为每 IP 30 秒 2 次（`middleware/email-verification-rate-limit.go:12`）。Resend 免费额度 100 封/天，单 IP 约 25 分钟即可耗尽 | 注册常开后，真实用户的注册验证码与找回密码邮件会被静默打瘫 | 待决策：加按域名/日期维度的发送上限，或改用带配额的邮件服务 |
| N2 | `User-Agent: Python-urllib/x.y` 被 CF 拦（403 error 1010） | 主流 SDK 不受影响；仅裸 urllib 脚本受影响，改 UA 或加 `Accept` 头即可绕开 | 已记录，不修 |
| N3 | `/api/user-agreement`、`/api/privacy-policy` 返回空字符串 | 协议未写入配置段 | S26 |
| N4 | `docs.openbridger.com` 返回 403 | 静态目录 `/srv/openbridger/docs-site` 为空，未部署 | S27 |

---

## 阶段 2：挂上游渠道后的转接与定价

前置：`compose.release.yml` 渠道 Base 填 `https://<上游域名>/v1`（**不要再拼一层 `/v1`**，重复拼接会 `404 Invalid URL`）。

| # | 用例 | 通过标准 |
| --- | --- | --- |
| 2.1 | 渠道连通性测试 | 后台点测通过 |
| 2.2 | 首批模型逐个非流式调用 | 9 个模型全部返回 200。注意 **`deepseek-flash` 在上游不存在**，须用 `deepseek-v4-flash` |
| 2.3 | 逐个模型跑 SSE 流式 | 经 CF + Nginx 后 `text/event-stream` 完整、`[DONE]` 收尾正确、无缓冲堆积（`proxy_buffering off` 已配） |
| 2.4 | 动态定价同步 | 连续两次同步后本地倍率与上游 `group_ratio` 一致；报价有效期 2 小时 |
| 2.5 | 账单对账 | 调用前后取 `GET /v1/dashboard/billing/usage`，`quota = Δtotal_usage × 5000`，与本地扣费比对 |
| 2.6 | 毛利与人工审批 | 定价方案可生成、可审核、可发布；发布后前台售价与方案一致 |
| 2.7 | 禁用 / 超时 / 上游报错时的行为 | 走正规状态接口，不能靠发坏 key 触发切换（`RetryTimes=0`，相邻渠道不会被自动重试） |

> 倍率已经漂移过（deepseek、kimi 6.8→1，glm 6→0.9），**同步用账单差值重核，不要信历史文档里的数字**。

---

## 阶段 3：上线前最后验收

| # | 用例 | 通过标准 |
| --- | --- | --- |
| 3.1 | OPS-003 备份与恢复 | 备份传到异地，并在临时实例上恢复一次验证可用；纳入定时任务 |
| 3.2 | OPS-004 回滚演练 | 记录当前 tag → 部署上一个 tag → 确认恢复 → 切回 |
| 3.3 | CI 真正跑起来 | fork 需手动启用工作流；推 commit 后出现真实运行记录，且三库矩阵日志中出现 `mysql` 与 `postgres` 两个 PASS 行 |
| 3.4 | 4xx/5xx 告警 | 制造一次错误，确认告警通路可达 |
| 3.5 | 完整 P0（按 `production-acceptance.md` 主线） | AUTH-001~005、API-001~005、RELAY-001、BILL-001~005、LEGAL-001、UI-001、DOC-001、OBS-001、OPS-001~004、E2E-001 全部有结果；PAY-001~003 记「首版不适用」并附阻断证据 |
| 3.6 | 邮件配额保护（N1） | 上线前必须定方案 |

---

## 依赖关系

```
阶段 1.1 ✅ ──► 阶段 1.2（需人工 + 测试令牌）──► 阶段 2（需上游渠道）
                                                      │
                      阶段 1.3 解除阻塞 ◄────────────┘
                                                      │
                               阶段 3 ◄── 阶段 2 全绿 + S26/S27 完成
```

阶段 3 的绝大部分（3.1~3.4）**不依赖阶段 2**，可以并行推进。
