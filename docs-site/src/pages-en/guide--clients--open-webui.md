# Open WebUI Step-by-Step Setup Guide (Lab & Self-Hosted Hub)

Open WebUI is a popular self-hosted web interface for research labs, study groups, and personal servers. Once an administrator connects OpenBridger in the Admin Panel, authorized users can chat with your selected models in the browser.

<div class="callout warning"><strong>Shared key quota reminder</strong><p>Connections added in the Admin Panel are shared across authorized Open WebUI users, and all usage bills against that single API key. Create a dedicated key with a spending limit in <a href="{{CONSOLE_URL}}/keys">API Keys</a>.</p></div>

---

## Method 1: Visual Setup in the Admin Panel (Recommended)

Sign in with an **Admin** account and follow steps **1 -> 2 -> 3 -> 4 -> 5**:

<div class="steps">
  <div class="step">
    <strong>Step 1: Open Admin Panel -> Settings -> Connections</strong>
    <p>Click your <strong>user avatar</strong> (bottom-left or top-right) -> select <strong>Admin Panel</strong> -> click the <strong>Settings</strong> tab at the top -> open <strong>Connections</strong>.</p>
  </div>
  <div class="step">
    <strong>Step 2: Add an OpenAI API Connection</strong>
    <p>Under the <strong><code>OpenAI API</code></strong> section (ensure its toggle is ON), click the <strong><code>+</code> (Add Connection)</strong> button on the right.</p>
  </div>
  <div class="step">
    <strong>Step 3: Enter URL, API Key & Model Allowlist</strong>
    <p>Fill in the connection modal:<br>
    • <strong>URL</strong>: <strong><code>{{API_BASE_URL}}</code></strong> (ends with <code>/v1</code>, no trailing slash)<br>
    • <strong>API Key</strong>: paste your dedicated <code>sk-...</code> key<br>
    • <strong>Model IDs (Recommended)</strong>: add the exact model IDs you want exposed (e.g., <code>gpt-6.2-astra</code>, <code>claude-sonnet-5-5</code>, <code>deepseek-v4.1</code>) by typing each ID and clicking <code>+</code>. Filtering keeps the model picker clean and speeds up page loads.</p>
  </div>
  <div class="step">
    <strong>Step 4: Verify & Click Both Save Buttons</strong>
    <p>Click the <strong>🔄 Verify Connection</strong> icon next to the URL field to confirm the green toast, click <strong>Save</strong> inside the modal, and <strong>click the main Save button at the bottom-right of the Connections page</strong> so the settings persist.</p>
  </div>
  <div class="step">
    <strong>Step 5: Start a New Chat & Check Logs</strong>
    <p>Click <strong>New Chat</strong> in the top-left corner, choose one of your added models, send a short prompt, and verify the entry in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p>
  </div>
</div>

### Quick Reference Card

<div class="config-card">
  <div><span>Menu Path</span>Avatar -> <code>Admin Panel</code> -> <code>Settings</code> -> <code>Connections</code> -> <code>OpenAI API</code></div>
  <div><span>URL</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Key</span>Dedicated <code>sk-...</code> key (with quota cap)</div>
  <div><span>Model IDs</span>Add exact IDs such as <code>gpt-6.2-astra</code>, <code>claude-sonnet-5-5</code></div>
</div>

---

## Method 2: Docker Environment Variables

When deploying Open WebUI via Docker (disabling unused local Ollama probes prevents startup lag):

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

> After editing `docker-compose.yml`, recreate the container (`docker compose up -d`) — refreshing the browser alone does not reload container environment variables.

---

## Common Issues & Fixes

| Symptom | Cause | Fix |
| --- | --- | --- |
| Slow initial page load (10+ seconds spinner) | Unused `Ollama API` probe timing out, or fetching hundreds of unfiltered models | Turn off **Ollama API** under Connections and specify a **Model IDs** allowlist |
| Non-admin users see no models | Newly added models default to private/admin-only visibility in newer Open WebUI releases | Go to **Admin Panel -> Workspace -> Models** and set model visibility to Public or your user group |
| Connection disappears after page refresh | Clicked Save in the modal but missed the bottom page Save button | Click **Save** at the bottom-right of the Connections page |

## Related

- [Create an API key](/guide/api-key/)
- [Dify, n8n & Coze](/guide/clients/dify-workflows/)
- [Usage and logs](/guide/account/usage/)
