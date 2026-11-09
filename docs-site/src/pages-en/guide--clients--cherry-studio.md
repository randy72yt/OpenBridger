# Use OpenBridger in Cherry Studio

Cherry Studio can add an OpenAI custom provider. In current builds the entry is under model services or the provider list. If you cannot find a custom OpenAI provider, follow that version's settings page. Do not change system network settings to force this URL.

## Values

<div class="config-card"><div><span>Provider name</span><code>OpenBridger</code></div><div><span>Type</span><code>OpenAI</code></div><div><span>API address</span><code>{{API_BASE_URL}}</code></div><div><span>API key</span><code>a dedicated sk- key</code></div><div><span>Model ID</span><code>full name from Model Square</code></div></div>

<div class="steps"><div class="step"><strong>Add a provider</strong><p>Type OpenAI, name OpenBridger.</p></div><div class="step"><strong>Enter the address and key</strong><p>Do not append <code>/v1</code> or <code>/chat/completions</code>. <code>{{API_BASE_URL}}</code> already includes <code>/v1</code>.</p></div><div class="step"><strong>Add models</strong><p>If the automatic list is empty, add the full ID from <a href="{{CONSOLE_URL}}/pricing">Model Square</a> by hand.</p></div><div class="step"><strong>Send a short message and check the log</strong><p>Select this provider, send one short message, and find the key in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div></div>

You do not sign out of other providers in Cherry Studio. Each chat uses the provider you selected.

## Common problems

| What you see | What to do |
| --- | --- |
| The key check fails | Compare with curl against <code>{{API_BASE_URL}}/models</code>. If curl works, the address has an extra path |
| The model list is empty | Add the model ID by hand |
| 503 or model not found | Check case and the version suffix |
| No streaming | Check the client's stream toggle, then try the same model without streaming |

## Related

- [Chatbox](/guide/clients/chatbox/)
- [LobeChat](/guide/clients/lobechat/)
- [Error codes](/guide/errors/)
