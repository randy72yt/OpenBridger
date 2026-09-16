# Make your first call in five minutes

Sign in, create an API key, and test a model in the Playground before connecting a tool or SDK.

<div class="steps"><div class="step"><strong>Open the console</strong><p>Sign in with an available method.</p></div><div class="step"><strong>Choose a model</strong><p>Check available models and pricing in Model Square.</p></div><div class="step"><strong>Create an API key</strong><p>Use a separate key for the current tool or project.</p></div><div class="step"><strong>Send a test request</strong><p>Configure the Base URL, key, and exact model ID.</p></div></div>

```bash
curl {{API_BASE_URL}}/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"MODEL_ID_FROM_CONSOLE","messages":[{"role":"user","content":"Hello OpenBridger"}]}'
```

After a successful call, confirm the time, model, key, and charge in Usage Logs.
