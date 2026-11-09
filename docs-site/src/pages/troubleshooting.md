# 故障排查

先用 curl 把网关和客户端分开。curl 成功、客户端失败，就查客户端的地址和缓存。curl 也失败，就查密钥、额度和模型 ID。

## 自检

```bash
curl -i "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer ${OPENBRIDGER_API_KEY}"
```

- HTTP 200 且有模型列表：密钥和地址可用。
- 401：密钥本身不可用。见 [错误码](/guide/errors/)。
- 403 且提到额度：去 [钱包]({{CONSOLE_URL}}/wallet)。
- 连接不上：确认主机名是 `openbridger.com` 或 `api.openbridger.com`，路径不要写成 `/v1/v1/models`。

然后再发一条 [快速开始](/guide/quickstart/) 里的短对话，到 [使用日志]({{CONSOLE_URL}}/usage-logs) 对请求 ID。

## 对照表

| 现象 | 常见原因 | 处理 |
| --- | --- | --- |
| 401「无效的令牌」 | 密钥错、禁用、过期，或密钥额度用尽；多写了 `Bearer` | 从 API 密钥页重新复制，检查该密钥剩余额度和过期时间 |
| 403 额度不足 | 钱包或订阅不够预扣 | 充值，或缩短输入后再试 |
| 403 分组 / IP | 密钥分组不可用，或 IP 白名单不含当前出口 | 改分组或白名单 |
| 503 无可用渠道 | 模型 ID 不属于当前分组 | 从模型广场复制完整 ID |
| 上游 404 / model not found | 上游不认识这个名字，或客户端改写了 ID | 换广场里的另一个 ID，并看日志里实际提交的名字 |
| 429 | 本地频率限制或上游限流 | 看文案是否包含「您已达到请求数限制」。是则等窗口；否则是上游，稍后重试 |
| 超时 | 上下文很长或上游慢 | 用一句短提示、减小输出长度再对比 |
| 流式中断 | 网络或客户端对流的处理 | 同一模型先发 `"stream": false` |
| 扣费偏高 | 历史消息重复、工具调用多 | 在日志里看 token，而不是只看客户端气泡 |
| 改了配置仍走旧地址 | 环境变量或另一个配置文件还在 | 新开终端，检查 `OPENAI_*` 和 `ANTHROPIC_*` |
| 模型列表拉不下来 | 客户端要的不是 `/v1/models`，或它过滤了 ID | 手工填模型 ID |

## 地址填法

| 客户端类型 | 填写 |
| --- | --- |
| OpenAI SDK、Chatbox、Cherry Studio、LobeChat、Cursor 的 OpenAI Base URL、Open WebUI | `{{API_BASE_URL}}` |
| 只有「完整 Endpoint」一栏 | `{{API_BASE_URL}}/chat/completions` |
| Claude Code、Anthropic SDK | `{{CONSOLE_URL}}`（不要加 `/v1`） |
| Codex | Base URL `{{API_BASE_URL}}`，协议 Responses |

## 相关

- [错误码](/guide/errors/)
- [联系与故障申报](/guide/support/)
