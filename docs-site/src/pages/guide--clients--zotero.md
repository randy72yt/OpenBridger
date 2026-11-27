# Zotero 配置教程（论文划词与 PDF 文献问答）

在 Zotero 中配合 **Awesome GPT（`zotero-gpt`）** 插件，可在读 PDF 论文时直接划词翻译、解释数学公式、提炼全文创新点。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `zotero` 的密钥，复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制长上下文对话模型 ID（如 `gemini-4-neon`、`claude-sonnet-5-5`、`gpt-6.2-sol`）；若要用「整篇 PDF 检索问答」，再记一个向量模型 `text-embedding-3-small`。

## 配置 4 步

<div class="steps">
  <div class="step"><strong>第 1 步：安装 Awesome GPT 插件（已装可跳过）</strong><p>从 GitHub <code>MuiseDestiny/zotero-gpt</code> 的 Releases 下载最新 <code>.xpi</code> 文件 -> 打开 Zotero 顶部菜单 <strong>「工具（Tools）-> 插件（Plugins）」</strong> -> 点右上角 ⚙️ 小齿轮选 <strong><code>Install Plugin From File...</code></strong> 选刚下载的 <code>.xpi</code> 安装。</p></div>
  <div class="step"><strong>第 2 步：打开 Zotero 的 GPT 设置页</strong><p>Windows 点顶部菜单 <strong>「编辑（Edit）-> 设置（Settings）」</strong>（macOS 点左上角 <strong><code>Zotero -> Settings</code></strong>），在左侧切到 <strong><code>GPT</code></strong> 标签页。</p></div>
  <div class="step"><strong>第 3 步：填写接口地址、密钥与模型</strong>
    <div class="config-card">
      <div><span>API（接口地址）</span>先填 <code>{{CONSOLE_URL}}</code>（若测试报 <code>404</code> 则改为 <code>{{API_BASE_URL}}</code>）</div>
      <div><span>Secret Key</span>粘贴 <code>sk-...</code> 完整密钥</div>
      <div><span>Model（对话模型）</span>粘贴模型广场完整 ID（如 <code>gpt-6.2-sol</code> 或 <code>gemini-4-neon</code>）</div>
      <div><span>Embedding Model（全文问答用）</span><code>text-embedding-3-small</code></div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：打开论文实测</strong><p>双击打开任意一篇 PDF 论文，选中一段英文摘要唤出 GPT 悬浮条点翻译（或按快捷键打开 GPT 窗口输入 <code>hi</code>），能正常流式输出即配置完成。</p></div>
</div>

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 报 `404` | 不同版本插件自动补 `/v1` 规则不同，把 `API` 一栏在 `{{CONSOLE_URL}}` 和 `{{API_BASE_URL}}` 之间换一下即可 |
| 划词正常，但点 `Meet Paper / 全文问答` 报错 | 检查 `Embedding Model` 是否填了当前密钥分组可用的向量模型（如 `text-embedding-3-small`） |
| 报 `403 用户额度不足` | 整篇论文一次发送数万 Token 预扣较高，去 [钱包]({{CONSOLE_URL}}/wallet) 检查余额 |

## 相关

- [沉浸式翻译（文献双语）](/guide/clients/immersive-translate/)
- [Embeddings 向量与 Rerank](/api/embeddings-rerank/)
- [错误码](/guide/errors/)
