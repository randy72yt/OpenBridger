# SDKs and LangChain

Server code should use the official OpenAI SDK, or any library that accepts a custom Base URL. Put the key in an environment variable or a secret store. Do not commit it.

```bash
export OPENBRIDGER_API_KEY="sk-..."
export OPENBRIDGER_BASE_URL="{{API_BASE_URL}}"
export OPENBRIDGER_MODEL="MODEL_ID_FROM_CONSOLE"
```

Node and Python OpenAI SDK samples are on [OpenAI-compatible calls](/guide/clients/openai-compatible/).

## Before production

- Use a different key for development, test, and production.
- Set connect and read timeouts. Retry `429` and temporary upstream 5xx a limited number of times with backoff. Do not retry `401` or a `403` for insufficient credit.
- Log the request id, model, status, and latency. Drop the key and sensitive prompt text.
- Put your own limit on input length and `max_tokens`. The gateway rejects some illegal values. That is not your quota policy.

## LangChain (Python)

```python
import os
from langchain_openai import ChatOpenAI

model = ChatOpenAI(
    model=os.environ["OPENBRIDGER_MODEL"],
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url=os.environ["OPENBRIDGER_BASE_URL"],
)

print(model.invoke("Hello OpenBridger").content)
```

The package is `langchain-openai`. It uses the OpenAI SDK, so `base_url` still includes `/v1`.

## LangChain (JavaScript)

```js
import { ChatOpenAI } from "@langchain/openai";

const model = new ChatOpenAI({
  apiKey: process.env.OPENBRIDGER_API_KEY,
  configuration: {
    baseURL: process.env.OPENBRIDGER_BASE_URL,
  },
  model: process.env.OPENBRIDGER_MODEL,
});

const response = await model.invoke("Hello OpenBridger");
console.log(response.content);
```

This `baseURL` is also `{{API_BASE_URL}}`. If a package version moved the field to a top-level `baseURL`, follow that version. The finished request host should be `openbridger.com` (or `api.openbridger.com`) and the path should end in `/v1/chat/completions` with a single `/v1`.

## Common problems

| What you see | What to do |
| --- | --- |
| Requests still go to official OpenAI | The environment variable is not in the process, or the parameter name is wrong |
| Browser CORS | Move the call to your server |
| Tool calls fail | `invoke` plain text first, then check whether the model supports tools |
| LangChain wraps a credit error as a generic exception | Print `status` and the body, then use [Error codes](/guide/errors/) |

## Related

- [OpenAI-compatible calls](/guide/clients/openai-compatible/)
- [Create an API key](/guide/api-key/)
