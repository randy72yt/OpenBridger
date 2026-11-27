# Gemini CLI

Google's open-source `@google/gemini-cli` supports custom gateway endpoints via environment variables, routing directly to OpenBridger's native `/v1beta` Gemini interface.

## Before you start

- Node.js 20+ installed on your machine.
- A dedicated API key named `gemini-cli` created under [API Keys]({{CONSOLE_URL}}/keys).
- A Gemini model ID available in [Model Square]({{CONSOLE_URL}}/pricing) (e.g. `gemini-4-neon`, `gemini-4-carbon`, or `gemini-4-hydrogen`).

## 1. Install Gemini CLI

```bash
npm install -g @google/gemini-cli
```

## 2. Export environment variables & launch

Set `GOOGLE_GEMINI_BASE_URL` to the root host `{{CONSOLE_URL}}` (**without** `/v1` or `/v1beta` — the CLI appends `/v1beta/models/...` automatically):

```bash
export GEMINI_API_KEY="sk-your-api-key"
export GOOGLE_GEMINI_BASE_URL="{{CONSOLE_URL}}"

cd /path/to/your/project
gemini -m gemini-4-neon
```

If prompted for an authentication method on first launch, choose **Use Gemini API Key**. Always pass `-m <MODEL_ID>` when you want to pin a specific model from Model Square.

## 3. One-shot non-interactive mode

```bash
gemini -m gemini-4-hydrogen -p "Summarize the key changes in git diff"
```

## Common issues

| Symptom | Fix |
| --- | --- |
| Prompted for Google OAuth sign-in | `GEMINI_API_KEY` is not exported in the current shell |
| `404` error | Remove any trailing `/v1` or `/v1beta` from `GOOGLE_GEMINI_BASE_URL`; keep `{{CONSOLE_URL}}` |
| `503` `model_not_found` | Pass `-m` with an exact Gemini model ID visible in [Model Square]({{CONSOLE_URL}}/pricing) |

## Related

- [Google Gemini Native API](/api/gemini-native/)
- [OpenAI Codex](/guide/clients/codex-cli/)
- [Claude Code](/guide/clients/claude-code/)
