# CodeBuddy Setup Guide (IDE & VS Code / JetBrains Plugin)

Add OpenBridger as a custom OpenAI-compatible model in **CodeBuddy** for **Chat** and **Craft / Agent** coding while staying signed in to your CodeBuddy account.

## Before You Begin

1. Create a key named `codebuddy` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
2. Copy an exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`).

## 5 Setup Steps

<div class="steps">
  <div class="step"><strong>Step 1: Open Model Management</strong><p>Click <strong>⚙️ Settings -> Models</strong> in the CodeBuddy panel (or click <strong><code>+ Add Model</code></strong> in the chat model dropdown).</p></div>
  <div class="step"><strong>Step 2: Select OpenAI Compatible</strong><p>Click <strong><code>+ Add Custom Model</code></strong> and choose <strong><code>OpenAI Compatible</code></strong>.</p></div>
  <div class="step"><strong>Step 3: Fill In Parameters</strong>
    <div class="config-card">
      <div><span>Display Name (if shown)</span>e.g., <code>OB - GPT-6.2</code></div>
      <div><span>Model ID</span>Exact ID from Models & Pricing (e.g., <code>gpt-6.2-astra</code>)</div>
      <div><span>Base URL</span><code>{{API_BASE_URL}}</code> (with <code>/v1</code>)</div>
      <div><span>API Key</span>Your <code>sk-...</code> key</div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Save</strong><p>Click <strong>Test</strong> (if shown) and <strong>Save</strong>.</p></div>
  <div class="step"><strong>Step 5: Select Model in Chat / Craft</strong><p>Start a new thread and select your custom model in the picker (<strong>Chat</strong> and <strong>Craft</strong> track models separately).</p></div>
</div>

## Common Issues

| Symptom | Fix |
| --- | --- |
| Custom Model option missing | Update CodeBuddy or pair VS Code with [Cline & Roo Code](/guide/clients/cline/) |
| `404` / `503` | `404`: set Base URL to `{{API_BASE_URL}}`; `503`: ensure `Model ID` holds the exact console ID |

## Related

- [Cursor](/guide/clients/cursor/)
- [Cline & Roo Code](/guide/clients/cline/)
- [Troubleshooting](/troubleshooting/)
