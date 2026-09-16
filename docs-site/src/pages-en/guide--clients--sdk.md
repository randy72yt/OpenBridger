# SDK integration

Server applications can use the official OpenAI SDK or a framework that supports a custom Base URL.

```bash
export OPENBRIDGER_API_KEY="YOUR_API_KEY"
export OPENBRIDGER_BASE_URL="{{API_BASE_URL}}"
```

- Keep keys in server environment variables or a secret manager.
- Use separate keys for development, testing, and production.
- Set timeouts and use limited retries for 429 and temporary upstream errors.
- Log request ID, model, latency, and status while hiding secrets and sensitive content.
- Bound user input, maximum output, and multimodal quantities.

## LangChain JavaScript

```js
import { ChatOpenAI } from "@langchain/openai";
const model = new ChatOpenAI({
  apiKey: process.env.OPENBRIDGER_API_KEY,
  configuration: { baseURL: process.env.OPENBRIDGER_BASE_URL },
  model: "MODEL_ID_FROM_CONSOLE",
});
```

| Symptom | Check |
| --- | --- |
| SDK calls official OpenAI | Use `baseURL` in JS or `base_url` in Python |
| Browser reports CORS | Move the call to your server; do not expose the key |
| Tool-call arguments fail | Test plain text, then verify tools support for the model and route |
| 429 or temporary failures | Use limited retries with backoff |
