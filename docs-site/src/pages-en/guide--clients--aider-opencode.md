# OpenCode & Aider

OpenCode and Aider are lightweight terminal coding agents that work well with OpenBridger's unified endpoint, letting you mix frontier planners (such as `gpt-6.2-astra` or `claude-opus-5-5`) with fast editor models (such as `deepseek-v4.1-flash`).

## Before you start

- Create a dedicated CLI key under [API Keys]({{CONSOLE_URL}}/keys).
- Copy the exact model IDs you plan to use from [Model Square]({{CONSOLE_URL}}/pricing).

---

## 1. OpenCode setup

1. Install OpenCode:

```bash
npm install -g opencode-ai
```

2. Edit `~/.config/opencode/opencode.json` to register OpenBridger as an OpenAI-compatible provider reading your key from an environment variable:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "openbridger": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "OpenBridger",
      "options": {
        "baseURL": "{{API_BASE_URL}}",
        "apiKey": "{env:OPENBRIDGER_API_KEY}"
      },
      "models": {
        "gpt-6.2-astra": {
          "name": "GPT-6.2 Astra"
        },
        "claude-sonnet-5-5": {
          "name": "Claude Sonnet 5.5"
        },
        "deepseek-v4.1": {
          "name": "DeepSeek V4.1"
        }
      }
    }
  }
}
```

3. Export your key and start `opencode`:

```bash
export OPENBRIDGER_API_KEY="sk-your-api-key"
opencode
```

---

## 2. Aider setup

Prefix any model ID with `openai/` so Aider (LiteLLM) sends requests through the OpenAI-compatible wire format to `{{API_BASE_URL}}`:

```bash
pipx install aider-chat

export OPENAI_API_BASE="{{API_BASE_URL}}"
export OPENAI_API_KEY="sk-your-api-key"

aider --model openai/claude-sonnet-5-5
```

For Architect mode (pairing a deep reasoning architect model with a fast editor model):

```bash
aider --architect \
  --model openai/gpt-6.2-astra \
  --editor-model openai/deepseek-v4.1-flash
```

## Common issues

| Symptom | Fix |
| --- | --- |
| Aider asks for `ANTHROPIC_API_KEY` | Add the `openai/` prefix (e.g. `openai/claude-sonnet-5-5`) so LiteLLM uses `OPENAI_API_BASE` + `OPENAI_API_KEY` |
| `503` `model_not_found` | Verify the model ID after `openai/` matches [Model Square]({{CONSOLE_URL}}/pricing) |

## Related

- [OpenAI Codex](/guide/clients/codex-cli/)
- [Claude Code](/guide/clients/claude-code/)
- [Base URL reference](/guide/base-urls/)
