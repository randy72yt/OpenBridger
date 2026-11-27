# Zotero（论文阅读与文献问答）

Zotero 搭配 AI 插件（如 **Awesome GPT / Zotero GPT**）可以在阅读 PDF 论文时直接划词解释公式、提炼研究方法、对比多篇文献创新点。该类插件走标准 OpenAI 兼容接口，填入网关地址即可使用。

## 开始之前

- 已安装 Zotero 7（或 Zotero 6）及 Awesome GPT（`zotero-gpt`）插件。
- 在 [API 密钥]({{CONSOLE_URL}}/keys) 新建一把名为 `zotero` 的密钥。
- 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制一个长上下文对话模型 ID（如 `gpt-6.2-sol`、`claude-sonnet-5-5`、`gemini-4-neon`）；若开启本地 PDF 向量检索，再准备一个向量模型 ID（如 `text-embedding-3-small`）。

## 配置步骤

1. 打开 Zotero 顶部菜单「编辑（Edit）-> 设置（Settings）」（macOS 为 `Zotero -> Settings`），切到 **GPT** 标签页。
2. 按插件界面字段填入以下配置：

<div class="config-card">
  <div><span>API（接口地址）</span><code>{{CONSOLE_URL}}</code> 或 <code>{{API_BASE_URL}}</code>（见下方说明）</div>
  <div><span>Secret Key</span><code>sk-你的密钥</code></div>
  <div><span>Model（对话模型）</span><code>从模型广场复制的模型 ID</code></div>
  <div><span>Embedding Model（可选）</span><code>text-embedding-3-small</code>（需分组内已开放该向量模型）</div>
</div>

<div class="callout"><strong>关于 API 地址是否带 /v1</strong><p>主流版本的 <code>zotero-gpt</code> 会在你填的 <code>API</code> 后面自动拼接 <code>/v1/chat/completions</code>，此时应填站点根地址 <code>{{CONSOLE_URL}}</code>；如果所用衍生插件明确提示填写 OpenAI Base URL，则填 <code>{{API_BASE_URL}}</code>。填完点测试或在论文页面发一句 <code>hi</code>，到 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 看到记录即配置成功。</p></div>

## 推荐科研模型搭配

- **全文精读与长文献综述**：优先选支持超长上下文的模型（如 `gemini-4-neon`、`claude-sonnet-5-5`、`gpt-6.2-astra`）。
- **日常划词翻译与术语解释**：选择轻量快速模型（如 `gpt-6.2-sol`、`deepseek-v4.1-flash`），响应更快、单次消耗更省。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 报错 `404` | 切换 `{{CONSOLE_URL}}` 与 `{{API_BASE_URL}}` 两种写法，避免出现 `/v1/v1/` |
| 划词正常但点击「PDF 全文问答」报错 | 全文检索依赖 Embedding 模型，检查插件里的 Embedding 模型名称是否在 [模型广场]({{CONSOLE_URL}}/pricing) 可用 |
| `403` 用户额度不足 | 长论文一次会发送数万 Token，预扣额度较高，检查 [钱包]({{CONSOLE_URL}}/wallet) 余额 |

## 相关

- [沉浸式翻译（文献双语）](/guide/clients/immersive-translate/)
- [Embeddings 向量与 Rerank](/api/embeddings-rerank/)
- [错误码](/guide/errors/)
