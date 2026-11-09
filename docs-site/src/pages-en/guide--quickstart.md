# Make the first call

Do four things: register and sign in, confirm the wallet has credits, create an API key, and call `POST /v1/chat/completions` with curl.

## Before you start

- An email address for registration and account recovery.
- A browser that can open [{{CONSOLE_URL}}]({{CONSOLE_URL}}).
- `curl` on your machine. Python and Node samples are on [OpenAI-compatible calls](/guide/clients/openai-compatible/).

The registration form is the source of truth. The program default is a username or email plus a password. If the page asks for an email code or a human check, finish that first. Third-party login such as GitHub is available only when the sign-in page shows it.

<div class="callout"><strong>Screenshot to add</strong><p>The console after sign-in, with Wallet, API Keys, Model Square, and Usage Logs marked. Replace this note with a capture of the current UI.</p></div>

## Steps

<div class="steps"><div class="step"><strong>Register and sign in</strong><p>Open <a href="{{CONSOLE_URL}}">{{CONSOLE_URL}}</a> and enter the console. Too many failed sign-ins are throttled. Wait, then try again.</p></div><div class="step"><strong>Check the wallet</strong><p>Open the <a href="{{CONSOLE_URL}}/wallet">wallet</a>. You need enough credit for one short chat. If the balance is zero, use the top-up or redemption code on that page. See <a href="/guide/account/redeem/">Top up and redeem</a>. A welcome credit, if any, is whatever extra balance appears. This guide does not promise a fixed gift.</p></div><div class="step"><strong>Copy a model ID</strong><p>Open <a href="{{CONSOLE_URL}}/pricing">Model Square</a>, pick a chat model your group can use, and copy the full model ID. Do not shorten the version.</p></div><div class="step"><strong>Create an API key</strong><p>Open <a href="{{CONSOLE_URL}}/keys">API Keys</a> and create a key named <code>curl-test</code>. The copied value starts with <code>sk-</code>. Details are on <a href="/guide/api-key/">Create an API key</a>.</p></div><div class="step"><strong>Send one short request</strong><p>Replace the key and model ID, then run the command.</p></div></div>

```bash
export OPENBRIDGER_API_KEY="sk-paste-the-key-from-the-console"
export OPENBRIDGER_MODEL="full-model-id-from-model-square"

curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer ${OPENBRIDGER_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"${OPENBRIDGER_MODEL}\",
    \"messages\": [{\"role\": \"user\", \"content\": \"Reply with one word: pong\"}]
  }"
```

A successful body is JSON with `choices[0].message.content`. The response header `X-Oneapi-Request-Id` is the id to keep if the call fails.

## Confirm the call hit the gateway

1. Open [Usage Logs]({{CONSOLE_URL}}/usage-logs).
2. Filter by the time and the key name `curl-test`.
3. Check the model ID, status, and charge. A matching row means the client did not send the request somewhere else.

The [Playground]({{CONSOLE_URL}}/playground) can test the same model with your login session. It does not use the API key, so it does not replace the curl above.

## Errors at this step

| What you see | What to do |
| --- | --- |
| `401`, message says the token is invalid | The key was copied wrong, disabled, expired, or its own credit is used up. Copy it again. Do not type a second `Bearer` or `sk-`. |
| `403`, message says user quota is insufficient | The wallet or the current funding source cannot cover the reserve. Add credits or send a smaller request. See [Credits and billing](/guide/billing/). |
| `503`, message says no channel is available | The model ID is wrong, or this group has no route. Copy it again from Model Square. |
| Connection failed or timed out | The URL must be `{{API_BASE_URL}}/chat/completions`, not `/v1/v1/`. Then see [Troubleshooting](/troubleshooting/). |

The full table is on [Error codes](/guide/errors/).

## Next

- Editors and chat apps: [Cursor](/guide/clients/cursor/), [Claude Code](/guide/clients/claude-code/), [Cherry Studio](/guide/clients/cherry-studio/), [Chatbox](/guide/clients/chatbox/), [LobeChat](/guide/clients/lobechat/).
- Your own program: [OpenAI-compatible calls](/guide/clients/openai-compatible/) and [SDKs and LangChain](/guide/clients/sdk/).
