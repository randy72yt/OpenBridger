# Open WebUI Setup Guide (Lab & Self-Hosted Hub)

Add OpenBridger once in Open WebUI's Admin Panel to share selected models with authorized users.

<div class="callout warning"><strong>Shared key reminder</strong><p>All users on the instance bill against this API key — create a dedicated key with a spending cap in <a href="{{CONSOLE_URL}}/keys">API Keys</a>.</p></div>

## Method 1: Admin Panel Setup (Recommended)

<div class="steps">
  <div class="step"><strong>Step 1: Open Connections</strong><p>Sign in as Admin -> click your avatar -> <strong>Admin Panel</strong> -> <strong>Settings</strong> -> <strong>Connections</strong>.</p></div>
  <div class="step"><strong>Step 2: Add OpenAI Connection</strong><p>Click <strong><code>+</code></strong> on the <strong><code>OpenAI API</code></strong> row (and turn OFF unused <code>Ollama API</code> below to prevent load delays).</p></div>
  <div class="step"><strong>Step 3: Enter URL, Key & Model Allowlist</strong>
    <div class="config-card">
      <div><span>URL</span><code>{{API_BASE_URL}}</code> (with <code>/v1</code>)</div>
      <div><span>API Key</span>Dedicated <code>sk-...</code> key</div>
      <div><span>Model IDs (Recommended)</span>Add exact IDs (e.g., <code>gpt-6.2-astra</code>, <code>claude-sonnet-5-5</code>) via <code>+</code></div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Verify & Click Both Save Buttons</strong><p>Click <strong>🔄 Verify</strong> -> click <strong>Save</strong> in the modal -> <strong>click the main Save button at the bottom-right of the Connections page</strong>.</p></div>
  <div class="step"><strong>Step 5: Grant Member Visibility (Multi-User)</strong><p>If non-admin users can't see the models, go to <strong>Admin Panel -> Workspace -> Models</strong> and set visibility to Public or your group.</p></div>
</div>

## Method 2: Docker Command

```bash
docker run -d -p 3000:8080 \
  -e OPENAI_API_BASE_URL="{{API_BASE_URL}}" \
  -e OPENAI_API_KEY="sk-your-key" \
  -e ENABLE_OLLAMA_API="false" \
  -v open-webui:/app/backend/data \
  --name open-webui \
  --restart always \
  ghcr.io/open-webui/open-webui:main
```

## Common Issues

| Symptom | Fix |
| --- | --- |
| 10+ second spinner when opening chat | Disable unused `Ollama API` and configure a `Model IDs` allowlist |
| Settings vanish on refresh | Click the main **Save** button at the bottom of the Connections page after closing the modal |

## Related

- [Create an API key](/guide/api-key/)
- [Dify, n8n & Coze](/guide/clients/dify-workflows/)
- [Usage and logs](/guide/account/usage/)
