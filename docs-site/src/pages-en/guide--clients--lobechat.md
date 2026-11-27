# LobeChat Step-by-Step Setup Guide (Web, Desktop & Docker)

LobeChat (LobeHub) supports rich agent personas, plugins, and multimodal chat. You can connect it to OpenBridger in five quick steps using its built-in OpenAI provider or a custom provider.

---

## Before You Begin

1. Open [API Keys]({{CONSOLE_URL}}/keys), create a dedicated key named `lobechat`, and copy your `sk-...` token.
2. Open [Models & Pricing]({{CONSOLE_URL}}/pricing) and copy the exact model IDs you plan to use (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`, `gemini-4-neon`).

---

## Step-by-Step UI Configuration (1-2-3-4-5)

<div class="steps">
  <div class="step">
    <strong>Step 1: Open Application Settings -> AI Provider</strong>
    <p>Launch LobeChat, click your <strong>avatar in the top-left corner</strong>, select <strong>⚙️ Settings</strong>, and click <strong>AI Provider (Language Model)</strong> in the left sidebar.</p>
  </div>
  <div class="step">
    <strong>Step 2: Select OpenAI (or Add a Custom Provider)</strong>
    <p>Click <strong><code>OpenAI</code></strong> at the top of the provider list and make sure its <strong>enable switch is ON</strong>.<br>
    <em>(Alternatively, scroll to the bottom and click <code>+ Add Custom Provider</code> with SDK type <code>OpenAI</code> and name <code>OpenBridger</code>.)</em></p>
  </div>
  <div class="step">
    <strong>Step 3: Fill In API Key & API Proxy URL</strong>
    <p>Enter the following values:<br>
    • <strong>API Key</strong>: paste your <code>sk-...</code> key.<br>
    • <strong>API Proxy URL (Endpoint)</strong>: enter <strong><code>{{API_BASE_URL}}</code></strong> (ends with <code>/v1</code>, no <code>/chat/completions</code>).</p>
  </div>
  <div class="step">
    <strong>Step 4: Add Models & Run Connectivity Check</strong>
    <p>Scroll down to the <strong>Model List</strong> section:<br>
    1. Click <strong><code>+ Add Model</code></strong> (or fetch the model list), paste the exact Model ID from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a> (such as <code>gpt-6.2-astra</code>), and toggle Vision / Tool Use if supported.<br>
    2. In the <strong>Connectivity Check</strong> row above, change the test model dropdown to the model you just added and click <strong><code>Check</code></strong>. A green checkmark confirms everything is wired up.</p>
  </div>
  <div class="step">
    <strong>Step 5: Start a New Session & Verify</strong>
    <p>Return to the main chat workspace, open a new session ("Just Chat"), click the <strong>model badge</strong> above the message input to select your custom model, send a test message, and confirm it appears in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p>
  </div>
</div>

### Quick Reference Card

<div class="config-card">
  <div><span>Menu Path</span>Top-left Avatar -> <code>Settings</code> -> <code>AI Provider</code> -> <code>OpenAI</code></div>
  <div><span>API Key</span><code>sk-...</code></div>
  <div><span>API Proxy URL</span><code>{{API_BASE_URL}}</code></div>
  <div><span>Model ID</span>Exact ID from Models & Pricing (e.g., <code>gpt-6.2-astra</code>)</div>
</div>

---

## Optional: Self-Hosted Docker Environment Variables

If you deploy LobeChat with Docker, pass the gateway URL and key via environment variables:

```bash
docker run -d -p 3210:3210 \
  -e OPENAI_API_KEY="sk-your-key" \
  -e OPENAI_PROXY_URL="{{API_BASE_URL}}" \
  -e ACCESS_CODE="your-passcode" \
  lobehub/lobe-chat
```

## Common Issues & Fixes

| Symptom | Cause | Fix |
| --- | --- | --- |
| `Check` fails complaining about a default model | Connectivity check dropdown defaults to an unconfigured model ID | Pick your newly added custom model ID in the dropdown next to `Check` first |
| `404 Not Found` | Missing `/v1` or duplicate `/v1/v1` in Proxy URL | Set API Proxy URL to `{{API_BASE_URL}}` |
| Browser network / CORS error | Browser extension blocking client-side requests | Turn off **Use Client-Side Fetching** in the provider settings |
| Requests still go to `api.openai.com` | Chat session is still bound to another provider's model | Select the model listed under your configured `OpenAI` / `OpenBridger` section |

## Related

- [Cherry Studio](/guide/clients/cherry-studio/)
- [Chatbox](/guide/clients/chatbox/)
- [Troubleshooting](/troubleshooting/)
