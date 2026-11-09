# Subscriptions

Terms v1.1 on the main site say **plan purchase and auto-renewal are not open yet.** If the wallet or the [subscriptions page]({{CONSOLE_URL}}/subscriptions) shows nothing you can buy, there is no plan for sale. Do not pay against a price list from somewhere else.

The gateway code already has subscription credit, plan records, and a billing preference. Those apply to buyers only after an operator publishes plans and the terms are updated. This page lists no plan names and no prices.

## Billing preference, when the wallet shows it

With no subscription credit, calls use the wallet. Once the account has an active subscription, the default is Subscription First.

| Preference | Meaning |
| --- | --- |
| Subscription First (default) | With no active subscription, the wallet is charged. With one, subscription credit is used first. The wallet is used afterward only when that subscription allows wallet overflow. Otherwise the error is insufficient subscription credit. |
| Wallet First | The wallet is charged first. If wallet credit is insufficient, the gateway tries the subscription. |
| Subscription Only | Subscription credit only. If it is not enough, HTTP `403` and a message that subscription credit is insufficient or no subscription is configured. |
| Wallet Only | Wallet only. |

If the wallet lets you switch, the selected option applies to later requests.

## If purchases open later

Trust the confirmation page and the terms in force. At least check:

- Price, currency, validity, and included credit
- Models and groups that are included
- Whether the wallet is charged after the plan credit runs out
- Whether renewal is automatic, and where to cancel
- What happens to unused credit at expiry

Do not pay twice while an order is processing. If a purchase fails, Usage Logs show whether the problem is credit or a model outside the plan.

<div class="callout"><strong>Do not assume auto-renewal today</strong><p>The terms say auto-renewal is not open. If the page does not say auto-renewal, you will not be charged on a cycle.</p></div>

## Related

- [Credits and billing](/guide/billing/)
- [Top up and redeem](/guide/account/redeem/)
- [Terms of service](/legal/terms/)
