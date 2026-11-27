# Chatbox 保姆级配置教程（桌面端 / 手机 App / 网页版）

Chatbox 界面轻巧简洁，同时支持 macOS、Windows、iOS 和 Android，非常适合作为日常随手提问与识图的跨端聊天软件。按下面 **1 -> 2 -> 3 -> 4 -> 5** 五步即可完成配置。

---

## 开始之前

1. 打开控制台 [API 密钥]({{CONSOLE_URL}}/keys)，新建一把名为 `chatbox` 的密钥并复制 `sk-...`。
2. 打开 [模型广场]({{CONSOLE_URL}}/pricing)，复制你想使用的完整模型 ID（例如 `gpt-6.2-astra`、`claude-sonnet-5-5`、`gemini-4-neon`）。
3. 注意：**不要**点击 Chatbox 里的「登录 Chatbox AI 官方订阅」，我们走「添加自定义提供方」即可。

---

## 详细配置步骤（1-2-3-4-5）

<div class="steps">
  <div class="step">
    <strong>第 1 步：打开左下角「设置」</strong>
    <p>打开 Chatbox 客户端，点击界面左下角的 <strong>⚙️「设置（Settings）」</strong>，切换到顶部的 <strong>「模型提供方（Model Provider）」</strong> 页面。</p>
  </div>
  <div class="step">
    <strong>第 2 步：点击「添加自定义提供方」</strong>
    <p>将左侧提供方列表拉到最底部，点击 <strong><code>+ 添加自定义提供方</code>（Add Custom Provider）</strong>：<br>
    • <strong>名称</strong>：填 <code>OpenBridger</code><br>
    • <strong>API 模式</strong>：下拉选择 <strong><code>OpenAI API 兼容</code></strong><br>
    <em>（如果你用的是较早版本的 Chatbox 没有「自定义提供方」按钮，直接在列表里选 <code>OpenAI API</code> 亦可。）</em></p>
  </div>
  <div class="step">
    <strong>第 3 步：逐项填写域名、路径与密钥</strong>
    <p>在右侧表单依次填入：<br>
    • <strong>API 域名（API Host）</strong>：填 <code>{{API_BASE_URL}}</code>（末尾带 <code>/v1</code>，不要带斜杠）<br>
    • <strong>API 路径（API Path）</strong>：保持默认的 <code>/chat/completions</code>（部分旧版没有此栏则不用管）<br>
    • <strong>API 密钥（API Key）</strong>：粘贴你的 <code>sk-...</code> 密钥</p>
  </div>
  <div class="step">
    <strong>第 4 步：添加模型并勾选视觉/推理能力</strong>
    <p>在下方的「模型」区域点击 <strong><code>+ 新建</code></strong>（或 <strong><code>获取</code></strong>）：<br>
    粘贴从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制的完整模型 ID（如 <code>gpt-6.2-astra</code>）。如果该模型支持看图或工具调用，勾选旁边的 <strong>「视觉（Vision）」</strong>、<strong>「推理」</strong> 选项，然后点击右下角 <strong>「保存」</strong>。</p>
  </div>
  <div class="step">
    <strong>第 5 步：新建一个对话窗口并选中模型</strong>
    <p>回到左侧主界面，点击 <strong><code>+ 新对话</code>（New Chat）</strong>（注意：旧对话可能还绑着之前的默认模型！），点击聊天输入框下方的<strong>模型名称按钮</strong>，选择 <code>OpenBridger</code> 下的模型发送一句测试消息，并在控制台 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 确认记录。</p>
  </div>
</div>

### 表单填写速查卡

<div class="config-card">
  <div><span>提供方类型 / API 模式</span>自定义提供方 -> <code>OpenAI API 兼容</code></div>
  <div><span>API 域名（API Host）</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API 路径（API Path）</span><code>/chat/completions</code>（若有此栏）</div>
  <div><span>API 密钥（API Key）</span><code>sk-...</code></div>
  <div><span>模型（Model）</span>模型广场完整 ID（如 <code>gpt-6.2-astra</code>、<code>claude-sonnet-5-5</code>）</div>
</div>

---

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 报错 `404`，提示路径里出现 `/v1/v1/chat/completions` | API 域名或 API 路径里重复写了 `/v1` | 确认 `API 域名` 为 `{{API_BASE_URL}}` 且 `API 路径` 为 `/chat/completions`（不要两边都带 `/v1`） |
| 报错 `401` | API Key 前后带了空格或粘贴不完整 | 重新去控制台复制粘贴 `sk-...` |
| 发消息没走 OpenBridger，使用日志无记录 | 还在旧对话窗口里提问，旧对话仍绑定着先前的内置模型 | 点击左上角 **`+ 新对话`**，在输入框下方重新点选 `OpenBridger` 的模型 |
| 不能发图片识图 | 新建自定义模型时未勾选「视觉（Vision）」开关 | 回到模型设置点编辑该模型，勾选「视觉」后保存 |

## 相关

- [Cherry Studio](/guide/clients/cherry-studio/)
- [LobeChat](/guide/clients/lobechat/)
- [Base URL 填写速查](/guide/base-urls/)
