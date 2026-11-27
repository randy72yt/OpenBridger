# Error Codes

When the gateway rejects a request directly, it returns a JSON payload containing a `request id` (matching the `X-Oneapi-Request-Id` response header):

```json
{
  "error": {
    "message": "Error description (request id: 2026...)",
    "type": "new_api_error",
    "code": "error_code"
  }
}
```

## 1. Authentication & Permission Errors (401 / 403)

| HTTP | `code` | Typical Message | How to Fix |
| --- | --- | --- | --- |
| `401` | empty | Invalid token | Wrong/disabled/expired key, or the key's own spending cap is exhausted. Check or re-copy in [API Keys]({{CONSOLE_URL}}/keys). |
| `401` | empty | Token not provided | Missing API key or missing `Authorization: Bearer sk-...` header. |
| `403` | empty | Group access denied / deprecated | Reset the key's group to default or `auto`. |
| `403` | `access_denied` | IP not in allowlist | Clear or update the key's IP allowlist. |
| `403` | empty | Account suspended | Email [support@openbridger.com](mailto:support@openbridger.com) from your registered address. |

## 2. Balance & Subscription Errors (403)

| HTTP | `code` | Typical Message | How to Fix |
| --- | --- | --- | --- |
| `403` | `insufficient_user_quota` | Insufficient user quota / pre-hold failed | Wallet balance is too low for the estimated pre-hold. Top up in [Wallet]({{CONSOLE_URL}}/wallet). |
| `403` | `insufficient_user_quota` | Insufficient subscription quota | Billing mode is set to Subscription Only or plan quota is exhausted. Switch preference in Wallet or top up. |

## 3. Model, URL & Rate-Limit Errors (400 / 404 / 429 / 503)

| HTTP | `code` | Typical Message | How to Fix |
| --- | --- | --- | --- |
| `400` | Bad request | Invalid JSON, `max_tokens` out of range, or image sent to text-only model | Check request payload or switch to a vision-capable model. |
| `404` | Route not found | `404 page not found` | Extra `/v1/v1` or missing `/v1` in Base URL — check [Base URL Reference](/guide/base-urls/). |
| `429` | Rate limit | Request rate limit reached | Lower concurrency and retry with exponential backoff. |
| `503` | `model_not_found` | No available channel for model in group | Copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) and verify key group. |
| `501` | `api_not_implemented` | `API not implemented` | Endpoint not supported (such as `/v1/fine-tunes` or `/v1/files`). |
| `500` | Internal / upstream error | Server error | Retry shortly or switch models; share the `request id` with support if persistent. |

## Related

- [Troubleshooting](/troubleshooting/)
- [Base URL reference](/guide/base-urls/)
- [Contact and incidents](/guide/support/)
