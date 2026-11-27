# CC Switch Step-by-Step Guide (Recommended for Beginners)

If you do not want to hunt for hidden folders or edit `config.toml` / `settings.json` by hand, **CC Switch** gives you a clean desktop GUI. Fill in the form, click **Enable**, and CC Switch writes the underlying configuration files for **Codex (Desktop App, VS Code Extension & CLI)** and **Claude Code** automatically.

---

## 1. Before you start (2 minutes)

1. **Create an API key**
   - Open [API Keys]({{CONSOLE_URL}}/keys) in the console and click **Create Key**.
   - Name it `cc-switch`, select a group that covers your target models (such as `default` or `codex`), and save.
   - Click **Copy** next to the key to copy the full `sk-...` string.
2. **Copy an exact model ID**
   - Open [Model Square]({{CONSOLE_URL}}/pricing) and copy the exact model ID you want to run (for example `gpt-6.2-astra` or `gpt-6.2-sol` for Codex, and `claude-opus-5-5` or `claude-sonnet-5-5` for Claude Code).
3. **Download & install CC Switch**
   - Download the installer for your OS from `https://ccswitch.io` (or the `farion1231/cc-switch` GitHub Releases page).
   - **macOS**: Drag CC Switch from the `.dmg` into **Applications**. If macOS shows an unverified developer prompt, open **System Settings -> Privacy & Security**, scroll down, and click **Open Anyway**.
   - **Windows**: Run the `.msi` or `.exe` installer and launch **CC Switch**.

<div class="callout warning"><strong>Do not mix up the two URLs</strong><p>For <strong>Codex</strong>, the Base URL <strong>must end with <code>/v1</code></strong> (<code>{{API_BASE_URL}}</code>). For <strong>Claude Code</strong>, the Base URL <strong>must NOT include <code>/v1</code></strong> (<code>{{CONSOLE_URL}}</code>).</p></div>

---

## 2. Configure Codex (Desktop App, VS Code & CLI) in 5 Steps

The Codex Desktop App, VS Code extension, and CLI all read the same local `.codex` profile — configuring it once in CC Switch configures all three:

<div class="steps">
  <div class="step"><strong>Step 1: Select the Codex tab</strong><p>Open CC Switch and click <strong><code>Codex</code></strong> in the top or left application switcher.</p></div>
  <div class="step"><strong>Step 2: Click "+" (Add Provider)</strong><p>Click the <strong><code>+</code></strong> button in the top-right corner, and choose <strong>Custom</strong> from the preset selector.</p></div>
  <div class="step"><strong>Step 3: Fill in the provider fields</strong><p>Enter each field using the table below (avoid trailing spaces or extra slashes).</p></div>
  <div class="step"><strong>Step 4: Save & click "Enable"</strong><p>Click <strong>Save / Add</strong>, find the new <code>OpenBridger</code> card in your provider list, and click <strong>Enable</strong> so the card shows the active badge.</p></div>
  <div class="step"><strong>Step 5: Restart Codex completely</strong><p>Fully quit and reopen the <strong>Codex Desktop App</strong> (<code>Cmd + Q</code> on macOS, or exit from system tray on Windows). In VS Code, press <code>Cmd/Ctrl + Shift + P</code> and run <code>Developer: Reload Window</code>. Start a new thread and send a short message.</p></div>
</div>

### Codex field values

<div class="config-card">
  <div><span>Provider Name</span><code>OpenBridger</code></div>
  <div><span>Base URL / Endpoint</span><code>{{API_BASE_URL}}</code> (<strong>must include /v1</strong>)</div>
  <div><span>API Key</span>Your <code>sk-...</code> key copied from the console</div>
  <div><span>Wire API</span><code>responses</code></div>
  <div><span>Model</span><code>gpt-6.2-astra</code> or <code>gpt-6.2-sol</code> (exact ID from Model Square)</div>
  <div><span>Reasoning Effort (optional)</span><code>medium</code> or <code>high</code></div>
</div>

---

## 3. Configure Claude Code in 5 Steps

<div class="steps">
  <div class="step"><strong>Step 1: Select the Claude Code tab</strong><p>Switch to <strong><code>Claude Code</code></strong> in CC Switch's top or left bar.</p></div>
  <div class="step"><strong>Step 2: Add a Custom Provider</strong><p>Click <strong><code>+</code> (Add Provider)</strong> in the top-right corner and select <strong>Custom</strong>.</p></div>
  <div class="step"><strong>Step 3: Fill in URL, Key & Model</strong><p>Use the root host <code>{{CONSOLE_URL}}</code> <strong>without <code>/v1</code></strong> (see table below).</p></div>
  <div class="step"><strong>Step 4: Save & click "Enable"</strong><p>Save the form, then click <strong>Enable</strong> on the <code>OpenBridger Claude</code> card.</p></div>
  <div class="step"><strong>Step 5: Open a fresh terminal or restart the app</strong><p>Close any existing terminal tabs, open a new terminal, and run <code>claude</code> (or relaunch the Claude desktop app). No official Anthropic login is needed.</p></div>
</div>

### Claude Code field values

<div class="config-card">
  <div><span>Provider Name</span><code>OpenBridger Claude</code></div>
  <div><span>Base URL</span><code>{{CONSOLE_URL}}</code> (<strong>NO /v1 suffix!</strong>)</div>
  <div><span>API Key / Auth Token</span>Your <code>sk-...</code> key copied from the console</div>
  <div><span>Default / Mapped Model (optional)</span><code>claude-sonnet-5-5</code> or <code>claude-opus-5-5</code></div>
</div>

---

## 4. Common beginner pitfalls

| Symptom | Cause | Quick fix |
| --- | --- | --- |
| App still asks to log in with an official account | Either **Enable** wasn't clicked on the card, or the app wasn't restarted | Click **Enable** in CC Switch, then fully quit (`Cmd+Q` / tray exit) and reopen Codex or Claude |
| `404` error on send | Missing `/v1` on Codex, or extra `/v1` on Claude Code | Codex uses `{{API_BASE_URL}}`; Claude Code uses `{{CONSOLE_URL}}` |
| `503` `model_not_found` | Typo in model name or mismatched key group | Re-copy the exact model ID from [Model Square]({{CONSOLE_URL}}/pricing) |
| `403` `insufficient_user_quota` | Account wallet starts at `0` before top-up | Top up or redeem a code in [Wallet]({{CONSOLE_URL}}/wallet) |

## Related

- [Codex (Desktop, IDE & CLI)](/guide/clients/codex-cli/)
- [Claude Code (App & CLI)](/guide/clients/claude-code/)
- [Usage and logs](/guide/account/usage/)
- [Error codes](/guide/errors/)
