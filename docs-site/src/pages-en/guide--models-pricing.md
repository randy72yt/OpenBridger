# Models and Pricing

All available models, supported protocols, and live prices are listed in [Models & Pricing]({{CONSOLE_URL}}/pricing).

<div class="actions"><a class="btn primary" href="{{CONSOLE_URL}}/pricing">Open Models & Pricing</a></div>

## 3 Rules When Choosing a Model

1. **Always Copy the Exact Model ID**
   - Click the model badge in Models & Pricing to copy its full identifier (case, hyphens, and suffixes included — e.g., `gpt-6.2-astra`, `claude-sonnet-5-5`). Never abbreviate it to `gpt-6` or `claude`.
2. **Match the Protocol to Your Client**
   - **Chat & translation apps** (Cherry Studio, Chatbox, Immersive Translate, Cursor): use `/v1/chat/completions`.
   - **Codex**: requires a model and group supporting `/v1/responses`.
   - **Claude Code**: requires a model and group supporting `/v1/messages`.
3. **Pick the Right Tier for the Job**
   - **Complex coding, debugging & math proofs**: use flagship / reasoning models (`gpt-6.2-astra`, `claude-opus-5-5`, `gemini-4-carbon`).
   - **Bilingual PDF/web translation & quick Q&A**: use fast lightweight models (`gpt-6.2-sol`, `deepseek-v4.1-flash`, `gemini-4-hydrogen`) for lower latency and lower cost.

## Why Is My Client's Model Dropdown Empty?

Many desktop clients either skip calling `/v1/models` or filter out newer model names they don't recognize. Simply click **"+ Add Custom Model"** in your client and paste the exact Model ID copied from Models & Pricing.

## Related

- [Credits and billing](/guide/billing/)
- [Groups and routes](/guide/account/groups/)
- [Error codes](/guide/errors/)
