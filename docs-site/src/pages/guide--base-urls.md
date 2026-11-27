# Base URL 填写速查

不同客户端拼接接口路径的规则不一样：有的会自动补上 `/v1/chat/completions`，有的会自动补上 `/v1/messages`，还有的工作流节点要求你填完整路径。如果填错，最常见的现象就是 `404`（路径变成了 `/v1/v1/...`）或返回空白网页。

接入任何新工具前，对照下面这张表选对格式。

## 三种地址写法

<div class="config-card">
  <div><span>写法 A：站点根地址（不带 /v1）</span><code>{{CONSOLE_URL}}</code></div>
  <div><span>写法 B：OpenAI 标准 Base URL（带 /v1）</span><code>{{API_BASE_URL}}</code></div>
  <div><span>写法 C：完整对话端点路径</span><code>{{API_BASE_URL}}/chat/completions</code></div>
</div>

## 常用工具与 SDK 对照表

| 客户端 / SDK | 该填哪一种 | 实际填写示例 | 鉴权方式 |
| --- | --- | --- | --- |
| **OpenAI Python / Node SDK** | 写法 B（带 `/v1`） | `{{API_BASE_URL}}` | `Authorization: Bearer sk-...` |
| **OpenAI Codex CLI** | 写法 B（带 `/v1`） | `{{API_BASE_URL}}`（`wire_api = "responses"`） | 环境变量 `OPENBRIDGER_API_KEY` |
| **Claude Code** | 写法 A（**不带** `/v1`） | `{{CONSOLE_URL}}` | `ANTHROPIC_AUTH_TOKEN=sk-...` |
| **Anthropic 官方 SDK** | 写法 A（**不带** `/v1`） | `{{CONSOLE_URL}}` | `x-api-key: sk-...` |
| **Google Gemini CLI** | 写法 A（**不带** `/v1`） | `{{CONSOLE_URL}}` | `GEMINI_API_KEY=sk-...` |
| **Cursor / Cline / Roo Code / Trae** | 写法 B（OpenAI 兼容模式） | `{{API_BASE_URL}}` | API Key 填 `sk-...` |
| **Cherry Studio / Chatbox / LobeChat** | 写法 A 或 B（看界面预览提示） | 界面预览末尾显示 `/v1/chat/completions` 即可 | API Key 填 `sk-...` |
| **沉浸式翻译** | 写法 C（自定义 OpenAI 完整路径） | `{{API_BASE_URL}}/chat/completions` | APIKEY 填 `sk-...` |
| **Zotero GPT** | 写法 B（`api` 字段通常填根或 `/v1`） | 按插件版本见 [Zotero](/guide/clients/zotero/) | Secret Key 填 `sk-...` |
| **Dify（OpenAI-API-compatible）** | 写法 B（带 `/v1`） | `{{API_BASE_URL}}` | API Key 填 `sk-...` |
| **n8n（OpenAI 节点）** | 写法 B（带 `/v1`） | `{{API_BASE_URL}}` | Credential API Key `sk-...` |
| **扣子 Coze / 通用 HTTP 请求节点** | 写法 C（完整 URL） | `{{API_BASE_URL}}/chat/completions` | Header `Authorization: Bearer sk-...` |

<div class="callout"><strong>一秒自检技巧</strong><p>大多数桌面客户端（如 Cherry Studio、Chatbox）在输入框下方会小字展示「最终请求地址预览」。只要预览结果是 <code>{{API_BASE_URL}}/chat/completions</code>（既没有少 <code>/v1</code>，也没有出现两个 <code>/v1/v1</code>），地址就是对的。</p></div>

## 排查地址填错的典型报错

| 报错现象 | 原因 | 修正方法 |
| --- | --- | --- |
| `404 page not found` 且日志里没有记录 | 多写了一层 `/v1`（变成 `/v1/v1/chat/completions`），或少写了 `/v1` | 对照上表切换写法 A / B / C |
| 返回了一整段 HTML 网页源码 | 把根地址 `{{CONSOLE_URL}}` 填进了不会自动补路径的 HTTP 工具，打到了前端网页 | 改成完整路径 `{{API_BASE_URL}}/chat/completions` |
| Claude Code 报 `404` | `ANTHROPIC_BASE_URL` 末尾多加了 `/v1` | 改成 `{{CONSOLE_URL}}` |
| `503` 文案含「无可用渠道」 | **地址已经通了**，是模型 ID 写错或当前密钥分组没有该模型线路 | 去 [模型广场]({{CONSOLE_URL}}/pricing) 核对完整模型 ID 与密钥分组 |

## 相关

- [快速开始](/guide/quickstart/)
- [OpenAI Codex](/guide/clients/codex-cli/)
- [Claude Code](/guide/clients/claude-code/)
- [错误码](/guide/errors/)
