# Responses 接口（Codex / 推理）

`POST /v1/responses` 是新一代 Agent 与深度推理模型的核心接口，也是 **OpenAI Codex CLI** 的默认通信协议（`wire_api = "responses"`）。网关同时支持上下文压缩路径 `POST /v1/responses/compact`。

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/responses</code></div>
<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{API_BASE_URL}}/responses/compact</code></div>

## Responses 与 Chat Completions 的区别

| 维度 | `POST /v1/responses` | `POST /v1/chat/completions` |
| --- | --- | --- |
| 典型场景 | Codex CLI、多步代码 Agent、长链路深度推理 | 通用聊天客户端、翻译插件、跨厂商统一接入 |
| 输入字段 | `input`（支持字符串或消息/工具输出数组）与 `instructions` | `messages` 数组 |
| 推理控制 | `"reasoning": {"effort": "low" \| "medium" \| "high"}` | `"reasoning_effort"` |
| 输出结构 | `output` 数组（区分 `reasoning` 项与 `message` 项） | `choices[0].message` |

<div class="callout"><strong>分组提示</strong><p>调用前请在 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 确认目标模型（如 <code>gpt-6.2-astra</code>、<code>gpt-6.2-sol</code>）及密钥分组已开放 Responses 端点支持，否则会返回 <code>503</code> 无可用渠道。</p></div>

## 1. 基础请求与推理强度控制

```bash
curl "{{API_BASE_URL}}/responses" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-astra",
    "instructions": "给出可直接运行的生产级 Go 代码，处理好边界情况。",
    "reasoning": {
      "effort": "high"
    },
    "input": "实现一个带并发安全的令牌桶限流器"
  }'
```

## 2. 流式返回（SSE）

开启 `"stream": true` 后，服务端会按事件类型推送 `response.created`、`response.output_text.delta`、`response.completed` 等标准事件流：

```bash
curl "{{API_BASE_URL}}/responses" \
  -H "Authorization: Bearer $OPENBRIDGER_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "stream": true,
    "input": "简述 TCP 三次握手与四次挥手的关键状态迁移"
  }'
```

## 3. OpenAI 官方 SDK 调用示例（Python）

```python
import os
from openai import OpenAI

client = OpenAI(
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url="{{API_BASE_URL}}",
)

resp = client.responses.create(
    model="gpt-6.2-astra",
    reasoning={"effort": "medium"},
    input="审查这段 SQL 是否有索引失效风险：SELECT * FROM orders WHERE YEAR(created_at) = 2026;",
)

print(resp.output_text)
```

## 相关

- [OpenAI Codex 配置教程](/guide/clients/codex-cli/)
- [Chat Completions 对话与识图](/api/openai-chat/)
- [思考推理参数与返回字段](/api/thinking-fields/)
