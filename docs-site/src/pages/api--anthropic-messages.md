# Anthropic Messages 原生接口

网关原生支持 Anthropic Messages 协议（`POST /v1/messages`），供 **Claude Code**、Anthropic 官方 Python / TypeScript SDK 以及各类原生 Claude 客户端直连使用，完整支持流式输出、Extended Thinking（扩展思考）、Prompt 缓存、识图与 PDF 解析。

<div class="endpoint-bar"><span class="method-badge method-post">POST</span><code>{{CONSOLE_URL}}/v1/messages</code></div>

<div class="callout warning"><strong>注意 Base URL 不要多加 /v1</strong><p>在 Anthropic 官方 SDK 或 Claude Code 中配置 <code>base_url</code> / <code>ANTHROPIC_BASE_URL</code> 时，请填站点根地址 <code>{{CONSOLE_URL}}</code>，SDK 会自动拼接 <code>/v1/messages</code>。详见 <a href="/guide/base-urls/">Base URL 填写速查</a>。</p></div>

## 请求头

| Header | 必填 | 说明 |
| --- | --- | --- |
| `x-api-key` | 是 | 直接填 `sk-...` 密钥（无需加 `Bearer ` 前缀；此路径同样兼容 `Authorization: Bearer`） |
| `anthropic-version` | 是 | 推荐填 `2023-06-01` |
| `content-type` | 是 | `application/json` |

## 1. 基础对话与流式输出

Anthropic 协议要求显式传入 `max_tokens`：

```bash
curl "{{CONSOLE_URL}}/v1/messages" \
  -H "x-api-key: $OPENBRIDGER_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-opus-5-5",
    "max_tokens": 2048,
    "stream": true,
    "system": "你是一名资深软件架构师。",
    "messages": [
      {"role": "user", "content": "对比 PostgreSQL 中 B-Tree 与 GIN 索引的适用边界"}
    ]
  }'
```

## 2. 开启扩展思考（Extended Thinking）

调用支持深度思考的 Claude 模型时，传入 `thinking` 对象，响应 `content` 数组中会先返回 `type: "thinking"` 思考过程块，再返回 `type: "text"` 最终答案（注意 `max_tokens` 必须大于 `budget_tokens`）：

```bash
curl "{{CONSOLE_URL}}/v1/messages" \
  -H "x-api-key: $OPENBRIDGER_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-opus-5-5",
    "max_tokens": 8192,
    "thinking": {
      "type": "enabled",
      "budget_tokens": 4096
    },
    "messages": [
      {"role": "user", "content": "证明调和级数发散，并给出两种不同思路的证明过程。"}
    ]
  }'
```

## 3. 图片与 PDF 文档输入（Base64）

```bash
curl "{{CONSOLE_URL}}/v1/messages" \
  -H "x-api-key: $OPENBRIDGER_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-sonnet-5-5",
    "max_tokens": 2048,
    "messages": [
      {
        "role": "user",
        "content": [
          {
            "type": "document",
            "source": {
              "type": "base64",
              "media_type": "application/pdf",
              "data": "<PDF_BASE64_DATA>"
            }
          },
          {"type": "text", "text": "总结这篇论文的核心贡献、实验数据集与局限性。"}
        ]
      }
    ]
  }'
```

## 4. Anthropic Python SDK 示例

```python
import os
import anthropic

client = anthropic.Anthropic(
    api_key=os.environ["OPENBRIDGER_API_KEY"],
    base_url="{{CONSOLE_URL}}",  # 注意不带 /v1
)

msg = client.messages.create(
    model="claude-sonnet-5-5",
    max_tokens=1024,
    messages=[{"role": "user", "content": "Hello from Anthropic SDK!"}],
)

print(msg.content[0].text)
```

<div class="callout"><strong>注</strong><p><code>/v1/messages/count_tokens</code> 当前未开放；Token 实际消耗以响应体中的 <code>usage</code> 字段及控制台使用日志为准。</p></div>

## 相关

- [Claude Code 配置教程](/guide/clients/claude-code/)
- [CC Switch（一键配置）](/guide/clients/cc-switch/)
- [错误码](/guide/errors/)
