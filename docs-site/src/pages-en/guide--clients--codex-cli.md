# Use OpenBridger with Codex

Codex can connect through a custom model provider. Keep the key in an environment variable.

```bash
export OPENBRIDGER_API_KEY="YOUR_API_KEY"
```

Edit `~/.codex/config.toml`:

```toml
model = "MODEL_ID_FROM_CONSOLE"
model_provider = "openbridger"

[model_providers.openbridger]
name = "OpenBridger"
base_url = "{{API_BASE_URL}}"
env_key = "OPENBRIDGER_API_KEY"
wire_api = "responses"
```

The exact model and route must support the Responses protocol. Start Codex, then check Usage Logs. For a 401, verify the environment variable; for `model not found`, copy the full model ID again.

<div class="callout warning"><strong>Keep the key out of your project</strong><p>Do not put a live key in config.toml or a repository.</p></div>
