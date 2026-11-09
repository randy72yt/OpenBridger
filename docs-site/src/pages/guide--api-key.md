# 创建和管理 API Key

API Key 是程序和客户端调用网关的凭证。控制台页面叫「API 密钥」，地址是 [{{CONSOLE_URL}}/keys]({{CONSOLE_URL}}/keys)。

## 开始之前

- 已经登录控制台。
- 知道这把密钥要给哪一个工具。一个工具一把密钥，停用和查账都更清楚。

## 创建

<div class="steps"><div class="step"><strong>打开 API 密钥</strong><p>进入 <a href="{{CONSOLE_URL}}/keys">{{CONSOLE_URL}}/keys</a>，新建密钥。</p></div><div class="step"><strong>填写名称</strong><p>名称最多 50 个字符，用工具或项目名，例如 <code>cursor</code>、<code>claude-code</code>、<code>server-prod</code>。</p></div><div class="step"><strong>按需限制范围</strong><p>页面可以设置分组、模型范围、密钥额度、过期时间和 IP 白名单。不确定就先保持默认，用短请求验证，再收紧。</p></div><div class="step"><strong>复制完整密钥</strong><p>控制台展示和复制的值带 <code>sk-</code> 前缀。请求时写成 <code>Authorization: Bearer sk-...</code>。网关会去掉这个前缀再查密钥，所以不要自己再加一次 <code>sk-</code>。</p></div></div>

<div class="callout"><strong>截图待补</strong><p>新建密钥抽屉：名称、分组、额度、过期时间、模型限制、IP 白名单。由运营用当前界面替换本段。</p></div>

创建数量有上限，达到上限时接口会拒绝再创建。具体上限由运营配置，页面报错里的数字就是当前值。

登录状态下可以再次查看某把密钥的完整值。这不表示它可以出现在仓库、前端包、截图或工单里。

## 页面上这几项实际会生效

| 设置 | 行为 |
| --- | --- |
| 额度不限 / 剩余额度 | 关闭「不限」后，这把密钥有自己的剩余额度。用尽后调用返回 `401`「无效的令牌」，即使钱包里还有余额。 |
| 过期时间 | 到期后同样是 `401`「无效的令牌」。`-1` 表示不过期。 |
| 模型限制 | 打开后，密钥只能调用列出的模型 ID。 |
| 分组 | 留空则用账户分组。填了分组时，该分组必须在你的可用分组里，并且没有被弃用，否则 `403`。 |
| 分组选 `auto` | 按你排列的分组顺序尝试。可以打开「跨分组重试」：当前分组的线路失败时再试下一组。 |
| IP 白名单 | 留空不限制。填写后，来源 IP 不在列表里会 `403`，`code` 为 `access_denied`。 |
| 禁用 | 禁用后的密钥不能调用。只读的用量查询接口同样拒绝已禁用的密钥。 |

普通用户不能在密钥后面加 `-渠道ID` 来指定上游渠道，那样会 `403`，文案是「普通用户不支持指定渠道」。

## 建议

<div class="grid"><div class="card"><h3>按用途拆开</h3><p>Cursor、Claude Code、线上服务各用一把。泄露或离职时只停一把。</p></div><div class="card"><h3>服务端持有</h3><p>不要把密钥放进浏览器、App 安装包或公开仓库。网页产品由你的服务器转发出去。</p></div><div class="card"><h3>轮换</h3><p>泄露、人员变动或项目结束时，禁用或删除旧密钥，再发一把新的。支持人员不会向你要完整密钥或密码。</p></div></div>

## 验证

```bash
curl "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer sk-你的密钥"
```

返回模型列表，说明这把密钥和网络是通的。列表是这把密钥当前分组能看到的模型，不是上游的全部目录。

## 相关

- [快速开始](/guide/quickstart/)
- [分组与线路](/guide/account/groups/)
- [错误码](/guide/errors/)
