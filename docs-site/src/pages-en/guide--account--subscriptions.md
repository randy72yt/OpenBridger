# Subscriptions

In addition to pay-as-you-go wallet top-ups, you can purchase fixed-term plans in [Subscriptions]({{CONSOLE_URL}}/subscriptions) or [Wallet]({{CONSOLE_URL}}/wallet). **Plans never auto-renew or auto-charge** when they expire.

## How to Subscribe

<div class="steps">
  <div class="step"><strong>Step 1: Pick a Plan</strong><p>Open <a href="{{CONSOLE_URL}}/subscriptions">Subscriptions</a>, review the duration, quota allowance, reset cycle, and included groups, then click <strong>Subscribe</strong>.</p></div>
  <div class="step"><strong>Step 2: Confirm & Pay</strong><p>Verify the RMB total in the confirmation dialog and pay via Alipay, WeChat Pay, or Wallet Balance (when enabled for that plan).</p></div>
  <div class="step"><strong>Step 3: Start Using Immediately</strong><p>Your existing API keys work right away once the plan is active.</p></div>
</div>

## Billing Order (Subscription vs. Wallet)

You can switch your billing preference anytime in [Wallet]({{CONSOLE_URL}}/wallet) (defaults to **Subscription First**):

| Preference | Behavior |
| --- | --- |
| **Subscription First (Default)** | Uses plan quota first; falls back to Wallet when plan quota runs out (if the plan permits fallback). |
| **Wallet First** | Uses Wallet balance first, then active plan quota. |
| **Subscription Only** | Uses plan quota exclusively — returns `403` when exhausted without touching Wallet funds. |
| **Wallet Only** | Uses Wallet balance exclusively. |

## Related

- [Credits and billing](/guide/billing/)
- [Top up and redeem](/guide/account/redeem/)
- [Terms of service](/legal/terms/)
