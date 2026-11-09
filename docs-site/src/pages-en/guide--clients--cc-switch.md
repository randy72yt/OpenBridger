# Switch to OpenBridger with CC Switch

CC Switch stores several providers on your machine and writes them into Codex or Claude Code. OpenBridger does not offer a one-click browser import. Create the provider by hand inside CC Switch.

## Do not mix the two base URLs

<div class="config-card"><div><span>Claude Code Base URL</span><code>{{CONSOLE_URL}}</code></div><div><span>Codex / OpenAI-compatible Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>Key</span><code>sk-...</code>, preferably its own key named <code>cc-switch</code></div><div><span>Model</span><code>full ID from Model Square</code></div></div>

## Steps

<div class="steps"><div class="step"><strong>Create a key in the console</strong><p>See <a href="/guide/api-key/">Create an API key</a>.</p></div><div class="step"><strong>Pick the target app in CC Switch</strong><p>Claude Code uses the Anthropic-compatible base, which is the site origin. Codex uses Responses, and its Base URL includes <code>/v1</code>.</p></div><div class="step"><strong>Enter the model ID</strong><p>Copy it from <a href="{{CONSOLE_URL}}/pricing">Model Square</a>.</p></div><div class="step"><strong>Enable it and restart the client</strong><p>Quit Codex or Claude Code completely, then open it again. Switching the UI without a restart leaves the old environment variables in place.</p></div></div>

Keep one Base URL active. A leftover `ANTHROPIC_BASE_URL` or `OPENAI_BASE_URL` in the terminal overrides the profile you just selected.

## The same values as environment variables

Claude Code:

```bash
export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
export ANTHROPIC_AUTH_TOKEN="sk-your-key"
```

OpenAI-compatible:

```bash
export OPENAI_BASE_URL="{{API_BASE_URL}}"
export OPENAI_API_KEY="sk-your-key"
```

Field details are on [Claude Code](/guide/clients/claude-code/) and [Codex](/guide/clients/codex-cli/).

## Common problems

| What you see | What to do |
| --- | --- |
| The old provider is still used | Open a new terminal and inspect `OPENAI_` and `ANTHROPIC_` |
| 401 | Copy the key again and remove spaces |
| 503 no channel | The model ID or the protocol does not match the group |
| It fails immediately and the log says insufficient credit | Check the wallet and the key's own credit |

## Related

- [Troubleshooting](/troubleshooting/)
