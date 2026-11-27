# Gemini CLI Setup Guide

Google's open-source `gemini-cli` connects over the native `/v1beta` protocol for long-context repository analysis and terminal tasks.

## Before You Begin

1. Install Node.js 20+ and run `npm install -g @google/gemini-cli`.
2. Create a key named `gemini-cli` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
3. Copy an exact Gemini Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gemini-4-neon`, `gemini-4-carbon`, `gemini-4-hydrogen`).

<div class="callout warning"><strong>Do NOT append /v1 or /v1beta</strong><p>Set <code>GOOGLE_GEMINI_BASE_URL</code> to <strong><code>{{CONSOLE_URL}}</code></strong> — the CLI appends <code>/v1beta/...</code> automatically.</p></div>

---

## Method 1: 1-Click GUI via CC Switch

Open **CC Switch** ([CC Switch Guide](/guide/clients/cc-switch/)) -> switch to the **`Gemini`** tab -> click **`+`** -> enter `Base URL` = `{{CONSOLE_URL}}` and `API Key` = `sk-...` -> Save and click **Enable**, then run `gemini -m gemini-4-neon`.

---

## Method 2: Persistent `~/.gemini/.env` or Terminal Export

### 1. Persistent Config (`~/.gemini/.env`)

Create `.env` inside `~/.gemini/` (`%USERPROFILE%\.gemini\.env` on Windows):

```ini
GEMINI_API_KEY=sk-replace-with-your-key
GOOGLE_GEMINI_BASE_URL={{CONSOLE_URL}}
GEMINI_MODEL=gemini-4-neon
```

Then run `gemini` in any project directory (if asked for an auth method on first launch, choose **Use Gemini API Key**).

### 2. One-Session Terminal Export

- **macOS / Linux**:
  ```bash
  export GEMINI_API_KEY="sk-your-key"
  export GOOGLE_GEMINI_BASE_URL="{{CONSOLE_URL}}"
  gemini -m gemini-4-neon
  ```
- **Windows PowerShell**:
  ```powershell
  $env:GEMINI_API_KEY="sk-your-key"
  $env:GOOGLE_GEMINI_BASE_URL="{{CONSOLE_URL}}"
  gemini -m gemini-4-neon
  ```

## Common Issues

| Symptom | Fix |
| --- | --- |
| Opens browser asking for Google login | Select **Use Gemini API Key** and verify `GEMINI_API_KEY` is set |
| `404` error | Change `GOOGLE_GEMINI_BASE_URL` to `{{CONSOLE_URL}}` (no `/v1` or `/v1beta`) |
| `503` No available channel | Pass `-m <exact-model-id>` from [Models & Pricing]({{CONSOLE_URL}}/pricing) |

## Related

- [CC Switch (1-Click GUI)](/guide/clients/cc-switch/)
- [Google Gemini Native API](/api/gemini-native/)
- [Error codes](/guide/errors/)
