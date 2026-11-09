# Subscriptions

Plans are for sale. The list is on the [wallet]({{CONSOLE_URL}}/wallet) and the [subscriptions page]({{CONSOLE_URL}}/subscriptions). The name, price, validity, and credit on each card are what you buy. This page does not list prices.

## Buy one

<div class="steps"><div class="step"><strong>Open the plan list</strong><p>Use the wallet or the subscriptions page. Buy only a plan that is shown there.</p></div><div class="step"><strong>Read the confirmation</strong><p>After Subscribe Now, the dialog shows the plan name, validity period, plan quota, and amount due. A reset period or an upgraded group appears there when the plan has one.</p></div><div class="step"><strong>Choose how to pay</strong><p>When online payment is on, you can use the Alipay or WeChat method the wallet already enables. If the plan allows balance payment, the dialog has Pay with Balance. That button stays disabled when the balance is too low or the plan disallows it.</p></div><div class="step"><strong>Wait for the result</strong><p>Do not submit the same purchase again while payment is processing. After success, the wallet or subscriptions page shows the plan. If a plan sets a per-user purchase limit, you cannot buy it again once you reach that limit.</p></div></div>

The amount due is the figure in the dialog. A plan priced in yuan shows ¥. Do not pay against a price list from another page or a chat message.

## Billing preference

With no active subscription, calls use the wallet. After you buy a plan, the default is Subscription First.

| Preference | Meaning |
| --- | --- |
| Subscription First (default) | With no active subscription, the wallet is charged. With one, subscription credit is used first. The wallet is used afterward only when that subscription allows wallet overflow. Otherwise the error is insufficient subscription credit. |
| Wallet First | The wallet is charged first. If wallet credit is insufficient, the gateway tries the subscription. |
| Subscription Only | Subscription credit only. If it is not enough, HTTP `403` and a message that subscription credit is insufficient or no subscription is configured. |
| Wallet Only | Wallet only. |

If the wallet lets you switch, the selected option applies to later requests.

## Validity, credit, and reset

The validity period in the dialog is the length of this purchase, such as a number of months or days. If the plan says it changes your group on purchase or on expiry, that note on the plan is the one to follow.

A reset period other than no reset means the plan credit is restored on that schedule. It is not another charge. The dialog has no "charge again at expiry" line. The code has no switch that bills Alipay or WeChat again when the period ends.

What happens when plan credit runs out depends on the billing preference and on whether that plan allows wallet overflow. See the table above.

## Related

- [Credits and billing](/guide/billing/)
- [Top up and redeem](/guide/account/redeem/)
- [Terms of service](/legal/terms/)
