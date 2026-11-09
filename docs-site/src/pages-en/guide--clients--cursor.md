# Use OpenBridger in Cursor

When your Cursor version offers a custom OpenAI Base URL or an OpenAI-compatible provider, Chat or Agent can use OpenBridger. Menu names change between versions. If there is no Base URL field, this version cannot be connected from this page.

## Before you start

- An API key used only by Cursor. See [Create an API key](/guide/api-key/).
- A full model ID copied from [Model Square]({{CONSOLE_URL}}/pricing).
- A Cursor version that allows your own API key. That permission is Cursor's, by version and plan. OpenBridger does not grant it.

## Configure

<div class="steps"><div class="step"><strong>Open Cursor Settings</strong><p>Find Models, API Keys, or a custom provider.</p></div><div class="step"><strong>Fill the OpenAI-compatible fields</strong><p>Turn on Override OpenAI Base URL, or the equivalent. Base URL is <code>{{API_BASE_URL}}</code>. The key starts with <code>sk-</code>.</p></div><div class="step"><strong>Add the model ID</strong><p>Use the full ID from Model Square. Do not type only a vendor family name.</p></div><div class="step"><strong>Send one short message</strong><p>Send it in Chat, then match this key in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div></div>

<div class="config-card"><div><span>Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>API key</span><code>sk-...</code></div><div><span>Model</span><code>full ID from Model Square</code></div></div>

## Separate this from Cursor's own features

- You do not sign out of the Cursor account. Account sync and a custom model are different settings.
- Tab completion and similar Cursor features may keep using Cursor's own service. Those calls will not appear in OpenBridger logs.
- Check Chat, Agent, and completion separately.
- A global override also sends official models to this Base URL. Turn the override off before you switch back, then restart Cursor.

## Common problems

| What you see | What to do |
| --- | --- |
| There is no Verify button | Save, send a short message, and confirm it in Usage Logs |
| model not found, or 503 no channel | Copy the model ID again and check the key's group |
| Official models fail too | Turn off the Base URL override and restart |
| Nothing in the logs | This feature is not using the custom Base URL, or it is still on a Cursor model |
| 401 | The key has a space, or `Bearer` was added twice |

## Related

- [OpenAI-compatible calls](/guide/clients/openai-compatible/)
- [Troubleshooting](/troubleshooting/)
