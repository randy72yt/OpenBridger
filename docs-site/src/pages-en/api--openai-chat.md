# Chat Completions & Vision

`POST /v1/chat/completions` is the standard OpenAI-compatible chat endpoint. Any model in [Model Square]({{CONSOLE_URL}}/pricing) that exposes the OpenAI chat protocol — across GPT, Claude, Gemini, DeepSeek, Qwen, and Grok — can be called with the same request format.

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/chat/completions</code></div>
<div class="endpoint-bar"><span class="method-badge method-get">GET</span><code>{{API_BASE_URL}}/models</code></div>

## Headers

| Header | Required | Description |
| --- | --- | --- |
| `Authorization` | Yes | `Bearer sk-...` (created under [API Keys]({{CONSOLE_URL}}/keys)) |
| `Content-Type` | Yes | `application/json` |

## 1. Standard chat completion

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "messages": [
      {"role": "system", "content": "You are a concise engineering assistant."},
      {"role": "user", "content": "Explain SSE streaming in three sentences."}
    ],
    "temperature": 0.7
  }'
```

The assistant reply is returned in `choices[0].message.content`, token accounting in `usage`, and the gateway trace ID in the `X-Oneapi-Request-Id` response header for lookup in [Usage Logs]({{CONSOLE_URL}}/usage-logs).

## 2. Streaming responses (SSE)

Pass `"stream": true` for real-time token streaming. Include `stream_options.include_usage` if your client reads final token usage from the last SSE chunk.

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "stream": true,
    "stream_options": {"include_usage": true},
    "messages": [
      {"role": "user", "content": "Write a clean Python quicksort with brief comments."}
    ]
  }'
```

## 3. Vision (image URL and Base64)

With vision-capable models (such as `gpt-6.2-astra`, `claude-opus-5-5`, or `gemini-4-neon`), pass an array of content parts containing either an HTTPS image URL or a `data:image/...;base64,...` data URI:

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
          {"type": "text", "text": "Transcribe the equation in this image to LaTeX."},
          {
            "type": "image_url",
            "image_url": {"url": "https://example.com/formula.png"}
          }
        ]
      }
    ]
  }'
```

## 4. Tool calling (Function Calling)

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "messages": [{"role": "user", "content": "What is the weather in Tokyo today?"}],
    "tools": [
      {
        "type": "function",
        "function": {
          "name": "get_weather",
          "description": "Get current weather for a city",
          "parameters": {
            "type": "object",
            "properties": {
              "city": {"type": "string", "description": "City name"}
            },
            "required": ["city"]
          }
        }
      }
    ]
  }'
```

## 5. List models visible to your key

```bash
curl "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY"
```

## Common errors

| Status | Cause and fix |
| --- | --- |
| `400` | Invalid JSON, `max_tokens` out of bounds, or image input sent to a text-only model |
| `401` / `403` | Invalid key or insufficient balance — see [Error codes](/guide/errors/) |
| `503` | No route for that model in your key's group — copy the full model ID from [Model Square]({{CONSOLE_URL}}/pricing) |

## Related

- [Responses API](/api/openai-responses/)
- [Reasoning & Thinking Fields](/api/thinking-fields/)
- [Python & Node SDKs](/guide/clients/openai-compatible/)
