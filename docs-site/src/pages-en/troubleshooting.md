# Troubleshooting

Check these items in order: Base URL, API key, exact model ID, group access, balance, and Usage Logs.

| Error | Common cause | Action |
| --- | --- | --- |
| 401 | Invalid, disabled, or malformed key | Recopy the key or create a test key |
| 403 | Group, model, or route access mismatch | Check Model Square and account group |
| 404 / model not found | Wrong model name | Copy the complete model ID |
| 429 | Rate, concurrency, or upstream limit | Reduce concurrency and retry later |
| Timeout | Network, slow upstream, or long context | Test a short prompt and reduce output |
| High usage | Repeated context or agent tools | Clear history and inspect logs |

```bash
curl {{API_BASE_URL}}/models -H "Authorization: Bearer YOUR_API_KEY"
```

A successful model-list response confirms basic network and key access. If the client still fails, check its URL format, model name, and cache.
