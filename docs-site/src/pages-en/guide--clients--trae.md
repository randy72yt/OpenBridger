# Trae Setup Guide

Add OpenBridger as a custom OpenAI-compatible model in Trae IDE's **Chat** and **Builder / Agent** panels without signing out of your Trae account.

## Before You Begin

1. Create a key named `trae` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
2. Copy an exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`).

## 5 Setup Steps

<div class="steps">
  <div class="step"><strong>Step 1: Open Model Settings</strong><p>Click the <strong>⚙️ gear icon</strong> in the AI sidebar -> <strong><code>Models</code></strong> (or click <strong><code>+ Add Model</code></strong> at the bottom of the chat model dropdown).</p></div>
  <div class="step"><strong>Step 2: Choose OpenAI Compatible</strong><p>Set Provider to <strong><code>OpenAI Compatible</code></strong> (update Trae or install <a href="/guide/clients/cline/">Cline / Roo Code</a> if your build lacks custom Base URLs).</p></div>
  <div class="step"><strong>Step 3: Fill In Parameters</strong>
    <div class="config-card">
      <div><span>Base URL</span><code>{{API_BASE_URL}}</code> (with <code>/v1</code>)</div>
      <div><span>Model ID</span>Exact ID from Models & Pricing (e.g., <code>gpt-6.2-astra</code>)</div>
      <div><span>API Key</span>Your <code>sk-...</code> key</div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Save</strong><p>Click <strong>Add / Save</strong> (repeat to add multiple models).</p></div>
  <div class="step"><strong>Step 5: Select Model in Chat / Builder</strong><p>Open a new thread and select your custom model (<strong>Chat</strong> and <strong>Builder</strong> store their selected models separately).</p></div>
</div>

## Common Issues

| Symptom | Fix |
| --- | --- |
| Inline Tab autocomplete isn't in Usage Logs | Inline completions run on Trae's built-in engine |
| `404` / `503` | `404`: set Base URL to `{{API_BASE_URL}}`; `503`: re-copy exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) |

## Related

- [Cursor](/guide/clients/cursor/)
- [Cline & Roo Code](/guide/clients/cline/)
- [Troubleshooting](/troubleshooting/)
