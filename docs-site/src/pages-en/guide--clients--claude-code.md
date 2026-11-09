# Use OpenBridger in Claude Code

Claude Code speaks the Anthropic Messages protocol. The gateway path is `POST /v1/messages`. The key may be sent as `x-api-key`, which is what Claude Code does, and the gateway accepts that header on this path.

The model name must be a full ID from [Model Square]({{CONSOLE_URL}}/pricing), and the current group needs a route that can serve Messages. A name on the square is not enough. Send one short request and confirm it in [Usage Logs]({{CONSOLE_URL}}/usage-logs).

## Before you start

- The `claude` command is installed.
- An API key named `claude-code`.
- The model ID you intend to use.

## This terminal only

```bash
export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
export ANTHROPIC_AUTH_TOKEN="sk-your-key"

cd /path/to/your/project
claude
```

`ANTHROPIC_BASE_URL` is the site origin. **Do not** set it to `{{API_BASE_URL}}`. Claude Code requests `{BASE}/v1/messages`. A base that already contains `/v1` becomes `/v1/v1/messages`.

Closing the terminal drops the variables. To use the official service again:

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_AUTH_TOKEN
```

Open a new terminal before you run `claude`.

## Persist it in Claude Code settings

`~/.claude/settings.json`:

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{CONSOLE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-key"
  }
}
```

That file contains the key. Do not commit it or screenshot it. Do not put a real key in a project `.claude/settings.json` either.

If the client lets you set the model, use the full ID from Model Square. If Claude Code sends a default name that is not on the square, you get `503` (no channel) or an upstream model-not-found error. The model ID in the log is the one that was sent.

Check the variables in this terminal:

```bash
env | grep ANTHROPIC_
```

Confirm only that they exist. Do not paste the output into chat or a ticket.

## Common problems

| What you see | What to do |
| --- | --- |
| It still asks you to sign in to the official account | Open a new terminal, confirm both `ANTHROPIC_` variables, then start `claude` |
| 401 | Copy the key again. Do not put `Bearer` in front of the token |
| 503 no channel | The model ID Claude Code sends is not the square's ID, or this group has no Messages route |
| The request never reaches Usage Logs | Base URL still points at the official host. Check `ANTHROPIC_BASE_URL` |

<div class="callout warning"><strong>The protocol has to match</strong><p>A model that works on Chat Completions does not automatically work in Claude Code. The model, the group, and <code>/v1/messages</code> all have to line up.</p></div>

## Related

- [CC Switch](/guide/clients/cc-switch/)
- [Error codes](/guide/errors/)
