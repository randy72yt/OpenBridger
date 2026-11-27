# CC Switch 1-Click Setup Guide (Recommended)

**CC Switch** is an open-source GUI manager that writes the local config files for **Codex (Desktop App, VS Code & CLI)** and **Claude Code** automatically when you click **Enable**.

## Before You Begin

1. Create a key in [API Keys]({{CONSOLE_URL}}/keys) (e.g., `cc-switch`) and copy `sk-...`.
2. Copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gpt-6.2-astra` or `claude-sonnet-5-5`).
3. Download and install **CC Switch** from `https://ccswitch.io` (or GitHub `farion1231/cc-switch` Releases):
   - **macOS**: Drag to Applications; if Gatekeeper blocks first launch, go to **System Settings -> Privacy & Security** and click **Open Anyway**.
   - **Windows**: Run the `.msi` / `.exe` installer.

<div class="callout warning"><strong>Don't swap the two URL formats</strong><p><strong>Codex</strong> requires <strong><code>/v1</code></strong> (<code>{{API_BASE_URL}}</code>). <strong>Claude Code</strong> must have <strong>NO <code>/v1</code></strong> (<code>{{CONSOLE_URL}}</code>).</p></div>

---

## 1. Configure Codex (Desktop App, VS Code & CLI)

<div class="steps">
  <div class="step"><strong>Step 1: Select Codex</strong><p>Open CC Switch and click the <strong><code>Codex</code></strong> tab at the top.</p></div>
  <div class="step"><strong>Step 2: Add Custom Provider</strong><p>Click <strong><code>+</code></strong> in the top-right corner -> select <strong>Custom</strong>.</p></div>
  <div class="step"><strong>Step 3: Fill In the Form</strong>
    <div class="config-card">
      <div><span>Name</span><code>OpenBridger</code></div>
      <div><span>Base URL</span><code>{{API_BASE_URL}}</code> (<strong>with /v1</strong>)</div>
      <div><span>API Key</span>Your <code>sk-...</code> key</div>
      <div><span>Wire API</span><code>responses</code></div>
      <div><span>Model</span><code>gpt-6.2-astra</code> (exact ID from Models & Pricing)</div>
      <div><span>Reasoning Effort (optional)</span><code>medium</code> or <code>high</code></div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Save & Click Enable</strong><p>Click <strong>Save</strong>, then click <strong>Enable</strong> on the card so it shows active.</p></div>
  <div class="step"><strong>Step 5: Restart Codex</strong><p>Fully quit and relaunch the Codex Desktop App (<code>Cmd + Q</code> on macOS / tray exit on Windows), or run <code>Developer: Reload Window</code> in VS Code, and open a new thread.</p></div>
</div>

---

## 2. Configure Claude Code (Desktop, VS Code & CLI)

<div class="steps">
  <div class="step"><strong>Step 1: Select Claude Code</strong><p>Click the <strong><code>Claude Code</code></strong> tab in CC Switch.</p></div>
  <div class="step"><strong>Step 2: Add Custom Provider</strong><p>Click <strong><code>+</code></strong> -> select <strong>Custom</strong>.</p></div>
  <div class="step"><strong>Step 3: Fill In the Form</strong>
    <div class="config-card">
      <div><span>Name</span><code>OpenBridger Claude</code></div>
      <div><span>Base URL</span><code>{{CONSOLE_URL}}</code> (<strong>NO /v1!</strong>)</div>
      <div><span>API Key / Auth Token</span>Your <code>sk-...</code> key</div>
      <div><span>Model (optional)</span><code>claude-sonnet-5-5</code> or <code>claude-opus-5-5</code></div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Save & Click Enable</strong><p>Click <strong>Save</strong>, then click <strong>Enable</strong> on the card.</p></div>
  <div class="step"><strong>Step 5: Open a Fresh Terminal</strong><p>Close old terminal windows, open a new terminal, and run <code>claude</code>.</p></div>
</div>

## Common Issues

| Symptom | Fix |
| --- | --- |
| Client still prompts for official login | Click **Enable** on the card in CC Switch and fully quit/relaunch the client |
| `404` error | Codex needs `{{API_BASE_URL}}`; Claude Code needs `{{CONSOLE_URL}}` |
| `503` No available channel | Re-copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) and check key group |
| `403` Insufficient quota | Top up or redeem credits in [Wallet]({{CONSOLE_URL}}/wallet) |

## Related

- [Codex (Desktop, IDE & CLI)](/guide/clients/codex-cli/)
- [Claude Code (App & CLI)](/guide/clients/claude-code/)
- [Error codes](/guide/errors/)
