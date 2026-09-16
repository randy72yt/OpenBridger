# 在 CodeBuddy 中使用 OpenBridger

当 CodeBuddy 提供 Custom Model、OpenAI Compatible 或自定义 API 地址时，可以添加 OpenBridger。

<div class="steps"><div class="step"><strong>打开模型设置</strong><p>查找 Model、Provider、API Key 或 Custom Model。</p></div><div class="step"><strong>新增服务商</strong><p>选择 OpenAI Compatible；没有该选项时查看是否允许覆盖 OpenAI Base URL。</p></div><div class="step"><strong>配置并测试</strong><p>填写 <code>{{API_BASE_URL}}</code>、独立 API Key 和完整模型 ID，发送短消息验证。</p></div></div>

不同版本的菜单名称可能不同。无法配置自定义 Base URL 的版本不能通过本方式接入。

<div class="config-card"><div><span>Provider Type</span><code>OpenAI Compatible</code></div><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>YOUR_API_KEY</code></div><div><span>Model</span><code>控制台中的完整模型 ID</code></div></div>

无需退出 CodeBuddy 或编辑器账号。Chat、Agent 和代码补全可能采用不同设置，应分别检查供应商。恢复内置服务时切回官方模型、停用自定义 Provider 并重启客户端。
