# 在 Chatbox 中使用 OpenBridger

在 Chatbox 的模型服务设置中选择 OpenAI API 或 OpenAI Compatible 自定义服务。

<div class="config-card"><div><span>API Host / Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>YOUR_API_KEY</code></div><div><span>Model</span><code>控制台中的完整模型 ID</code></div></div>

## 配置步骤

<div class="steps"><div class="step"><strong>打开模型设置</strong><p>在 Settings 中找到 Model Provider、AI Provider 或 API 设置。</p></div><div class="step"><strong>选择兼容服务</strong><p>选择 OpenAI API 或自定义 OpenAI Compatible 服务，不要选择网页账号登录。</p></div><div class="step"><strong>填写地址、密钥和模型</strong><p>使用上面的配置；自动模型列表为空时，从控制台手动添加完整模型 ID。</p></div><div class="step"><strong>新建对话验证</strong><p>选择 OpenBridger 模型发送短消息，并在使用日志中核对请求。</p></div></div>

无需退出 Chatbox 账号。出现 404 时检查是否把 `/chat/completions` 重复追加到 Base URL；旧对话仍使用原模型时，新建对话并重新选择 Provider。
