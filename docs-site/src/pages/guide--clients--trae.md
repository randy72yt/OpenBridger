# 在 Trae 中使用 OpenBridger

如果当前 Trae 版本支持自定义模型或 OpenAI Compatible Provider，可按通用方式接入。

<div class="config-card"><div><span>Provider</span><code>OpenAI Compatible</code></div><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>YOUR_API_KEY</code></div><div><span>Model</span><code>控制台中的完整模型 ID</code></div></div>

保存后先发送一条短消息，再到使用日志确认请求。若客户端要求填写完整 Endpoint，才追加 `/chat/completions`。

<div class="callout warning"><strong>版本差异</strong><p>菜单名称和自定义模型能力以当前客户端为准。内置模型与自定义 Provider 可以并存，切换时注意当前选中的供应商。</p></div>

## 使用与切换

- 无需退出 Trae 账号；编辑器账号与自定义 Provider 是不同配置。
- 如果只有聊天可用而代码补全不可用，分别检查 Chat、Agent 和 Completion 的模型入口。
- 只有字段明确要求完整 Endpoint 时，才在地址后追加 `/chat/completions`。
- 恢复内置服务时切回官方模型并停用 OpenBridger Provider。
