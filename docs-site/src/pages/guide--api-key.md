# 创建和管理 API Key

API Key 是软件调用模型的凭证，统一在控制台 [API 密钥]({{CONSOLE_URL}}/keys) 页面管理。建议**一个软件单独建一把密钥**（如 `codex`、`cherry`、`zotero`），方便查账与随时停用。

## 3 步创建密钥

<div class="steps">
  <div class="step">
    <strong>第 1 步：点击「创建密钥」</strong>
    <p>打开 <a href="{{CONSOLE_URL}}/keys">{{CONSOLE_URL}}/keys</a>，点击右上角新建按钮。</p>
  </div>
  <div class="step">
    <strong>第 2 步：填写名称与分组（其余可保持默认）</strong>
    <p>• <strong>名称</strong>：填软件名（如 <code>cursor</code>、<code>claude-code</code>）。<br>
    • <strong>分组</strong>：保持默认或选择你要调用的模型所在分组（不确定可选 <code>auto</code> 自动按序匹配）。<br>
    • <strong>额度 / 过期时间</strong>：个人自用保持「无限额度 / 永不过期」即可；多人共享（如课题组 Open WebUI）建议设一个额度上限。</p>
  </div>
  <div class="step">
    <strong>第 3 步：复制 <code>sk-...</code> 密钥</strong>
    <p>保存后点击列表右侧复制按钮。复制出的密钥自带 <code>sk-</code> 前缀，填进客户端时直接粘贴，<strong>不要</strong>重复手打 <code>sk-</code> 或在客户端密码框里多写 <code>Bearer </code>。</p>
  </div>
</div>

## 高级限制项说明（按需开启）

| 选项 | 作用与注意事项 |
| --- | --- |
| **剩余额度** | 关闭「无限」后，这把密钥用完设定额度就会报 `401 无效的令牌`（哪怕钱包里还有钱）。 |
| **分组选择 `auto`** | 按你设定的分组顺序寻找可用线路；开启「跨分组重试」后，当前组失败会自动尝试下一组。 |
| **模型限制** | 勾选后该密钥只能调用指定模型 ID，适合给特定业务脚本锁权限。 |
| **IP 白名单** | 留空不限制；填写后非白名单 IP 调用会返回 `403 access_denied`（家用宽带/校园网 IP 会变，个人请留空）。 |

<div class="callout warning"><strong>安全提醒</strong><p>请勿将 <code>sk-...</code> 密钥写进前端网页代码、提交到公开 GitHub 仓库或截图发到群聊。一旦疑似泄露，直接在列表点击「禁用」或「删除」并换一把新密钥即可。</p></div>

## 相关

- [3 分钟快速开始](/guide/quickstart/)
- [分组与线路](/guide/account/groups/)
- [错误码](/guide/errors/)
