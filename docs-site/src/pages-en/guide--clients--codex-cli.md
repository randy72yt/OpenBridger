# OpenAI Codex Setup Guide (Desktop App, VS Code Extension & CLI)

OpenAI Codex comes in three forms: the **Codex Desktop App (macOS / Windows)**, the **VS Code / Cursor Codex Extension**, and the **terminal `codex` CLI**.

All three share the **exact same local `.codex` configuration folder**. Once you configure it, the Desktop App, VS Code extension, and CLI all connect to OpenBridger automatically. Choose **Method 1 (CC Switch GUI — easiest for beginners)** or **Method 2 (Manual 5-Step Config Files)**.

---

## Before you start (2 quick checks)

1. Open [API Keys]({{CONSOLE_URL}}/keys), create a key named `codex` under a group that supports the Responses protocol (such as `default` or `codex`), and copy the `sk-...` key.
2. Open [Model Square]({{CONSOLE_URL}}/pricing) and copy the exact model ID you want to use (for example `gpt-6.2-astra` or `gpt-6.2-sol`).

---

## Method 1: 1-Minute GUI Setup with CC Switch (Recommended)

1. Install and launch **CC Switch** (full walkthrough on [CC Switch (1-Click GUI)](/guide/clients/cc-switch/)).
2. Select the **`Codex`** tab -> click **`+` (Add Provider)** -> choose **Custom**.
3. Fill in:
   - **Name**: `OpenBridger`
   - **Base URL**: `{{API_BASE_URL}}` (includes `/v1`)
   - **API Key**: paste your `sk-...` key
   - **Wire API**: `responses`
   - **Model**: `gpt-6.2-astra` (or any Responses-capable model ID from Model Square)
4. Click **Save**, then click **Enable** on the `OpenBridger` card.
5. **Quit and relaunch the Codex Desktop App** (or reload VS Code) and start a new thread!

---

## Method 2: Manual 5-Step Setup (`auth.json` + `config.toml`)

<div class="callout warning"><strong>Why use <code>auth.json</code> for the Desktop App?</strong><p>Launching the <strong>Codex Desktop App</strong> from macOS Dock/Spotlight or the Windows Start Menu does not load shell variables from <code>~/.zshrc</code>. Placing your key in <code>~/.codex/auth.json</code> works reliably across the Desktop App, VS Code extension, and CLI simultaneously.</p></div>

<div class="steps">
  <div class="step">
    <strong>Step 1: Install Codex (any or all of the three)</strong>
    <p><strong>• Desktop App</strong>: Install the OpenAI Codex app for macOS or Windows (if it prompts you to sign in with ChatGPT on first launch, simply close the app first).<br>
    <strong>• VS Code Extension</strong>: Search for <code>OpenAI Codex</code> (<code>openai.chatgpt</code>) in the VS Code Extensions Marketplace and click Install.<br>
    <strong>• CLI</strong>: With Node.js 22+ installed, run <code>npm install -g @openai/codex</code>.</p>
  </div>
  <div class="step">
    <strong>Step 2: Open your <code>.codex</code> folder</strong>
    <p><strong>• macOS</strong>: Open <strong>Finder</strong>, press <code>Cmd + Shift + G</code>, type <code>~/.codex</code>, and press Enter. If it does not exist yet, go to <code>~/</code>, press <code>Cmd + Shift + .</code> to show hidden folders, create a folder named <code>.codex</code>, and open it.<br>
    <strong>• Windows</strong>: Open <strong>File Explorer</strong>, click the address bar, type <code>%USERPROFILE%</code> and press Enter (e.g. <code>C:\Users\YourName</code>). Create a folder named <code>.codex</code> and open it.</p>
  </div>
  <div class="step">
    <strong>Step 3: Create <code>auth.json</code></strong>
    <p>Inside <code>.codex</code>, create a plain-text file named <strong><code>auth.json</code></strong> (on Windows, check <strong>View -> File name extensions</strong> to make sure it is not saved as <code>auth.json.txt</code>) and paste the JSON snippet below with your key.</p>
  </div>
  <div class="step">
    <strong>Step 4: Create <code>config.toml</code></strong>
    <p>In the same <code>.codex</code> folder, create a file named <strong><code>config.toml</code></strong> (not <code>config.toml.txt</code>) and paste the TOML configuration below.</p>
  </div>
  <div class="step">
    <strong>Step 5: Restart Codex & verify in Usage Logs</strong>
    <p><strong>• Desktop App</strong>: Fully quit Codex (<code>Cmd + Q</code> on macOS, or right-click tray icon -> Quit on Windows), reopen it, and start a new thread.<br>
    <strong>• VS Code Extension</strong>: Press <code>Cmd/Ctrl + Shift + P</code>, run <code>Developer: Reload Window</code>, and open a new Codex chat.<br>
    <strong>• CLI</strong>: Run <code>codex</code> inside your project directory.<br>
    Send a test prompt and confirm the call appears under [Usage Logs]({{CONSOLE_URL}}/usage-logs).</p>
  </div>
</div>

### Step 3 file: `auth.json`

```json
{
  "OPENAI_API_KEY": "sk-paste-your-openbridger-key-here"
}
```

### Step 4 file: `config.toml`

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

- `model`: exact model ID from [Model Square]({{CONSOLE_URL}}/pricing).
- `model_reasoning_effort`: `"low"`, `"medium"`, or `"high"`.
- `wire_api`: **must remain `"responses"`**.
- Keep `~/.codex/auth.json` private and never commit it to Git.

---

## Troubleshooting

| Symptom | Cause | Fix |
| --- | --- | --- |
| Desktop App still shows the ChatGPT sign-in screen | Missing `preferred_auth_method = "apikey"` or file accidentally named `config.toml.txt` | Show file extensions, verify exact filenames in `~/.codex/`, and fully quit & relaunch the app |
| Custom model does not appear in UI dropdown after an update | UI dropdown only lists built-in presets, while `config.toml` still controls the active model | Edit `model = "..."` in `config.toml` (or CC Switch), restart Codex, and start a **New Thread** |
| `401` invalid token | Extra whitespace in `auth.json` or exhausted key quota | Re-copy the `sk-...` key without `Bearer ` |
| `503` `model_not_found` | Key group does not expose Responses for that model, or typo in model ID | Copy the exact model ID from [Model Square]({{CONSOLE_URL}}/pricing) |

## Related

- [CC Switch (1-Click GUI)](/guide/clients/cc-switch/)
- [Responses API](/api/openai-responses/)
- [Error codes](/guide/errors/)
