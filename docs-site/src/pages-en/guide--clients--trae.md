# Trae Step-by-Step Setup Guide

Trae IDE allows adding custom OpenAI-compatible models for its **Chat** and **Builder / Agent** panels without signing out of your Trae account. Follow steps **1 -> 2 -> 3 -> 4 -> 5** below.

---

## Before You Begin

1. Open [API Keys]({{CONSOLE_URL}}/keys), create a dedicated key named `trae`, and copy your `sk-...` token.
2. Open [Models & Pricing]({{CONSOLE_URL}}/pricing) and copy the exact model ID you want to use (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`, `deepseek-v4.1`).

---

## Step-by-Step UI Configuration (1-2-3-4-5)

<div class="steps">
  <div class="step">
    <strong>Step 1: Open Trae AI Model Settings</strong>
    <p>In Trae, click the <strong>⚙️ gear icon</strong> in the top-right corner of the AI sidebar, or click the <strong>model selector dropdown</strong> at the bottom of the chat box and select <strong><code>+ Add Model</code></strong> (or open <strong><code>Models</code></strong> in AI Settings).</p>
  </div>
  <div class="step">
    <strong>Step 2: Choose OpenAI Compatible Provider</strong>
    <p>In the Add Model dialog, set <strong>Provider</strong> to <strong><code>OpenAI Compatible</code></strong>.<br>
    <em>(If an older build of Trae only lists fixed presets without a Base URL field, update Trae or install the <a href="/guide/clients/cline/">Cline / Roo Code extension</a> inside Trae.)</em></p>
  </div>
  <div class="step">
    <strong>Step 3: Fill In Base URL, Model ID & API Key</strong>
    <p>Enter the fields below:<br>
    • <strong>Base URL</strong>: <code>{{API_BASE_URL}}</code> (only append <code>/chat/completions</code> if the input placeholder explicitly asks for the full endpoint)<br>
    • <strong>Model ID</strong>: exact model ID copied from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a> (e.g., <code>gpt-6.2-astra</code>)<br>
    • <strong>API Key</strong>: your <code>sk-...</code> key</p>
  </div>
  <div class="step">
    <strong>Step 4: Save the Custom Model</strong>
    <p>Click <strong>Add / Save</strong>. Repeat steps 2–4 if you want to add multiple models (for example, both a GPT model and a Claude model).</p>
  </div>
  <div class="step">
    <strong>Step 5: Select the Model in Chat / Builder & Verify</strong>
    <p>Start a new thread in the AI panel and select your newly added model from the picker (note that <strong>Chat</strong> and <strong>Builder</strong> remember their selected models separately). Send a test message and confirm it appears in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p>
  </div>
</div>

### Quick Reference Card

<div class="config-card">
  <div><span>Provider</span><code>OpenAI Compatible</code></div>
  <div><span>Base URL</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Key</span><code>sk-...</code></div>
  <div><span>Model ID</span>Exact ID from Models & Pricing (e.g., <code>gpt-6.2-astra</code>)</div>
</div>

---

## Common Issues & Fixes

| Symptom | Cause | Fix |
| --- | --- | --- |
| Inline Tab autocomplete doesn't hit OpenBridger | Trae's inline ghost-text engine runs on its own completion service | Expected — Chat and Builder panels route through OpenBridger |
| `404 Not Found` | Missing `/v1` or extra path segment in Base URL | Check the input placeholder; default to `{{API_BASE_URL}}` |
| `503` no available channel / `model not found` | Typo in Model ID or key group mismatch | Re-copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) |
| Nothing recorded in Usage Logs | Switching between Chat and Builder reset the dropdown to a built-in model | Re-select your custom model in the active mode's picker |

## Related

- [Cursor](/guide/clients/cursor/)
- [Cline & Roo Code](/guide/clients/cline/)
- [Troubleshooting](/troubleshooting/)
