# SDK 接入

服务端应用可以使用 OpenAI 官方 SDK 或支持自定义 Base URL 的框架。API Key 应存放在服务端环境变量或密钥管理服务中。

## 环境变量

```bash
export OPENBRIDGER_API_KEY="YOUR_API_KEY"
export OPENBRIDGER_BASE_URL="{{API_BASE_URL}}"
```

## 上线前检查

- 使用服务端调用，禁止把 Key 打包进浏览器或移动端公开代码。
- 为开发、测试和生产环境使用不同 Key。
- 设置连接、读取和总请求超时，并对 429 与临时上游错误做有限重试。
- 记录请求 ID、模型、耗时和状态；日志中隐藏凭证与敏感内容。
- 对用户输入、最大输出和多模态数量设置业务边界。

通用 Node.js 与 Python 代码见“OpenAI 兼容客户端”。

## LangChain JavaScript 示例

```js
import { ChatOpenAI } from "@langchain/openai";

const model = new ChatOpenAI({
  apiKey: process.env.OPENBRIDGER_API_KEY,
  configuration: { baseURL: process.env.OPENBRIDGER_BASE_URL },
  model: "MODEL_ID_FROM_CONSOLE",
});
```

## 常见问题

| 现象 | 检查方式 |
| --- | --- |
| SDK 仍请求官方 OpenAI | JS 检查 `baseURL`，Python 检查 `base_url` |
| 浏览器出现 CORS | 把调用移到自己的服务端，不要让前端携带 Key |
| tools 参数失败 | 先验证普通文本，再确认模型和线路支持工具调用 |
| 429 或临时错误 | 使用带退避的有限重试，避免无限循环 |
