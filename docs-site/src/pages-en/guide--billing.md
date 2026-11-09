# Credits and billing

The wallet balance is the credit used for calls. Model Square prices decide how a request converts into credit. The usage row is the settled result.

## The credit unit

The ledger stores an integer. The code default is **500000 credits = 1 US dollar** (`QuotaPerUnit`). The console can display that as dollars, yuan, or the raw integer. The currency symbol you see in the wallet and the logs is the current display. You do not multiply by this factor yourself.

When the display is yuan, balances are converted with the configured USD exchange rate. That rate and the top-up unit price are operator settings. The amount on the confirmation page is the amount you pay. Do not recompute it from a default in this guide.

## How one request is charged

<div class="steps"><div class="step"><strong>Reserve</strong><p>The gateway estimates the credit for this call. If the wallet or subscription cannot cover it, the call fails with HTTP <code>403</code>, <code>code</code> <code>insufficient_user_quota</code>, and a message about insufficient user quota or a failed pre-charge. Upstream is not called.</p></div><div class="step"><strong>Forward</strong><p>The request is sent only after the reserve succeeds. A failed request refunds the reserve.</p></div><div class="step"><strong>Settle</strong><p>On success the gateway charges actual usage. It debits more if the reserve was low and refunds the difference if the reserve was high. Usage Logs show the settled amount.</p></div></div>

The default preference is Subscription First (`subscription_first`). With no active subscription, the wallet is charged. With one, subscription credit is used first, and a shortfall falls back to the wallet only when that subscription allows it. See [Subscriptions](/guide/account/subscriptions/). The wallet can also switch to Wallet First, Subscription Only, or Wallet Only. The current terms say plan purchases are not open yet.

A key can also have its own remaining credit. When that key balance is used up the result is `401` and an invalid-token message, which is not the wallet's `403`.

## What changes the amount

- Input, output, and cache tokens. A longer context costs more input. An agent resends history and tool results.
- The group ratio. The same model can settle differently in another group.
- Image count, media duration, and resolution, only when that model bills those units. Extreme counts are rejected at validation. They are not stored as a negative charge.

Do not assume an extra unit is billed when Model Square does not list it. Use that model's page and this request's log.

## Check one charge

<div class="steps"><div class="step"><strong>Use a separate key</strong><p>Do not share a test key with production.</p></div><div class="step"><strong>Note the time</strong><p>Look at the clock before you send.</p></div><div class="step"><strong>Open Usage Logs</strong><p>Filter <a href="{{CONSOLE_URL}}/usage-logs">Usage Logs</a> by time, key, and model. Drawing and video tasks may be in the task log, not the chat log.</p></div><div class="step"><strong>Read the model and group</strong><p>Confirm the charge used the model ID you selected, not a name the client substituted.</p></div></div>

If a charge looks high, check whether the full history was sent again or whether an agent called tools repeatedly. The token counts in the log are the settlement basis.

## Where new credit comes from

Online top-ups and redemption codes both add wallet credit. Minimums and refunds are on [Top up and redeem](/guide/account/redeem/).

<div class="callout warning"><strong>Do not reconcile against numbers in this guide</strong><p>Unit prices, exchange rates, welcome credit, and the minimum top-up can be changed in the admin settings. Reconcile against Model Square, the confirmation page, and Usage Logs.</p></div>

## Related

- [Models and pricing](/guide/models-pricing/)
- [Usage and logs](/guide/account/usage/)
- [Error codes](/guide/errors/)
