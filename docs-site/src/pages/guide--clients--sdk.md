# SDK 与 LangChain

服务端使用官方 OpenAI SDK，或任何接受自定义 Base URL 的库。密钥放在环境变量或密钥管理服务里，不要写进仓库。

```bash
export OPENBRIDGER_API_KEY="sk-..."
export OPENBRIDGER_BASE_URL="{{API_BASE_URL}}"
export OPENBRIDGER_MODEL="MODEL_ID_FROM_CONSOLE"
```

Node 与 Python 的 OpenAI SDK 示例在 [OpenAI 兼容调用](/guide/clients/openai-compatible/)。

## 上线前

- 开发、测试、生产各用一把密钥。
- 设置连接超时和读取超时。对 `429` 和上游的临时 5xx 做有限次退避，不要无限重试 `401` 和额度不足的 `403`。
- 日志记录请求 ID、模型、状态和耗时。丢掉密钥和提示词里的敏感内容。
- 对用户输入长度和 `max_tokens` 设业务上限。网关会拒绝明显不合法的请求，但这不是你的配额策略。

## LangChain（Python）

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

依赖是 `langchain-openai`，它使用 OpenAI SDK，因此 `base_url` 同样要包含 `/v1`。

## LangChain（JavaScript）

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

这里的 `baseURL` 也是 `{{API_BASE_URL}}`。如果某个版本的包把参数改成了顶层 `baseURL` 而不是 `configuration.baseURL`，以该版本文档为准，只要最终请求主机是 `openbridger.com`（或 `api.openbridger.com`）且路径以 `/v1/chat/completions` 结尾即可。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 仍然请求官方 OpenAI | 环境变量没进进程，或传错了参数名 |
| 浏览器 CORS | 把调用移到服务端 |
| 工具调用失败 | 先 `invoke` 一句纯文本，再查模型是否支持 tools |
| 额度错误被 LangChain 包成普通异常 | 打印 `status` 和响应体，对照 [错误码](/guide/errors/) |

## 相关

- [OpenAI 兼容调用](/guide/clients/openai-compatible/)
- [创建 API Key](/guide/api-key/)
