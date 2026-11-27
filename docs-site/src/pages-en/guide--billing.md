# Credits and Billing

Balances, model prices, and per-request costs across [Wallet]({{CONSOLE_URL}}/wallet), [Models & Pricing]({{CONSOLE_URL}}/pricing), and [Usage Logs]({{CONSOLE_URL}}/usage-logs) are all displayed directly in **RMB (¥)**.

## How Each Request Is Billed

Every API call follows a **Pre-hold -> Execute -> Settle** flow:

<div class="steps">
  <div class="step"><strong>1. Pre-hold</strong><p>Before forwarding your request, the gateway reserves an estimated amount. If your balance cannot cover the pre-hold, the request stops immediately with <code>403</code> (<code>insufficient_user_quota</code>) and nothing is charged.</p></div>
  <div class="step"><strong>2. Upstream Execution</strong><p>If an upstream failure prevents completion, the pre-hold is automatically refunded.</p></div>
  <div class="step"><strong>3. Exact Settlement</strong><p>Once finished, exact charges are calculated from actual input tokens, output tokens (including thinking tokens on reasoning models), cache hits, and group multipliers. Unused pre-hold is refunded immediately and the final cost is recorded in <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a>.</p></div>
</div>

## Wallet vs. Subscription Priority

- Without an active subscription, requests deduct from your **Wallet**.
- With an active subscription, billing defaults to **Subscription First** (configurable in Wallet between Subscription First, Wallet First, Subscription Only, or Wallet Only — see [Subscriptions](/guide/account/subscriptions/)).
- Note the difference in error codes: an empty wallet/subscription returns `403` (`insufficient_user_quota`), whereas hitting a per-key spending cap set in API Keys returns `401` Invalid token.

## Why Did My Usage Spike?

Check the input and output token columns in [Usage Logs]({{CONSOLE_URL}}/usage-logs):
1. **Long chat history**: Clients resend the entire conversation history on every turn. Click **New Chat** when switching topics.
2. **Multi-step coding agents**: Tools like Codex, Claude Code, and Cline automatically run multiple file-read and tool-call turns to finish one task.

## Related

- [Models and pricing](/guide/models-pricing/)
- [Top up and redeem](/guide/account/redeem/)
- [Usage and logs](/guide/account/usage/)
