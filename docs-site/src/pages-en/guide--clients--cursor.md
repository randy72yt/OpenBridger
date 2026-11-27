# Cursor Step-by-Step Setup Guide

Cursor lets you route **Chat** and **Agent** requests through OpenBridger using its built-in **Override OpenAI Base URL** setting. No extra extension is needed — follow steps **1 -> 2 -> 3 -> 4 -> 5** below.

---

## Before You Begin

1. Open [API Keys]({{CONSOLE_URL}}/keys), create a dedicated key named `cursor`, and copy your `sk-...` token.
2. Open [Models & Pricing]({{CONSOLE_URL}}/pricing) and copy the exact model ID you want to use (for example `gpt-6.2-astra`, `claude-sonnet-5-5`, or `deepseek-v4.1`).
3. Keep your existing Cursor account signed in — custom API keys work alongside your Cursor login (subject to your Cursor version's BYOK policy).

---

## Step-by-Step UI Configuration (1-2-3-4-5)

<div class="steps">
  <div class="step">
    <strong>Step 1: Open Cursor Settings</strong>
    <p>In Cursor, click the <strong>⚙️ gear icon</strong> in the top-right corner (or press <code>Cmd + Shift + J</code> on macOS / <code>Ctrl + Shift + J</code> on Windows) to open <strong>Cursor Settings</strong>.</p>
  </div>
  <div class="step">
    <strong>Step 2: Go to Models & Add Custom Model IDs</strong>
    <p>Select <strong><code>Models</code></strong> in the left sidebar.<br>
    Scroll below the model list, click <strong><code>+ Add Model</code> (or Add Custom Model)</strong>, paste the exact model ID copied from <a href="{{CONSOLE_URL}}/pricing">Models & Pricing</a> (such as <code>gpt-6.2-astra</code>), and make sure its toggle is switched on. <em>(Tip: temporarily uncheck unused built-in OpenAI models so Cursor's Verify check tests your custom model.)</em></p>
  </div>
  <div class="step">
    <strong>Step 3: Enter OpenAI API Key & Override Base URL</strong>
    <p>Scroll down to the <strong><code>OpenAI API Key</code></strong> section:<br>
    1. Paste your <code>sk-...</code> key into the <strong>OpenAI API Key</strong> input.<br>
    2. Expand and turn on <strong><code>Override OpenAI Base URL</code></strong>.<br>
    3. Enter <strong><code>{{API_BASE_URL}}</code></strong> (keep <code>/v1</code>, with no trailing slash or <code>/chat/completions</code>).</p>
  </div>
  <div class="step">
    <strong>Step 4: Save & Verify</strong>
    <p>Click the <strong>Save</strong> button next to the Base URL input, then click <strong><code>Verify</code></strong> (or toggle the OpenAI API Key switch on).</p>
  </div>
  <div class="step">
    <strong>Step 5: Start a New Chat, Select Your Model & Check Logs</strong>
    <p>Press <code>Cmd + L</code> (Windows: <code>Ctrl + L</code>) to open the AI panel, click <strong><code>+</code> (New Chat)</strong>, turn off <code>Auto</code> in the model picker at the bottom of the input box, select your newly added model (e.g., <code>gpt-6.2-astra</code>), and send a short test message. Confirm it appears in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p>
  </div>
</div>

### Quick Reference Card

<div class="config-card">
  <div><span>Menu Path</span>Top-right ⚙️ -> <code>Cursor Settings</code> -> <code>Models</code></div>
  <div><span>OpenAI API Key</span>Your dedicated <code>sk-...</code> key</div>
  <div><span>Override OpenAI Base URL</span>Enabled -> <code>{{API_BASE_URL}}</code></div>
  <div><span>+ Add Model</span>Exact model ID from Models & Pricing (e.g., <code>gpt-6.2-astra</code>)</div>
</div>

---

## Important Notes on Built-in Cursor Features

- **Cursor Tab Autocomplete stays on Cursor**: Inline ghost-text completions continue to use Cursor's proprietary completion engine and will not appear in OpenBridger logs.
- **Turn off Base URL override before switching back to Cursor's built-in models**: While `Override OpenAI Base URL` is active, Cursor sends all OpenAI-routed models to `{{API_BASE_URL}}`. Disable that toggle and restart Cursor if you want to return to Cursor's bundled models.

## Common Issues & Fixes

| Symptom | Cause | Fix |
| --- | --- | --- |
| Clicking `Verify` complains about an unfamiliar model | Cursor picks the first checked model in your list to test | Uncheck other built-in OpenAI models so only your custom model ID is checked, then click `Verify` |
| `model not found` or `503` no available channel | Typo in model name or key group lacks that model | Re-copy the exact ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) and check key group |
| `404 Not Found` | Missing `/v1` or extra `/chat/completions` in Base URL | Set Base URL to `{{API_BASE_URL}}` |
| Nothing shows up in Usage Logs | Model selector is still set to `Auto` or a Cursor built-in model | Disable `Auto`, pick your custom model ID in a fresh Chat tab, and try again |

## Related

- [Cline & Roo Code (Extension Alternative)](/guide/clients/cline/)
- [Base URL reference](/guide/base-urls/)
- [Troubleshooting](/troubleshooting/)
