# Chat Completions 对话与识图

`POST /v1/chat/completions` 是最通用的对话入口。只要模型在 [模型广场]({{CONSOLE_URL}}/pricing) 里支持 OpenAI 对话协议，无论是 GPT、Claude、Gemini、DeepSeek、Qwen 还是 Grok，都可以用同一套请求体调用。

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/chat/completions</code></div>
<div class="endpoint-bar"><span class="method-badge method-get">GET</span><code>{{API_BASE_URL}}/models</code></div>

## 请求头

| Header | 必填 | 说明 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer sk-...`（在 [API 密钥]({{CONSOLE_URL}}/keys) 创建） |
| `Content-Type` | 是 | `application/json` |

## 1. 基础对话（非流式）

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "messages": [
      {"role": "system", "content": "你是一名严谨的技术助手。"},
      {"role": "user", "content": "用三句话解释什么是 SSE 流式传输。"}
    ],
    "temperature": 0.7
  }'
```

响应结构与 OpenAI 官方一致，回复文本在 `choices[0].message.content`，Token 用量在 `usage` 对象中，响应头包含 `X-Oneapi-Request-Id` 便于在 [使用日志]({{CONSOLE_URL}}/usage-logs) 核对。

## 2. 流式对话（SSE）

生产环境与长文本生成建议开启 `"stream": true`，降低首字等待时间（TTFT）。如需在最后一个流式分片拿到 Token 统计，一并传入 `stream_options.include_usage`。

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "stream": true,
    "stream_options": {"include_usage": true},
    "messages": [
      {"role": "user", "content": "写一段 Python 快速排序并附简短注释"}
    ]
  }'
```

## 3. 多模态识图（图片 URL 与 Base64）

选择具备视觉能力的模型（如 `gpt-6.2-astra`、`claude-opus-5-5`、`gemini-4-neon`），将 `content` 改为数组即可传入公网图片链接或 `data:image/...;base64,...`：

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-astra",
    "messages": [
      {
        "role": "user",
        "content": [
          {"type": "text", "text": "请提取这张图里的公式并转成 LaTeX"},
          {
            "type": "image_url",
            "image_url": {"url": "https://example.com/formula.png"}
          }
        ]
      }
    ]
  }'
```

本地图片用 Base64 时，把 `url` 写成 `"data:image/png;base64,iVBORw0KGgo..."` 即可，无需先上传到图床。

## 4. 工具调用（Function Calling）与结构化输出

支持工具调用的模型可传入 `tools` 数组，或通过 `response_format` 强制输出符合 JSON Schema 的结构化数据：

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "messages": [{"role": "user", "content": "查一下杭州今天天气"}],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "查询指定城市天气",
          "parameters": {
            "type": "object",
            "properties": {
              "city": {"type": "string", "description": "城市名称"}
            },
            "required": ["city"]
          }
        }
      }
    ]
  }'
```

## 5. 查询当前密钥可用模型

```bash
curl "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY"
```

返回结果仅包含当前 API Key 所属分组可路由的模型列表。

## 常见错误

| 状态码 | 原因与处理 |
| --- | --- |
| `400` | `max_tokens` 超出上限、JSON 语法错误，或向纯文本模型传了图片内容 |
| `401` / `403` | 密钥无效或账户额度不足，对照 [错误码](/guide/errors/) |
| `503` | 当前密钥分组下无此模型可用渠道，去 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID |

## 相关

- [Responses 接口（Codex / 推理）](/api/openai-responses/)
- [思考推理参数与返回字段](/api/thinking-fields/)
- [Python / Node SDK 接入](/guide/clients/openai-compatible/)
