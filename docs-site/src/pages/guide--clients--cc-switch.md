# CC Switch 一键配置教程（新手首选）

**CC Switch** 是开源图形化配置面板：在界面上填好地址和密钥点 **「启用」**，即可自动写好 **Codex（桌面版 / VS Code / CLI）** 与 **Claude Code** 的底层配置文件，不用手找隐藏文件夹。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建密钥（如 `cc-switch`），复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 点击复制完整模型 ID（Codex 常用 `gpt-6.2-astra`、`gpt-6.2-sol`；Claude 常用 `claude-sonnet-5-5`、`claude-opus-5-5`）。
3. 从 `https://ccswitch.io`（或 GitHub `farion1231/cc-switch` Releases）下载安装 **CC Switch**：
   - **macOS**：拖入「应用程序」；首次打开若提示无法验证开发者，去系统「设置 -> 隐私与安全性」底部点 **「仍要打开」**。
   - **Windows**：双击 `.msi` / `.exe` 安装并打开。

<div class="callout warning"><strong>地址规则别填反</strong><p><strong>Codex</strong> 地址末尾<strong>必须带 <code>/v1</code></strong>（填 <code>{{API_BASE_URL}}</code>）；<strong>Claude Code</strong> 地址末尾<strong>绝不带 <code>/v1</code></strong>（填 <code>{{CONSOLE_URL}}</code>）。</p></div>

---

## 一、配置 Codex（桌面版 / VS Code 插件 / CLI 一次全通）

<div class="steps">
  <div class="step"><strong>第 1 步：切到 Codex 页</strong><p>打开 CC Switch，点击顶部应用栏的 <strong><code>Codex</code></strong>。</p></div>
  <div class="step"><strong>第 2 步：新建自定义供应商</strong><p>点击右上角 <strong><code>+</code></strong> 按钮 -> 选择 <strong>「自定义配置（Custom）」</strong>。</p></div>
  <div class="step"><strong>第 3 步：对照填写表单</strong>
    <div class="config-card">
      <div><span>名称（Name）</span><code>OpenBridger</code></div>
      <div><span>Base URL</span><code>{{API_BASE_URL}}</code>（<strong>带 /v1</strong>）</div>
      <div><span>API Key</span>粘贴 <code>sk-...</code> 完整密钥</div>
      <div><span>协议（Wire API）</span><code>responses</code></div>
      <div><span>模型（Model）</span><code>gpt-6.2-astra</code>（或从模型广场复制的 ID）</div>
      <div><span>推理强度（可选）</span><code>medium</code> 或 <code>high</code></div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：保存并点「启用」</strong><p>点击右下角 <strong>「保存」</strong>，回到列表点击该卡片上的 <strong>「启用（Enable）」</strong>，确认显示「已启用」。</p></div>
  <div class="step"><strong>第 5 步：重启客户端生效</strong><p>若 Codex 桌面版已打开，按 <code>Cmd + Q</code>（Windows 右下角托盘退出）<strong>彻底退出再重开</strong>；VS Code 按 <code>Cmd/Ctrl + Shift + P</code> 运行 <code>Reload Window</code>。新建会话发消息即可。</p></div>
</div>

---

## 二、配置 Claude Code（桌面端 / VS Code / CLI 通用）

<div class="steps">
  <div class="step"><strong>第 1 步：切到 Claude Code 页</strong><p>在 CC Switch 顶部点击 <strong><code>Claude Code</code></strong>。</p></div>
  <div class="step"><strong>第 2 步：新建自定义供应商</strong><p>点击右上角 <strong><code>+</code></strong> -> 选择 <strong>「自定义配置（Custom）」</strong>。</p></div>
  <div class="step"><strong>第 3 步：对照填写表单</strong>
    <div class="config-card">
      <div><span>名称（Name）</span><code>OpenBridger Claude</code></div>
      <div><span>Base URL</span><code>{{CONSOLE_URL}}</code>（<strong>不要加 /v1！</strong>）</div>
      <div><span>API Key / Auth Token</span>粘贴 <code>sk-...</code> 完整密钥</div>
      <div><span>模型（Model，可选）</span><code>claude-sonnet-5-5</code> 或 <code>claude-opus-5-5</code></div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：保存并点「启用」</strong><p>点击 <strong>「保存」</strong>，再点卡片上的 <strong>「启用（Enable）」</strong>。</p></div>
  <div class="step"><strong>第 5 步：新开终端或重启桌面端</strong><p>关掉旧终端窗口，新开终端运行 <code>claude</code> 即可直接提问，无需登录官方账号。</p></div>
</div>

> **多模型切换技巧**：你可以在 CC Switch 里建多张不同模型的卡片（例如一张旗舰卡写复杂代码、一张轻量卡做日常修改），点「启用」切换后在客户端**新建对话**即可。

## 常见报错速查

| 现象 | 原因与一步解决 |
| --- | --- |
| 打开客户端仍弹官方登录框 | 忘了点卡片上的 **「启用」**，或客户端没有彻底退出重开 |
| 报 `404` | Codex 漏了 `/v1`，或 Claude Code 多加了 `/v1` |
| 报 `503 无可用渠道` | 去 [模型广场]({{CONSOLE_URL}}/pricing) 重新复制完整模型 ID，确认密钥分组支持该模型 |
| 报 `403 用户额度不足` | 先去 [钱包]({{CONSOLE_URL}}/wallet) 充值或兑换额度 |

## 相关

- [Codex（桌面版 / 插件 / CLI）](/guide/clients/codex-cli/)
- [Claude Code（桌面端 / CLI）](/guide/clients/claude-code/)
- [错误码](/guide/errors/)
