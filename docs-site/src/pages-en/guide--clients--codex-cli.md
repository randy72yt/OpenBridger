# Use OpenBridger in Codex

Codex calls a custom model provider with the Responses protocol. The gateway path is `POST /v1/responses`. Keep the key in an environment variable. The config file only names that variable.

## Before you start

- The Codex CLI is installed.
- An API key named `codex`.
- A full model ID from Model Square. That model needs a route on the Responses path, or the call returns `503` and no channel.

## Configure

```bash
export OPENBRIDGER_API_KEY="sk-your-key"
```

`~/.codex/config.toml`:

```toml
model = "MODEL_ID_FROM_CONSOLE"
model_provider = "openbridger"

[model_providers.openbridger]
name = "OpenBridger"
base_url = "{{API_BASE_URL}}"
env_key = "OPENBRIDGER_API_KEY"
wire_api = "responses"
```

`base_url` includes `/v1`. `wire_api` must be `responses`, so the client calls `/v1/responses` instead of Chat Completions.

Do not write the `sk-` key into `config.toml`.

## Start

```bash
cd /path/to/your/project
codex
```

Match the model and the key name `codex` in [Usage Logs]({{CONSOLE_URL}}/usage-logs).

To use the official provider again, set `model_provider` back or delete this provider, then restart Codex. You do not need to sign out of the ChatGPT account.

## Check

- The model shown in the UI matches `model` in `config.toml`.
- Use a check that does not print the secret, such as testing that the variable is non-empty, instead of echoing it onto the screen.
- A personal shell profile is a fine place for the variable. The project repository is not.

## Common problems

| What you see | What to do |
| --- | --- |
| 401 | The environment variable is empty, or the key is disabled |
| 503 no channel | The model ID is not the full ID from Model Square, or this group has no Responses route |
| The request uses another provider | `model_provider` is not `openbridger`, or an old process is still running |

## Related

- [CC Switch](/guide/clients/cc-switch/)
- [OpenAI-compatible calls](/guide/clients/openai-compatible/)
