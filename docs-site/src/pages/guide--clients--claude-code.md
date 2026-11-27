# Claude Code 配置教程（桌面端 / VS Code / CLI）

Claude Code（命令行 `claude`、VS Code 插件、桌面端 Code 模式）走 Anthropic 原生 `/v1/messages` 协议，共用本地 `.claude` 目录。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `claude-code` 的密钥，复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID（如 `claude-sonnet-5-5`、`claude-opus-5-5`）。

<div class="callout warning"><strong>Base URL 绝不要带 /v1</strong><p>Claude Code 会自动拼接 <code>/v1/messages</code>，地址必须填 <strong><code>{{CONSOLE_URL}}</code></strong>（写成带 <code>/v1</code> 会报 <code>404</code>）。</p></div>

---

## 方式一：用 CC Switch 图形界面一键配（推荐）

1. 打开 **CC Switch**（见 [CC Switch 教程](/guide/clients/cc-switch/)），顶部切到 **`Claude Code`** -> 点右上角 **`+`** -> 选 **「自定义配置」**。
2. 填 `Base URL` = `{{CONSOLE_URL}}`（不带 `/v1`）、`API Key` = `sk-...`、`Model` = `claude-sonnet-5-5`。
3. 点 **「保存」** -> 点卡片 **「启用」** -> 新开终端运行 `claude`（或重载 VS Code）即可。

---

## 方式二：手动配置 `.claude/settings.json`

<div class="steps">
  <div class="step"><strong>第 1 步：安装客户端</strong><p>• <strong>CLI 命令行</strong>：需 Node.js 18+，运行 <code>npm install -g @anthropic-ai/claude-code</code>。<br>• <strong>VS Code / 桌面端</strong>：安装官方 <code>Claude Code</code> 扩展或桌面端。</p></div>
  <div class="step"><strong>第 2 步：打开 <code>.claude</code> 隐藏文件夹</strong><p>• <strong>macOS</strong>：访达按 <code>Command + Shift + G</code> 输入 <code>~/</code> 回车，按 <code>Command + Shift + .</code> 显示隐藏文件，新建或打开 <code>.claude</code>。<br>• <strong>Windows</strong>：资源管理器地址栏输 <code>%USERPROFILE%</code> 回车，勾选显示「文件扩展名」，新建或打开 <code>.claude</code>。</p></div>
  <div class="step"><strong>第 3 步：新建 <code>settings.json</code></strong><p>在 <code>.claude</code> 内保存 <strong><code>settings.json</code></strong>（见下方模板，替换密钥和模型 ID）。</p></div>
  <div class="step"><strong>第 4 步：跳过首次引导（如卡在登录页）</strong><p>若全新安装首次启动仍弹登录引导，在主目录 <code>~/.claude.json</code> 写入 <code>{"hasCompletedOnboarding": true}</code>（或用 CC Switch 点一次启用自动完成）。</p></div>
  <div class="step"><strong>第 5 步：重开终端 / 重载 VS Code 验证</strong><p>新开终端进入项目目录输入 <code>claude</code>（VS Code 执行 <code>Developer: Reload Window</code>），发一句提问并在 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 核对。</p></div>
</div>

### `settings.json` 模板

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{CONSOLE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-替换成你在控制台复制的密钥",
    "ANTHROPIC_MODEL": "claude-sonnet-5-5",
    "API_TIMEOUT_MS": "300000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

### 附：单次终端临时命令（不改配置文件）

- **macOS / Linux**：`export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}" ANTHROPIC_AUTH_TOKEN="sk-密钥" ANTHROPIC_MODEL="claude-sonnet-5-5" && claude`
- **Windows PowerShell**：`$env:ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"; $env:ANTHROPIC_AUTH_TOKEN="sk-密钥"; $env:ANTHROPIC_MODEL="claude-sonnet-5-5"; claude`

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 仍提示 `/login` 登录官方账号 | 关掉旧终端新开一个，检查文件名不是 `settings.json.txt`，或用 [CC Switch](/guide/clients/cc-switch/) 一键启用 |
| 右上角显示红色 `offline` | 客户端探测遥测域名的提示，**只要发消息能正常回复即可直接忽略** |
| 报 `404` / `503` | `404` 删掉地址末尾的 `/v1`；`503` 去 [模型广场]({{CONSOLE_URL}}/pricing) 核对完整模型 ID |

## 相关

- [CC Switch（新手一键配置）](/guide/clients/cc-switch/)
- [Anthropic Messages 原生接口](/api/anthropic-messages/)
- [错误码](/guide/errors/)
