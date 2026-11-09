# FAQ

## What do I put in Base URL?

OpenAI-compatible clients use `{{API_BASE_URL}}`. Add `/chat/completions` only when the form has a single field that asks for the full endpoint.

Claude Code and the Anthropic SDK use `{{CONSOLE_URL}}` without `/v1`. See [Claude Code](/guide/clients/claude-code/).

`https://api.openbridger.com/v1` and `{{API_BASE_URL}}` are the same gateway. Examples use the latter.

## Do I add Bearer to the key?

The header is `Authorization: Bearer sk-...`. The key already includes `sk-`. Do not send `Bearer Bearer` or `Bearer sk-sk-`.

## Why is the model list empty when the key is valid?

Many clients never call `/v1/models`, or they drop IDs they do not recognize. If curl lists models, paste the full ID from [Model Square]({{CONSOLE_URL}}/pricing).

## How do I know traffic went through OpenBridger?

Create a key used only by that client, send one short message, and find it in [Usage Logs]({{CONSOLE_URL}}/usage-logs) by key and time. The model name in the client UI is not enough.

## Do I sign out of the official account?

Usually no. The editor account is for sign-in and sync. A custom Base URL affects only requests that use this key. To go back, turn off the custom Base URL or select the built-in provider, clear the related environment variables, and restart the client.

## How do I tell 401 from 403?

`401` and an invalid-token message means the key is unusable, including when that key's own credit is used up. `403` with `code` `insufficient_user_quota` means the wallet or subscription cannot cover the reserve. Group, IP allowlist, and a banned account are also `403`, with different text. See [Error codes](/guide/errors/).

## Which clients are supported?

Any OpenAI-compatible client that can set a Base URL, an API key, and a model ID can follow [the general guide](/guide/clients/openai-compatible/). Separate steps exist for Cursor, Claude Code, Codex, CC Switch, Cherry Studio, Chatbox, LobeChat, Open WebUI, Trae, CodeBuddy, the OpenAI SDK, and LangChain.

Whether a model supports tools, images, or the protocol that client requires is decided by Model Square and one short test call.

## Which balance is charged first, wallet or plan?

The default is Subscription First. With no subscription credit, the wallet is charged. The four preferences are on [Subscriptions](/guide/account/subscriptions/). Under terms v1.1, plan purchase is not open yet.

## How can I add credit?

Use the online methods the wallet shows (Alipay and WeChat by default) or a redemption code. The minimum and the credited amount are on the confirmation page. See [Top up and redeem](/guide/account/redeem/).

## Does registration include free credit?

The program default is none (new-user credit is 0). An operator can configure a gift. After you register, look at the wallet. An extra credit row means a gift is configured. No row means there is no gift right now.

## Should I retry a failed request?

Retries do not fix `401`, a `403` for insufficient credit, or a `400` for a bad body. `429` and a temporary upstream 5xx can be retried a limited number of times with backoff. Do not loop forever.

## Related

- [Quick start](/guide/quickstart/)
- [Troubleshooting](/troubleshooting/)
- [Contact and incidents](/guide/support/)
