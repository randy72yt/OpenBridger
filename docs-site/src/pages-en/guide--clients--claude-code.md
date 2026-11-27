# Claude Code Setup Guide (Desktop, VS Code & CLI)

Claude Code — including the `claude` terminal CLI, the official VS Code extension, and Code mode inside the Claude desktop app — communicates over the native Anthropic Messages protocol (`POST /v1/messages`).

All three surfaces share the same `.claude` configuration directory. Below are two setup methods: **Method 1 (CC Switch GUI)** is recommended for beginners, while **Method 2** walks through creating `~/.claude/settings.json` step by step.

---

## Before You Begin (2 Quick Steps)

1. Open [API Keys]({{CONSOLE_URL}}/keys), click **Create Key**, name it `claude-code`, select a group that supports Messages routes (such as `default` or `claude`), and copy your `sk-...` key.
2. Open [Models & Pricing]({{CONSOLE_URL}}/pricing) and copy the exact model ID you want to run (for example `claude-sonnet-5-5` or `claude-opus-5-5`).

<div class="callout warning"><strong>Critical rule: Do NOT append /v1 to the Base URL</strong><p>Claude Code automatically appends <code>/v1/messages</code>. Set the Base URL to <strong><code>{{CONSOLE_URL}}</code></strong> — NOT <code>{{API_BASE_URL}}</code> — otherwise requests hit <code>/v1/v1/messages</code> and fail with <code>404</code>.</p></div>

---

## Method 1: 1-Click GUI Setup with CC Switch (Recommended)

No hidden folders or manual JSON editing required:

1. Download and open **CC Switch** (see the [CC Switch Step-by-Step Guide](/guide/clients/cc-switch/)).
2. Switch to the **`Claude Code`** tab -> click **`+`** in the top-right corner -> select **Custom**.
3. Fill in each field:
   - **Name**: `OpenBridger Claude`
   - **Base URL**: `{{CONSOLE_URL}}` (**no `/v1`**)
   - **API Key / Auth Token**: paste your `sk-...` key
   - **Model (optional)**: paste the exact model ID from Models & Pricing (e.g., `claude-sonnet-5-5`)
4. Click **Save**, then click **Enable** on the newly created card.
5. Close any old terminal windows (or reload VS Code / restart the desktop app) and run `claude`.

---

## Method 2: Manual `.claude/settings.json` Setup (Works Across CLI, VS Code & App)

Follow steps **1 -> 2 -> 3 -> 4 -> 5** below to configure Claude Code manually:

<div class="steps">
  <div class="step">
    <strong>Step 1: Install Node.js & Claude Code</strong>
    <p><strong>• Terminal CLI</strong>: Ensure Node.js 18+ is installed (<code>node -v</code>), then run <code>npm install -g @anthropic-ai/claude-code</code>.<br>
    <strong>• VS Code Extension / Desktop App</strong>: Install the official <code>Claude Code</code> extension from the VS Code Marketplace or install the desktop client.</p>
  </div>
  <div class="step">
    <strong>Step 2: Open Your <code>.claude</code> Hidden Folder</strong>
    <p><strong>• macOS</strong>: Open Finder, press <code>Command + Shift + G</code>, type <code>~/.claude</code>, and press Enter. If it does not exist yet, navigate to <code>~/</code>, press <code>Command + Shift + .</code> to show hidden files, and create a folder named <code>.claude</code>.<br>
    <strong>• Windows</strong>: Open File Explorer, type <code>%USERPROFILE%</code> in the address bar, press Enter (e.g., <code>C:\Users\YourName</code>), create a folder named <code>.claude</code>, and open it.</p>
  </div>
  <div class="step">
    <strong>Step 3: Create or Edit <code>settings.json</code></strong>
    <p>Inside <code>.claude</code>, create a plain-text file named <strong><code>settings.json</code></strong> (on Windows, enable "File name extensions" under View so it is not saved as <code>settings.json.txt</code>). Paste the JSON template below, replace the placeholder key and model ID, and save.</p>
  </div>
  <div class="step">
    <strong>Step 4: Bypass Initial Onboarding (If Prompted)</strong>
    <p>On a fresh install, if Claude Code still pauses on the first-run login screen, add <code>{"hasCompletedOnboarding": true}</code> to <code>~/.claude.json</code> (in your home directory), or toggle the profile once via CC Switch.</p>
  </div>
  <div class="step">
    <strong>Step 5: Restart Your Terminal / Reload VS Code & Verify</strong>
    <p>Close old terminal windows, open a fresh terminal inside your project folder, and run <code>claude</code> (in VS Code, press <code>Cmd/Ctrl + Shift + P</code> -> <code>Developer: Reload Window</code>). Send a short prompt and confirm the request appears in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p>
  </div>
</div>

### Step 3 `settings.json` Template

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{CONSOLE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-replace-with-your-openbridger-key",
    "ANTHROPIC_MODEL": "claude-sonnet-5-5",
    "API_TIMEOUT_MS": "300000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

<div class="config-card">
  <div><span>Config Path (macOS / Linux)</span><code>~/.claude/settings.json</code></div>
  <div><span>Config Path (Windows)</span><code>%USERPROFILE%\.claude\settings.json</code></div>
  <div><span>ANTHROPIC_BASE_URL</span><code>{{CONSOLE_URL}}</code> (<strong>NO /v1</strong>)</div>
  <div><span>ANTHROPIC_AUTH_TOKEN</span><code>sk-...</code> (do not prepend <code>Bearer </code>)</div>
  <div><span>ANTHROPIC_MODEL</span>Exact model ID from Models & Pricing (e.g., <code>claude-sonnet-5-5</code>)</div>
</div>

> **Security reminder**: Never commit `~/.claude/settings.json` or project-level `.claude/settings.json` containing real API keys to Git.

---

## Optional: Temporary One-Session Terminal Export

If you only want to test in your current terminal window without saving a config file:

- **macOS / Linux (zsh / bash)**:
  ```bash
  export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
  export ANTHROPIC_AUTH_TOKEN="sk-your-key"
  export ANTHROPIC_MODEL="claude-sonnet-5-5"
  claude
  ```
- **Windows PowerShell**:
  ```powershell
  $env:ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
  $env:ANTHROPIC_AUTH_TOKEN="sk-your-key"
  $env:ANTHROPIC_MODEL="claude-sonnet-5-5"
  claude
  ```

---

## Common Issues & Fixes

| Symptom | Cause | Fix |
| --- | --- | --- |
| Still asks to `/login` to an Anthropic account | Old terminal session or `settings.json.txt` filename / invalid JSON syntax | Open a new terminal, check `env \| grep ANTHROPIC_`, or apply via [CC Switch](/guide/clients/cc-switch/) |
| Red `offline` indicator in top-right corner | Non-essential connectivity probe to upstream telemetry hosts | Safe to ignore if prompts stream replies normally; keep `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1"` |
| `404 Not Found` | `/v1` was appended to `ANTHROPIC_BASE_URL` | Change `ANTHROPIC_BASE_URL` to `{{CONSOLE_URL}}` and restart `claude` |
| `401 Unauthorized` | Extra `Bearer ` prefix or whitespace in `ANTHROPIC_AUTH_TOKEN` | Paste only the raw `sk-...` key |
| `503` no available channel / `model not found` | Model name mismatch or key group lacks `/v1/messages` routes | Copy the exact ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) or run `/model <exact-id>` |

## Related

- [CC Switch (1-Click GUI)](/guide/clients/cc-switch/)
- [Anthropic Messages API](/api/anthropic-messages/)
- [Error codes](/guide/errors/)
