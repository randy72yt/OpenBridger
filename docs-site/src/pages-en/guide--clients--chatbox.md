# Chatbox Setup Guide (Desktop & Mobile)

Add OpenBridger as a custom OpenAI-compatible provider in Chatbox on macOS, Windows, iOS, or Android (skip Chatbox's official subscription sign-in).

## Before You Begin

1. Create a key named `chatbox` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
2. Copy exact Model IDs from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`).

## 5 Setup Steps

<div class="steps">
  <div class="step"><strong>Step 1: Open Settings</strong><p>Click <strong>⚙️ Settings</strong> (bottom-left) -> <strong>Model Provider</strong>.</p></div>
  <div class="step"><strong>Step 2: Add Custom Provider</strong><p>Scroll down and click <strong><code>+ Add Custom Provider</code></strong>, name it <code>OpenBridger</code>, and set API Mode to <strong><code>OpenAI API Compatible</code></strong> (or pick <code>OpenAI API</code> on older versions).</p></div>
  <div class="step"><strong>Step 3: Enter Host & Key</strong>
    <div class="config-card">
      <div><span>API Host</span><code>{{API_BASE_URL}}</code> (with <code>/v1</code>)</div>
      <div><span>API Path (if shown)</span><code>/chat/completions</code></div>
      <div><span>API Key</span>Your <code>sk-...</code> key</div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Add Model & Save</strong><p>Click <strong><code>+ New</code></strong> under Models, paste the exact Model ID (check <strong>Vision</strong> if supported), and click <strong>Save</strong>.</p></div>
  <div class="step"><strong>Step 5: Start a New Chat</strong><p>Click <strong><code>+ New Chat</code></strong> in the sidebar, select your <code>OpenBridger</code> model below the input box, and send a test message.</p></div>
</div>

## Common Issues

| Symptom | Fix |
| --- | --- |
| `404` (`/v1/v1/chat/completions`) | Keep API Host as `{{API_BASE_URL}}` and API Path as `/chat/completions` |
| `401` Unauthorized | Re-copy `sk-...` without extra spaces |
| Nothing in Usage Logs | Click **`+ New Chat`** and explicitly select your custom model |

## Related

- [Cherry Studio](/guide/clients/cherry-studio/)
- [LobeChat](/guide/clients/lobechat/)
- [Base URL reference](/guide/base-urls/)
