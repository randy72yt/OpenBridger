# Use OpenBridger in Trae

If this Trae version has a custom model or an OpenAI-compatible provider, use the values below. If the menus have no Base URL and no custom provider, this version cannot be connected from this page.

<div class="config-card"><div><span>Provider</span><code>OpenAI Compatible</code></div><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API key</span><code>sk-...</code></div><div><span>Model</span><code>full ID from Model Square</code></div></div>

<div class="steps"><div class="step"><strong>Open model or provider settings</strong><p>Find custom OpenAI or Base URL.</p></div><div class="step"><strong>Enter the four values</strong><p>Use <code>{{API_BASE_URL}}/chat/completions</code> only when the field says it wants the full endpoint.</p></div><div class="step"><strong>Check Chat, Agent, and completion separately</strong><p>Each can have its own model entry. Completion being unavailable does not mean chat skipped the gateway.</p></div><div class="step"><strong>Confirm in Usage Logs</strong><p>The short message should appear under this key.</p></div></div>

You do not sign out of the Trae account. To use the built-in service again, select the official provider and turn OpenBridger off.

## Related

- [Cursor](/guide/clients/cursor/)
- [Troubleshooting](/troubleshooting/)
