# Create and manage an API key

An API key is the credential a program or client uses. The console page is API Keys: [{{CONSOLE_URL}}/keys]({{CONSOLE_URL}}/keys).

## Before you start

- You are signed in.
- You know which tool this key is for. One tool, one key, so you can disable it and read its logs on their own.

## Create one

<div class="steps"><div class="step"><strong>Open API Keys</strong><p>Go to <a href="{{CONSOLE_URL}}/keys">{{CONSOLE_URL}}/keys</a> and create a key.</p></div><div class="step"><strong>Name it</strong><p>The name is at most 50 characters. Use the tool or project, for example <code>cursor</code>, <code>claude-code</code>, or <code>server-prod</code>.</p></div><div class="step"><strong>Limit it if you need to</strong><p>The form can set the group, model list, key credit, expiry, and IP allowlist. If you are unsure, keep the defaults, send a short request, then tighten them.</p></div><div class="step"><strong>Copy the full key</strong><p>The value shown in the console starts with <code>sk-</code>. Send <code>Authorization: Bearer sk-...</code>. The gateway strips that prefix before it looks the key up, so do not add <code>sk-</code> a second time.</p></div></div>

<div class="callout"><strong>Screenshot to add</strong><p>The create-key drawer: name, group, credit, expiry, model limit, and IP allowlist. Replace this note with the current UI.</p></div>

There is a maximum number of keys. The error text includes the current limit when you hit it.

While you are signed in you can reveal a key again. That does not make it safe to put in a repository, a frontend bundle, a screenshot, or a support ticket.

## What each field does

| Setting | Behavior |
| --- | --- |
| Unlimited credit / remaining credit | With unlimited turned off, the key has its own balance. When that balance is used up, calls return `401` and an invalid-token message, even if the wallet still has credit. |
| Expiry | An expired key also returns `401`. `-1` means it does not expire. |
| Model limit | When enabled, the key can call only the listed model IDs. |
| Group | Empty uses the account group. A filled group must be one you can use and must not be retired, or the call returns `403`. |
| Group `auto` | Tries groups in the order you saved. Cross-group retry continues with the next group when the current group's routes fail. |
| IP allowlist | Empty means no IP limit. A source IP outside the list returns `403` with `code` `access_denied`. |
| Disabled | A disabled key cannot call the API. Read-only usage queries reject it too. |

A normal user cannot append `-channelId` to pin an upstream channel. That returns `403`.

## Habits that avoid leaks

<div class="grid"><div class="card"><h3>Split by use</h3><p>Cursor, Claude Code, and production each get a key. A leak or a departure disables one of them.</p></div><div class="card"><h3>Keep it on a server</h3><p>Do not ship the key in a browser, an app package, or a public repository. A website should call your server, and the server should call OpenBridger.</p></div><div class="card"><h3>Rotate it</h3><p>Disable or delete the old key when it leaks, when people change, or when the project ends. Support will not ask for the full key or your password.</p></div></div>

## Check

```bash
curl "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer sk-your-key"
```

A model list means this key and the network work. The list is what this key's group can see, not every upstream model.

## Related

- [Quick start](/guide/quickstart/)
- [Groups and routes](/guide/account/groups/)
- [Error codes](/guide/errors/)
