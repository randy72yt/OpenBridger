# CodeBuddy Step-by-Step Setup Guide (IDE & VS Code / JetBrains Plugin)

**CodeBuddy** (both the standalone CodeBuddy IDE and the CodeBuddy extension for VS Code and JetBrains IDEs) supports adding custom OpenAI-compatible models for **Chat** and **Craft / Agent** workflows. Follow steps **1 -> 2 -> 3 -> 4 -> 5** below.

---

## Before You Begin

1. Open [API Keys]({{CONSOLE_URL}}/keys), create a dedicated key named `codebuddy`, and copy your `sk-...` token.
2. Open [Models & Pricing]({{CONSOLE_URL}}/pricing) and copy the exact model ID you want to use (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`, `deepseek-v4.1`).
3. Stay signed in to your CodeBuddy account — custom models run alongside built-in inline completions.

---

## Step-by-Step UI Configuration (1-2-3-4-5)

<div class="steps">
  <div class="step">
    <strong>Step 1: Open CodeBuddy Model Settings</strong>
    <p>Open the CodeBuddy panel in your editor, click the <strong>⚙️ Settings icon</strong> at the top-right (or click the <strong>model dropdown</strong> below the chat input and choose <strong><code>+ Add Model / Manage Models</code></strong>).</p>
  </div>
  <div class="step">
    <strong>Step 2: Add a Custom OpenAI-Compatible Model</strong>
    <p>Click <strong><code>+ Add Custom Model</code></strong> and set the provider / protocol type to <strong><code>OpenAI Compatible</code></strong>.</p>
  </div>
  <div class="step">
    <strong>Step 3: Enter Base URL, Model ID & API Key</strong>
    <p>Fill in the form:<br>
    • <strong>Display Name (if shown)</strong>: e.g., <code>OB - GPT-6.2</code><br>
    • <strong>Model ID</strong>: paste the exact Model ID from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a> (e.g., <code>gpt-6.2-astra</code>)<br>
    • <strong>Base URL</strong>: <strong><code>{{API_BASE_URL}}</code></strong> (ends with <code>/v1</code>, without <code>/chat/completions</code>)<br>
    • <strong>API Key</strong>: paste your <code>sk-...</code> key</p>
  </div>
  <div class="step">
    <strong>Step 4: Test & Save</strong>
    <p>Click <strong>Test</strong> (if available in your version) and then click <strong>Save</strong>.</p>
  </div>
  <div class="step">
    <strong>Step 5: Select the Model in Chat / Craft & Verify</strong>
    <p>Start a new thread in the CodeBuddy sidebar, select your newly added model in the picker at the bottom of the input box (note that <strong>Chat</strong> and <strong>Craft</strong> modes track their selected models independently), send a test prompt, and verify it in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p>
  </div>
</div>

### Quick Reference Card

<div class="config-card">
  <div><span>Provider Type</span><code>OpenAI Compatible</code></div>
  <div><span>Base URL</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Key</span><code>sk-...</code></div>
  <div><span>Model ID</span>Exact ID from Models & Pricing (e.g., <code>gpt-6.2-astra</code>)</div>
</div>

---

## Common Issues & Fixes

| Symptom | Cause | Fix |
| --- | --- | --- |
| Custom Model menu is missing | Older plugin build or enterprise-locked edition | Update CodeBuddy to the latest version, or pair VS Code with [Cline & Roo Code](/guide/clients/cline/) |
| `404 Not Found` | Missing `/v1` or extra `/chat/completions` in Base URL | Set Base URL to `{{API_BASE_URL}}` |
| `503` no available channel / `model not found` | Swapped Display Name and Model ID, or typo in Model ID | Ensure `Model ID` matches the exact identifier from [Models & Pricing]({{CONSOLE_URL}}/pricing) |
| No entries in Usage Logs | Inline code completion uses CodeBuddy's built-in model, or picker wasn't switched | Explicitly select your custom model in the Chat / Craft input bar |

## Related

- [Cursor](/guide/clients/cursor/)
- [Cline & Roo Code](/guide/clients/cline/)
- [Troubleshooting](/troubleshooting/)
