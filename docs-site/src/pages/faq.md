# 常见问题

## 1. Base URL 到底填哪个？要不要带 `/v1`？

- **绝大多数软件（Codex、Cursor、Cherry Studio、Chatbox、LobeChat、OpenAI SDK）**：填 `{{API_BASE_URL}}`（带 `/v1`）。
- **Claude Code / Gemini CLI / Anthropic SDK**：填 `{{CONSOLE_URL}}`（**不带 `/v1`**）。
- **沉浸式翻译 / Coze HTTP 节点**：填完整路径 `{{API_BASE_URL}}/chat/completions`。
- 拿不准直接查 [Base URL 填写速查](/guide/base-urls/)。

## 2. 填密钥时前面要加 `Bearer` 吗？

- **在客户端图形界面的 API Key 输入框里**：只贴 `sk-...` 本身，**不要**加 `Bearer `。
- **自己写 HTTP 请求头或 `curl` 时**：写 `Authorization: Bearer sk-...`。

## 3. 新注册账号送体验金吗？怎么充值？

注册初始余额为 ¥0。前往控制台 [钱包]({{CONSOLE_URL}}/wallet) 使用支付宝、微信支付或兑换码充值即可立即调用，详见 [充值与兑换](/guide/account/redeem/)。

## 4. 客户端里点「获取模型列表」是空的怎么办？

很多软件会自动过滤它不认识的新模型名字。不用管自动获取，直接点客户端里的 **「+ 添加模型 / 自定义模型」**，去 [模型广场]({{CONSOLE_URL}}/pricing) 点击复制完整模型 ID 贴进去就能用。

## 5. 需要先退出 Cursor / Trae / CodeBuddy 原来的账号吗？

不需要。保留原账号登录完全不影响自定义 API Key 调用。

## 6. 怎么确认软件真的连上了 OpenBridger，而不是还在走旧模型？

在软件里新开一个对话窗口发一句短消息，然后打开控制台 [使用日志]({{CONSOLE_URL}}/usage-logs)，能看到刚才这把密钥新增了一条消耗记录，就说明 100% 走通了。

## 7. `401`、`403`、`404`、`503` 分别代表什么问题？

- `401`：API Key 填错了，或这把密钥自己设的限额到了。
- `403`：账户钱包没余额了（去 [钱包]({{CONSOLE_URL}}/wallet) 充值），或密钥分组无权限。
- `404`：Base URL 多写了或少写了 `/v1`。
- `503`：地址和密钥都通了，但模型名字打错了或密钥分组不含该模型。

完整对照表见 [错误码](/guide/errors/) 与 [故障排查](/troubleshooting/)。
