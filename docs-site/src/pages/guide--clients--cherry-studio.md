# 在 Cherry Studio 中使用 OpenBridger

Cherry Studio 可以添加 OpenAI 类型的自定义服务商。按你安装的版本，入口一般在设置里的模型服务或供应商列表。找不到「自定义 / OpenAI」时，以该版本的设置页为准，不要改系统网络配置来凑这个地址。

## 配置

<div class="config-card"><div><span>服务商名称</span><code>OpenBridger</code></div><div><span>类型</span><code>OpenAI</code></div><div><span>API 地址</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>单独创建的 sk- 密钥</code></div><div><span>模型 ID</span><code>模型广场中的完整名称</code></div></div>

<div class="steps"><div class="step"><strong>新增服务商</strong><p>类型选 OpenAI，名称填 OpenBridger。</p></div><div class="step"><strong>填写地址和密钥</strong><p>API 地址不要再加 <code>/v1</code> 或 <code>/chat/completions</code>。<code>{{API_BASE_URL}}</code> 已经包含 <code>/v1</code>。</p></div><div class="step"><strong>添加模型</strong><p>自动列表为空时，从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 手工添加完整 ID。</p></div><div class="step"><strong>发短消息并核对日志</strong><p>选择该服务商后发送一句短消息，在 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 中按密钥确认。</p></div></div>

不需要退出 Cherry Studio 里的其他服务商账号。每条对话用你当前选中的服务商。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 检查密钥失败 | 对照 curl <code>{{API_BASE_URL}}/models</code>。curl 成功则是地址多写了一层路径 |
| 模型列表为空 | 手工添加模型 ID |
| 503 或 model not found | 核对大小写和版本后缀 |
| 没有流式输出 | 看客户端的流式开关；用非流式再试一次同一模型 |

## 相关

- [Chatbox](/guide/clients/chatbox/)
- [LobeChat](/guide/clients/lobechat/)
- [错误码](/guide/errors/)
