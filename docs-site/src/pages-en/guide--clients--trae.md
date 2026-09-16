# Use OpenBridger with Trae

If your Trae version supports custom models or an OpenAI-compatible provider, use these settings:

<div class="config-card"><div><span>Provider</span><code>OpenAI Compatible</code></div><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API key</span><code>YOUR_API_KEY</code></div><div><span>Model</span><code>Exact ID from the console</code></div></div>

Send a short request and confirm it in Usage Logs. Add `/chat/completions` only when the client explicitly asks for a complete endpoint.
