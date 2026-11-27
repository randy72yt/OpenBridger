# 额度与计费

控制台中的 [钱包]({{CONSOLE_URL}}/wallet)、[模型广场]({{CONSOLE_URL}}/pricing) 与 [使用日志]({{CONSOLE_URL}}/usage-logs) 全部直接以 **人民币（¥）** 展示金额与消耗，所见即所得。

## 一次请求怎么扣费？

每次调用按 **「预扣 -> 实际调用 -> 多退少补结算」** 三步自动完成：

<div class="steps">
  <div class="step"><strong>1. 调用前预扣</strong><p>发送请求时，系统先按预估上限冻结一小笔额度。如果余额连预扣都不够，直接返回 <code>403</code>（<code>insufficient_user_quota</code>，提示「用户额度不足」），不会产生扣费。</p></div>
  <div class="step"><strong>2. 转发模型响应</strong><p>请求若因上游异常报错失败，预扣额度自动全额退回。</p></div>
  <div class="step"><strong>3. 按实际用量结算</strong><p>生成结束后，按本次真实的输入 Token、输出 Token（含推理模型的思考 Token）、缓存命中率及分组倍率算出最终金额，多退少补，并写入 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a>。</p></div>
</div>

## 钱包与订阅套餐的扣费顺序

- 没有购买订阅套餐时，直接扣 **钱包余额**。
- 同时有订阅套餐和钱包余额时，默认 **优先扣订阅套餐额度**；也可以在钱包页手动切换为「优先钱包」「仅用订阅」或「仅用钱包」（详见 [订阅套餐](/guide/account/subscriptions/)）。
- 注意区分两种报错：**钱包/订阅没钱**返回 `403 用户额度不足`；而你给某把 **API Key 单独设的限额用完**返回 `401 无效的令牌`。

## 为什么有时候觉得消耗变快了？

最常见的原因有两个，打开 [使用日志]({{CONSOLE_URL}}/usage-logs) 看一眼每条请求的 **提示（输入）Token** 与 **补全（输出）Token** 就能定位：
1. **一个对话窗口聊得太长**：客户端每次发新消息都会把前面几十轮历史记录（甚至整份代码文件）重新发一遍。开启新话题时顺手点 **「新建对话」** 最省额度。
2. **编程 Agent 自动多轮调工具**：Codex、Claude Code、Cline 在改一个需求时，后台可能自动连续发起 5～15 次读写文件请求。

## 相关

- [模型与价格](/guide/models-pricing/)
- [充值与兑换](/guide/account/redeem/)
- [用量与日志](/guide/account/usage/)
