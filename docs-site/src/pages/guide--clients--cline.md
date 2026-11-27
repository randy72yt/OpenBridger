# 在 Cline / Roo Code 中使用 OpenBridger

Cline 与 Roo Code 是 VS Code 中开源免费的自主编程 Agent 插件，支持自动读写文件、执行终端命令与多步调试。接入 OpenBridger 有两种常用协议方式：**OpenAI Compatible**（通用性最强，可切换全系模型）或 **Anthropic**（直连原生 Messages 协议）。

## 开始之前

- VS Code 已安装 **Cline** 或 **Roo Code** 扩展。
- 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建一把名为 `cline` 的密钥。
- 在 [模型广场]({{CONSOLE_URL}}/pricing) 选择**支持工具调用（Tool Calling / Function Calling）**的编程模型（如 `claude-opus-5-5`、`claude-sonnet-5-5`、`gpt-6.2-astra`、`deepseek-v4.1`）。

<div class="callout warning"><strong>关键要求：必须支持工具调用</strong><p>Cline / Roo Code 读写代码文件依赖模型的工具调用能力。请确保所选模型与密钥分组支持原生 Tool Calling，否则插件无法自动修改文件。</p></div>

## 方式一：OpenAI Compatible 模式（推荐，可自由切换各家模型）

1. 打开插件右上角齿轮进入设置（Settings）。
2. **API Provider** 选择 `OpenAI Compatible`。
3. 按下表填写：

<div class="config-card">
  <div><span>Base URL</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Key</span><code>sk-你的密钥</code></div>
  <div><span>Model ID</span><code>claude-sonnet-5-5</code> 或 <code>gpt-6.2-astra</code>（从模型广场复制）</div>
</div>

如果所选模型具备识图能力，勾选下方的 **Supports Images**，即可在对话框里直接拖入网页报错截图让 Agent 对照修复。

## 方式二：Anthropic 原生协议模式（专用于 Claude 模型）

如果你在当前分组下只跑 Claude 系列模型并希望使用原生 Prompt 缓存与 Extended Thinking：

1. **API Provider** 选择 `Anthropic`。
2. 勾选 **Use custom base URL**。
3. **Base URL** 填站点根地址 `{{CONSOLE_URL}}`（**不要**带 `/v1`）。
4. **Anthropic API Key** 填 `sk-你的密钥`，并选择模型广场中对应的 Claude 模型 ID。

## 验证与常见问题

发一句简单的任务（如「读取当前目录下的 README 并用两句话总结」），随后到 [使用日志]({{CONSOLE_URL}}/usage-logs) 核对请求状态与缓存命中情况。

| 现象 | 处理 |
| --- | --- |
| `404` | OpenAI Compatible 模式填 `{{API_BASE_URL}}`；Anthropic 模式填 `{{CONSOLE_URL}}`，切勿混用 |
| `503` 无可用渠道 | 插件默认选中的模型名与模型广场里的完整 ID 不一致，手动粘贴完整模型 ID |
| Agent 只输出代码块、不自动改文件 | 当前选用的轻量模型不支持工具调用，切换到旗舰编程模型 |

## 相关

- [Cursor](/guide/clients/cursor/)
- [OpenAI Codex](/guide/clients/codex-cli/)
- [Claude Code](/guide/clients/claude-code/)
