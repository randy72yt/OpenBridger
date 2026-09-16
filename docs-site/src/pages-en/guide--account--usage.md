# Usage and logs

Usage Logs show whether a request reached OpenBridger, which model handled it, how much credit it used, and the returned status.

- Create separate keys for Codex, Claude Code, test scripts, and production.
- Filter by time and key name instead of model alone.
- Record the request ID, time, model, and status when reporting an error. Never send the full key.
- Unexpected usage may come from repeated conversation history or agent tool calls.

<div class="callout"><strong>Task logs</strong><p>Asynchronous image or video jobs may appear under Task Logs, while chat requests normally appear under Usage Logs.</p></div>
