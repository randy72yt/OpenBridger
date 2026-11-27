# Chatbox Step-by-Step Setup Guide (Desktop, Mobile & Web)

Chatbox is a lightweight cross-platform chat client for macOS, Windows, iOS, and Android. Follow steps **1 -> 2 -> 3 -> 4 -> 5** below to connect OpenBridger.

---

## Before You Begin

1. Open [API Keys]({{CONSOLE_URL}}/keys), create a dedicated key named `chatbox`, and copy your `sk-...` token.
2. Open [Models & Pricing]({{CONSOLE_URL}}/pricing) and copy the exact model IDs you want to use (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`, `gemini-4-neon`).
3. Do **not** choose "Sign in to Chatbox AI official subscription" — add a custom OpenAI-compatible provider instead.

---

## Step-by-Step UI Configuration (1-2-3-4-5)

<div class="steps">
  <div class="step">
    <strong>Step 1: Open Settings</strong>
    <p>Launch Chatbox, click <strong>⚙️ Settings</strong> in the bottom-left corner, and select the <strong>Model Provider</strong> tab.</p>
  </div>
  <div class="step">
    <strong>Step 2: Add a Custom Provider</strong>
    <p>Scroll to the bottom of the provider sidebar and click <strong><code>+ Add Custom Provider</code></strong>:<br>
    • <strong>Name</strong>: <code>OpenBridger</code><br>
    • <strong>API Mode</strong>: select <strong><code>OpenAI API Compatible</code></strong><br>
    <em>(On older Chatbox versions without custom providers, select <code>OpenAI API</code> directly.)</em></p>
  </div>
  <div class="step">
    <strong>Step 3: Enter API Host, Path & Key</strong>
    <p>Fill in the provider fields:<br>
    • <strong>API Host</strong>: <code>{{API_BASE_URL}}</code> (ends with <code>/v1</code>, no trailing slash)<br>
    • <strong>API Path</strong>: keep the default <code>/chat/completions</code> (if shown)<br>
    • <strong>API Key</strong>: paste your <code>sk-...</code> key</p>
  </div>
  <div class="step">
    <strong>Step 4: Add Models & Enable Capabilities</strong>
    <p>Under <strong>Models</strong>, click <strong><code>+ New</code></strong> (or <strong><code>Fetch</code></strong>):<br>
    Paste the exact Model ID from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a> (such as <code>gpt-6.2-astra</code>). Enable <strong>Vision</strong> or <strong>Reasoning</strong> if supported by that model, then click <strong>Save</strong>.</p>
  </div>
  <div class="step">
    <strong>Step 5: Start a New Chat & Verify</strong>
    <p>Click <strong><code>+ New Chat</code></strong> in the main sidebar (older threads remember their previous provider!), click the <strong>model selector</strong> below the message input box, pick your <code>OpenBridger</code> model, send a short message, and verify it in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p>
  </div>
</div>

### Quick Reference Card

<div class="config-card">
  <div><span>Provider / API Mode</span>Custom Provider -> <code>OpenAI API Compatible</code></div>
  <div><span>API Host</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Path</span><code>/chat/completions</code> (if shown)</div>
  <div><span>API Key</span><code>sk-...</code></div>
  <div><span>Model</span>Exact ID from Models & Pricing (e.g., <code>gpt-6.2-astra</code>)</div>
</div>

---

## Common Issues & Fixes

| Symptom | Cause | Fix |
| --- | --- | --- |
| `404` error mentioning `/v1/v1/chat/completions` | `/v1` was entered in both API Host and API Path | Keep API Host as `{{API_BASE_URL}}` and API Path as `/chat/completions` |
| `401 Unauthorized` | Extra spaces or truncated `sk-...` key | Re-copy and paste your API key |
| No entries in Usage Logs | Still chatting inside an older thread bound to another provider | Click **`+ New Chat`** and explicitly select your `OpenBridger` model |
| Cannot attach images | Vision toggle wasn't checked on the custom model | Edit the model in Chatbox settings and enable **Vision** |

## Related

- [Cherry Studio](/guide/clients/cherry-studio/)
- [LobeChat](/guide/clients/lobechat/)
- [Base URL reference](/guide/base-urls/)
