# LangChain 接入（Python & TypeScript）

通过 `langchain-openai` / `@langchain/openai` 将 `base_url` 指向 `{{API_BASE_URL}}` 即可接入全部对话与工具调用模型。

## 1. Python（`pip install langchain-openai`）

```python
import os
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="gpt-6.2-sol",
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url="{{API_BASE_URL}}",
)

print(llm.invoke("用三点概括 RAG 的检索流程").content)
```

## 2. JavaScript / TypeScript（`npm install @langchain/openai`）

```ts
import { ChatOpenAI } from "@langchain/openai";

const llm = new ChatOpenAI({
  model: "gpt-6.2-sol",
  apiKey: process.env.OPENBRIDGER_API_KEY,
  configuration: {
    baseURL: "{{API_BASE_URL}}",
  },
});

const res = await llm.invoke("用三点概括 RAG 的检索流程");
console.log(res.content);
```

## 生产环境建议

- **按环境隔离密钥**：本地开发与线上服务使用不同 API Key，便于按密钥查看日志与限额。
- **合理设置重试**：仅对 `429` 和临时 `5xx` 开启有限次指数退避；遇到 `401`（密钥无效）或 `403`（余额不足）应立即告警停止重试。

## 相关

- [Python / Node SDK 接入](/guide/clients/openai-compatible/)
- [Embeddings 向量与 Rerank](/api/embeddings-rerank/)
- [错误码](/guide/errors/)
