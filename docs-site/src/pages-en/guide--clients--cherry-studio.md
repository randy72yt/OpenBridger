# Cherry Studio Step-by-Step Setup Guide

Cherry Studio is a full-featured desktop AI client supporting multi-model comparison, local RAG knowledge bases, vision, and reasoning models. Follow steps **1 -> 2 -> 3 -> 4 -> 5** below to connect OpenBridger in under a minute.

---

## Before You Begin

1. Open [API Keys]({{CONSOLE_URL}}/keys), create a dedicated key named `cherry-studio`, and copy your `sk-...` token.
2. Open [Models & Pricing]({{CONSOLE_URL}}/pricing) and copy the exact model IDs you plan to use (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`, `gemini-4-neon`, `deepseek-v4.1`).

---

## Step-by-Step UI Configuration (1-2-3-4-5)

<div class="steps">
  <div class="step">
    <strong>Step 1: Open Model Provider Settings</strong>
    <p>Launch Cherry Studio, click the <strong>⚙️ Settings gear icon</strong> at the bottom-left corner, and select <strong>Model Provider</strong> at the top of the settings sidebar.</p>
  </div>
  <div class="step">
    <strong>Step 2: Add a Custom Provider</strong>
    <p>Scroll to the bottom of the provider list and click <strong><code>+ Add</code></strong>:<br>
    • <strong>Provider Name</strong>: <code>OpenBridger</code><br>
    • <strong>Provider Type</strong>: select <strong><code>OpenAI</code></strong><br>
    Click <strong>OK</strong>, and make sure the toggle switch at the top-right of the <code>OpenBridger</code> panel is <strong>switched ON</strong>.</p>
  </div>
  <div class="step">
    <strong>Step 3: Fill In API Key & API Host (Check the Gray Preview URL!)</strong>
    <p>In the right-hand configuration pane:<br>
    1. <strong>API Key</strong>: paste your <code>sk-...</code> key.<br>
    2. <strong>API Host</strong>: enter <code>{{API_BASE_URL}}</code> or <code>{{CONSOLE_URL}}</code> — <strong>pro tip</strong>: look at the small gray preview line directly underneath the input box. As long as the preview reads <strong><code>{{API_BASE_URL}}/chat/completions</code></strong> (with exactly one <code>/v1</code>), it is 100% correct! If you see <code>/v1/v1</code>, remove the trailing <code>/v1</code> from the input.</p>
  </div>
  <div class="step">
    <strong>Step 4: Add Models & Click "Check"</strong>
    <p>Scroll down to the <strong>Models</strong> section:<br>
    • Click <strong><code>+ Add</code></strong> (or click <strong><code>Manage</code></strong> to fetch models automatically and click <code>+</code> next to the ones you want).<br>
    • When adding manually, paste the exact Model ID from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a> (such as <code>gpt-6.2-astra</code>). Check <strong>Vision</strong>, <strong>Reasoning</strong>, or <strong>Function Calling</strong> if the model supports them.<br>
    • Click the <strong><code>Check</code></strong> button next to your API Key at the top, pick the model you added, and confirm the green connection success toast.</p>
  </div>
  <div class="step">
    <strong>Step 5: Start Chatting</strong>
    <p>Click the 💬 <strong>Chat icon</strong> at the top-left sidebar, click the <strong>model selector dropdown</strong> at the top bar, choose your model under <code>OpenBridger</code>, and send a message.</p>
  </div>
</div>

### Quick Reference Card

<div class="config-card">
  <div><span>Provider Name</span><code>OpenBridger</code> (ensure toggle switch is ON)</div>
  <div><span>Provider Type</span><code>OpenAI</code></div>
  <div><span>API Key</span><code>sk-...</code></div>
  <div><span>API Host</span>Verify gray preview reads <code>{{API_BASE_URL}}/chat/completions</code></div>
  <div><span>Model ID</span>Exact ID from Models & Pricing (e.g., <code>gpt-6.2-astra</code>, <code>claude-sonnet-5-5</code>)</div>
</div>

---

## Bonus: Local PDF Knowledge Base (RAG) in Cherry Studio

To index PDF papers or lecture notes locally:
1. Under the `OpenBridger` provider, click **`+ Add`**, enter `text-embedding-3-large` (or `text-embedding-3-small`), and check the **Embedding** capability badge.
2. Open the **Knowledge Base** tab in the left sidebar -> Create Knowledge Base -> pick that embedding model, then drag and drop your PDFs or Markdown files (see [Embeddings & Rerank](/api/embeddings-rerank/)).

## Common Issues & Fixes

| Symptom | Cause | Fix |
| --- | --- | --- |
| `Check` fails with `404` | Gray preview URL shows `/v1/v1/chat/completions` or misses `/v1` | Adjust API Host until the gray preview line below the input shows `{{API_BASE_URL}}/chat/completions` |
| `OpenBridger` doesn't appear in the chat model dropdown | Provider toggle is OFF or no models have been added yet | Enable the toggle at the top-right of the `OpenBridger` provider page and add at least one model |
| Image upload icon is missing in chat | Vision capability was not enabled on the manually added model | Click the ⚙️ gear next to the model in Cherry Studio and check **Vision** |
| `503` no available channel | Typo in Model ID or key group mismatch | Re-copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) |

## Related

- [Chatbox](/guide/clients/chatbox/)
- [LobeChat](/guide/clients/lobechat/)
- [Embeddings & Rerank](/api/embeddings-rerank/)
- [Error codes](/guide/errors/)
