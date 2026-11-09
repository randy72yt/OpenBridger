# 模型与价格

可售模型、协议和单价都在 [模型广场]({{CONSOLE_URL}}/pricing)。客户端里填写的模型名必须和广场上的完整模型 ID 一致，包括大小写、短横线和版本后缀。

<div class="actions"><a class="btn primary" href="{{CONSOLE_URL}}/pricing">打开模型广场</a></div>

文档不刊登单价表。单价会随上游报价和运营发布变化，写在这里会过时。广场上每个模型的输入、输出、缓存和其他计量单位，就是那次调用的计价依据。

## 选择时看这四件事

- **协议。** 聊天客户端通常要 Chat Completions（`/v1/chat/completions`）。Codex 要 Responses（`/v1/responses`）。Claude Code 要 Anthropic Messages（`/v1/messages`）。模型广场里有这个名字，不表示三种协议都能用。
- **分组。** 同一模型 ID 在不同分组下的线路和价格可以不同。密钥上的分组决定你实际能选到哪条。
- **计量。** 有的按输入和输出 token，有的还计缓存、工具调用、图像张数或音视频时长。先看该模型在广场上的说明，再用一条短请求到 [使用日志]({{CONSOLE_URL}}/usage-logs) 核对。
- **可用性。** 没有可用线路时，网关返回 `503`，文案含「无可用渠道」，`code` 为 `model_not_found`。这和上游返回的 `404` 不是同一件事，见 [错误码](/guide/errors/)。

## 名称怎么写

从广场复制。不要把展示名、厂商口语名或你记得的旧 ID 填进客户端。网关按你提交的 `model` 字段选线路，不会根据「大概是哪个模型」去猜。

```bash
curl "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer ${OPENBRIDGER_API_KEY}"
```

这个列表适合核对密钥能看见哪些 ID。部分客户端不读取它，或者会滤掉自己不认识的名字。滤掉时，把广场上的 ID 手工加进客户端。

## 和计费的关系

一次成功调用的扣费顺序是：按模型价格估算并预扣，请求结束后按实际用量结算，多退少补。预扣失败不会发出上游请求。公式和额度单位见 [额度与计费](/guide/billing/)。

生产调用应准备模型不可用时的处理：换一个广场里仍可用的模型 ID，或稍后重试。不要在客户端里写死一个文档中的示例 ID。

## 相关

- [额度与计费](/guide/billing/)
- [分组与线路](/guide/account/groups/)
- [OpenAI 兼容调用](/guide/clients/openai-compatible/)
