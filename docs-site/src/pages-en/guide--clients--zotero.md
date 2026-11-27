# Zotero

Pairing Zotero with an AI add-on such as **Awesome GPT (`zotero-gpt`)** lets you summarize PDF papers, explain equations inline, and query across references in your library via OpenBridger.

## Before you start

- Zotero 7 (or Zotero 6) with the **Awesome GPT / Zotero GPT** add-on installed.
- A dedicated API key named `zotero` created under [API Keys]({{CONSOLE_URL}}/keys).
- A chat model ID from [Model Square]({{CONSOLE_URL}}/pricing) (e.g. `gpt-6.2-sol`, `claude-sonnet-5-5`, or `gemini-4-neon`), plus an optional embedding model ID (`text-embedding-3-small`) if you use local PDF semantic search.

## Setup steps

1. Open Zotero **Settings / Preferences** and select the **GPT** tab.
2. Fill in the connection fields:

<div class="config-card">
  <div><span>API</span><code>{{CONSOLE_URL}}</code> or <code>{{API_BASE_URL}}</code> (see note below)</div>
  <div><span>Secret Key</span><code>sk-your-api-key</code></div>
  <div><span>Model</span><code>Full model ID from Model Square</code></div>
  <div><span>Embedding Model (optional)</span><code>text-embedding-3-small</code></div>
</div>

<div class="callout"><strong>Root URL vs. /v1</strong><p>Standard <code>zotero-gpt</code> builds append <code>/v1/chat/completions</code> automatically — enter <code>{{CONSOLE_URL}}</code>. If your fork prompts specifically for an OpenAI Base URL, enter <code>{{API_BASE_URL}}</code>. Send a short test prompt and verify the request appears in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div>

## Common issues

| Symptom | Fix |
| --- | --- |
| `404` error | Switch between `{{CONSOLE_URL}}` and `{{API_BASE_URL}}` so `/v1` is neither missing nor doubled |
| Selection chat works, but full-PDF Q&A fails | Full-PDF Q&A calls `/v1/embeddings`; ensure the configured Embedding Model exists in [Model Square]({{CONSOLE_URL}}/pricing) |
| `403` `insufficient_user_quota` | Long papers send tens of thousands of prompt tokens; check your [Wallet]({{CONSOLE_URL}}/wallet) balance |

## Related

- [Immersive Translate](/guide/clients/immersive-translate/)
- [Embeddings & Rerank](/api/embeddings-rerank/)
- [Error codes](/guide/errors/)
