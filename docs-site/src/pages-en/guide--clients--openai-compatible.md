# OpenAI-compatible clients

Tools that allow a custom Base URL, API key, and model ID can usually connect to OpenBridger.

<div class="config-card"><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API key</span><code>YOUR_API_KEY</code></div><div><span>Chat endpoint</span><code>/chat/completions</code></div></div>

```js
import OpenAI from "openai";
const client = new OpenAI({ apiKey: process.env.OPENBRIDGER_API_KEY, baseURL: "{{API_BASE_URL}}" });
const response = await client.chat.completions.create({
  model: "MODEL_ID_FROM_CONSOLE",
  messages: [{ role: "user", content: "Hello OpenBridger" }],
});
```

```python
from openai import OpenAI
client = OpenAI(api_key="YOUR_API_KEY", base_url="{{API_BASE_URL}}")
response = client.chat.completions.create(model="MODEL_ID_FROM_CONSOLE", messages=[{"role": "user", "content": "Hello OpenBridger"}])
print(response.choices[0].message.content)
```

<div class="callout warning"><strong>Do not call from a public web page</strong><p>A browser bundle would expose the key. Call OpenBridger from your own server.</p></div>
