# OpenAI-compatible calls

A client that can set a Base URL, an API key, and a model ID uses the three values below. Images, audio, and tool calls depend on that model and the current route, not on the Base URL alone.

<div class="config-card"><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API key</span><code>sk-...</code> in a Bearer header</div><div><span>Chat path</span><code>POST /v1/chat/completions</code></div></div>

Keep the key in an environment variable. The samples use `YOUR_API_KEY` as a placeholder. Replace it, or read the environment variable, before you run them.

## curl

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "MODEL_ID_FROM_CONSOLE",
    "messages": [{"role": "user", "content": "Hello OpenBridger"}]
  }'
```

Add `"stream": true` for a stream. Whether usage is included in the stream depends on the model and the route. Check the response and [Usage Logs]({{CONSOLE_URL}}/usage-logs).

## Node.js

```js
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.OPENBRIDGER_API_KEY,
  baseURL: process.env.OPENBRIDGER_BASE_URL, // {{API_BASE_URL}}
});

const response = await client.chat.completions.create({
  model: process.env.OPENBRIDGER_MODEL,
  messages: [{ role: "user", content: "Hello OpenBridger" }],
});

console.log(response.choices[0].message.content);
```

```bash
export OPENBRIDGER_API_KEY="sk-..."
export OPENBRIDGER_BASE_URL="{{API_BASE_URL}}"
export OPENBRIDGER_MODEL="MODEL_ID_FROM_CONSOLE"
```

The official SDK appends `/chat/completions` to `baseURL`. `baseURL` must include `/v1`. Do not set it to `{{API_BASE_URL}}/chat/completions`.

## Python

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url=os.environ["OPENBRIDGER_BASE_URL"],
)

response = client.chat.completions.create(
    model=os.environ["OPENBRIDGER_MODEL"],
    messages=[{"role": "user", "content": "Hello OpenBridger"}],
)

print(response.choices[0].message.content)
```

The Python parameter is `base_url`, not `baseURL`.

## Other paths

Use the same Base URL and swap in a path the gateway implements:

```bash
curl "{{API_BASE_URL}}/embeddings" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"MODEL_ID_FROM_CONSOLE","input":"hello"}'
```

List models:

```bash
curl "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

The path list is on the [introduction](/). Unimplemented paths return `501` and `api_not_implemented`.

<div class="callout warning"><strong>Do not call from a public web page with the key</strong><p>A key in the browser is visible to visitors. A website should call your backend, and the backend should call OpenBridger.</p></div>

## Common problems

| What you see | What to do |
| --- | --- |
| The request still goes to `api.openai.com` | Check `baseURL` in Node and `base_url` in Python |
| 404 and the path contains `/v1` twice | The Base URL has an extra `/v1` |
| `tools` fails | Drop tools, send plain text, then check whether that model supports tools |
| 401 / 403 / 503 | Match [Error codes](/guide/errors/). Do not stop at the client's "network error" |

## Related

- [SDKs and LangChain](/guide/clients/sdk/)
- [Quick start](/guide/quickstart/)
