# Cursor Setup Guide

Connect Cursor's **Chat** and **Agent** panels to OpenBridger via **Override OpenAI Base URL** (no need to sign out of your Cursor account; BYOK availability depends on your Cursor build/plan).

## Before You Begin

1. Create a key named `cursor` in [API Keys]({{CONSOLE_URL}}/keys) and copy `sk-...`.
2. Copy an exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) (e.g., `gpt-6.2-astra` or `claude-sonnet-5-5`).

## 5 Setup Steps

<div class="steps">
  <div class="step"><strong>Step 1: Open Cursor Settings</strong><p>Click the top-right <strong>⚙️ gear icon</strong> (<code>Cmd/Ctrl + Shift + J</code>) -> select <strong><code>Models</code></strong>.</p></div>
  <div class="step"><strong>Step 2: Add Your Custom Model ID First</strong><p>Click <strong><code>+ Add Model</code></strong> below the list, paste the exact Model ID (e.g., <code>gpt-6.2-astra</code>), and keep it checked. Uncheck unused built-in OpenAI models so <code>Verify</code> tests your custom model.</p></div>
  <div class="step"><strong>Step 3: Enter Key & Override Base URL</strong>
    <div class="config-card">
      <div><span>OpenAI API Key</span>Your <code>sk-...</code> key</div>
      <div><span>Override OpenAI Base URL</span>Toggle ON -> <code>{{API_BASE_URL}}</code> (with <code>/v1</code>)</div>
    </div>
  </div>
  <div class="step"><strong>Step 4: Save & Verify</strong><p>Click <strong>Save</strong> next to the Base URL field, then click <strong><code>Verify</code></strong> (or enable the OpenAI API Key toggle).</p></div>
  <div class="step"><strong>Step 5: Open a New Chat & Disable Auto</strong><p>Press <code>Cmd/Ctrl + L</code> -> click <strong><code>+</code> (New Chat)</strong> -> turn off <code>Auto</code> in the model picker, select your custom model, send a message, and check <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div>
</div>

> **Notes**: Cursor Tab inline autocomplete stays on Cursor's own engine. Turn off `Override OpenAI Base URL` and restart Cursor before switching back to Cursor's bundled subscription models.

## Common Issues

| Symptom | Fix |
| --- | --- |
| `Verify` complains about an unfamiliar model | Uncheck other built-in OpenAI models so only your custom Model ID remains checked |
| `404` error | Set Base URL strictly to `{{API_BASE_URL}}` |
| `503` or `model not found` | Re-copy the exact Model ID from [Models & Pricing]({{CONSOLE_URL}}/pricing) |
| Nothing in Usage Logs | Turn off `Auto` in the chat model picker and start a fresh Chat tab |

## Related

- [Cline & Roo Code (Extension Alternative)](/guide/clients/cline/)
- [Troubleshooting](/troubleshooting/)
