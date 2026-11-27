# 3 分钟快速开始

不管你要配编程工具、翻译插件还是日常聊天客户端，准备工作都只有下面 **4 步**。

## 上手 4 步

<div class="steps">
  <div class="step">
    <strong>第 1 步：注册并登录控制台</strong>
    <p>打开 <a href="{{CONSOLE_URL}}">{{CONSOLE_URL}}</a> 注册账号并登录。</p>
  </div>
  <div class="step">
    <strong>第 2 步：充值或兑换额度</strong>
    <p>打开 <a href="{{CONSOLE_URL}}/wallet">钱包</a>（新注册账号初始余额为 ¥0）。选择面额在线支付，或输入兑换码充值（详见 <a href="/guide/account/redeem/">充值与兑换</a>）。哪怕先充最低档，也足够完成大量日常对话测试。</p>
  </div>
  <div class="step">
    <strong>第 3 步：创建一把 API Key</strong>
    <p>打开 <a href="{{CONSOLE_URL}}/keys">API 密钥</a> -> 点击右上角 <strong>「创建密钥」</strong> -> 填一个好记的名称（例如 <code>codex</code> 或 <code>cherry</code>）-> 点击确认，复制生成的 <code>sk-...</code> 完整密钥。</p>
  </div>
  <div class="step">
    <strong>第 4 步：开始使用（任选一种）</strong>
    <p><strong>• 不装软件直接在网页聊</strong>：打开控制台左侧 <a href="{{CONSOLE_URL}}/playground">GPT 对话</a>，选模型直接提问。<br>
    <strong>• 接入常用客户端（推荐）</strong>：去 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 点复制你想用的模型完整 ID，然后按下方对应教程填入即可：</p>
  </div>
</div>

### 热门客户端直达

- **写代码**：[CC Switch（新手一键配置）](/guide/clients/cc-switch/) · [Codex（桌面版 / 插件 / CLI）](/guide/clients/codex-cli/) · [Claude Code](/guide/clients/claude-code/) · [Cursor](/guide/clients/cursor/) · [Cline / Roo Code](/guide/clients/cline/)
- **看论文 / 聊天**：[沉浸式翻译](/guide/clients/immersive-translate/) · [Zotero](/guide/clients/zotero/) · [Cherry Studio](/guide/clients/cherry-studio/) · [Chatbox](/guide/clients/chatbox/) · [LobeChat](/guide/clients/lobechat/)

---

## 附：开发者命令行 10 秒连通测试（可选）

如果你写代码或习惯用终端，可在终端执行一条 `curl` 验证密钥与模型：

```bash
curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer sk-替换成你的密钥" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-6.2-sol",
    "messages": [{"role": "user", "content": "回复 pong"}]
  }'
```

返回 JSON 中出现 `choices[0].message.content`，且在 [使用日志]({{CONSOLE_URL}}/usage-logs) 能看到这条记录，即代表完全跑通。Python / Node.js 代码示例见 [Python / Node SDK 接入](/guide/clients/openai-compatible/)。

## 新手最常见的 3 个报错

| 报错提示 | 原因 | 解决方法 |
| --- | --- | --- |
| `401 无效的令牌` | 密钥复制漏了字符、前后多了空格，或重复写了 `Bearer` | 回 [API 密钥]({{CONSOLE_URL}}/keys) 重新点击复制按钮粘贴 |
| `403 用户额度不足` | 账号余额为 ¥0，尚未充值或兑换 | 前往 [钱包]({{CONSOLE_URL}}/wallet) 充值或兑换额度 |
| `503 无可用渠道` | 手打了缩写的模型名，或当前密钥分组不含此模型 | 去 [模型广场]({{CONSOLE_URL}}/pricing) 点击模型名一键复制完整 ID |

更多状态码见 [错误码](/guide/errors/)。
