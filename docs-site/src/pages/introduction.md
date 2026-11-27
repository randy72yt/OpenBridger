# OpenBridger 使用文档

一个账号、一把 API Key，统一调用 **GPT-6.2 / Codex、Claude 5.5、Gemini 4、DeepSeek** 等主流模型。按实际用量以人民币（¥）结算，无需分别绑卡或维护多家官方账号。

<div class="actions"><a class="btn primary" href="/guide/quickstart/">3 分钟快速开始</a><a class="btn secondary" href="{{CONSOLE_URL}}">打开控制台</a></div>

## 按你的使用场景直达教程

<div class="grid">
  <div class="card"><h3>💻 AI 编程与命令行</h3><p><a href="/guide/clients/cc-switch/">CC Switch 一键配置</a> · <a href="/guide/clients/codex-cli/">Codex 桌面版/插件/CLI</a> · <a href="/guide/clients/claude-code/">Claude Code</a> · <a href="/guide/clients/cursor/">Cursor</a> · <a href="/guide/clients/cline/">Cline / Roo Code</a></p></div>
  <div class="card"><h3>📚 论文阅读与双语翻译</h3><p><a href="/guide/clients/immersive-translate/">沉浸式翻译（PDF / 网页双语）</a> · <a href="/guide/clients/zotero/">Zotero 文献精读</a> · <a href="/guide/clients/cherry-studio/">Cherry Studio 本地知识库</a></p></div>
  <div class="card"><h3>💬 日常跨端聊天与共享</h3><p><a href="{{CONSOLE_URL}}/playground">网页版 GPT 对话（免配置）</a> · <a href="/guide/clients/chatbox/">Chatbox（电脑/手机）</a> · <a href="/guide/clients/lobechat/">LobeChat</a> · <a href="/guide/clients/open-webui/">Open WebUI（课题组面板）</a></p></div>
  <div class="card"><h3>🔌 代码接入与自动化工作流</h3><p><a href="/api/openai-chat/">OpenAI 对话接口</a> · <a href="/api/openai-responses/">Responses 接口</a> · <a href="/api/anthropic-messages/">Anthropic 原生接口</a> · <a href="/guide/clients/dify-workflows/">Dify / n8n / Coze</a></p></div>
</div>

## 最常填的三个核心参数

绝大多数软件只需要填这三项（不同软件的地址格式差异见 [Base URL 填写速查](/guide/base-urls/)）：

<div class="config-card">
  <div><span>OpenAI 兼容 Base URL</span><code>{{API_BASE_URL}}</code>（Claude Code 填 <code>{{CONSOLE_URL}}</code>，不带 <code>/v1</code>）</div>
  <div><span>API Key（密钥）</span><code>sk-...</code>（在控制台 <a href="{{CONSOLE_URL}}/keys">API 密钥</a> 创建）</div>
  <div><span>Model（模型 ID）</span>在 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 点击复制完整英文 ID</div>
</div>

## 控制台常用入口

| 功能 | 入口地址 | 用途 |
| --- | --- | --- |
| **GPT 对话** | [{{CONSOLE_URL}}/playground]({{CONSOLE_URL}}/playground) | 网页端直接选模型对话，无需配置客户端 |
| **模型广场** | [{{CONSOLE_URL}}/pricing]({{CONSOLE_URL}}/pricing) | 查看在售模型、支持协议与实时单价，一键复制模型 ID |
| **API 密钥** | [{{CONSOLE_URL}}/keys]({{CONSOLE_URL}}/keys) | 为不同软件创建独立 `sk-...` 密钥、设置额度上限 |
| **钱包 / 订阅** | [{{CONSOLE_URL}}/wallet]({{CONSOLE_URL}}/wallet) | 支付宝 / 微信充值、兑换码充值、购买订阅套餐 |
| **使用日志** | [{{CONSOLE_URL}}/usage-logs]({{CONSOLE_URL}}/usage-logs) | 逐笔查看调用时间、模型、Token 数与扣费明细 |

## 开发者接口一览

| 协议 / 能力 | 请求路径 | 参考文档 |
| --- | --- | --- |
| 模型列表 | `GET /v1/models` | [Chat Completions](/api/openai-chat/) |
| OpenAI 对话与识图 | `POST /v1/chat/completions` | [Chat Completions](/api/openai-chat/) |
| OpenAI Responses（Codex） | `POST /v1/responses` | [Responses 接口](/api/openai-responses/) |
| Anthropic Messages（Claude Code） | `POST /v1/messages` | [Anthropic 原生接口](/api/anthropic-messages/) |
| Google Gemini 原生协议 | `POST /v1beta/models/{model}:generateContent` | [Gemini 原生接口](/api/gemini-native/) |
| 图像 / 视频 / 语音 / 实时语音 | `/v1/images/*`、`/v1/videos`、`/v1/audio/*`、`/v1/realtime` | [图像、视频与语音](/api/multimodal-media/) |
| 向量化与重排序（RAG） | `POST /v1/embeddings`、`POST /v1/rerank` | [向量与 Rerank](/api/embeddings-rerank/) |

<div class="callout warning"><strong>合规提示</strong><p>本平台仅面向非中国大陆地区用户提供服务，访问与使用须符合你所在地法律法规及 <a href="/legal/terms/">用户协议</a>。</p></div>
