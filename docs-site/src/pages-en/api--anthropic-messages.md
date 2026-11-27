# Anthropic Messages API

OpenBridger natively supports the Anthropic Messages wire protocol (`POST /v1/messages`) for **Claude Code**, the official Anthropic Python / TypeScript SDKs, and native Claude integrations — including streaming, Extended Thinking, prompt caching, vision, and PDF document inputs.

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{CONSOLE_URL}}/v1/messages</code></div>

<div class="callout warning"><strong>Do not append /v1 to the SDK base URL</strong><p>When configuring <code>base_url</code> in the Anthropic SDK or <code>ANTHROPIC_BASE_URL</code> in Claude Code, use the root host <code>{{CONSOLE_URL}}</code>. The client appends <code>/v1/messages</code> automatically. See <a href="/guide/base-urls/">Base URL reference</a>.</p></div>

## Headers

| Header | Required | Description |
| --- | --- | --- |
| `x-api-key` | Yes | Your `sk-...` key without `Bearer ` (`Authorization: Bearer` is also accepted) |
| `anthropic-version` | Yes | Recommended: `2023-06-01` |
| `content-type` | Yes | `application/json` |

## 1. Chat & streaming

Anthropic's protocol requires an explicit `max_tokens` field:

```bash
curl "{{CONSOLE_URL}}/v1/messages" \
  -H "x-api-key: $OPENBRIDGER_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-opus-5-5",
    "max_tokens": 2048,
    "stream": true,
    "system": "You are a principal database architect.",
    "messages": [
      {"role": "user", "content": "Compare B-Tree and GIN indexes in PostgreSQL."}
    ]
  }'
```

## 2. Extended Thinking

Pass a `thinking` configuration (`max_tokens` must exceed `budget_tokens`). The response `content` array returns a `thinking` block followed by the final `text` block:

```bash
curl "{{CONSOLE_URL}}/v1/messages" \
  -H "x-api-key: $OPENBRIDGER_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-opus-5-5",
    "max_tokens": 8192,
    "thinking": {
      "type": "enabled",
      "budget_tokens": 4096
    },
    "messages": [
      {"role": "user", "content": "Prove the divergence of the harmonic series using two distinct approaches."}
    ]
  }'
```

## 3. PDF & image inputs (Base64)

```bash
curl "{{CONSOLE_URL}}/v1/messages" \
  -H "x-api-key: $OPENBRIDGER_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-sonnet-5-5",
    "max_tokens": 2048,
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "document",
            "source": {
              "type": "base64",
              "media_type": "application/pdf",
              "data": "<PDF_BASE64_DATA>"
            }
          },
          {"type": "text", "text": "Summarize this paper'\''s key contributions, datasets, and limitations."}
        ]
      }
    ]
  }'
```

## 4. Official Anthropic Python SDK

```python
import os
import anthropic

client = anthropic.Anthropic(
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url="{{CONSOLE_URL}}",  # Root URL without /v1
)

msg = client.messages.create(
    model="claude-sonnet-5-5",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello from the Anthropic SDK!"}],
)

print(msg.content[0].text)
```

<div class="callout"><strong>Note</strong><p><code>/v1/messages/count_tokens</code> is not enabled; refer to the response <code>usage</code> object and Usage Logs for exact token counts.</p></div>

## Related

- [Claude Code](/guide/clients/claude-code/)
- [CC Switch](/guide/clients/cc-switch/)
- [Error codes](/guide/errors/)
