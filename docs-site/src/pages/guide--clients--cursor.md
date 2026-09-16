# 在 Cursor 中使用 OpenBridger

如果当前 Cursor 版本提供自定义 OpenAI 兼容服务入口，可以填写 OpenBridger 的 Base URL、API Key 和模型 ID。

<div class="steps"><div class="step"><strong>打开模型设置</strong><p>在 Cursor Settings 中查找 Models、API Keys 或自定义 Provider。</p></div><div class="step"><strong>添加兼容服务</strong><p>选择 OpenAI Compatible，或使用允许覆盖 Base URL 的 OpenAI 配置。</p></div><div class="step"><strong>填写模型</strong><p>Base URL 使用 <code>{{API_BASE_URL}}</code>，并从控制台复制完整模型 ID。</p></div><div class="step"><strong>发送短请求</strong><p>保存后在 Chat 或 Agent 中测试，并在 OpenBridger 使用日志中核对。</p></div></div>

<div class="callout"><strong>版本差异</strong><p>Cursor 的菜单和自带 API Key 能力可能随版本或套餐变化。Tab Completion 等专用功能也可能继续使用 Cursor 自有服务。</p></div>

## 账号、内置模型与切换

- 不需要退出 Cursor 账号；账号同步和编辑器功能与自定义模型配置相互独立。
- 如果启用了全局 `Override OpenAI Base URL`，切回官方模型前先关闭该开关。
- Chat、Agent 与 Tab Completion 可能使用不同模型入口，应分别检查。
- Cursor 对自带 API Key 的使用范围可能受当前版本或套餐限制。

## 常见问题

| 现象 | 处理方式 |
| --- | --- |
| 没有 Verify 按钮 | 保存后直接发送短消息，并在 OpenBridger 使用日志确认 |
| model not found | 重新复制完整模型 ID，并核对分组权限 |
| 官方模型也报错 | 关闭全局 Base URL 覆盖并重启 Cursor |
| Network / TLS 错误 | 检查地址与网络；客户端支持时尝试 HTTP/1.1 兼容模式 |
| Tab 没有走 OpenBridger | Tab Completion 可能仍使用 Cursor 自有服务 |
