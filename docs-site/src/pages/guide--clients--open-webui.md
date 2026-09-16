# 在 Open WebUI 中使用 OpenBridger

通过界面添加全局连接通常需要 Open WebUI 管理员权限。普通用户只能使用管理员已经启用的模型。

<div class="steps"><div class="step"><strong>进入 Connections</strong><p>在 Admin Panel / Settings 中打开 OpenAI API 连接设置。</p></div><div class="step"><strong>添加连接</strong><p>名称填写 OpenBridger，URL 使用 <code>{{API_BASE_URL}}</code>，Key 使用独立 API Key。</p></div><div class="step"><strong>刷新模型</strong><p>保存后刷新列表；无法自动发现时手动限制或填写模型 ID。</p></div><div class="step"><strong>核对日志</strong><p>发送短消息，并在 OpenBridger 使用日志中确认请求。</p></div></div>

<div class="callout warning"><strong>共享范围</strong><p>管理员保存的连接可能供多个 Open WebUI 用户使用。请设置适当的模型、额度和访问限制。</p></div>

## 容器环境变量方式

```bash
OPENAI_API_BASE_URL={{API_BASE_URL}}
OPENAI_API_KEY=YOUR_API_KEY
```

修改 Docker Compose 或容器环境变量后，需要重新创建或重启 Open WebUI 容器；只刷新浏览器不会加载新值。多人共享同一个全局 Key 时，用量会汇总到该 Key，建议按实例或团队单独创建。
