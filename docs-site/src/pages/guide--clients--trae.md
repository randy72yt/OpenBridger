# 在 Trae 中使用 OpenBridger

Trae 的版本如果提供自定义模型或 OpenAI Compatible 供应商，就用下面的值。菜单里没有 Base URL 或自定义供应商时，这个版本不能按本页接入。

<div class="config-card"><div><span>Provider</span><code>OpenAI Compatible</code></div><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>sk-...</code></div><div><span>Model</span><code>模型广场中的完整 ID</code></div></div>

<div class="steps"><div class="step"><strong>打开模型或供应商设置</strong><p>找到自定义 OpenAI 或 Base URL。</p></div><div class="step"><strong>填入上面四项</strong><p>只有字段写明要完整 Endpoint 时，才使用 <code>{{API_BASE_URL}}/chat/completions</code>。</p></div><div class="step"><strong>分别检查 Chat、Agent 和补全</strong><p>它们可能各有一个模型入口。补全不可用时，不代表对话也没有走网关。</p></div><div class="step"><strong>用使用日志确认</strong><p>短消息应出现在对应密钥下。</p></div></div>

不需要退出 Trae 账号。恢复内置服务时切回官方供应商并关掉 OpenBridger 这一项。

## 相关

- [Cursor](/guide/clients/cursor/)
- [故障排查](/troubleshooting/)
