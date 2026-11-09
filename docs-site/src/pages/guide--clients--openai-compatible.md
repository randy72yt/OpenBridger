# OpenAI 兼容调用

能填写 Base URL、API Key 和模型 ID 的客户端，按下面三个值接入。图像、语音、工具调用是否可用，取决于模型广场里的该模型和当前线路，不是 Base URL 本身。

<div class="config-card"><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>sk-...</code>，请求头用 Bearer</div><div><span>对话路径</span><code>POST /v1/chat/completions</code></div></div>

密钥放在环境变量里。下面示例用 `YOUR_API_KEY` 占位，运行前替换，或改成读取环境变量。

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

流式输出加上 `"stream": true`。是否在流里返回用量，看模型和线路，以响应和 [使用日志]({{CONSOLE_URL}}/usage-logs) 为准。

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

官方 SDK 会在 `baseURL` 后面拼接 `/chat/completions`。`baseURL` 必须含 `/v1`，不要再写成 `{{API_BASE_URL}}/chat/completions`。

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

Python SDK 的参数名是 `base_url`，不是 `baseURL`。

## 其他常用路径

在同样的 Base URL 下，把路径换成网关已实现的接口即可，例如：

```bash
curl "{{API_BASE_URL}}/embeddings" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"MODEL_ID_FROM_CONSOLE","input":"hello"}'
```

模型列表：

```bash
curl "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer YOUR_API_KEY"
```

路径清单见 [产品介绍](/)。未实现的路径返回 `501` `api_not_implemented`。

<div class="callout warning"><strong>不要从公开网页直接带密钥</strong><p>浏览器里的密钥对访问者可见。网页应用请调用你自己的后端，由后端再请求 OpenBridger。</p></div>

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 请求仍发往 `api.openai.com` | Node 检查 `baseURL`，Python 检查 `base_url` |
| 404 且路径里有两个 `/v1` | Base URL 多写了一层 `/v1` |
| `tools` 报错 | 先去掉工具，用纯文本打通，再确认该模型是否支持工具调用 |
| 401 / 403 / 503 | 按 [错误码](/guide/errors/) 对照，不要只看客户端的「网络错误」 |

## 相关

- [SDK 与 LangChain](/guide/clients/sdk/)
- [快速开始](/guide/quickstart/)
