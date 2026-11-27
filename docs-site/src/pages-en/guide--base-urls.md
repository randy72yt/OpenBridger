# Base URL reference

Different tools append API paths differently: some automatically append `/v1/chat/completions`, Anthropic clients append `/v1/messages`, and raw HTTP workflow nodes require the full endpoint URL. Entering the wrong variant typically causes a `404` (`/v1/v1/...`) or returns an HTML page instead of JSON.

Use this table whenever connecting a new tool or SDK.

## Three URL formats

<div class="config-card">
  <div><span>Format A: Root host (no /v1)</span><code>{{CONSOLE_URL}}</code></div>
  <div><span>Format B: OpenAI Base URL (with /v1)</span><code>{{API_BASE_URL}}</code></div>
  <div><span>Format C: Full chat endpoint</span><code>{{API_BASE_URL}}/chat/completions</code></div>
</div>

## Client and SDK lookup table

| Tool / SDK | Format | Exact value to enter | Authentication |
| --- | --- | --- | --- |
| **OpenAI Python / Node SDK** | Format B (with `/v1`) | `{{API_BASE_URL}}` | `Authorization: Bearer sk-...` |
| **OpenAI Codex CLI** | Format B (with `/v1`) | `{{API_BASE_URL}}` (`wire_api = "responses"`) | Env `OPENBRIDGER_API_KEY` |
| **Claude Code** | Format A (**no** `/v1`) | `{{CONSOLE_URL}}` | `ANTHROPIC_AUTH_TOKEN=sk-...` |
| **Anthropic Official SDK** | Format A (**no** `/v1`) | `{{CONSOLE_URL}}` | `x-api-key: sk-...` |
| **Google Gemini CLI** | Format A (**no** `/v1`) | `{{CONSOLE_URL}}` | `GEMINI_API_KEY=sk-...` |
| **Cursor / Cline / Roo Code / Trae** | Format B (OpenAI-compatible) | `{{API_BASE_URL}}` | API Key `sk-...` |
| **Cherry Studio / Chatbox / LobeChat** | Format A or B (check URL preview) | Final preview should end with `/v1/chat/completions` | API Key `sk-...` |
| **Immersive Translate** | Format C (full endpoint) | `{{API_BASE_URL}}/chat/completions` | API Key `sk-...` |
| **Zotero GPT** | Format B | See [Zotero](/guide/clients/zotero/) | Secret Key `sk-...` |
| **Dify (OpenAI-API-compatible)** | Format B (with `/v1`) | `{{API_BASE_URL}}` | API Key `sk-...` |
| **n8n (OpenAI node)** | Format B (with `/v1`) | `{{API_BASE_URL}}` | Credential API Key `sk-...` |
| **Coze / Raw HTTP Request node** | Format C (full endpoint) | `{{API_BASE_URL}}/chat/completions` | Header `Authorization: Bearer sk-...` |

<div class="callout"><strong>Quick preview check</strong><p>Many desktop apps (such as Cherry Studio and Chatbox) display a live preview of the full request URL beneath the input field. Confirm the preview resolves to <code>{{API_BASE_URL}}/chat/completions</code> — neither missing <code>/v1</code> nor doubling it into <code>/v1/v1</code>.</p></div>

## Typical URL symptoms

| Symptom | Cause | Fix |
| --- | --- | --- |
| `404 page not found` and nothing in Usage Logs | Doubled `/v1/v1/chat/completions` or missing `/v1` | Switch between Format A, B, and C using the table above |
| Response body is HTML instead of JSON | Root URL `{{CONSOLE_URL}}` was pasted into a raw HTTP client that does not append `/v1/...` | Use Format C (`{{API_BASE_URL}}/chat/completions`) |
| Claude Code returns `404` | `ANTHROPIC_BASE_URL` includes `/v1` | Set `ANTHROPIC_BASE_URL` to `{{CONSOLE_URL}}` |
| `503` with `model_not_found` | **The URL is right** — the model ID or key group has no matching upstream route | Copy the exact model ID from [Model Square]({{CONSOLE_URL}}/pricing) and check key group |

## Related

- [Quick start](/guide/quickstart/)
- [OpenAI Codex](/guide/clients/codex-cli/)
- [Claude Code](/guide/clients/claude-code/)
- [Error codes](/guide/errors/)
