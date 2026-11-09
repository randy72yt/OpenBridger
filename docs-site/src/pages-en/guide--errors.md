# Error codes

When the gateway itself rejects a call, the body looks like this:

```json
{
  "error": {
    "message": "explanation (request id: REQUEST_ID)",
    "type": "new_api_error",
    "code": "error code, or an empty string"
  }
}
```

The `X-Oneapi-Request-Id` response header and the request id in the body are the same value. Send that id to support. Do not send the full API key.

An upstream error can keep the upstream `type` and HTTP status. It is not always `new_api_error`. This table is only what callers hit most often from the gateway itself.

## Authentication and permission

| HTTP | code | Meaning | What to do |
| --- | --- | --- | --- |
| 401 | empty | Invalid token | The key is wrong, disabled, expired, or that key's own credit is used up. A funded wallet does not change this. Copy the `sk-` key again, or raise that key's credit. |
| 401 | empty | Token was not provided | There is no `Authorization` header. It should be `Bearer` plus the key. |
| 403 | empty | No access to this group, or the group is retired | Change the key group back to the account group or `auto`. |
| 403 | `access_denied` | The IP is not on the key allowlist, or the client IP cannot be parsed | Edit or clear the allowlist. |
| 403 | empty | The account is banned | Email support from the registration address. Do not keep retrying the same key. |
| 403 | empty | A normal user cannot pin a channel | Do not append `-channelId` to the key. |

An Anthropic client may send the key in `x-api-key`. The gateway accepts that only on `/v1/messages` and `/v1/models`. Other paths need `Authorization: Bearer`.

## Credit and billing

| HTTP | code | Message | What to do |
| --- | --- | --- | --- |
| 403 | `insufficient_user_quota` | User quota is insufficient, or the pre-charge failed | The wallet cannot cover this reserve. Top up, redeem a code, or send a shorter input. |
| 403 | `insufficient_user_quota` | Subscription credit is insufficient or no subscription is configured | The preference is Subscription Only, or the subscription balance is too low. Switch to Wallet First or add wallet credit. See [Subscriptions](/guide/account/subscriptions/). |

Insufficient credit is **403**, not 429.

## Model and request

| HTTP | code | Message | What to do |
| --- | --- | --- | --- |
| 400 | empty or a request error | The model name is empty, or the body is not valid JSON | Check the JSON and the `model` field. |
| 503 | `model_not_found` | No channel is available for this group and model | Copy the ID from Model Square again and check the key's group. This is not a wrong-URL 404. |
| 501 | `api_not_implemented` | `API not implemented` | That path is not implemented. Image variations, files, and fine-tunes are in this set. |
| 429 | empty | You reached the request limit, including failed tries when the message says so | This appears only when the application rate limit is enabled. The window and the count are in the message. Lower concurrency and wait. |
| 500 | empty | An internal failure such as a database error | Keep the request id and contact support. Do not retry the same key in a tight loop. |

An upstream limit can be the upstream's own `429` and body. That is not the local rate-limit rule above. Retry later, or pick another model that Model Square still lists.

## Looks like a gateway error, and is not

| What you see | Actual cause |
| --- | --- |
| The path is `/v1/v1/chat/completions` | The Base URL already contains `/v1`, and the client added it again. OpenAI-compatible clients use `{{API_BASE_URL}}`. |
| Claude Code calls the wrong host | `ANTHROPIC_BASE_URL` must be `{{CONSOLE_URL}}`, without `/v1`. The client requests `/v1/messages` itself. |
| The model list is empty, but curl `/v1/models` returns data | The client filtered unknown IDs. Paste the full ID from Model Square. |
| The browser reports CORS, or the key leaks | Do not put the key in a web page. Call from your own server. |

## Related

- [Troubleshooting](/troubleshooting/)
- [Create an API key](/guide/api-key/)
- [Credits and billing](/guide/billing/)
