# OpenBridger 是一个 API 网关

OpenBridger 把多家模型服务放到同一套账号、API Key 和额度下面。你用自己的密钥调用网关，网关再把请求转发到当前可用的上游线路，并按控制台展示的价格记一笔消耗。

它是给开发者和日常调用准备的 API 工具：注册、充值、发密钥、看日志。模型是否在售、某次请求走哪条线路，以控制台当时的数据为准，不在本文事先列清单。

<div class="actions"><a class="btn primary" href="/guide/quickstart/">从第一次调用开始</a><a class="btn secondary" href="{{CONSOLE_URL}}">进入控制台</a></div>

<div class="grid"><div class="card"><h3>一个入口</h3><p>OpenAI 兼容调用使用同一 Base URL 和 <code>Authorization: Bearer</code>。Anthropic 兼容调用走同一站点上的 <code>/v1/messages</code>。</p></div><div class="card"><h3>按次计费</h3><p>调用前会预扣额度，结束后按实际用量结算差额。价格看模型广场，不在文档里写死单价。</p></div><div class="card"><h3>按密钥记账</h3><p>给每个工具单独建 API Key，再到使用日志里按密钥核对模型、状态和消耗。</p></div></div>

## 你要记住的三个值

<div class="config-card"><div><span>控制台</span><code>{{CONSOLE_URL}}</code></div><div><span>OpenAI 兼容 Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>sk-</code> 开头，在「API 密钥」里创建</div></div>

文档里的示例统一用 `{{API_BASE_URL}}`。Anthropic 客户端会自己在主机名后面加 `/v1/messages`，那种情况只填 `{{CONSOLE_URL}}`，不要再加 `/v1`。详见 [Claude Code](/guide/clients/claude-code/)。

## 网关实际提供的调用面

下面这些路径在网关代码里有对应处理。某个模型能不能用其中一条，还要看模型广场和当前分组有没有可用线路。

| 用途 | 方法与路径 |
| --- | --- |
| 列出当前密钥可见的模型 | `GET /v1/models` |
| 对话 | `POST /v1/chat/completions` |
| 文本补全 | `POST /v1/completions` |
| Responses（Codex 使用） | `POST /v1/responses` |
| Anthropic Messages | `POST /v1/messages` |
| Embeddings | `POST /v1/embeddings` |
| 图像生成 / 编辑 | `POST /v1/images/generations`、`POST /v1/images/edits` |
| 语音转写、翻译、合成 | `POST /v1/audio/transcriptions`、`/v1/audio/translations`、`/v1/audio/speech` |
| Rerank、审核 | `POST /v1/rerank`、`POST /v1/moderations` |
| Gemini 原生路径 | `POST /v1beta/models/...` |
| 实时语音 | `GET /v1/realtime`（WebSocket） |

`POST /v1/images/variations`、`/v1/files` 和 `/v1/fine-tunes` 会返回 `501`，`code` 为 `api_not_implemented`。`/v1/messages/count_tokens` 当前没有开放。视频和异步任务走 `/v1/videos`、`/v1/tasks`，只有对应能力已启用时才会成功。

## 控制台入口

| 要做的事 | 地址 |
| --- | --- |
| 模型广场 | [{{CONSOLE_URL}}/pricing]({{CONSOLE_URL}}/pricing) |
| API 密钥 | [{{CONSOLE_URL}}/keys]({{CONSOLE_URL}}/keys) |
| 钱包 | [{{CONSOLE_URL}}/wallet]({{CONSOLE_URL}}/wallet) |
| AI 体验台 | [{{CONSOLE_URL}}/playground]({{CONSOLE_URL}}/playground) |
| 使用日志 | [{{CONSOLE_URL}}/usage-logs]({{CONSOLE_URL}}/usage-logs) |
| 用户协议 | [{{CONSOLE_URL}}/user-agreement]({{CONSOLE_URL}}/user-agreement) |

<div class="callout warning"><strong>服务范围以用户协议为准</strong><p>主站公布的用户协议写明：本平台仅面向非中国大陆地区用户。访问和使用须符合你所在地法律、上游服务条款和平台规则。协议全文见 <a href="/legal/terms/">用户协议</a>。</p></div>

## 相关

- [快速开始](/guide/quickstart/)
- [额度与计费](/guide/billing/)
- [错误码](/guide/errors/)
