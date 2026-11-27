# Python & Node.js SDK Quick Reference

Use the official `openai` SDK by pointing `base_url` / `baseURL` to `{{API_BASE_URL}}`.

<div class="config-card">
  <div><span>Base URL</span><code>{{API_BASE_URL}}</code> (with <code>/v1</code>, without <code>/chat/completions</code>)</div>
  <div><span>API Key</span><code>sk-...</code> via environment variable</div>
  <div><span>Model</span>Exact ID from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a></div>
</div>

## 1. Python (`pip install openai`)

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url="{{API_BASE_URL}}",
)

resp = client.chat.completions.create(
    model="gpt-6.2-sol",
    messages=[{"role": "user", "content": "Summarize quicksort in one sentence"}],
)
print(resp.choices[0].message.content)
```

## 2. Node.js / TypeScript (`npm install openai`)

```ts
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.OPENBRIDGER_API_KEY,
  baseURL: "{{API_BASE_URL}}",
});

const resp = await client.chat.completions.create({
  model: "gpt-6.2-sol",
  messages: [{ role: "user", content: "Summarize quicksort in one sentence" }],
});
console.log(resp.choices[0].message.content);
```

## 3. Raw `curl`

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

<div class="callout warning"><strong>Keep keys on your server</strong><p>Never ship <code>sk-...</code> keys in browser JavaScript — proxy requests through your own backend.</p></div>

## Common Issues

| Symptom | Fix |
| --- | --- |
| Requests still hit `api.openai.com` | Python uses `base_url`; Node.js uses `baseURL` |
| `404` (`/v1/v1`) | Pass `{{API_BASE_URL}}` — the SDK appends `/chat/completions` automatically |

## Related

- [Chat Completions & Vision](/api/openai-chat/)
- [Responses API](/api/openai-responses/)
- [LangChain](/guide/clients/sdk/)
