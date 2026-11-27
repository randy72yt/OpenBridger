# FAQ

## 1. Which Base URL should I enter — with or without `/v1`?

- **Most clients (Codex, Cursor, Cherry Studio, Chatbox, LobeChat, OpenAI SDKs)**: `{{API_BASE_URL}}` (with `/v1`).
- **Claude Code / Gemini CLI / Anthropic SDK**: `{{CONSOLE_URL}}` (**no `/v1`**).
- **Immersive Translate / Coze HTTP nodes**: `{{API_BASE_URL}}/chat/completions` (full path).
- Check the [Base URL Reference](/guide/base-urls/) whenever in doubt.

## 2. Should I type `Bearer ` in front of my API key?

- **In GUI client API Key fields**: paste **only** `sk-...` (no `Bearer `).
- **In raw HTTP headers or `curl`**: send `Authorization: Bearer sk-...`.

## 3. Do new accounts include free credits? How do I top up?

New accounts start at ¥0. Open [Wallet]({{CONSOLE_URL}}/wallet) to top up via Alipay / WeChat Pay or redeem a voucher code — see [Top up and redeem](/guide/account/redeem/).

## 4. Why is the model list empty in my client?

Many apps filter out newer model IDs they don't recognize. Click **"+ Add Custom Model"** in your client and paste the exact Model ID copied from [Models & Pricing]({{CONSOLE_URL}}/pricing).

## 5. Do I need to log out of Cursor / Trae / CodeBuddy first?

No. Keep your editor account signed in — custom API keys work alongside your existing login.

## 6. How do I confirm requests are actually going through OpenBridger?

Open a **New Chat** in your client, send one short message, and check [Usage Logs]({{CONSOLE_URL}}/usage-logs) for a new entry under that key.

## 7. What do `401`, `403`, `404`, and `503` mean at a glance?

- `401`: Wrong API key, extra spaces, or the key's own spending cap ran out.
- `403`: Wallet balance is empty (top up in [Wallet]({{CONSOLE_URL}}/wallet)) or group/IP restriction blocked it.
- `404`: Base URL has an extra `/v1/v1` or is missing `/v1`.
- `503`: Key and URL work, but the Model ID has a typo or isn't in your key's group.

See [Error codes](/guide/errors/) and [Troubleshooting](/troubleshooting/) for full details.
