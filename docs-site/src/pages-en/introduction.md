# OpenBridger Documentation

One account and one API key to access **GPT-6.2 / Codex, Claude 5.5, Gemini 4, DeepSeek**, and more — billed pay-as-you-go in RMB (¥) without managing separate vendor accounts.

<div class="actions"><a class="btn primary" href="/guide/quickstart/">3-Minute Quick Start</a><a class="btn secondary" href="{{CONSOLE_URL}}">Open Console</a></div>

## Jump Straight to Your Tool

<div class="grid">
  <div class="card"><h3>💻 Coding & CLI Agents</h3><p><a href="/guide/clients/cc-switch/">CC Switch (1-Click GUI)</a> · <a href="/guide/clients/codex-cli/">Codex (App / IDE / CLI)</a> · <a href="/guide/clients/claude-code/">Claude Code</a> · <a href="/guide/clients/cursor/">Cursor</a> · <a href="/guide/clients/cline/">Cline & Roo Code</a></p></div>
  <div class="card"><h3>📚 Papers & Bilingual Translation</h3><p><a href="/guide/clients/immersive-translate/">Immersive Translate (PDF & Web)</a> · <a href="/guide/clients/zotero/">Zotero</a> · <a href="/guide/clients/cherry-studio/">Cherry Studio RAG</a></p></div>
  <div class="card"><h3>💬 Daily Chat & Lab Sharing</h3><p><a href="{{CONSOLE_URL}}/playground">Web GPT Chat (Zero Setup)</a> · <a href="/guide/clients/chatbox/">Chatbox (Desktop & Mobile)</a> · <a href="/guide/clients/lobechat/">LobeChat</a> · <a href="/guide/clients/open-webui/">Open WebUI</a></p></div>
  <div class="card"><h3>🔌 APIs, SDKs & Workflows</h3><p><a href="/api/openai-chat/">Chat Completions</a> · <a href="/api/openai-responses/">Responses API</a> · <a href="/api/anthropic-messages/">Anthropic Messages</a> · <a href="/guide/clients/dify-workflows/">Dify, n8n & Coze</a></p></div>
</div>

## The Three Values You'll Use Everywhere

Most apps only ask for these three fields (see [Base URL Reference](/guide/base-urls/) for per-client rules):

<div class="config-card">
  <div><span>OpenAI-Compatible Base URL</span><code>{{API_BASE_URL}}</code> (Claude Code uses <code>{{CONSOLE_URL}}</code> without <code>/v1</code>)</div>
  <div><span>API Key</span><code>sk-...</code> (created in <a href="{{CONSOLE_URL}}/keys">API Keys</a>)</div>
  <div><span>Model ID</span>Exact ID copied from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a></div>
</div>

## Console Shortcuts

| Page | Link | What It Does |
| --- | --- | --- |
| **GPT Chat** | [{{CONSOLE_URL}}/playground]({{CONSOLE_URL}}/playground) | Chat with models directly in the browser — no client setup needed |
| **Models & Pricing** | [{{CONSOLE_URL}}/pricing]({{CONSOLE_URL}}/pricing) | Browse live models, supported protocols, and prices; copy exact Model IDs |
| **API Keys** | [{{CONSOLE_URL}}/keys]({{CONSOLE_URL}}/keys) | Create per-app `sk-...` keys and spending caps |
| **Wallet & Plans** | [{{CONSOLE_URL}}/wallet]({{CONSOLE_URL}}/wallet) | Top up, redeem codes, or purchase subscription plans |
| **Usage Logs** | [{{CONSOLE_URL}}/usage-logs]({{CONSOLE_URL}}/usage-logs) | Inspect every request's timestamp, model, tokens, and cost |

## Supported API Endpoints

| Protocol / Capability | Endpoint | Reference |
| --- | --- | --- |
| List visible models | `GET /v1/models` | [Chat Completions](/api/openai-chat/) |
| OpenAI Chat & Vision | `POST /v1/chat/completions` | [Chat Completions](/api/openai-chat/) |
| OpenAI Responses (Codex) | `POST /v1/responses` | [Responses API](/api/openai-responses/) |
| Anthropic Messages (Claude Code) | `POST /v1/messages` | [Anthropic Messages](/api/anthropic-messages/) |
| Google Gemini Native | `POST /v1beta/models/{model}:generateContent` | [Gemini Native API](/api/gemini-native/) |
| Image, Video, Audio & Realtime | `/v1/images/*`, `/v1/videos`, `/v1/audio/*`, `/v1/realtime` | [Image, Video & Audio](/api/multimodal-media/) |
| Embeddings & Rerank | `POST /v1/embeddings`, `POST /v1/rerank` | [Embeddings & Rerank](/api/embeddings-rerank/) |

<div class="callout warning"><strong>Service scope</strong><p>Per our published <a href="/legal/terms/">Terms of Service</a>, this platform is intended only for users outside Mainland China and must be used in compliance with applicable local laws.</p></div>
