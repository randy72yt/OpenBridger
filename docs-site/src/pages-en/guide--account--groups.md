# Groups and routes

A group decides which models this key can call, which kind of route it uses, and which ratio is billed. You do not need the upstream channel's name. Pick model IDs that Model Square shows for the current group.

## The group on a key

| Value | Result |
| --- | --- |
| Empty | Uses the account group. |
| A named group | That group must be one you can use. Otherwise `403`. A retired group is also `403`. |
| `auto` | Tries groups in the order you saved. With cross-group retry, a failure in the current group continues with the next one. |

Names and ratios are operator settings. This guide does not define product names such as Standard or Premium. The names on Model Square and on the key form are the real ones.

## While you use it

- A client's model list should come from what this key can access. `GET /v1/models` is that set.
- The same model ID can differ in price and stability across groups. Changing the group changes the price. Check Model Square first.
- A `503` that says no channel is available is a model or group mismatch. Do not rotate keys to fix it.
- In a team, write down the group and model ID each project uses.

## Related

- [Create an API key](/guide/api-key/)
- [Models and pricing](/guide/models-pricing/)
- [Error codes](/guide/errors/)
