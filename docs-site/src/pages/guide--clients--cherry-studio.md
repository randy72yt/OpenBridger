# Cherry Studio 配置教程

在 Cherry Studio 中添加 OpenAI 兼容服务商，即可调用对话、识图、推理及本地 PDF 知识库（RAG）。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `cherry-studio` 的密钥，复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID（如 `gpt-6.2-astra`、`claude-sonnet-5-5`、`gemini-4-neon`）。

## 配置 5 步

<div class="steps">
  <div class="step"><strong>第 1 步：打开「模型服务」</strong><p>点击左下角 <strong>⚙️ 设置</strong> -> 左侧点 <strong>「模型服务」</strong>。</p></div>
  <div class="step"><strong>第 2 步：新增服务商并打开开关</strong><p>拉到服务商列表底部点 <strong><code>+ 添加</code></strong>，名称填 <code>OpenBridger</code>，类型选 <strong><code>OpenAI</code></strong>，确定后<strong>打开右侧总启用开关</strong>。</p></div>
  <div class="step"><strong>第 3 步：填密钥与地址（核对下方灰色预览！）</strong>
    <div class="config-card">
      <div><span>API 密钥</span>粘贴 <code>sk-...</code> 完整密钥</div>
      <div><span>API 地址</span>填 <code>{{API_BASE_URL}}</code> 或 <code>{{CONSOLE_URL}}</code>，使下方灰色小字预览刚好显示 <code>{{API_BASE_URL}}/chat/completions</code></div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：添加模型并点「检测」</strong><p>在下方模型区点 <strong><code>+ 添加</code></strong>，粘贴完整模型 ID（按需勾选<strong>视觉</strong>、<strong>推理</strong>、<strong>工具</strong>图标）；再点密钥右侧的 <strong>「检测」</strong> 按钮提示连接成功即可。</p></div>
  <div class="step"><strong>第 5 步：回聊天页选模型对话</strong><p>点左上角 💬 返回聊天页，在顶部模型下拉框选中 <code>OpenBridger</code> 下的模型开始对话。</p></div>
</div>

> **搭建本地 PDF 知识库**：在同一服务商下点 `+ 添加` 填入 `text-embedding-3-small`（勾选「嵌入」标签），即可在左侧「知识库」导入论文 PDF 做检索问答（见 [向量与 Rerank](/api/embeddings-rerank/)）。

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 点「检测」报 `404` | 看 API 地址下方灰色预览是否出现了 `/v1/v1`，删掉多余的 `/v1` |
| 顶部下拉框看不到 OpenBridger | 回「设置 -> 模型服务 -> OpenBridger」打开右上角总开关，并至少添加 1 个模型 |
| 聊天框没有上传图片按钮 | 在模型列表点该模型右侧 ⚙️，勾选 **「视觉（Vision）」** 保存 |
| 报 `503 无可用渠道` | 从 [模型广场]({{CONSOLE_URL}}/pricing) 重新复制完整模型 ID |

## 相关

- [Chatbox](/guide/clients/chatbox/)
- [Embeddings 向量与 Rerank](/api/embeddings-rerank/)
- [错误码](/guide/errors/)
