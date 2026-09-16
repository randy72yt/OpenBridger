# OpenAI Compatible 接入

支持自定义 Base URL、API Key 和模型 ID 的工具通常可以连接 OpenBridger。特定能力仍取决于客户端协议、所选模型和线路。

<div class="config-card"><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>YOUR_API_KEY</code></div><div><span>Chat Endpoint</span><code>/chat/completions</code></div></div>

## Node.js 示例

```js
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.OPENBRIDGER_API_KEY,
  baseURL: "{{API_BASE_URL}}",
});

const response = await client.chat.completions.create({
  model: "MODEL_ID_FROM_CONSOLE",
  messages: [{ role: "user", content: "Hello OpenBridger" }],
});
```

## Python 示例

```python
from openai import OpenAI

client = OpenAI(
    api_key="YOUR_API_KEY",
    base_url="{{API_BASE_URL}}",
)

response = client.chat.completions.create(
    model="MODEL_ID_FROM_CONSOLE",
    messages=[{"role": "user", "content": "Hello OpenBridger"}],
)

print(response.choices[0].message.content)
```

<div class="callout warning"><strong>不要从公开网页直接调用</strong><p>浏览器前端携带 API Key 会把凭证暴露给访问者。网页应用应通过自己的服务端调用。</p></div>
