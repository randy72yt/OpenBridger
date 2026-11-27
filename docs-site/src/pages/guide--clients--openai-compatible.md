# Python / Node.js 官方 SDK 接入

直接复用官方 `openai` SDK，只需把 `base_url` / `baseURL` 指向 `{{API_BASE_URL}}`，即可无缝调用 GPT、Claude、Gemini、DeepSeek 等在售模型。

<div class="config-card">
  <div><span>Base URL</span><code>{{API_BASE_URL}}</code>（必须带 <code>/v1</code>，不要加 <code>/chat/completions</code>）</div>
  <div><span>API Key</span>环境变量传入 <code>sk-...</code></div>
  <div><span>Model</span>从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制完整 ID</div>
</div>

## 1. Python（`pip install openai`）

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url="{{API_BASE_URL}}",
)

resp = client.chat.completions.create(
    model="gpt-6.2-sol",
    messages=[{"role": "user", "content": "用一句话介绍快速排序"}],
)
print(resp.choices[0].message.content)
```

## 2. Node.js / TypeScript（`npm install openai`）

```ts
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.OPENBRIDGER_API_KEY,
  baseURL: "{{API_BASE_URL}}",
});

const resp = await client.chat.completions.create({
  model: "gpt-6.2-sol",
  messages: [{ role: "user", content: "用一句话介绍快速排序" }],
});
console.log(resp.choices[0].message.content);
```

## 3. curl 原生请求

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

<div class="callout warning"><strong>安全提示</strong><p>网页或小程序前端切勿直接放置 <code>sk-...</code> 密钥，请由你自己的后端服务中转调用。</p></div>

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 请求仍打到 `api.openai.com` | Python 参数名是小写 `base_url`，Node.js 参数名是驼峰 `baseURL` |
| 报 `404`（出现 `/v1/v1`） | `base_url` 只写到 `{{API_BASE_URL}}`，SDK 会自动拼接 `/chat/completions` |

## 相关

- [Chat Completions 对话与识图](/api/openai-chat/)
- [Responses 接口](/api/openai-responses/)
- [LangChain 接入](/guide/clients/sdk/)
