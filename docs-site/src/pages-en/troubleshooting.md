# Troubleshooting (30-Second Fix Guide)

Check the **3-digit HTTP status code** in your error message — 95% of setup issues match one row in the table below.

## 1. Match Your Error Code

| Error | Real Cause | Immediate Fix |
| --- | --- | --- |
| **`401` Invalid token** | Truncated `sk-...` key, extra `Bearer ` prefix in a GUI field, or the key's own spending cap ran out | Re-copy the raw `sk-...` key from [API Keys]({{CONSOLE_URL}}/keys) and check its cap |
| **`403` Insufficient quota** | Wallet / plan balance is ¥0 or below the request pre-hold | Top up or redeem a code in [Wallet]({{CONSOLE_URL}}/wallet) |
| **`404` Not Found** | Duplicate `/v1/v1/...` or missing `/v1` in Base URL | Check [Base URL Reference](/guide/base-urls/) |
| **`503` No available channel** | **URL and key work!** Either the Model ID has a typo or your key's group doesn't carry that model | Re-copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) or set key group to `auto` |
| **`429` Rate limit** | Burst concurrency (e.g., translating a 50-page PDF at once) or temporary upstream congestion | Lower requests-per-second in your plugin or retry shortly |
| **Still asks for official login / nothing in logs** | Client wasn't restarted or you're typing in an old thread | **Quit and relaunch the app**, click **+ New Chat**, and select your custom model |

---

## 2. Double-Check Your Base URL Format

<div class="config-card">
  <div><span>Codex / Cursor / Cherry Studio / Chatbox / LobeChat / Open WebUI</span><code>{{API_BASE_URL}}</code> (with <code>/v1</code>)</div>
  <div><span>Claude Code / Gemini CLI / Anthropic SDK</span><code>{{CONSOLE_URL}}</code> (<strong>NO /v1</strong>)</div>
  <div><span>Immersive Translate / Coze HTTP Node</span><code>{{API_BASE_URL}}/chat/completions</code> (full path)</div>
</div>

---

## 3. Optional Terminal Check (For Developers)

To isolate a client UI quirk from your API key, run:

```bash
curl -i "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer sk-your-key"
```

- `HTTP/2 200` with JSON models: your key and network are healthy — check the client's Base URL or active thread.
- Still stuck? Copy the `request id` from the error response and see [Contact and incidents](/guide/support/).
