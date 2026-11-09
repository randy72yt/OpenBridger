# Models and pricing

Models on sale, protocols, and unit prices are on [Model Square]({{CONSOLE_URL}}/pricing). The model string in a client must match that full ID, including case, hyphens, and the version suffix.

<div class="actions"><a class="btn primary" href="{{CONSOLE_URL}}/pricing">Open Model Square</a></div>

This guide does not publish a price table. Prices change when upstream quotes and published offers change. The input, output, cache, and other units shown for a model are the basis of that call's charge.

## Four things to check

- **Protocol.** Chat apps usually need Chat Completions (`/v1/chat/completions`). Codex needs Responses (`/v1/responses`). Claude Code needs Anthropic Messages (`/v1/messages`). A name on Model Square does not mean all three work.
- **Group.** The same model ID can have different routes and prices in different groups. The group on the key decides which one you get.
- **Units.** Some models bill input and output tokens. Others also bill cache, tool calls, image count, or media duration. Read that model's card, then send one short request and check [Usage Logs]({{CONSOLE_URL}}/usage-logs).
- **Availability.** With no route, the gateway returns `503`, a message that no channel is available, and `code` `model_not_found`. That is not the same as an upstream `404`. See [Error codes](/guide/errors/).

## How to write the name

Copy it from Model Square. Do not type a display name, a vendor nickname, or an ID you remember. The gateway selects a route from the `model` field you send. It does not guess.

```bash
curl "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer ${OPENBRIDGER_API_KEY}"
```

Use this list to see which IDs the key can access. Some clients never call it, or they drop IDs they do not recognize. When that happens, paste the ID into the client by hand.

## How this connects to billing

A successful call is reserved at the model's price, forwarded, then settled from actual usage. A failed reserve never calls upstream. The unit and the formula are on [Credits and billing](/guide/billing/).

For production, plan for a model being unavailable: switch to another ID that Model Square still lists, or retry later. Do not hard-code an example ID from this guide.

## Related

- [Credits and billing](/guide/billing/)
- [Groups and routes](/guide/account/groups/)
- [OpenAI-compatible calls](/guide/clients/openai-compatible/)
