# 在 Chatbox 中使用 OpenBridger

在 Chatbox 的模型提供方里选择 OpenAI API 或 OpenAI Compatible，不要选择「用网页账号登录」。

<div class="config-card"><div><span>API Host / Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>sk-...</code></div><div><span>Model</span><code>模型广场中的完整 ID</code></div></div>

## 步骤

<div class="steps"><div class="step"><strong>打开设置里的模型提供方</strong><p>名称可能是 Model Provider、AI Provider 或 API。</p></div><div class="step"><strong>选择 OpenAI 兼容</strong><p>填写上面三项。Host 不要包含 <code>/chat/completions</code>，除非这一栏的说明写明要完整 Endpoint。</p></div><div class="step"><strong>模型列表为空就手工填</strong><p>从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制。</p></div><div class="step"><strong>新建对话</strong><p>旧对话可能仍绑着上一个提供方。新建一条再选 OpenBridger，然后到使用日志核对。</p></div></div>

不需要退出 Chatbox 账号。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 404，路径里出现两次 chat/completions 或 /v1 | Host 多写了路径。改回 `{{API_BASE_URL}}` |
| 401 | 重新粘贴密钥 |
| 日志没有记录 | 当前对话还在用内置模型。新建对话并重新选择提供方 |

## 相关

- [Cherry Studio](/guide/clients/cherry-studio/)
- [OpenAI 兼容调用](/guide/clients/openai-compatible/)
