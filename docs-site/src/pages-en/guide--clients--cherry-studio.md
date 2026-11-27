# Cherry Studio Setup Guide

Add OpenBridger as a custom OpenAI provider in Cherry Studio for chat, vision, reasoning, and local PDF knowledge bases (RAG).

## Before You Begin

1. Create a key named `cherry-studio` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
2. Copy exact Model IDs from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`, `gemini-4-neon`).

## 5 Setup Steps

<div class="steps">
  <div class="step"><strong>Step 1: Open Model Provider</strong><p>Click <strong>⚙️ Settings</strong> (bottom-left) -> <strong>Model Provider</strong>.</p></div>
  <div class="step"><strong>Step 2: Add Provider & Switch It ON</strong><p>Scroll to the bottom and click <strong><code>+ Add</code></strong>, name it <code>OpenBridger</code>, pick type <strong><code>OpenAI</code></strong>, and toggle the top-right provider switch <strong>ON</strong>.</p></div>
  <div class="step"><strong>Step 3: Enter Key & Host (Check Gray Preview!)</strong>
    <div class="config-card">
      <div><span>API Key</span>Your <code>sk-...</code> key</div>
      <div><span>API Host</span>Enter <code>{{API_BASE_URL}}</code> or <code>{{CONSOLE_URL}}</code> so the gray preview below reads <code>{{API_BASE_URL}}/chat/completions</code></div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Add Models & Click Check</strong><p>Under Models, click <strong><code>+ Add</code></strong>, paste the exact Model ID (enable <strong>Vision</strong> / <strong>Reasoning</strong> / <strong>Tools</strong> badges if supported), then click <strong><code>Check</code></strong> next to API Key.</p></div>
  <div class="step"><strong>Step 5: Select Model in Chat</strong><p>Click 💬 at the top-left, open the top model dropdown, and pick your model under <code>OpenBridger</code>.</p></div>
</div>

> **Local PDF RAG**: Add `text-embedding-3-small` (with the **Embedding** badge checked) under `OpenBridger`, then select it when creating a Knowledge Base (see [Embeddings & Rerank](/api/embeddings-rerank/)).

## Common Issues

| Symptom | Fix |
| --- | --- |
| `Check` fails with `404` | Remove duplicate `/v1` so the gray preview line shows `{{API_BASE_URL}}/chat/completions` |
| `OpenBridger` missing from chat dropdown | Turn ON the top-right provider switch and add at least one model |
| No image upload icon | Click ⚙️ next to the model and check **Vision** |
| `503` No available channel | Re-copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) |

## Related

- [Chatbox](/guide/clients/chatbox/)
- [Embeddings & Rerank](/api/embeddings-rerank/)
- [Error codes](/guide/errors/)
