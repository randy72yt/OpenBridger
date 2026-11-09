# 在 CodeBuddy 中使用 OpenBridger

CodeBuddy 提供 Custom Model、OpenAI Compatible 或可覆盖的 API 地址时，可以添加 OpenBridger。没有这些字段的版本不能按本页接入。

<div class="steps"><div class="step"><strong>打开模型设置</strong><p>查找 Model、Provider、API Key 或 Custom Model。</p></div><div class="step"><strong>选择 OpenAI Compatible</strong><p>没有该类型时，看是否允许覆盖 OpenAI Base URL。</p></div><div class="step"><strong>填写并测试</strong><p>Base URL 为 <code>{{API_BASE_URL}}</code>，密钥单独创建，模型 ID 从模型广场复制。保存后发短消息，到使用日志核对。</p></div></div>

<div class="config-card"><div><span>类型</span><code>OpenAI Compatible</code></div><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>sk-...</code></div><div><span>Model</span><code>模型广场中的完整 ID</code></div></div>

不需要退出 CodeBuddy 账号。Chat、Agent 和补全要分别确认供应商。恢复内置服务时切回官方模型，关闭自定义地址，并重启客户端。

## 相关

- [OpenAI 兼容调用](/guide/clients/openai-compatible/)
- [故障排查](/troubleshooting/)
