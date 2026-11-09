# Use OpenBridger in Chatbox

In Chatbox's model providers, choose OpenAI API or OpenAI Compatible. Do not choose sign-in with a web account.

<div class="config-card"><div><span>API Host / Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API key</span><code>sk-...</code></div><div><span>Model</span><code>full ID from Model Square</code></div></div>

## Steps

<div class="steps"><div class="step"><strong>Open the model provider settings</strong><p>The label may be Model Provider, AI Provider, or API.</p></div><div class="step"><strong>Choose OpenAI-compatible</strong><p>Enter the three values above. Do not include <code>/chat/completions</code> in the host unless that field says it wants the full endpoint.</p></div><div class="step"><strong>Type the model if the list is empty</strong><p>Copy it from <a href="{{CONSOLE_URL}}/pricing">Model Square</a>.</p></div><div class="step"><strong>Start a new chat</strong><p>An old chat can stay bound to the previous provider. Create a new one, select OpenBridger, and confirm it in Usage Logs.</p></div></div>

You do not sign out of the Chatbox account.

## Common problems

| What you see | What to do |
| --- | --- |
| 404, and the path contains chat/completions or /v1 twice | The host includes an extra path. Set it back to `{{API_BASE_URL}}` |
| 401 | Paste the key again |
| Nothing in the logs | This chat is still on a built-in model. Start a new chat and select the provider again |

## Related

- [Cherry Studio](/guide/clients/cherry-studio/)
- [OpenAI-compatible calls](/guide/clients/openai-compatible/)
