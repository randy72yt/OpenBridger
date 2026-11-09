# Use OpenBridger in Open WebUI

Adding a global OpenAI connection requires an Open WebUI administrator. Other users can only select models an administrator has already enabled.

## Add it in the admin UI

<div class="steps"><div class="step"><strong>Open connections</strong><p>Admin Panel or Settings, then OpenAI API connections.</p></div><div class="step"><strong>Add one</strong><p>Name it OpenBridger. URL is <code>{{API_BASE_URL}}</code>. Use a dedicated API key.</p></div><div class="step"><strong>Refresh models</strong><p>Save, then refresh. If discovery fails, enter model IDs from Model Square using Open WebUI's filter rules.</p></div><div class="step"><strong>Confirm with a short message</strong><p>That key should show up in OpenBridger <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div></div>

<div class="callout warning"><strong>One key is shared by the instance</strong><p>When several Open WebUI users share the connection an administrator saved, all of that usage is on this key. Create a key per instance or team, and set a credit cap on the key.</p></div>

## Environment variables

```bash
OPENAI_API_BASE_URL={{API_BASE_URL}}
OPENAI_API_KEY=sk-your-key
```

These are container environment variables, not browser settings. After you edit Docker Compose, recreate or restart the Open WebUI container. Reloading the page does not load the new values.

If your Open WebUI version uses `OPENAI_API_BASE_URLS` (with an S) or expects a URL list, follow that version. The value is still `{{API_BASE_URL}}`.

## Related

- [Create an API key](/guide/api-key/)
- [Usage and logs](/guide/account/usage/)
