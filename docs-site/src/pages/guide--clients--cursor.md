# Cursor 详细配置教程

Cursor 支持通过「自定义 OpenAI Base URL」接入 OpenBridger 的大模型进行 **Chat 对话**与 **Agent 编码**。无需安装额外插件，按下面 **1 -> 2 -> 3 -> 4 -> 5** 五步即可配好。

---

## 开始之前

1. 打开控制台 [API 密钥]({{CONSOLE_URL}}/keys)，单独创建一把名为 `cursor` 的密钥并复制 `sk-...`。
2. 打开 [模型广场]({{CONSOLE_URL}}/pricing)，复制你想在 Cursor 里调用的完整模型 ID（例如 `gpt-6.2-astra`、`claude-sonnet-5-5`、`deepseek-v4.1`）。
3. 说明：是否允许使用自定义 API Key（BYOK）取决于你当前安装的 Cursor 版本与账号规则，不需要退出你原有的 Cursor 登录账号。

---

## 详细配置步骤（1-2-3-4-5）

<div class="steps">
  <div class="step">
    <strong>第 1 步：打开 Cursor 设置面板（Cursor Settings）</strong>
    <p>打开 Cursor 编辑器，点击界面右上角的 <strong>⚙️ 齿轮图标</strong>（或按快捷键：macOS <code>Command + Shift + J</code> / Windows <code>Ctrl + Shift + J</code>），进入 <strong>Cursor Settings</strong>（注意不是 VS Code 原本的 Editor Settings）。</p>
  </div>
  <div class="step">
    <strong>第 2 步：进入 Models 页面并添加自定义模型名</strong>
    <p>在左侧菜单点击 <strong><code>Models</code></strong>。<br>
    先拉到模型列表下方点击 <strong><code>+ Add Model</code>（或 Add Custom Model）</strong>，粘贴你在 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制的<strong>完整模型 ID</strong>（如 <code>gpt-6.2-astra</code>），点击添加并确保它右侧的开关处于勾选启用状态。<em>（建议先把不用的官方自带模型取消勾选，避免验证时请求混淆。）</em></p>
  </div>
  <div class="step">
    <strong>第 3 步：填写 OpenAI API Key 与 Override Base URL</strong>
    <p>在同一页面向下滚动找到 <strong><code>OpenAI API Key</code></strong> 区域：<br>
    1. 在 <strong>OpenAI API Key</strong> 输入框粘贴你的 <code>sk-...</code> 密钥。<br>
    2. 点击展开下方的 <strong><code>Override OpenAI Base URL</code></strong>（覆盖基础地址），把开关拨到开启状态。<br>
    3. 在地址输入框填入 <strong><code>{{API_BASE_URL}}</code></strong>（末尾带 <code>/v1</code>，不要多加斜杠或 <code>/chat/completions</code>）。</p>
  </div>
  <div class="step">
    <strong>第 4 步：点击保存 / Verify 验证</strong>
    <p>先点击地址输入框右侧的 <strong>保存（Save / 勾号图标）</strong>，再点击 <strong><code>Verify</code></strong> 按钮（或把 OpenAI API Key 右侧的总开关打开）。如果部分 Cursor 版本点击 Verify 时默认拿你没权限的旧模型去测导致弹窗告警，也可直接保存开启后进入下一步实测。</p>
  </div>
  <div class="step">
    <strong>第 5 步：新建对话切换模型并核对日志</strong>
    <p>按 <code>Command + L</code>（Windows 按 <code>Ctrl + L</code>）或 <code>Command + I</code> 打开右侧 AI 面板，点击右上角 <strong><code>+</code>（New Chat）</strong> 新建对话，在输入框左下角的<strong>模型选择下拉框</strong>里选中你刚添加的模型（如 <code>gpt-6.2-astra</code>），发送一句 <code>你好</code>。随后在控制台 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 能看到记录即代表接入成功！</p>
  </div>
</div>

### 表单填写速查卡

<div class="config-card">
  <div><span>设置入口</span>右上角 ⚙️ -> <code>Cursor Settings</code> -> <code>Models</code></div>
  <div><span>OpenAI API Key</span>粘贴控制台复制的 <code>sk-...</code> 密钥</div>
  <div><span>Override OpenAI Base URL</span>开启开关并填 <code>{{API_BASE_URL}}</code></div>
  <div><span>+ Add Model</span>填入模型广场中的完整 ID（如 <code>gpt-6.2-astra</code>）</div>
</div>

---

## 必读注意事项（和 Cursor 自带功能的关系）

- **代码自动补全（Cursor Tab）仍走官方**：行内灰色幽灵字补全由 Cursor 自有小模型负责，不走自定义 Base URL，也不会消耗 OpenBridger 额度。
- **切回 Cursor 官方套餐模型前请关掉开关**：只要 `Override OpenAI Base URL` 开着，Cursor 会把所有走 OpenAI 通道的模型都发往 `{{API_BASE_URL}}`。如果你后续想切回 Cursor 会员自带模型，记得先把 `Override OpenAI Base URL` 关掉再重启 Cursor。

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 点 `Verify` 报错，提示某个没见过的模型不存在 | Cursor 点 Verify 时会自动挑列表里第一个勾着的模型发测试请求 | 在 Models 列表里把其他官方模型先关掉，只保留你从模型广场添加的那个模型再点 Verify |
| 对话报错 `model not found` 或 `503 无可用渠道` | 手打模型名字大小写或后缀错了，或密钥分组不包含该模型 | 去 [模型广场]({{CONSOLE_URL}}/pricing) 点复制完整 ID，核对密钥分组 |
| 报错 `404` | Base URL 漏写了 `/v1` 或多写了 `/chat/completions` | 严格改为 `{{API_BASE_URL}}` |
| 控制台使用日志里完全没有记录 | 右下角对话模型下拉框还选在 `Auto` 或官方内置模型上 | 取消 `Auto` 自动选模型，手动点选你刚添加的模型并新开一个 Chat |

## 相关

- [Cline / Roo Code（VS Code / Cursor 内插件方案）](/guide/clients/cline/)
- [Base URL 填写速查](/guide/base-urls/)
- [故障排查](/troubleshooting/)
