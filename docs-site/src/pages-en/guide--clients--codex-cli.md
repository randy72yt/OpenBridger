# Codex Setup Guide (Desktop App, VS Code & CLI)

The **Codex Desktop App**, **VS Code extension**, and **`codex` CLI** all read from the same local `.codex` folder — configure it once and all three work together.

## Before You Begin

1. Create a key named `codex` in [API Keys]({{CONSOLE_URL}}/keys) (pick a Responses-capable group such as `default` or `codex`) and copy `sk-...`.
2. Copy an exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gpt-6.2-astra` or `gpt-6.2-sol`).

---

## Method 1: 1-Click GUI via CC Switch (Recommended)

1. Open **CC Switch** (see [CC Switch Guide](/guide/clients/cc-switch/)).
2. Click **`Codex`** -> click **`+`** -> choose **Custom**.
3. Enter `Base URL` = `{{API_BASE_URL}}`, `API Key` = `sk-...`, `Wire API` = `responses`, `Model` = `gpt-6.2-astra`.
4. Click **Save** -> click **Enable** -> fully quit and relaunch Codex Desktop (or reload VS Code).

---

## Method 2: Manual `.codex` Config Files

<div class="callout warning"><strong>Desktop App note</strong><p>Launching the GUI Desktop App does not load shell exports from <code>~/.zshrc</code>. Saving your key in <code>auth.json</code> alongside <code>config.toml</code> works across Desktop, VS Code, and CLI.</p></div>

<div class="steps">
  <div class="step"><strong>Step 1: Install Codex</strong><p>• <strong>Desktop App</strong>: Install the official Codex app (close the initial ChatGPT login prompt without signing in).<br>• <strong>VS Code Extension</strong>: Install <code>OpenAI Codex</code> (<code>openai.chatgpt</code>).<br>• <strong>CLI</strong>: Requires Node.js 22+; run <code>npm install -g @openai/codex</code>.</p></div>
  <div class="step"><strong>Step 2: Open Your <code>.codex</code> Folder</strong><p>• <strong>macOS</strong>: In Finder, press <code>Cmd + Shift + G</code>, enter <code>~/</code>, press <code>Cmd + Shift + .</code> to show hidden files, and create/open <code>.codex</code>.<br>• <strong>Windows</strong>: In File Explorer, enter <code>%USERPROFILE%</code> in the address bar, check <strong>File name extensions</strong> under View, and create/open <code>.codex</code>.</p></div>
  <div class="step"><strong>Step 3: Create <code>auth.json</code></strong><p>Inside <code>.codex</code>, create <strong><code>auth.json</code></strong> (not <code>.json.txt</code>) with the JSON below.</p></div>
  <div class="step"><strong>Step 4: Create <code>config.toml</code></strong><p>Inside <code>.codex</code>, create <strong><code>config.toml</code></strong> with the TOML config below.</p></div>
  <div class="step"><strong>Step 5: Restart & Verify</strong><p>Quit and relaunch Codex Desktop (<code>Cmd + Q</code> / tray exit), run <code>Developer: Reload Window</code> in VS Code, or run <code>codex</code> in terminal. Start a new thread and check <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div>
</div>

### `auth.json`

```json
{
  "OPENAI_API_KEY": "sk-replace-with-your-key"
}
```

### `config.toml`

```toml
model_provider = "openbridger"
model = "gpt-6.2-astra"
model_reasoning_effort = "high"
disable_response_storage = true
preferred_auth_method = "apikey"

[model_providers.openbridger]
name = "OpenBridger"
base_url = "{{API_BASE_URL}}"
wire_api = "responses"
```

> Set `model_reasoning_effort` to `"medium"` for faster everyday edits or `"high"` for complex debugging. Never commit `auth.json` to Git.

## Common Issues

| Symptom | Fix |
| --- | --- |
| Desktop App still opens browser login | Verify `preferred_auth_method = "apikey"` in `config.toml`, confirm no `.txt` extension, and quit/relaunch |
| Custom model isn't in the UI picker after an app update | Edit `model` in `config.toml` (or CC Switch), restart, and open a **New Thread** instead of picking bundled UI presets |
| `401` / `503` | `401`: re-copy `sk-...`; `503`: verify exact Model ID and key group in [Models & Pricing]({{CONSOLE_URL}}/pricing) |

## Related

- [CC Switch (1-Click GUI)](/guide/clients/cc-switch/)
- [Responses API](/api/openai-responses/)
- [Error codes](/guide/errors/)
