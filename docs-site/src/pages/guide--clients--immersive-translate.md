# 沉浸式翻译（文献与网页双语对照）

沉浸式翻译（Immersive Translate）支持外文网页、ArXiv 论文以及本地 PDF 文献的双语对照翻译，保留排版与数学公式。接入 OpenBridger 后，你可以按需选用响应快、性价比高的轻量大模型完成全文翻译。

## 开始之前

- 浏览器已安装「沉浸式翻译」扩展。
- 在 [API 密钥]({{CONSOLE_URL}}/keys) 单独创建一把名为 `immersive-translate` 的密钥。
- 在 [模型广场]({{CONSOLE_URL}}/pricing) 挑选一个**非深度思考**的快速模型（如 `gpt-6.2-sol`、`gemini-4-hydrogen`、`deepseek-v4.1-flash`），复制完整模型 ID。

<div class="callout warning"><strong>模型选择建议</strong><p>网页与 PDF 翻译追求毫秒级响应，请勿选择默认开启长思考链的推理模型，否则每个段落都会先生成一大段思考 Token，既慢又增加不必要的消耗。</p></div>

## 配置步骤

<div class="steps">
  <div class="step"><strong>打开翻译服务设置</strong><p>点击浏览器工具栏的沉浸式翻译图标 ->「设置」->「翻译服务」，找到 <strong>OpenAI</strong>（或「自定义 OpenAI 兼容服务」）并点击配置。</p></div>
  <div class="step"><strong>填写密钥与完整接口地址</strong><p>沉浸式翻译的自定义地址框要求填写<strong>完整对话路径</strong>（不要只填根域名）：</p></div>
  <div class="step"><strong>填入模型 ID 并测试</strong><p>在「模型（Model）」中选择自定义模型，粘贴从模型广场复制的完整模型 ID，点击页面右上角「验证服务」确认通过。</p></div>
</div>

<div class="config-card">
  <div><span>APIKEY</span><code>sk-你的密钥</code></div>
  <div><span>自定义 API 接口地址</span><code>{{API_BASE_URL}}/chat/completions</code></div>
  <div><span>自定义模型名称</span><code>从模型广场复制的快速模型 ID</code></div>
</div>

## 整篇 PDF 论文翻译防限流设置

翻译几十页的英文 PDF 时，插件默认会瞬间并发大量请求。为避免触发速率限制（`429`），建议在同一设置页展开「更多设置 / 限流设置」调整：

- **每秒最大请求数**：设为 `3` ～ `5`
- **每次请求最大段落数**：保持默认或适当调大（合并段落可显著减少请求次数与系统提示词重复计费）

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 点「验证」报 `404` | 检查接口地址是否漏写了末尾的 `/v1/chat/completions`（见 [Base URL 填写速查](/guide/base-urls/)） |
| 报 `503` 无可用渠道 | 自定义模型名称手打错了，从 [模型广场]({{CONSOLE_URL}}/pricing) 重新复制 |
| 翻译中途部分段落报 `429` | PDF 并发过高，调低每秒最大请求数后重试未译段落 |

## 相关

- [Zotero（论文阅读）](/guide/clients/zotero/)
- [Base URL 填写速查](/guide/base-urls/)
- [错误码](/guide/errors/)
