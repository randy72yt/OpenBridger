# Create and Manage API Keys

API keys authenticate your apps and SDKs. Manage them in [API Keys]({{CONSOLE_URL}}/keys). Creating **one dedicated key per tool** (e.g., `codex`, `cherry`, `zotero`) makes usage tracking and key rotation effortless.

## Create a Key in 3 Steps

<div class="steps">
  <div class="step">
    <strong>Step 1: Click Create Key</strong>
    <p>Open <a href="{{CONSOLE_URL}}/keys">{{CONSOLE_URL}}/keys</a> and click the button to create a new key.</p>
  </div>
  <div class="step">
    <strong>Step 2: Name It & Pick a Group</strong>
    <p>• <strong>Name</strong>: name it after the client (e.g., <code>cursor</code>, <code>claude-code</code>).<br>
    • <strong>Group</strong>: keep your default group, pick a specific route group, or choose <code>auto</code>.<br>
    • <strong>Quota / Expiry</strong>: keep unlimited / never expire for personal tools; set a spending cap when sharing a key (such as a lab Open WebUI server).</p>
  </div>
  <div class="step">
    <strong>Step 3: Copy Your <code>sk-...</code> Key</strong>
    <p>Click Copy next to your new key. The copied token already starts with <code>sk-</code> — paste it directly into client API Key fields without adding a second <code>sk-</code> or <code>Bearer </code> prefix.</p>
  </div>
</div>

## Optional Key Restrictions

| Setting | Behavior |
| --- | --- |
| **Key Quota** | When a key's own cap runs out, requests return `401` Invalid token even if your wallet still has funds. |
| **Group = `auto`** | Tries your available groups in order; enabling cross-group retry automatically falls back on transient upstream errors. |
| **Model Allowlist** | Restricts the key to only the listed Model IDs. |
| **IP Allowlist** | Leave blank for personal laptops/campus Wi-Fi (dynamic IPs); non-matching IPs receive `403 access_denied`. |

<div class="callout warning"><strong>Keep keys private</strong><p>Never embed <code>sk-...</code> keys in browser frontend code, public Git repositories, or screenshots. If a key is exposed, disable or delete it immediately and create a replacement.</p></div>

## Related

- [Quick start](/guide/quickstart/)
- [Groups and routes](/guide/account/groups/)
- [Error codes](/guide/errors/)
