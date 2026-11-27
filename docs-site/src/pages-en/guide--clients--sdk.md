# LangChain Integration (Python & TypeScript)

Point `langchain-openai` or `@langchain/openai` at `{{API_BASE_URL}}` to use any supported chat or tool-calling model.

## 1. Python (`pip install langchain-openai`)

```python
import os
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="gpt-6.2-sol",
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url="{{API_BASE_URL}}",
)

print(llm.invoke("Summarize a RAG pipeline in 3 bullet points").content)
```

## 2. JavaScript / TypeScript (`npm install @langchain/openai`)

```ts
import { ChatOpenAI } from "@langchain/openai";

const llm = new ChatOpenAI({
  model: "gpt-6.2-sol",
  apiKey: process.env.OPENBRIDGER_API_KEY,
  configuration: {
    baseURL: "{{API_BASE_URL}}",
  },
});

const res = await llm.invoke("Summarize a RAG pipeline in 3 bullet points");
console.log(res.content);
```

## Production Tips

- **Separate keys per environment**: Use distinct keys for dev, staging, and prod.
- **Retry only transient errors**: Retry `429` and transient `5xx` with exponential backoff; fail fast on `401` and `403`.

## Related

- [Python & Node SDKs](/guide/clients/openai-compatible/)
- [Embeddings & Rerank](/api/embeddings-rerank/)
- [Error codes](/guide/errors/)
