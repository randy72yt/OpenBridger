# Cursor 配置教程

通过 Cursor 内置的 **Override OpenAI Base URL** 接入 OpenBridger 进行 Chat 与 Agent 编码（无需退出你原有的 Cursor 账号；是否支持自带 Key 取决于你安装的 Cursor 版本策略）。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `cursor` 的密钥，复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID（如 `gpt-6.2-astra`、`claude-sonnet-5-5`）。

## 配置 5 步

<div class="steps">
  <div class="step"><strong>第 1 步：打开 Cursor Settings</strong><p>点击 Cursor 右上角 <strong>⚙️ 齿轮图标</strong>（快捷键 <code>Cmd/Ctrl + Shift + J</code>）-> 左侧点 <strong><code>Models</code></strong>。</p></div>
  <div class="step"><strong>第 2 步：先添加自定义模型名</strong><p>拉到模型列表下方点 <strong><code>+ Add Model</code></strong>，粘贴完整模型 ID（如 <code>gpt-6.2-astra</code>）并勾选启用；建议顺手把其他不用的内置 OpenAI 模型取消勾选，防止验证时测错模型。</p></div>
  <div class="step"><strong>第 3 步：填写 Key 并开启 Override Base URL</strong>
    <div class="config-card">
      <div><span>OpenAI API Key</span>粘贴 <code>sk-...</code> 完整密钥</div>
      <div><span>Override OpenAI Base URL</span>打开开关 -> 填 <code>{{API_BASE_URL}}</code>（带 <code>/v1</code>）</div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：点击 Save 与 Verify</strong><p>先点地址框右侧的 <strong>Save</strong>，再打开 OpenAI API Key 右侧开关（或点 <code>Verify</code>）。</p></div>
  <div class="step"><strong>第 5 步：新建对话关掉 Auto 并选模型</strong><p>按 <code>Cmd/Ctrl + L</code> 打开右侧面板 -> 点 <strong><code>+</code>（New Chat）</strong> -> 在输入框左下角关掉 <code>Auto</code>，选中你刚添加的模型发消息，到 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 核对即可。</p></div>
</div>

> **补充说明**：行内代码补全（Cursor Tab）仍走 Cursor 自有服务，不消耗 OpenBridger 额度。后续若想切回 Cursor 会员内置模型，请先关闭 `Override OpenAI Base URL` 并重启 Cursor。

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 点 `Verify` 提示某个没见过的模型不存在 | 先在 `Models` 列表里取消勾选其他内置模型，只留你添加的模型再点 `Verify` |
| 报 `404` | Base URL 改为 `{{API_BASE_URL}}`（末尾不要加 `/` 或 `/chat/completions`） |
| 报 `503` 或 `model not found` | 从 [模型广场]({{CONSOLE_URL}}/pricing) 重新复制完整模型 ID |
| 使用日志里没记录 | 对话框左下角模型还停在 `Auto` 或内置模型上，手动切换并新开 Chat |

## 相关

- [Cline / Roo Code（Cursor 内插件替代方案）](/guide/clients/cline/)
- [故障排查](/troubleshooting/)
