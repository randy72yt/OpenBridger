# Responses API

`POST /v1/responses` is OpenAI's agentic and reasoning wire protocol, required by **OpenAI Codex CLI** (`wire_api = "responses"`). The gateway also supports context compaction via `POST /v1/responses/compact`.

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/responses</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/responses/compact</code></div>

## Responses API vs. Chat Completions

| Aspect | `POST /v1/responses` | `POST /v1/chat/completions` |
| --- | --- | --- |
| Primary use | Codex CLI, multi-step coding agents, deep reasoning | General chat apps, translation plugins, cross-vendor SDKs |
| Input field | `input` (string or item array) + `instructions` | `messages` array |
| Reasoning control | `"reasoning": {"effort": "low" \| "medium" \| "high"}` | `"reasoning_effort"` |
| Output shape | `output` item array (`reasoning` + `message` items) | `choices[0].message` |

<div class="callout"><strong>Check endpoint support</strong><p>Confirm in <a href="{{CONSOLE_URL}}/pricing">Model Square</a> that your target model (e.g. <code>gpt-6.2-astra</code>, <code>gpt-6.2-sol</code>) and key group expose the Responses endpoint; otherwise the gateway returns <code>503</code>.</p></div>

## 1. Request with reasoning effort

```bash
curl "{{API_BASE_URL}}/responses" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-astra",
    "instructions": "Write production-ready Go code and handle edge cases.",
    "reasoning": {
      "effort": "high"
    },
    "input": "Implement a concurrency-safe token bucket rate limiter."
  }'
```

## 2. Streaming events (SSE)

Setting `"stream": true` emits typed SSE events including `response.created`, `response.output_text.delta`, and `response.completed`:

```bash
curl "{{API_BASE_URL}}/responses" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "stream": true,
    "input": "Summarize the TCP 3-way handshake state transitions."
  }'
```

## 3. Official OpenAI Python SDK

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url="{{API_BASE_URL}}",
)

resp = client.responses.create(
    model="gpt-6.2-astra",
    reasoning={"effort": "medium"},
    input="Check why this SQL query misses an index: SELECT * FROM orders WHERE YEAR(created_at) = 2026;",
)

print(resp.output_text)
```

## Related

- [OpenAI Codex](/guide/clients/codex-cli/)
- [Chat Completions & Vision](/api/openai-chat/)
- [Reasoning & Thinking Fields](/api/thinking-fields/)
