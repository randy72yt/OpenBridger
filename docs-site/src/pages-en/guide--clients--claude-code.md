# Claude Code Setup Guide (Desktop, VS Code & CLI)

Claude Code (`claude` CLI, VS Code extension, and Desktop Code mode) uses the native Anthropic `/v1/messages` protocol and reads from `.claude`.

## Before You Begin

1. Create a key named `claude-code` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
2. Copy an exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `claude-sonnet-5-5` or `claude-opus-5-5`).

<div class="callout warning"><strong>Never append /v1 to Base URL</strong><p>Claude Code appends <code>/v1/messages</code> automatically — enter <strong><code>{{CONSOLE_URL}}</code></strong> (adding <code>/v1</code> causes <code>404</code>).</p></div>

---

## Method 1: 1-Click GUI via CC Switch (Recommended)

1. Open **CC Switch** ([CC Switch Guide](/guide/clients/cc-switch/)), select **`Claude Code`** -> click **`+`** -> choose **Custom**.
2. Enter `Base URL` = `{{CONSOLE_URL}}` (no `/v1`), `API Key` = `sk-...`, `Model` = `claude-sonnet-5-5`.
3. Click **Save** -> click **Enable** -> open a fresh terminal and run `claude` (or reload VS Code).

---

## Method 2: Manual `.claude/settings.json`

<div class="steps">
  <div class="step"><strong>Step 1: Install Claude Code</strong><p>• <strong>CLI</strong>: Requires Node.js 18+; run <code>npm install -g @anthropic-ai/claude-code</code>.<br>• <strong>VS Code / Desktop</strong>: Install the official <code>Claude Code</code> extension or desktop app.</p></div>
  <div class="step"><strong>Step 2: Open Your <code>.claude</code> Folder</strong><p>• <strong>macOS</strong>: Finder -> <code>Cmd + Shift + G</code> -> <code>~/</code> -> <code>Cmd + Shift + .</code> -> create/open <code>.claude</code>.<br>• <strong>Windows</strong>: File Explorer -> <code>%USERPROFILE%</code> -> enable File name extensions -> create/open <code>.claude</code>.</p></div>
  <div class="step"><strong>Step 3: Save <code>settings.json</code></strong><p>Create <strong><code>settings.json</code></strong> inside <code>.claude</code> using the template below.</p></div>
  <div class="step"><strong>Step 4: Skip Initial Onboarding (If Prompted)</strong><p>If a fresh install pauses on first-run login, add <code>{"hasCompletedOnboarding": true}</code> to <code>~/.claude.json</code> (or click Enable once in CC Switch).</p></div>
  <div class="step"><strong>Step 5: Restart & Verify</strong><p>Open a new terminal in your project folder and run <code>claude</code> (or <code>Developer: Reload Window</code> in VS Code), then confirm the call in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div>
</div>

### `settings.json` Template

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{CONSOLE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-replace-with-your-key",
    "ANTHROPIC_MODEL": "claude-sonnet-5-5",
    "API_TIMEOUT_MS": "300000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

### Optional: One-Session Terminal Command

- **macOS / Linux**: `export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}" ANTHROPIC_AUTH_TOKEN="sk-key" ANTHROPIC_MODEL="claude-sonnet-5-5" && claude`
- **Windows PowerShell**: `$env:ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"; $env:ANTHROPIC_AUTH_TOKEN="sk-key"; $env:ANTHROPIC_MODEL="claude-sonnet-5-5"; claude`

## Common Issues

| Symptom | Fix |
| --- | --- |
| Still asks for `/login` | Open a new terminal, check filename isn't `settings.json.txt`, or apply via [CC Switch](/guide/clients/cc-switch/) |
| Red `offline` badge in top-right | Telemetry probe indicator — safe to ignore if chat replies stream normally |
| `404` / `503` | `404`: remove `/v1` from Base URL; `503`: copy exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) |

## Related

- [CC Switch (1-Click GUI)](/guide/clients/cc-switch/)
- [Anthropic Messages API](/api/anthropic-messages/)
- [Error codes](/guide/errors/)
