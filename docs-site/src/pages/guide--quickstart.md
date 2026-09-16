# 5 分钟完成第一次调用

完成登录、创建 API Key 后，可以先在 AI 体验台验证模型，再接入工具或 SDK。

<div class="steps"><div class="step"><strong>登录控制台</strong><p>使用平台当前开放的登录方式进入控制台。</p></div><div class="step"><strong>选择模型</strong><p>在模型广场确认账户可用的模型及计费说明。</p></div><div class="step"><strong>创建 API Key</strong><p>在 API 密钥页面为当前工具或项目创建独立密钥，并妥善保存。</p></div><div class="step"><strong>发送测试请求</strong><p>把 Base URL、API Key 和完整模型 ID 配置到客户端，先发送一条短消息。</p></div></div>

```bash
curl {{API_BASE_URL}}/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MODEL_ID_FROM_CONSOLE",
    "messages": [{"role": "user", "content": "Hello OpenBridger"}]
  }'
```

请求成功后，到控制台的使用日志中核对时间、模型、API Key 和消耗，确认调用确实经过 OpenBridger。
