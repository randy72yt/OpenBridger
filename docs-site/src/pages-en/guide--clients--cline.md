# Cline & Roo Code

Cline and Roo Code are free, open-source autonomous coding agents for VS Code. You can connect both extensions to OpenBridger via either **OpenAI Compatible** mode (works across all model families) or native **Anthropic** mode.

## Before you start

- Install **Cline** or **Roo Code** in VS Code.
- Create a dedicated API key named `cline` under [API Keys]({{CONSOLE_URL}}/keys).
- Copy a coding model ID from [Model Square]({{CONSOLE_URL}}/pricing) that supports **tool / function calling** (such as `claude-opus-5-5`, `claude-sonnet-5-5`, `gpt-6.2-astra`, or `deepseek-v4.1`).

<div class="callout warning"><strong>Tool calling required</strong><p>Cline and Roo Code edit files and run terminal commands through structured tool calls. Pick a model and key group that support native tool calling.</p></div>

## Option 1: OpenAI Compatible mode (recommended for multi-vendor switching)

1. Click the gear icon in Cline / Roo Code to open **Settings**.
2. Set **API Provider** to `OpenAI Compatible`.
3. Fill in:

<div class="config-card">
  <div><span>Base URL</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Key</span><code>sk-your-api-key</code></div>
  <div><span>Model ID</span><code>claude-sonnet-5-5</code> or <code>gpt-6.2-astra</code> (from Model Square)</div>
</div>

Enable **Supports Images** if the chosen model supports vision so you can paste UI screenshots directly into prompts.

## Option 2: Native Anthropic mode (for Claude models)

1. Set **API Provider** to `Anthropic`.
2. Check **Use custom base URL** and enter `{{CONSOLE_URL}}` (**without** `/v1`).
3. Paste `sk-your-api-key` into **Anthropic API Key** and select a Claude model ID available in your group.

## Common issues

| Symptom | Fix |
| --- | --- |
| `404` | Use `{{API_BASE_URL}}` for OpenAI Compatible mode, and `{{CONSOLE_URL}}` for Anthropic mode |
| `503` `model_not_found` | Paste the exact model ID from [Model Square]({{CONSOLE_URL}}/pricing) |
| Agent prints code instead of applying diffs | Switch to a frontier coding model with reliable tool-use support |

## Related

- [Cursor](/guide/clients/cursor/)
- [OpenAI Codex](/guide/clients/codex-cli/)
- [Claude Code](/guide/clients/claude-code/)
