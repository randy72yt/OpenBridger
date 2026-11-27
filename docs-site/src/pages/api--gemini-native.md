# Google Gemini 原生接口

你既可以通过 OpenAI 兼容接口（`POST /v1/chat/completions`）调用 Gemini 模型，也可以直接调用 Google 原生 `/v1beta` 端点，使用 Gemini CLI、`@google/genai` SDK、原生多模态理解、思考配置（`thinkingConfig`）及原生生图能力。

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{CONSOLE_URL}}/v1beta/models/{model}:generateContent</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{CONSOLE_URL}}/v1beta/models/{model}:streamGenerateContent?alt=sse</code></div>
<div class="endpoint-bar"><span class="method-badge method-get">GET</span><code>{{CONSOLE_URL}}/v1beta/models</code></div>

## 鉴权方式

推荐在请求头传入 `x-goog-api-key: sk-...`（或 `Authorization: Bearer sk-...`）。避免把密钥拼在 URL 查询参数 `?key=` 里以免被代理或客户端日志记录。

## 1. 原生文本生成

```bash
curl "{{CONSOLE_URL}}/v1beta/models/gemini-4-neon:generateContent" \
  -H "x-goog-api-key: $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "用通俗语言解释注意力机制中的 Q、K、V 矩阵作用"}]
      }
    ]
  }'
```

## 2. 流式输出与思考模式（thinkingConfig）

流式端点使用 `:streamGenerateContent?alt=sse`，并可在 `generationConfig.thinkingConfig` 中控制思考输出：

```bash
curl "{{CONSOLE_URL}}/v1beta/models/gemini-4-carbon:streamGenerateContent?alt=sse" \
  -H "x-goog-api-key: $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "设计一套支持百万在线连接的校园选课秒杀系统架构"}]
      }
    ],
    "generationConfig": {
      "thinkingConfig": {
        "includeThoughts": true
      }
    }
  }'
```

## 3. 多模态输入（图片 / PDF / 音频 inlineData）

通过 `inlineData` 传入 Base64 编码的图片、PDF 或音频片段进行跨模态理解：

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
          {"text": "提取图中实验数据表格并输出 Markdown 表格"}
        ]
      }
    ]
  }'
```

## 4. Google GenAI SDK（Python）接入

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
    contents="用三点总结 Transformer 相比 RNN 的并行化优势",
)
print(response.text)
```

## 相关

- [Gemini CLI 配置教程](/guide/clients/gemini-cli/)
- [Chat Completions 对话与识图](/api/openai-chat/)
- [错误码](/guide/errors/)
