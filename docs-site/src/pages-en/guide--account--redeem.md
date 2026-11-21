# Top up and redeem

The wallet is [{{CONSOLE_URL}}/wallet]({{CONSOLE_URL}}/wallet). Two ways add credit: the checkout the wallet itself opens, or a redemption code. A payment button that is not on the page is not enabled. Do not pay through a link from a chat message.

<div class="callout"><strong>Screenshot to add</strong><p>Wallet balance, amount choices, Alipay and WeChat buttons, the redemption entry, and the balance after a successful payment. Replace this note with the current UI.</p></div>

## Before you start

- You are signed in to the account you want to fund.
- You use only the checkout opened from the wallet.
- You can pay with Alipay or WeChat. The code defaults include Alipay (`alipay`) and WeChat (`wxpay`). Any extra button is an additional configured method. Use only the buttons the wallet shows.

When online payment is enabled, the wallet shows top-up. When it is disabled, only redemption remains and creating an order is rejected.

## Pay online

<div class="steps"><div class="step"><strong>Open the wallet</strong><p>Confirm the account. When top-up is available the card title is Add Funds.</p></div><div class="step"><strong>Pick an amount</strong><p>Amounts are in yuan. The code default minimum is ¥10, and the default shortcuts are ¥10, ¥20, ¥50, ¥100, ¥200, and ¥500. If an administrator changed the list or the floor, the wallet page wins. A smaller order fails with a message that the amount is below the minimum.</p></div><div class="step"><strong>Read the confirmation</strong><p>The yuan on the confirmation page is what the cashier charges. That payment buys US-dollar credit at the recharge price (Price): amount divided by Price. The wallet, Model Square, and logs then show that credit in yuan using the display rate (USDExchangeRate). Those two settings are not the same number, so the yuan you pay and the yuan added to the balance can differ. Compare the confirmation page with the balance after payment. Do not multiply by a rate yourself, and this guide does not publish the current values.</p></div><div class="step"><strong>Finish payment</strong><p>You return to <code>/wallet?pay=success</code>. Refresh the balance and the top-up history. Do not pay the same order again while it is still processing.</p></div></div>

Payment uses the epay protocol to the Alipay or WeChat checkout you see. Support will not ask for a payment password, an SMS code, or a full card number.

## Redemption codes

<div class="steps"><div class="step"><strong>Use a code OpenBridger issued</strong><p>Ignore codes from unknown sources.</p></div><div class="step"><strong>Redeem in the wallet</strong><p>Submit the whole code from the Redeem action.</p></div><div class="step"><strong>Check the result</strong><p>A success increases the balance and the history. On failure, do not send the full code to anyone, and do not hammer the button.</p></div></div>

## Paid, but the balance did not move

1. See whether the checkout says success, processing, cancelled, or failed.
2. Return to the wallet and refresh. The notification can arrive after the page.
3. Do not pay again just to refresh the balance.
4. If it is still missing, email [support@openbridger.com](mailto:support@openbridger.com) with the account email, time, amount, method, and any order id you can see. Do not include a payment password or verification code.

## Refunds

The [terms]({{CONSOLE_URL}}/user-agreement) on the main site control this. Version 1.1 says:

- A top-up becomes credit. Credit you already spent is not refunded.
- Unused top-up balance can be requested by email. After review it is returned on the original payment path.
- Fees the payment channel already kept are not refunded.
- You get a result within 7 business days.

## Related

- [Credits and billing](/guide/billing/)
- [Terms of service](/legal/terms/)
- [Contact and incidents](/guide/support/)
