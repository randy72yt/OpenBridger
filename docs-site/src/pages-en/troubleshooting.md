# Troubleshooting

Separate the gateway from the client with curl. If curl works and the client does not, the client's URL or cache is wrong. If curl fails too, check the key, the credit, and the model ID.

## Check

```bash
curl -i "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer ${OPENBRIDGER_API_KEY}"
```

- HTTP 200 and a model list: the key and the URL work.
- 401: the key itself is unusable. See [Error codes](/guide/errors/).
- 403 that mentions quota: open the [wallet]({{CONSOLE_URL}}/wallet).
- Cannot connect: the host is `openbridger.com` or `api.openbridger.com`. The path must not be `/v1/v1/models`.

Then send the short chat from [Quick start](/guide/quickstart/) and match the request id in [Usage Logs]({{CONSOLE_URL}}/usage-logs).

## Table

| What you see | Usual cause | What to do |
| --- | --- | --- |
| 401 invalid token | Wrong, disabled, or expired key, key credit used up, or `Bearer` typed twice | Copy the key again and check its remaining credit and expiry |
| 403 insufficient quota | Wallet or subscription cannot cover the reserve | Add credit, or shorten the input |
| 403 group or IP | The key's group is not usable, or the allowlist excludes this egress IP | Change the group or the allowlist |
| 503 no channel | The model ID is not in this group | Copy the full ID from Model Square |
| Upstream 404 / model not found | Upstream does not know that name, or the client rewrote the ID | Try another ID from Model Square and read the name stored in the log |
| 429 | Local rate limit or an upstream limit | If the message says you reached the request limit, wait for that window. Otherwise it is upstream. Retry later |
| Timeout | A long context or a slow upstream | Send one short prompt and a smaller output, then compare |
| Stream cut off | The network, or the client's stream handling | Send the same model with `"stream": false` |
| Charge looks high | History was resent, or tools ran many times | Read tokens in the log, not only the chat bubbles |
| An old URL is still used | Another env var or config file is still set | Open a new terminal and inspect `OPENAI_*` and `ANTHROPIC_*` |
| Model list will not load | The client does not call `/v1/models`, or it filters IDs | Enter the model ID by hand |

## Which URL to type

| Client | Value |
| --- | --- |
| OpenAI SDK, Chatbox, Cherry Studio, LobeChat, Cursor's OpenAI Base URL, Open WebUI | `{{API_BASE_URL}}` |
| A single field that asks for the full endpoint | `{{API_BASE_URL}}/chat/completions` |
| Claude Code, Anthropic SDK | `{{CONSOLE_URL}}` (no `/v1`) |
| Codex | Base URL `{{API_BASE_URL}}`, protocol Responses |

## Related

- [Error codes](/guide/errors/)
- [Contact and incidents](/guide/support/)
