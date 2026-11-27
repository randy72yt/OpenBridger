# Google Gemini Native API

In addition to calling Gemini models via OpenAI-compatible `POST /v1/chat/completions`, you can call Google's native `/v1beta` endpoints directly for **Gemini CLI**, the `@google/genai` / `google-genai` SDKs, `thinkingConfig`, and native multimodal inputs.

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{CONSOLE_URL}}/v1beta/models/{model}:generateContent</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{CONSOLE_URL}}/v1beta/models/{model}:streamGenerateContent?alt=sse</code></div>
<div class="endpoint-bar"><span class="method-badge method-get">GET</span><code>{{CONSOLE_URL}}/v1beta/models</code></div>

## Authentication

Pass `x-goog-api-key: sk-...` (or `Authorization: Bearer sk-...`) in request headers rather than placing your key in URL query parameters.

## 1. Native content generation

```bash
curl "{{CONSOLE_URL}}/v1beta/models/gemini-4-neon:generateContent" \
  -H "x-goog-api-key: $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "Explain the roles of Q, K, and V matrices in self-attention."}]
      }
    ]
  }'
```

## 2. Streaming & thinking configuration

Use `:streamGenerateContent?alt=sse` for SSE streaming, and configure `generationConfig.thinkingConfig` when calling reasoning-capable Gemini models:

```bash
curl "{{CONSOLE_URL}}/v1beta/models/gemini-4-carbon:streamGenerateContent?alt=sse" \
  -H "x-goog-api-key: $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "Design a high-concurrency university course registration system."}]
      }
    ],
    "generationConfig": {
      "thinkingConfig": {
        "includeThoughts": true
      }
    }
  }'
```

## 3. Multimodal input (`inlineData`)

```bash
curl "{{CONSOLE_URL}}/v1beta/models/gemini-4-neon:generateContent" \
  -H "x-goog-api-key: $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [
          {
            "inlineData": {
              "mimeType": "image/jpeg",
              "data": "<IMAGE_BASE64_DATA>"
            }
          },
          {"text": "Extract the experimental table from this image into Markdown."}
        ]
      }
    ]
  }'
```

## 4. Google GenAI Python SDK

```python
import os
from google import genai
from google.genai import types

client = genai.Client(
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    http_options=types.HttpOptions(base_url="{{CONSOLE_URL}}"),
)

response = client.models.generate_content(
    model="gemini-4-neon",
    contents="List three parallelization advantages of Transformers over RNNs.",
)
print(response.text)
```

## Related

- [Gemini CLI](/guide/clients/gemini-cli/)
- [Chat Completions & Vision](/api/openai-chat/)
- [Error codes](/guide/errors/)
