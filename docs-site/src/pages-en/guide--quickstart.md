# 3-Minute Quick Start

Whether you are setting up a coding agent, a paper translator, or a desktop chat app, getting started takes **4 steps**.

## 4 Steps to Get Started

<div class="steps">
  <div class="step">
    <strong>Step 1: Sign Up & Sign In</strong>
    <p>Open <a href="{{CONSOLE_URL}}">{{CONSOLE_URL}}</a>, register an account, and sign in to the console.</p>
  </div>
  <div class="step">
    <strong>Step 2: Top Up or Redeem Credits</strong>
    <p>Open <a href="{{CONSOLE_URL}}/wallet">Wallet</a> (new accounts start with a ¥0 balance). Top up online or redeem a voucher code (see <a href="/guide/account/redeem/">Top up and redeem</a>).</p>
  </div>
  <div class="step">
    <strong>Step 3: Create an API Key</strong>
    <p>Open <a href="{{CONSOLE_URL}}/keys">API Keys</a> -> click <strong>Create Key</strong> -> give it a clear name (such as <code>codex</code> or <code>cherry</code>) -> save and copy the full <code>sk-...</code> key.</p>
  </div>
  <div class="step">
    <strong>Step 4: Start Using Models</strong>
    <p><strong>• Chat in Browser (No Install)</strong>: Open <a href="{{CONSOLE_URL}}/playground">GPT Chat</a> in the console sidebar and start chatting immediately.<br>
    <strong>• Connect Your Favorite App</strong>: Copy an exact Model ID from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a> and follow the step-by-step guide for your tool below:</p>
  </div>
</div>

### Popular Client Guides

- **Coding**: [CC Switch (1-Click GUI)](/guide/clients/cc-switch/) · [Codex (Desktop, IDE & CLI)](/guide/clients/codex-cli/) · [Claude Code](/guide/clients/claude-code/) · [Cursor](/guide/clients/cursor/) · [Cline & Roo Code](/guide/clients/cline/)
- **Research & Chat**: [Immersive Translate](/guide/clients/immersive-translate/) · [Zotero](/guide/clients/zotero/) · [Cherry Studio](/guide/clients/cherry-studio/) · [Chatbox](/guide/clients/chatbox/) · [LobeChat](/guide/clients/lobechat/)

---

## Optional: 10-Second Terminal `curl` Check (For Developers)

If you are integrating via code or terminal, run a quick request to verify your key:

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer sk-replace-with-your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "messages": [{"role": "user", "content": "Reply with pong"}]
  }'
```

When you see `choices[0].message.content` in the JSON response and a matching entry in [Usage Logs]({{CONSOLE_URL}}/usage-logs), everything is working. For SDK code, see [Python & Node SDKs](/guide/clients/openai-compatible/).

## Top 3 First-Run Errors

| Error | Cause | Fix |
| --- | --- | --- |
| `401` Invalid token | Truncated key, extra spaces, or duplicate `Bearer` prefix | Copy the `sk-...` key fresh from [API Keys]({{CONSOLE_URL}}/keys) |
| `403` Insufficient quota | Account balance is ¥0 | Top up or redeem a code in [Wallet]({{CONSOLE_URL}}/wallet) |
| `503` No available channel | Abbreviated model name or key group mismatch | Copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) |

See [Error codes](/guide/errors/) for the full list.
