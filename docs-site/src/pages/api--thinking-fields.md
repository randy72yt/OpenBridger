# 思考推理参数与返回字段

当前主流旗舰模型（GPT-6.2、Claude 5.5、Gemini 4、DeepSeek Reasoner、Qwen4、Grok 4.7）普遍内置深度思考（Reasoning / Thinking）能力。自己写代码解析响应或排查客户端不显示“思考过程”时，对照本页的请求参数与返回字段。

## 一、如何控制思考强度（请求侧）

| 调用协议 | 请求参数写法 | 可选值 / 说明 |
| --- | --- | --- |
| **OpenAI Chat** (`/v1/chat/completions`) | `"reasoning_effort": "medium"` | `"low"`、`"medium"`、`"high"`（仅推理模型生效） |
| **OpenAI Responses** (`/v1/responses`) | `"reasoning": {"effort": "high"}` | `"low"`、`"medium"`、`"high"`，Codex CLI 在 `config.toml` 中通过 `model_reasoning_effort` 配置 |
| **Anthropic Messages** (`/v1/messages`) | `"thinking": {"type": "enabled", "budget_tokens": 4096}` | `budget_tokens` 必须小于 `max_tokens` |
| **Gemini 原生** (`/v1beta/models/...`) | `"generationConfig": {"thinkingConfig": {"includeThoughts": true}}` | 支持按模型设定思考预算 |

## 二、不同协议返回的“思考内容”在哪里（响应侧）

当你开发自定义前端或脚本时，思考文本可能出现在以下三种位置之一：

### 1. OpenAI 兼容格式：`reasoning_content` 字段（最常见）

通过 `/v1/chat/completions` 调用 DeepSeek Reasoner 及网关转换后的推理模型时，思考过程与最终正文是分开的两个字段：

- **非流式**：`choices[0].message.reasoning_content`（思考过程） + `choices[0].message.content`（最终回答）
- **流式 SSE**：先连续收到 `choices[0].delta.reasoning_content`，随后收到 `choices[0].delta.content`

```json
{
  "choices": [
    {
      "delta": {
        "reasoning_content": "先分析边界条件 n = 0 与 n = 1...",
        "content": ""
      }
    }
  ]
}
```

### 2. Anthropic 原生格式：`content` 数组中的 `thinking` 块

通过 `/v1/messages` 调用并开启 `thinking` 时，`content` 是一个多元素数组：

```json
{
  "content": [
    {
      "type": "thinking",
      "thinking": "逐步推导调和级数分组下界...",
      "signature": "..."
    },
    {
      "type": "text",
      "text": "证明如下：..."
    }
  ]
}
```

### 3. 正文内嵌 `<think>...</think>` 标签

极少数开源模型渠道会将思考链以 `<think>思考内容</think>最终正文` 的形式直接拼在 `content` 字符串开头。Cherry Studio、Chatbox、LobeChat 等现代客户端均已自动识别以上三种格式并折叠展示。

<div class="callout warning"><strong>计费说明</strong><p>模型在内部生成的思考 Token 同样消耗上游算力，按照行业标准统一计入输出 Token（<code>completion_tokens</code> / <code>output_tokens</code>）结算。若日常简单翻译或摘要不需要深度推理，选择 Flash / 轻量模型或调低 <code>reasoning_effort</code> 可显著降低耗时与消耗。</p></div>

## 相关

- [Chat Completions 对话与识图](/api/openai-chat/)
- [Responses 接口（Codex / 推理）](/api/openai-responses/)
- [Anthropic Messages 原生接口](/api/anthropic-messages/)
