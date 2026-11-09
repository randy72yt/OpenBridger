# 在 Cursor 中使用 OpenBridger

Cursor 在提供「自定义 OpenAI Base URL」或 OpenAI Compatible 供应商时，可以让 Chat 或 Agent 走 OpenBridger。菜单名字会随版本变化。找不到 Base URL 一项时，该版本不能按本页接入。

## 开始之前

- 一把只给 Cursor 用的 API Key，见 [创建 API Key](/guide/api-key/)。
- 从 [模型广场]({{CONSOLE_URL}}/pricing) 复制的完整模型 ID。
- 确认你的 Cursor 版本允许使用自带 API Key。这个范围由 Cursor 的版本和套餐决定，OpenBridger 不能代替这项许可。

## 配置

<div class="steps"><div class="step"><strong>打开 Cursor Settings</strong><p>找到 Models、API Keys 或自定义 Provider。</p></div><div class="step"><strong>填写 OpenAI 兼容项</strong><p>Override OpenAI Base URL 或同类开关打开后，Base URL 填 <code>{{API_BASE_URL}}</code>，Key 填 <code>sk-</code> 开头的密钥。</p></div><div class="step"><strong>添加模型 ID</strong><p>使用广场上的完整 ID。不要只写厂商家族名。</p></div><div class="step"><strong>发一句短消息</strong><p>在 Chat 里发送，然后到 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 按这把密钥核对。</p></div></div>

<div class="config-card"><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>sk-...</code></div><div><span>Model</span><code>模型广场中的完整 ID</code></div></div>

## 和 Cursor 自己的功能分开

- 不需要退出 Cursor 账号。账号同步和自定义模型是两件事。
- Tab Completion 等由 Cursor 提供的补全，可能仍然走 Cursor 自己的服务，不会出现在 OpenBridger 日志里。
- Chat、Agent、补全要分别看各自的模型设置。
- 全局 Override 打开时，官方模型也会被带到这个 Base URL。切回官方模型前先关掉它，再重启 Cursor。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 没有 Verify 按钮 | 保存后直接发短消息，用使用日志确认 |
| model not found 或 503 无可用渠道 | 重新复制模型 ID，核对密钥分组 |
| 官方模型也失败 | 关闭 Base URL 覆盖并重启 |
| 日志里没有请求 | 当前功能没有走自定义 Base URL，或仍在使用 Cursor 自有模型 |
| 401 | 密钥多了空格或写了两次 `Bearer` |

## 相关

- [OpenAI 兼容调用](/guide/clients/openai-compatible/)
- [故障排查](/troubleshooting/)
