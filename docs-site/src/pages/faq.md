# 常见问题

## Base URL 填什么？

OpenAI 兼容客户端填 `{{API_BASE_URL}}`。只有单独一个「完整 Endpoint」输入框时，才在后面加 `/chat/completions`。

Claude Code 和 Anthropic SDK 填 `{{CONSOLE_URL}}`，不带 `/v1`。说明见 [Claude Code](/guide/clients/claude-code/)。

`https://api.openbridger.com/v1` 与 `{{API_BASE_URL}}` 是同一套网关。文档示例用后者。

## 密钥要不要加 Bearer？

请求头是 `Authorization: Bearer sk-...`。密钥本身已经带 `sk-`。不要写成 `Bearer Bearer` 或 `Bearer sk-sk-`。

## 为什么模型列表是空的，密钥却没报错？

不少客户端不调用 `/v1/models`，或者会丢掉自己不认识的 ID。用 curl 能列出模型时，把 [模型广场]({{CONSOLE_URL}}/pricing) 的完整 ID 手工加进去。

## 怎样确认流量走了 OpenBridger？

给这个客户端单独一把密钥，发一句短消息，然后在 [使用日志]({{CONSOLE_URL}}/usage-logs) 里按密钥和时间查找。只看客户端界面上的模型名不够。

## 要退出原来的官方账号吗？

通常不用。编辑器账号负责登录和同步；自定义 Base URL 只影响走这把密钥的请求。恢复官方模型时，关掉自定义 Base URL 或换回内置供应商，并清掉对应环境变量，然后重启客户端。

## 401 和 403 怎么分？

`401`「无效的令牌」是密钥不可用，包括这把密钥自己的额度用尽。`403` 且 `code` 为 `insufficient_user_quota` 是钱包或订阅不够预扣。分组、IP 白名单和封禁也是 `403`，但文案不同。表在 [错误码](/guide/errors/)。

## 支持哪些客户端？

任何能自定义 Base URL、API Key 和模型 ID 的 OpenAI 兼容客户端都可以按 [通用说明](/guide/clients/openai-compatible/) 试。单独写了步骤的有 Cursor、Claude Code、Codex、CC Switch、Cherry Studio、Chatbox、LobeChat、Open WebUI、Trae、CodeBuddy，以及 OpenAI SDK 和 LangChain。

某个模型是否支持工具调用、图像或该客户端要求的协议，以模型广场和一次短请求的结果为准。

## 钱包和套餐哪个先扣？

默认是「优先订阅」。没有订阅额度时扣钱包。四种偏好见 [订阅套餐](/guide/account/subscriptions/)。协议 v1.1 下套餐购买还未开放。

## 可以怎样充值？

钱包页展示的在线支付（默认是支付宝和微信）以及兑换码。最低金额和到账数以确认页为准。见 [充值与兑换](/guide/account/redeem/)。

## 注册会送额度吗？

程序默认不送（新用户额度为 0）。运营可以另行配置赠送。注册后看钱包里的余额，有一笔入账就是当前的赠送，没有就说明当前没有。

## 请求失败要不要一直重试？

`401`、`403` 额度不足、`400` 参数错误，重试不会成功。`429` 和上游暂时失败可以有限次退避重试。不要写成死循环。

## 相关

- [快速开始](/guide/quickstart/)
- [故障排查](/troubleshooting/)
- [联系与故障申报](/guide/support/)
