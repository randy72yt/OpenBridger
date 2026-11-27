# Zotero Setup Guide (Paper Reading & PDF Q&A)

Pair Zotero with the **Awesome GPT (`zotero-gpt`)** plugin to translate passages, explain formulas, and query full PDFs while reading papers.

## Before You Begin

1. Create a key named `zotero` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
2. Copy a long-context chat Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gemini-4-neon`, `claude-sonnet-5-5`, `gpt-6.2-sol`), plus `text-embedding-3-small` if you use full-PDF vector search.

## 4 Setup Steps

<div class="steps">
  <div class="step"><strong>Step 1: Install Awesome GPT (If Needed)</strong><p>Download the latest <code>.xpi</code> from GitHub <code>MuiseDestiny/zotero-gpt</code> Releases -> in Zotero open <strong>Tools -> Plugins</strong> -> click the top-right ⚙️ gear -> <strong><code>Install Plugin From File...</code></strong> and select the <code>.xpi</code>.</p></div>
  <div class="step"><strong>Step 2: Open GPT Settings</strong><p>Open <strong>Edit -> Settings</strong> on Windows (or <strong><code>Zotero -> Settings</code></strong> on macOS) and select the <strong><code>GPT</code></strong> tab on the left.</p></div>
  <div class="step"><strong>Step 3: Fill In API, Key & Models</strong>
    <div class="config-card">
      <div><span>API</span><code>{{CONSOLE_URL}}</code> (switch to <code>{{API_BASE_URL}}</code> if a test returns <code>404</code>)</div>
      <div><span>Secret Key</span>Your <code>sk-...</code> key</div>
      <div><span>Model</span>Exact ID from Models & Pricing (e.g., <code>gpt-6.2-sol</code>)</div>
      <div><span>Embedding Model</span><code>text-embedding-3-small</code></div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Test Inside a PDF</strong><p>Open any PDF paper in Zotero, highlight a paragraph, and trigger a GPT action to verify streaming output.</p></div>
</div>

## Common Issues

| Symptom | Fix |
| --- | --- |
| `404` error | Toggle the `API` field between `{{CONSOLE_URL}}` and `{{API_BASE_URL}}` |
| Highlighting works, but full-PDF Q&A fails | Ensure `Embedding Model` is set to an active embedding model in your group (such as `text-embedding-3-small`) |
| `403` Insufficient quota | Full-paper prompts carry large token counts — check your balance in [Wallet]({{CONSOLE_URL}}/wallet) |

## Related

- [Immersive Translate](/guide/clients/immersive-translate/)
- [Embeddings & Rerank](/api/embeddings-rerank/)
- [Error codes](/guide/errors/)
