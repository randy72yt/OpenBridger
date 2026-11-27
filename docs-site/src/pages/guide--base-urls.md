# Base URL 填写速查

不同软件自动拼接接口路径的规则不同。填错最典型的现象就是报 `404`（拼成了 `/v1/v1/...`）或返回网页源码。接入前对照下表一键复制即可。

## 三种地址格式

<div class="config-card">
  <div><span>格式 A：根地址（不带 /v1）</span><code>{{CONSOLE_URL}}</code></div>
  <div><span>格式 B：标准 Base URL（带 /v1）</span><code>{{API_BASE_URL}}</code></div>
  <div><span>格式 C：完整对话路径</span><code>{{API_BASE_URL}}/chat/completions</code></div>
</div>

## 常用软件填写对照表

| 软件 / SDK | 选哪种 | 直接复制填入的值 | 备注 |
| --- | --- | --- | --- |
| **Codex（桌面版 / VS Code / CLI）** | 格式 B | `{{API_BASE_URL}}` | 协议选 `responses`，密钥写进 `auth.json` 或用 [CC Switch](/guide/clients/cc-switch/) |
| **Claude Code / Anthropic SDK** | 格式 A | `{{CONSOLE_URL}}` | **千万别带 `/v1`**，客户端会自动拼 `/v1/messages` |
| **Google Gemini CLI / GenAI SDK** | 格式 A | `{{CONSOLE_URL}}` | **不要带 `/v1` 或 `/v1beta`** |
| **Cursor / Cline / Roo Code / Trae / CodeBuddy** | 格式 B | `{{API_BASE_URL}}` | 选 OpenAI Compatible 模式 |
| **Cherry Studio / Chatbox / LobeChat** | 格式 B | `{{API_BASE_URL}}` | 输入框下方灰色预览显示 `.../v1/chat/completions` 即正确 |
| **Open WebUI / Dify / n8n** | 格式 B | `{{API_BASE_URL}}` | OpenAI API 连接模式 |
| **沉浸式翻译** | 格式 C | `{{API_BASE_URL}}/chat/completions` | 自定义 OpenAI 接口地址必填完整路径 |
| **扣子 Coze / 通用 HTTP 节点** | 格式 C | `{{API_BASE_URL}}/chat/completions` | Header 带 `Authorization: Bearer sk-...` |
| **Zotero GPT** | 格式 A 或 B | `{{CONSOLE_URL}}`（部分衍生版填 `{{API_BASE_URL}}`） | 报 `404` 时切换这两种即可 |
| **OpenAI Python / Node SDK / LangChain** | 格式 B | `{{API_BASE_URL}}` | SDK 自动补 `/chat/completions` 或 `/responses` |

## 报错 10 秒定位

| 现象 | 原因与改法 |
| --- | --- |
| 报 `404` 且日志里没记录 | 多写了 `/v1`（变成 `/v1/v1/...`）或少了 `/v1`，按上表切换格式 A / B / C |
| 返回了一大段 HTML 代码 | 工具不会自动补路径，改成格式 C：`{{API_BASE_URL}}/chat/completions` |
| 报 `503 无可用渠道` | **地址已经填对了！** 是模型名称写错或密钥分组无此模型，去 [模型广场]({{CONSOLE_URL}}/pricing) 重新复制完整 ID |

## 相关

- [CC Switch（新手一键配置）](/guide/clients/cc-switch/)
- [错误码](/guide/errors/)
