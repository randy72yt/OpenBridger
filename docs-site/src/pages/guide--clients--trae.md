# Trae 详细配置教程

Trae IDE 支持在内置的 AI 对话（Chat）与智能体构建（Builder / Agent）中添加自定义模型。无需退出你原有的 Trae 账号，按下面 **1 -> 2 -> 3 -> 4 -> 5** 五步即可接入 OpenBridger。

---

## 开始之前

1. 打开控制台 [API 密钥]({{CONSOLE_URL}}/keys)，新建一把名为 `trae` 的密钥并复制 `sk-...`。
2. 打开 [模型广场]({{CONSOLE_URL}}/pricing)，复制你想在 Trae 中调用的完整模型 ID（例如 `gpt-6.2-astra`、`claude-sonnet-5-5`、`deepseek-v4.1`）。

---

## 详细配置步骤（1-2-3-4-5）

<div class="steps">
  <div class="step">
    <strong>第 1 步：打开 Trae 的 AI 模型设置</strong>
    <p>打开 Trae 编辑器，点击右侧 AI 侧边栏右上角的 <strong>⚙️ 齿轮图标（AI Settings / 设置）</strong>，或者直接点击对话输入框左下角的<strong>当前模型名称下拉菜单</strong>，选择最底部的 <strong>「+ 添加模型（Add Model）」</strong> 或进入 <strong><code>Models</code></strong> 设置页。</p>
  </div>
  <div class="step">
    <strong>第 2 步：选择「自定义 / OpenAI 兼容」服务商</strong>
    <p>在添加模型的弹窗里：<br>
    • <strong>服务商（Provider）</strong>：选择 <strong><code>OpenAI Compatible</code>（自定义 / OpenAI 兼容）</strong>。<em>（如果当前安装的旧版 Trae 只提供内置固定厂商而没有自定义 Base URL 输入框，请先将 Trae 升级至最新版本，或在 Trae 内安装 <a href="/guide/clients/cline/">Cline / Roo Code 扩展</a> 接入。）</em></p>
  </div>
  <div class="step">
    <strong>第 3 步：填写 Base URL、模型 ID 与 API Key</strong>
    <p>对照下方卡片填入：<br>
    • <strong>API 地址 / Base URL</strong>：填 <code>{{API_BASE_URL}}</code>（末尾带 <code>/v1</code>；仅当输入框提示文字明确要求完整 Endpoint 时才填 <code>{{API_BASE_URL}}/chat/completions</code>）<br>
    • <strong>模型名称 / Model ID</strong>：粘贴你在 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制的完整 ID（如 <code>gpt-6.2-astra</code>）<br>
    • <strong>API Key</strong>：粘贴 <code>sk-...</code> 密钥</p>
  </div>
  <div class="step">
    <strong>第 4 步：点击保存 / 添加模型</strong>
    <p>点击弹窗底部的 <strong>「添加 / 保存（Add / Save）」</strong> 按钮完成绑定。如果你想同时配 Claude 和 GPT 两个模型，重复第 2-4 步再添加一条不同模型 ID 即可。</p>
  </div>
  <div class="step">
    <strong>第 5 步：在 Chat 与 Builder 面板分别切换模型验证</strong>
    <p>回到右侧 AI 面板，点击右上角 <strong>新建对话</strong>，把输入框下方的模型切换为你刚添加的模型（注意：<strong>Chat 模式</strong>和 <strong>Builder 模式</strong>各自独立记住所选模型，切换模式后记得看一眼模型名字！）。发一句测试指令，并到控制台 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 核对记录。</p>
  </div>
</div>

### 表单填写速查卡

<div class="config-card">
  <div><span>Provider（服务商）</span><code>OpenAI Compatible</code></div>
  <div><span>Base URL（接口地址）</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Key（密钥）</span><code>sk-...</code></div>
  <div><span>Model ID（模型）</span>模型广场完整 ID（如 <code>gpt-6.2-astra</code>、<code>claude-sonnet-5-5</code>）</div>
</div>

---

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 行内代码自动补全没走自定义模型 | Trae 的行内 Tab 补全由官方自带补全服务驱动，与右侧对话模型相互独立 | 属正常现象，Chat 与 Builder 会正常走 OpenBridger |
| 报错 `404` | Base URL 漏了 `/v1` 或多拼了一层路径 | 检查该版本输入框下方的占位提示，默认填 `{{API_BASE_URL}}` |
| 报错 `503 无可用渠道` 或 `model not found` | 模型名称手打有误或密钥所属分组不含此模型 | 去 [模型广场]({{CONSOLE_URL}}/pricing) 点击复制完整模型 ID |
| 日志里查不到请求 | 切到了 Builder 模式后模型自动跳回了 Trae 内置模型 | 在当前面板底部下拉框重新选中你添加的自定义模型 |

## 相关

- [Cursor](/guide/clients/cursor/)
- [Cline / Roo Code](/guide/clients/cline/)
- [故障排查](/troubleshooting/)
