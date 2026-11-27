# Dify, n8n & Coze Workflows

Low-code AI app builders and automation engines — including **Dify**, **n8n**, and **Coze** — connect cleanly to OpenBridger for chat, vision, RAG embeddings, reranking, and async media tasks.

## 1. Dify (LLM, Text Embedding & Rerank)

1. In Dify, open **Settings -> Model Providers** and install the official **OpenAI-API-compatible** plugin.
2. Click **Add Model** and fill in:

<div class="config-card">
  <div><span>Model Type</span><code>LLM</code>, <code>Text Embedding</code>, or <code>Rerank</code></div>
  <div><span>Model Name</span>Exact model ID from <a href="{{CONSOLE_URL}}/pricing">Model Square</a> (e.g. <code>gpt-6.2-sol</code>)</div>
  <div><span>API Key</span><code>sk-your-api-key</code></div>
  <div><span>API Endpoint URL</span><code>{{API_BASE_URL}}</code> (includes <code>/v1</code>)</div>
</div>

Enable **Vision** and **Function Calling** checkboxes when adding multimodal or tool-capable models.

---

## 2. n8n automation nodes

1. Add an **OpenAI** node (or **OpenAI Chat Model** sub-node under AI Agent).
2. Create a new **OpenAI API** credential:
   - **API Key**: `sk-your-api-key`
   - **Base URL**: replace `https://api.openai.com/v1` with `{{API_BASE_URL}}`
3. Switch the **Model** field to **By ID / Expression** and enter the exact model ID from Model Square.

---

## 3. Coze & generic HTTP Request nodes

On workflow platforms where you use raw HTTP Request nodes:

### Synchronous chat node

- **Method**: `POST`
- **URL**: `{{API_BASE_URL}}/chat/completions`
- **Headers**:
  - `Authorization`: `Bearer sk-your-api-key` (store in workflow secrets)
  - `Content-Type`: `application/json`
- **JSON Body**: keep `"stream": false` so downstream nodes can read `choices[0].message.content` directly.

### Async video polling pattern (`/v1/videos`)

1. **Create Task (`POST {{API_BASE_URL}}/videos`)**: parse `id` from the JSON response.
2. **Loop Node**: cap iterations (e.g. 40).
3. **Wait 5s -> Poll (`GET {{API_BASE_URL}}/videos/{{id}}`)**: check `status`.
4. **Condition Break**: exit the loop once `status` reaches `completed` (or `failed`) and pass the output video URL downstream.

## Related

- [Chat Completions & Vision](/api/openai-chat/)
- [Image, Video & Audio APIs](/api/multimodal-media/)
- [Embeddings & Rerank](/api/embeddings-rerank/)
