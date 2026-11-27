# 模型与价格

所有当前可用模型、支持协议与实时单价均以控制台 [模型广场]({{CONSOLE_URL}}/pricing) 为准。

<div class="actions"><a class="btn primary" href="{{CONSOLE_URL}}/pricing">打开模型广场</a></div>

## 选模型时看清这 3 件事

1. **必须复制完整模型 ID**
   - 在客户端填写模型时，直接在模型广场**点击模型名一键复制**（包括大小写、连字符与后缀，例如 `gpt-6.2-astra`、`claude-sonnet-5-5`），不要自己简写成 `gpt-6` 或 `claude`。
2. **协议要和你的软件匹配**
   - **通用聊天 / 翻译软件**（Cherry Studio、Chatbox、沉浸式翻译、Cursor）：走 `/v1/chat/completions` 对话协议。
   - **Codex**：需要模型及分组支持 `/v1/responses` 协议。
   - **Claude Code**：需要模型及分组支持 `/v1/messages` 协议。
3. **按任务挑轻重，省钱又快**
   - **架构设计 / 复杂 Debug / 论文深度推导**：选旗舰或深度思考模型（如 `gpt-6.2-astra`、`claude-opus-5-5`、`gemini-4-carbon`）。
   - **网页双语翻译 / 日常划词 / 简单脚本**：选轻量高速模型（如 `gpt-6.2-sol`、`deepseek-v4.1-flash`、`gemini-4-hydrogen`），首字响应快且单次消耗低得多。

## 为什么客户端拉不到模型列表？

不少客户端不会主动拉取 `/v1/models`，或者会自动隐藏它不认识的新模型名字。遇到列表为空时无需排障，直接在客户端点 **「+ 手动添加模型」**，把模型广场复制的完整 ID 贴进去即可正常调用。

## 相关

- [额度与计费](/guide/billing/)
- [分组与线路](/guide/account/groups/)
- [错误码](/guide/errors/)
