# Use OpenBridger in LobeChat

LobeChat's OpenAI provider can take a custom API address. These values are enough. No extra plugin is required.

<div class="config-card"><div><span>Provider</span><code>OpenAI</code></div><div><span>API key</span><code>sk-...</code></div><div><span>Endpoint / proxy address</span><code>{{API_BASE_URL}}</code></div><div><span>Model ID</span><code>full ID from Model Square</code></div></div>

Depending on the version, the field may be called API Proxy, Endpoint, or a proxy address. Enter `{{API_BASE_URL}}`. If the placeholder already ends in `/v1`, do not add another `/v1`. If the placeholder is a bare host and the client appends `/v1` itself, enter only `{{CONSOLE_URL}}`. The check is the final path: `/v1/chat/completions`, with a single `/v1`.

## Steps

<div class="steps"><div class="step"><strong>Open language-model settings</strong><p>Use the OpenAI provider, not a provider that requires an official account sign-in.</p></div><div class="step"><strong>Enter the key and address</strong><p>Use a key that belongs only to LobeChat.</p></div><div class="step"><strong>Add the model</strong><p>If the list does not load, add a custom model with the full ID from Model Square.</p></div><div class="step"><strong>Start a new session</strong><p>Send a short message and confirm it in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div></div>

LobeChat renames screens between versions. If you cannot find the labels above, use the form that is OpenAI plus a custom address. The three values stay the same.

## Common problems

| What you see | What to do |
| --- | --- |
| 401 | The key or the address is wrong. curl `/v1/models` first |
| The model list is empty | Add the model ID by hand |
| The request goes to `api.openai.com` | The custom address was not saved, or the session is still on an official model |

## Related

- [Chatbox](/guide/clients/chatbox/)
- [Troubleshooting](/troubleshooting/)
