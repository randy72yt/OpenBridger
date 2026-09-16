# Manage OpenBridger with CC Switch

Create the provider manually. OpenBridger does not currently claim one-click browser import support.

<div class="config-card"><div><span>Profile name</span><code>OpenBridger</code></div><div><span>Base URL</span><code>{{CONSOLE_URL}}</code></div><div><span>Auth token</span><code>YOUR_API_KEY</code></div><div><span>Model</span><code>Exact ID from the console</code></div></div>

Create a dedicated key, choose Codex or Claude Code in CC Switch, select the matching Responses or Anthropic configuration, save it, and restart the client.

<div class="callout warning"><strong>Troubleshooting</strong><p>Keep only one active Base URL and key. Old environment variables can override a new profile.</p></div>

## Environment variables

```bash
export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
export ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"
```

For OpenAI-compatible clients:

```bash
export OPENAI_BASE_URL="{{API_BASE_URL}}"
export OPENAI_API_KEY="YOUR_API_KEY"
```

| Symptom | Check |
| --- | --- |
| Old provider is still used | Reopen the terminal and inspect active OpenAI and Anthropic variables |
| Unauthorized | Recopy the key or create a dedicated CC Switch key |
| Model not found | Copy the exact model ID and check group access |
| Connection refused / timeout | Check endpoint access, proxy conflicts, and DNS |
| Default model remains active | Save the profile and restart the target client |
