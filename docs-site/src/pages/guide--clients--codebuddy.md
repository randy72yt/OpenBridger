# CodeBuddy 详细配置教程（独立 IDE / VS Code & JetBrains 插件）

腾讯云代码助手 **CodeBuddy**（含独立版 CodeBuddy IDE，以及安装在 VS Code、IntelliJ IDEA / PyCharm 里的 CodeBuddy 扩展）支持添加 OpenAI 兼容的自定义模型用于 **Chat 对话**与 **Craft / Agent 编码**。按下面 **1 -> 2 -> 3 -> 4 -> 5** 五步即可完成接入。

---

## 开始之前

1. 打开控制台 [API 密钥]({{CONSOLE_URL}}/keys)，创建一把名为 `codebuddy` 的密钥并复制 `sk-...`。
2. 打开 [模型广场]({{CONSOLE_URL}}/pricing)，复制你想调用的完整模型 ID（例如 `gpt-6.2-astra`、`claude-sonnet-5-5`、`deepseek-v4.1`）。
3. 保持你原有的 CodeBuddy 账号登录状态即可，自定义模型与内置补全互不冲突。

---

## 详细配置步骤（1-2-3-4-5）

<div class="steps">
  <div class="step">
    <strong>第 1 步：打开 CodeBuddy 模型设置</strong>
    <p>在编辑器里打开 CodeBuddy 侧边面板，点击面板右上角的 <strong>⚙️ 设置图标（或个人头像菜单 -> 设置 / Settings）</strong>，找到 <strong>「模型管理 / 自定义模型（Models / Custom Models）」</strong>（也可以直接点击对话输入框下方的<strong>模型切换下拉框</strong>，选择底部的 <strong>「+ 添加模型 / 管理模型」</strong>）。</p>
  </div>
  <div class="step">
    <strong>第 2 步：新增 OpenAI 兼容模型</strong>
    <p>点击 <strong><code>+ 添加自定义模型</code>（Add Custom Model）</strong>，在协议或供应商类型（Provider / API Type）下拉框中选择 <strong><code>OpenAI Compatible</code>（OpenAI 兼容）</strong>。</p>
  </div>
  <div class="step">
    <strong>第 3 步：填写接口地址、模型 ID 与密钥</strong>
    <p>在配置表单中依次填入：<br>
    • <strong>显示名称（Display Name，若有）</strong>：如 <code>OB - GPT-6.2</code>（方便在下拉框里一眼认出）<br>
    • <strong>模型名称 / Model ID</strong>：粘贴从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制的<strong>完整 ID</strong>（如 <code>gpt-6.2-astra</code>）<br>
    • <strong>接口地址 / Base URL</strong>：填 <strong><code>{{API_BASE_URL}}</code></strong>（末尾带 <code>/v1</code>，不要多加 <code>/chat/completions</code>）<br>
    • <strong>API Key</strong>：粘贴你在控制台复制的 <code>sk-...</code> 密钥</p>
  </div>
  <div class="step">
    <strong>第 4 步：测试并保存</strong>
    <p>如果弹窗底部有「测试连接（Test）」按钮可先点一下确认通过，然后点击 <strong>「保存 / 确定（Save）」</strong>。</p>
  </div>
  <div class="step">
    <strong>第 5 步：在 Chat 或 Craft 面板切换模型验证</strong>
    <p>回到 CodeBuddy 对话面板，点击右上角 <strong>新建会话</strong>，在输入框下方的模型下拉菜单里选中刚添加的模型（注意：<strong>Chat</strong> 与 <strong>Craft 智能体模式</strong>的模型选择相互独立，切换模式时留意确认），发一句测试问题，并在控制台 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 核对记录。</p>
  </div>
</div>

### 表单填写速查卡

<div class="config-card">
  <div><span>类型（Provider）</span><code>OpenAI Compatible</code></div>
  <div><span>Base URL（接口地址）</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Key（密钥）</span><code>sk-...</code></div>
  <div><span>Model ID（模型标识）</span>模型广场完整 ID（如 <code>gpt-6.2-astra</code>、<code>claude-sonnet-5-5</code>）</div>
</div>

---

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 菜单里找不到「自定义模型」入口 | 部分旧版企业内网插件禁用了外部自定义模型 | 将 CodeBuddy 插件/IDE 升级到最新版，或在 VS Code 中搭配 [Cline / Roo Code](/guide/clients/cline/) 使用 |
| 报错 `404` | Base URL 漏写了 `/v1` 或多写了 `/chat/completions` | 改为 `{{API_BASE_URL}}` 保存重试 |
| 报错 `503 无可用渠道` 或 `model not found` | 把「显示名称」和「Model ID」填混了，或模型名拼写不一致 | 确保 `Model ID` 一栏严格粘贴 [模型广场]({{CONSOLE_URL}}/pricing) 里的完整英文标识 |
| 使用日志没有消耗记录 | 行内代码补全走的是 CodeBuddy 自带免费小模型，或对话下拉框未切到自定义模型 | 在 Chat / Craft 输入框下方手动选中你添加的模型再提问 |

## 相关

- [Cursor](/guide/clients/cursor/)
- [Cline / Roo Code](/guide/clients/cline/)
- [故障排查](/troubleshooting/)
