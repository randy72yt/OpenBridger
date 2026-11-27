# LobeChat Setup Guide (Web, Desktop & Docker)

Connect LobeChat to OpenBridger using its built-in `OpenAI` provider (or a custom provider).

## Before You Begin

1. Create a key named `lobechat` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
2. Copy exact Model IDs from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`).

## 5 Setup Steps

<div class="steps">
  <div class="step"><strong>Step 1: Open AI Provider Settings</strong><p>Click your <strong>top-left avatar</strong> -> <strong>⚙️ Settings</strong> -> <strong>AI Provider</strong>.</p></div>
  <div class="step"><strong>Step 2: Enable OpenAI (or Add Custom Provider)</strong><p>Select <strong><code>OpenAI</code></strong> and toggle its switch ON (or click <code>+ Add Custom Provider</code> at the bottom with format <code>OpenAI</code>).</p></div>
  <div class="step"><strong>Step 3: Enter Key & Proxy URL</strong>
    <div class="config-card">
      <div><span>API Key</span>Your <code>sk-...</code> key</div>
      <div><span>API Proxy URL</span><code>{{API_BASE_URL}}</code> (with <code>/v1</code>)</div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Add Model & Click Check</strong><p>Click <strong><code>+ Add Model</code></strong> to paste your exact Model ID, select that model in the <strong>Connectivity Check</strong> dropdown above, and click <strong><code>Check</code></strong>.</p></div>
  <div class="step"><strong>Step 5: Start a New Session</strong><p>Return to chat, open a new session, and pick your custom model from the model badge above the input bar.</p></div>
</div>

### Optional: Self-Hosted Docker Command

```bash
docker run -d -p 3210:3210 \
  -e OPENAI_API_KEY="sk-your-key" \
  -e OPENAI_PROXY_URL="{{API_BASE_URL}}" \
  -e ACCESS_CODE="your-passcode" \
  lobehub/lobe-chat
```

## Common Issues

| Symptom | Fix |
| --- | --- |
| `Check` fails on default model | Select your newly added Model ID in the dropdown next to `Check` first |
| `404` Not Found | Set API Proxy URL to `{{API_BASE_URL}}` |
| Browser CORS / Network Error | Disable **Use Client-Side Fetching** in the provider settings |

## Related

- [Cherry Studio](/guide/clients/cherry-studio/)
- [Chatbox](/guide/clients/chatbox/)
- [Troubleshooting](/troubleshooting/)
