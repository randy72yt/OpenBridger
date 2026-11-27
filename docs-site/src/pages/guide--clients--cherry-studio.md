# Cherry Studio 保姆级配置教程

Cherry Studio 是目前桌面端最好用的全能 AI 客户端之一（支持知识库、联网、多模型同屏对比、论文阅读与画图）。按下面 **1 -> 2 -> 3 -> 4 -> 5** 五步，1 分钟即可配好 OpenBridger。

---

## 开始之前

1. 打开控制台 [API 密钥]({{CONSOLE_URL}}/keys)，创建一把名为 `cherry-studio` 的密钥并复制 `sk-...`。
2. 打开 [模型广场]({{CONSOLE_URL}}/pricing)，记下或复制你想常用的几个模型完整 ID（例如 `gpt-6.2-astra`、`claude-sonnet-5-5`、`gemini-4-neon`、`deepseek-v4.1`）。

---

## 详细配置步骤（1-2-3-4-5）

<div class="steps">
  <div class="step">
    <strong>第 1 步：进入「模型服务」设置页</strong>
    <p>打开 Cherry Studio 软件，点击界面<strong>左下角的 ⚙️（设置）图标</strong>，在设置页左侧栏点击第一项 <strong>「模型服务（Model Provider）」</strong>。</p>
  </div>
  <div class="step">
    <strong>第 2 步：新增自定义服务商</strong>
    <p>将中间的服务商列表拉到底部，点击 <strong><code>+ 添加</code>（Add）</strong> 按钮：<br>
    • <strong>提供商名称</strong>：输入 <code>OpenBridger</code><br>
    • <strong>提供商类型</strong>：下拉选择 <strong><code>OpenAI</code></strong>（如果列表里有 <code>New API</code> 也可以选 <code>OpenAI</code>，通用性最好）<br>
    点击弹窗里的 <strong>「确定」</strong>，并确保新出现的 <code>OpenBridger</code> 右侧<strong>总开关处于打开状态（绿色/高亮）</strong>。</p>
  </div>
  <div class="step">
    <strong>第 3 步：填写 API 密钥与 API 地址（看一眼灰色预览！）</strong>
    <p>在右侧配置面板依次填写：<br>
    1. <strong>API 密钥</strong>：粘贴你在控制台复制的 <code>sk-...</code>。<br>
    2. <strong>API 地址</strong>：填入 <code>{{API_BASE_URL}}</code>，或者填 <code>{{CONSOLE_URL}}</code> —— <strong>判断诀窍</strong>：看一眼输入框正下方 Cherry Studio 自动生成的那行灰色小字「预览地址」，只要灰色预览最终显示为 <strong><code>{{API_BASE_URL}}/chat/completions</code></strong>（里面有且只有一个 <code>/v1</code>）就完全正确！如果预览里出现了两个 <code>/v1/v1</code>，就把输入框末尾的 <code>/v1</code> 删掉，或在地址末尾加一个英文斜杠。</p>
  </div>
  <div class="step">
    <strong>第 4 步：添加模型并点「检测」连通性</strong>
    <p>向下滚动到该服务商底部的「模型」区域：<br>
    • 点击 <strong><code>+ 添加</code></strong>（或者点击 <strong><code>管理</code></strong> 自动拉取列表后点 <code>+</code> 勾选你想用的模型）。<br>
    • 手动添加时，在「模型 ID」栏粘贴从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制的完整名称（如 <code>gpt-6.2-astra</code>）。如果该模型支持看图识图或深度思考，顺手勾选弹窗里的 <strong>「视觉（眼睛图标）」</strong>、<strong>「推理」</strong> 与 <strong>「工具调用」</strong>。<br>
    • 回到上方 API 密钥右侧点击 <strong>「检测」</strong> 按钮，选择刚添加的模型点确定，提示绿色「连接成功」即全部配通！</p>
  </div>
  <div class="step">
    <strong>第 5 步：回到聊天界面切换模型开聊</strong>
    <p>点击界面最左侧顶部的 💬 <strong>对话图标</strong>返回主界面，点击顶部中间的<strong>模型名称下拉框</strong>，找到 <code>OpenBridger</code> 分组下的模型点击选中，发一句消息即可流畅对话。</p>
  </div>
</div>

### 表单填写速查卡

<div class="config-card">
  <div><span>提供商名称</span><code>OpenBridger</code>（确保右侧启用开关已打开）</div>
  <div><span>提供商类型</span><code>OpenAI</code></div>
  <div><span>API 密钥</span><code>sk-...</code>（不要带空格）</div>
  <div><span>API 地址</span>填后核对下方灰色预览为 <code>{{API_BASE_URL}}/chat/completions</code></div>
  <div><span>模型 ID</span>从模型广场复制完整 ID（如 <code>gpt-6.2-astra</code>、<code>claude-sonnet-5-5</code>）</div>
</div>

---

## 进阶用法：在 Cherry Studio 搭建本地文献知识库

如果你要用 Cherry Studio 导入 PDF 课件或论文建本地知识库（RAG）：
1. 在同一个 `OpenBridger` 服务商下点击 **`+ 添加`**，把向量模型 `text-embedding-3-large`（或 `text-embedding-3-small`）添加进来，并在模型属性里勾选 **「嵌入（Embedding）」** 标签。
2. 点击左侧栏的 **「知识库」** 图标 -> 新建知识库 -> 选择刚才添加的嵌入模型，即可拖入 PDF / Word 文档做精准问答（详见 [向量与重排接口](/api/embeddings-rerank/)）。

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 点「检测」提示 `404` | API 地址下方的灰色预览变成了 `/v1/v1/chat/completions` 或少了 `/v1` | 看着输入框下方的灰色预览小字调整地址，确保最终预览是 `{{API_BASE_URL}}/chat/completions` |
| 顶部聊天模型下拉框里看不到 OpenBridger | 第 2 步忘记把 `OpenBridger` 服务商右上角的启用开关打开，或还没添加任何模型 | 回到「设置 -> 模型服务 -> OpenBridger」，打开右上角开关并至少添加 1 个模型 |
| 想发图片但聊天框没有上传图片按钮 | 手动添加模型时没勾选该模型的「视觉」能力标签 | 在模型列表点击该模型右侧的 ⚙️ 小齿轮，勾选 **「视觉（Vision）」** 保存 |
| 报错 `503 无可用渠道` | 模型 ID 拼写错误或密钥分组无此模型 | 从 [模型广场]({{CONSOLE_URL}}/pricing) 重新复制完整模型 ID |

## 相关

- [Chatbox](/guide/clients/chatbox/)
- [LobeChat](/guide/clients/lobechat/)
- [Embeddings 向量与 Rerank](/api/embeddings-rerank/)
- [错误码](/guide/errors/)
