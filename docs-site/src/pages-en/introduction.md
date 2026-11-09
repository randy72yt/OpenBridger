# OpenBridger is an API gateway

OpenBridger puts multiple model providers behind one account, one API key, and one credit balance. You call the gateway with your key. The gateway forwards the request to a route that is available for that model, then records the charge shown in the console.

This is a developer API: register, add credits, create a key, and read logs. Which models are on sale, and which route a request uses, is whatever the console shows at that moment. This guide does not keep a model price list.

<div class="actions"><a class="btn primary" href="/guide/quickstart/">Make the first call</a><a class="btn secondary" href="{{CONSOLE_URL}}">Open console</a></div>

<div class="grid"><div class="card"><h3>One entry point</h3><p>OpenAI-compatible calls use one Base URL and <code>Authorization: Bearer</code>. Anthropic-compatible calls use <code>/v1/messages</code> on the same host.</p></div><div class="card"><h3>Pay for each call</h3><p>The gateway reserves credits before the call and settles the difference afterward. Prices live in Model Square, not in this guide.</p></div><div class="card"><h3>One key per tool</h3><p>Create a separate API key for each client, then match the model, status, and charge in Usage Logs.</p></div></div>

## Three values to keep

<div class="config-card"><div><span>Console</span><code>{{CONSOLE_URL}}</code></div><div><span>OpenAI-compatible Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API key</span>Starts with <code>sk-</code>. Create it under API Keys.</div></div>

Examples use `{{API_BASE_URL}}`. Anthropic clients append `/v1/messages` themselves, so those clients use `{{CONSOLE_URL}}` with no `/v1`. See [Claude Code](/guide/clients/claude-code/).

## What the gateway actually routes

These paths have handlers in the gateway. A model can use one of them only when Model Square and your group have a route for it.

| Use | Method and path |
| --- | --- |
| List models visible to this key | `GET /v1/models` |
| Chat | `POST /v1/chat/completions` |
| Text completion | `POST /v1/completions` |
| Responses (Codex) | `POST /v1/responses` |
| Anthropic Messages | `POST /v1/messages` |
| Embeddings | `POST /v1/embeddings` |
| Image generation / edits | `POST /v1/images/generations`, `POST /v1/images/edits` |
| Speech to text, translation, speech | `POST /v1/audio/transcriptions`, `/v1/audio/translations`, `/v1/audio/speech` |
| Rerank, moderation | `POST /v1/rerank`, `POST /v1/moderations` |
| Gemini native path | `POST /v1beta/models/...` |
| Realtime audio | `GET /v1/realtime` (WebSocket) |

`POST /v1/images/variations`, `/v1/files`, and `/v1/fine-tunes` return `501` with `code` `api_not_implemented`. `/v1/messages/count_tokens` is not enabled. Video and async tasks use `/v1/videos` and `/v1/tasks` only when that capability is turned on.

## Console pages

| Task | URL |
| --- | --- |
| Model Square | [{{CONSOLE_URL}}/pricing]({{CONSOLE_URL}}/pricing) |
| API Keys | [{{CONSOLE_URL}}/keys]({{CONSOLE_URL}}/keys) |
| Wallet | [{{CONSOLE_URL}}/wallet]({{CONSOLE_URL}}/wallet) |
| Playground | [{{CONSOLE_URL}}/playground]({{CONSOLE_URL}}/playground) |
| Usage Logs | [{{CONSOLE_URL}}/usage-logs]({{CONSOLE_URL}}/usage-logs) |
| Terms | [{{CONSOLE_URL}}/user-agreement]({{CONSOLE_URL}}/user-agreement) |

<div class="callout warning"><strong>Who the service is for</strong><p>The terms published on the main site say this service is only for users outside mainland China. Your use also has to follow the laws that apply to you, upstream terms, and the platform rules. See <a href="/legal/terms/">Terms of service</a>.</p></div>

## Related

- [Quick start](/guide/quickstart/)
- [Credits and billing](/guide/billing/)
- [Error codes](/guide/errors/)
