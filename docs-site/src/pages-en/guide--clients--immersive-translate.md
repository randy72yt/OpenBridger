# Immersive Translate

Immersive Translate provides side-by-side bilingual translation for web pages, ArXiv preprints, and local PDF papers while preserving layout and LaTeX equations. Connecting it to OpenBridger lets you use fast, cost-efficient models for high-volume reading.

## Before you start

- Install the **Immersive Translate** browser extension.
- Create a dedicated API key named `immersive-translate` under [API Keys]({{CONSOLE_URL}}/keys).
- Pick a **fast, non-reasoning** model from [Model Square]({{CONSOLE_URL}}/pricing) (such as `gpt-6.2-sol`, `gemini-4-hydrogen`, or `deepseek-v4.1-flash`) and copy its exact model ID.

<div class="callout warning"><strong>Model choice tip</strong><p>Paragraph-by-paragraph translation needs low latency. Avoid heavy chain-of-thought reasoning models here — reasoning tokens slow down every paragraph and add unnecessary output token charges.</p></div>

## Configuration steps

<div class="steps">
  <div class="step"><strong>Open Translation Services</strong><p>Click the Immersive Translate extension icon -> <strong>Settings</strong> -> <strong>Translation Services</strong>, then select <strong>OpenAI</strong> (or Custom OpenAI-compatible service).</p></div>
  <div class="step"><strong>Enter your key and full endpoint URL</strong><p>Immersive Translate expects the <strong>full chat completions path</strong> rather than just the host:</p></div>
  <div class="step"><strong>Set a custom model & verify</strong><p>Enable custom model name, paste the exact model ID from Model Square, and click <strong>Verify</strong> to test.</p></div>
</div>

<div class="config-card">
  <div><span>API Key</span><code>sk-your-api-key</code></div>
  <div><span>Custom API Endpoint URL</span><code>{{API_BASE_URL}}/chat/completions</code></div>
  <div><span>Custom Model</span><code>Model ID copied from Model Square</code></div>
</div>

## Rate-limit settings for full PDF papers

Translating a multi-page PDF fires many parallel requests at once. Under **Rate Limit / Advanced Settings** in the OpenAI service panel, set:

- **Max requests per second**: `3` to `5`
- **Paragraphs per request**: keep default or slightly higher so multiple short sentences are batched into one request

## Common issues

| Symptom | Fix |
| --- | --- |
| Verification fails with `404` | Confirm the URL ends with `/v1/chat/completions` — see [Base URL reference](/guide/base-urls/) |
| `503` `model_not_found` | Re-copy the exact model ID from [Model Square]({{CONSOLE_URL}}/pricing) and verify your key's group |
| Occasional `429` mid-PDF | Lower max requests per second in the extension's rate-limit settings |

## Related

- [Zotero](/guide/clients/zotero/)
- [Base URL reference](/guide/base-urls/)
- [Error codes](/guide/errors/)
