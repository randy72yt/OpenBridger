# Base URL Reference

Different clients append endpoint paths differently. Entering the wrong format usually causes a `404` (`/v1/v1/...`) or returns HTML instead of JSON. Match your tool in the table below.

## The Three Formats

<div class="config-card">
  <div><span>Format A: Root URL (no /v1)</span><code>{{CONSOLE_URL}}</code></div>
  <div><span>Format B: Standard Base URL (with /v1)</span><code>{{API_BASE_URL}}</code></div>
  <div><span>Format C: Full Chat Endpoint</span><code>{{API_BASE_URL}}/chat/completions</code></div>
</div>

## Client & SDK Lookup Table

| Client / SDK | Format | Exact Value to Enter | Notes |
| --- | --- | --- | --- |
| **Codex (Desktop / VS Code / CLI)** | Format B | `{{API_BASE_URL}}` | Set `wire_api = "responses"`; save key in `auth.json` or via [CC Switch](/guide/clients/cc-switch/) |
| **Claude Code / Anthropic SDK** | Format A | `{{CONSOLE_URL}}` | **Do NOT add `/v1`** — client appends `/v1/messages` |
| **Google Gemini CLI / GenAI SDK** | Format A | `{{CONSOLE_URL}}` | **Do NOT add `/v1` or `/v1beta`** |
| **Cursor / Cline / Roo Code / Trae / CodeBuddy** | Format B | `{{API_BASE_URL}}` | Use OpenAI Compatible mode |
| **Cherry Studio / Chatbox / LobeChat** | Format B | `{{API_BASE_URL}}` | Preview URL below input should read `.../v1/chat/completions` |
| **Open WebUI / Dify / n8n** | Format B | `{{API_BASE_URL}}` | OpenAI API connection |
| **Immersive Translate** | Format C | `{{API_BASE_URL}}/chat/completions` | Custom OpenAI endpoint requires the full path |
| **Coze / Generic HTTP Node** | Format C | `{{API_BASE_URL}}/chat/completions` | Header: `Authorization: Bearer sk-...` |
| **Zotero GPT** | Format A or B | `{{CONSOLE_URL}}` (or `{{API_BASE_URL}}`) | Toggle between A and B if you see `404` |
| **OpenAI Python / Node SDK / LangChain** | Format B | `{{API_BASE_URL}}` | SDK appends `/chat/completions` or `/responses` |

## 10-Second Error Check

| Symptom | Cause & Fix |
| --- | --- |
| `404` and nothing in Usage Logs | Duplicate `/v1/v1/...` or missing `/v1` — switch between Format A, B, and C above |
| Response is an HTML web page | Client does not append paths automatically — use Format C (`{{API_BASE_URL}}/chat/completions`) |
| `503` No available channel | **Your Base URL is already working!** Re-copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) |

## Related

- [CC Switch (1-Click GUI)](/guide/clients/cc-switch/)
- [Error codes](/guide/errors/)
