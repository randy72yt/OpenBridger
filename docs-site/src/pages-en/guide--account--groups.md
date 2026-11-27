# Groups and Routes

When creating or editing a key in [API Keys]({{CONSOLE_URL}}/keys), you can assign a **Group**. Different groups expose different models, upstream routes, and rate multipliers — check [Models & Pricing]({{CONSOLE_URL}}/pricing) to compare available models and prices per group.

## Which Group Should You Pick?

| Choice | When to Use |
| --- | --- |
| **Default / Blank** | Uses your account's default group — simplest for everyday use. |
| **`auto`** | Tries your available groups in priority order; enable cross-group retry to fail over automatically. |
| **Specific Group** | Locks a key to a dedicated route tier (for example, a dedicated Codex or Claude group). |

## Common Group Errors

- **`503` No available channel for model in group**: Your key's group doesn't carry that model or protocol. Check [Models & Pricing]({{CONSOLE_URL}}/pricing) for supported groups, then switch the key's group or set it to `auto`.
- **`403` Group access denied**: That group isn't available on your current account/plan — switch the key back to your default group.

## Related

- [Create an API key](/guide/api-key/)
- [Models and pricing](/guide/models-pricing/)
- [Error codes](/guide/errors/)
