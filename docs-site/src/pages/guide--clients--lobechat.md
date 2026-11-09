# 在 LobeChat 中使用 OpenBridger

LobeChat 的 OpenAI 供应商可以改 API 地址。用这组值即可，不需要别的插件。

<div class="config-card"><div><span>供应商</span><code>OpenAI</code></div><div><span>API Key</span><code>sk-...</code></div><div><span>接口地址 / 代理地址</span><code>{{API_BASE_URL}}</code></div><div><span>模型 ID</span><code>模型广场中的完整 ID</code></div></div>

不同版本里，这一栏可能叫 API Proxy、Endpoint 或「接口代理」。填 `{{API_BASE_URL}}`。如果输入框的占位符已经以 `/v1` 结尾，就不要再加一段 `/v1`。如果占位符是不带路径的主机名，并且客户端会自动补 `/v1`，则只填 `{{CONSOLE_URL}}`。判断标准：浏览器或客户端日志里的最终路径应是 `/v1/chat/completions`，且只有一个 `/v1`。

## 步骤

<div class="steps"><div class="step"><strong>打开语言模型设置</strong><p>进入 OpenAI 供应商，而不是需要官方账号登录的其他供应商。</p></div><div class="step"><strong>填写密钥和地址</strong><p>使用只给 LobeChat 的密钥。</p></div><div class="step"><strong>添加模型</strong><p>列表拉不下来时，用自定义模型填入广场上的完整 ID。</p></div><div class="step"><strong>新开会话验证</strong><p>发送短消息，在 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 确认。</p></div></div>

LobeChat 的界面会随版本改名。上面找不到对应字段时，以该版本里「OpenAI + 自定义地址」的实际表单为准，三个值不变。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 401 | 密钥或地址错误。先用 curl 打 `/v1/models` |
| 模型列表为空 | 手工添加模型 ID |
| 请求打到 `api.openai.com` | 自定义地址没有保存，或会话仍选着官方模型 |

## 相关

- [Chatbox](/guide/clients/chatbox/)
- [故障排查](/troubleshooting/)
