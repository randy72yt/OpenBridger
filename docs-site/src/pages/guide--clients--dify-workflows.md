# Dify / n8n / 扣子（Coze）工作流接入

无论是自托管的 **Dify** 知识库应用、**n8n** 自动化流水线，还是 **扣子（Coze）** 的 HTTP 请求节点，都可以通过标准接口接入 OpenBridger 的文本、识图、向量与视频能力。

## 一、Dify 添加模型（LLM / Embedding / Rerank）

1. 进入 Dify 右上角「设置 -> 模型供应商」，安装官方的 **OpenAI-API-compatible**（或 **New API**）插件。
2. 点击「添加模型」，按模型类型填写：

<div class="config-card">
  <div><span>模型类型</span><code>LLM</code> / <code>Text Embedding</code> / <code>Rerank</code></div>
  <div><span>模型名称（Model Name）</span>从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制完整 ID（如 <code>gpt-6.2-sol</code>）</div>
  <div><span>API Key</span><code>sk-你的密钥</code></div>
  <div><span>API Endpoint URL</span><code>{{API_BASE_URL}}</code>（带 <code>/v1</code>）</div>
</div>

如果是支持识图的模型，勾选「Vision / 视觉支持」；支持工具调用的勾选「Function Calling」。

<div class="callout"><strong>Docker 部署网络提示</strong><p>Dify 容器直接请求公网的 <code>{{API_BASE_URL}}</code> 即可；只有在你把网关和 Dify 跑在同一台内网机器上时，才需要注意不要填宿主机的 <code>127.0.0.1</code>。</p></div>

---

## 二、n8n 工作流配置

1. 在工作流中添加 **OpenAI** 节点（或 AI Agent 下的 **OpenAI Chat Model** 子节点）。
2. 新建 **OpenAI API** 凭据（Credential）：
   - **API Key**：`sk-你的密钥`
   - **Base URL**：将默认的 `https://api.openai.com/v1` 替换为 `{{API_BASE_URL}}`
3. 在节点的 Model 参数中切为 **Expression / 手动输入 ID**，填入模型广场中的完整模型 ID（避免下拉列表只过滤官方旧模型名）。

---

## 三、扣子（Coze）或通用 HTTP 节点调用

在不支持内置自定义供应商的云端工作流平台中，直接使用 **HTTP 请求节点** 最稳定：

### 1. 同步文本 / 识图节点配置

- **请求方法**：`POST`
- **完整请求地址**：`{{API_BASE_URL}}/chat/completions`
- **Headers**：
  - `Authorization`: `Bearer sk-你的密钥`（建议放在工作流密钥变量中引用）
  - `Content-Type`: `application/json`
- **Body（JSON）**：传入 `"stream": false` 以便工作流下游节点直接通过 `choices[0].message.content` 取值。

### 2. 异步视频生成轮询范式（`/v1/videos`）

在工作流里串联异步视频生成时，按以下 4 个节点编排即可：

1. **创建任务（HTTP POST）**：请求 `{{API_BASE_URL}}/videos`，从响应 JSON 提取任务 `id`。
2. **循环节点（Loop）**：设置最大循环次数（例如 40 次）。
3. **等待 + 查询状态（Wait 5s -> HTTP GET）**：每次循环先等待 5 秒，再请求 `{{API_BASE_URL}}/videos/{{id}}`。
4. **条件跳出（Break）**：当返回状态变为已完成或失败时跳出循环，输出最终视频地址。

## 相关

- [Chat Completions 对话与识图](/api/openai-chat/)
- [图像、视频与语音接口](/api/multimodal-media/)
- [Embeddings 向量与 Rerank](/api/embeddings-rerank/)
