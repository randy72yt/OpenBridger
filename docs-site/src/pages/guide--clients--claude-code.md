# Claude Code 完整配置教程（桌面端 / VS Code / CLI）

Claude Code（包括终端命令行 `claude`、VS Code 内的 Claude Code 扩展，以及 Claude 桌面端内置的 Code 模式）统一使用 Anthropic 原生 Messages 协议（`POST /v1/messages`）。

好消息是：它们都读取同一个本地配置目录 `.claude`。下面提供两种配置方式：**新手强烈推荐「方式一（CC Switch 图形界面一键配）」**；习惯自己管理配置文件的同学可按「方式二」逐步操作。

---

## 开始之前（必做 2 步）

1. 打开控制台 [API 密钥]({{CONSOLE_URL}}/keys)，点击 **「创建密钥」**，名称填 `claude-code`，分组选择支持 Messages 协议的分组（如 `default` 或 `claude`），复制生成的 `sk-...` 密钥。
2. 打开 [模型广场]({{CONSOLE_URL}}/pricing)，点击复制你要使用的完整模型 ID（例如 `claude-sonnet-5-5`、`claude-opus-5-5`）。

<div class="callout warning"><strong>最关键的一条规则：Base URL 不要带 /v1</strong><p>Claude Code 会自动在地址末尾拼接 <code>/v1/messages</code>。因此地址栏必须填站点根地址 <strong><code>{{CONSOLE_URL}}</code></strong>，<strong>千万不要</strong>写成 <code>{{API_BASE_URL}}</code>，否则会变成 <code>/v1/v1/messages</code> 导致 <code>404</code>。</p></div>

---

## 方式一：用 CC Switch 图形界面配置（零代码，新手首选）

不需要找隐藏文件夹，也不用担心 JSON 少写一个逗号：

1. 下载并打开 **CC Switch**（安装细节见 [CC Switch 新手教程](/guide/clients/cc-switch/)）。
2. 在顶部切换到 **`Claude Code`** 标签页 -> 点击右上角 **`+`** -> 选择 **「自定义配置（Custom）」**。
3. 对照填写：
   - **供应商名称**：`OpenBridger Claude`
   - **请求地址 / Base URL**：`{{CONSOLE_URL}}`（**不带 `/v1`**）
   - **API Key / Auth Token**：粘贴 `sk-...`
   - **主模型（Model，可选）**：粘贴从模型广场复制的完整 ID（如 `claude-sonnet-5-5`）
4. 点击 **「保存」**，回到列表点击该卡片上的 **「启用（Enable）」**。
5. 彻底关闭旧的终端窗口（或重启 VS Code / 桌面端），新开终端运行 `claude` 即可直接使用！

---

## 方式二：手动配置 `.claude/settings.json`（永久生效，三端通用）

如果你不想安装辅助软件，按下面 **1 -> 2 -> 3 -> 4 -> 5** 五步手动配置即可：

<div class="steps">
  <div class="step">
    <strong>第 1 步：安装 Node.js 与 Claude Code</strong>
    <p><strong>• 命令行版</strong>：确认电脑已安装 Node.js 18+（在终端输入 <code>node -v</code> 能看到版本号），运行 <code>npm install -g @anthropic-ai/claude-code</code>。<br>
    <strong>• VS Code 扩展 / 桌面端</strong>：在 VS Code 扩展商店搜索安装官方 <code>Claude Code</code> 插件，或安装官方桌面客户端。</p>
  </div>
  <div class="step">
    <strong>第 2 步：打开电脑上的 <code>.claude</code> 隐藏文件夹</strong>
    <p><strong>• macOS</strong>：打开「访达（Finder）」，按快捷键 <code>Command + Shift + G</code>，输入 <code>~/.claude</code> 回车。若提示不存在，先输入 <code>~/</code> 回车进入个人主目录，按 <code>Command + Shift + .</code> 显示隐藏文件，新建名为 <code>.claude</code> 的文件夹。<br>
    <strong>• Windows</strong>：打开「文件资源管理器」，在顶部地址栏输入 <code>%USERPROFILE%</code> 回车（例如 <code>C:\Users\你的用户名</code>），新建名为 <code>.claude</code> 的文件夹并双击进入。</p>
  </div>
  <div class="step">
    <strong>第 3 步：新建或编辑 <code>settings.json</code></strong>
    <p>在 <code>.claude</code> 文件夹内新建纯文本文件，命名为 <strong><code>settings.json</code></strong>（Windows 用户请在顶部「查看」勾选「文件扩展名」，确保不是 <code>settings.json.txt</code>），将下方的 JSON 模板复制进去，把密钥和模型名替换成你自己的值并保存。</p>
  </div>
  <div class="step">
    <strong>第 4 步：跳过首次引导（如有提示）</strong>
    <p>如果是全新安装的 Claude Code，首次启动若仍卡在官方登录引导页，可在同一个 <code>~/.claude.json</code>（主目录下的 <code>.claude.json</code> 文件）里加上 <code>{"hasCompletedOnboarding": true}</code>，或者直接用上方的 CC Switch 自动写入。</p>
  </div>
  <div class="step">
    <strong>第 5 步：重启终端 / 重载 VS Code 并验证</strong>
    <p>关掉之前开着的终端窗口，新开一个终端，<code>cd</code> 进入你的代码目录，输入 <code>claude</code> 回车；VS Code 用户按 <code>Cmd/Ctrl + Shift + P</code> 执行 <code>Developer: Reload Window</code>。发一句 <code>你好，请介绍一下当前目录结构</code>，再到控制台 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 核对记录。</p>
  </div>
</div>

### 第 3 步要复制的 `settings.json` 模板

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{CONSOLE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-把这里替换成你在控制台复制的密钥",
    "ANTHROPIC_MODEL": "claude-sonnet-5-5",
    "API_TIMEOUT_MS": "300000",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  }
}
```

<div class="config-card">
  <div><span>配置文件路径（macOS / Linux）</span><code>~/.claude/settings.json</code></div>
  <div><span>配置文件路径（Windows）</span><code>%USERPROFILE%\.claude\settings.json</code></div>
  <div><span>ANTHROPIC_BASE_URL</span><code>{{CONSOLE_URL}}</code>（<strong>不要加 /v1</strong>）</div>
  <div><span>ANTHROPIC_AUTH_TOKEN</span><code>sk-...</code>（不要加 <code>Bearer </code> 前缀）</div>
  <div><span>ANTHROPIC_MODEL</span>模型广场中的完整 ID（如 <code>claude-sonnet-5-5</code>、<code>claude-opus-5-5</code>）</div>
</div>

> **安全提醒**：`~/.claude/settings.json` 含有你的私密密钥，切勿提交到 Git 仓库或截图发到公开群聊。项目目录下的 `.claude/settings.json` 也不要写入真实密钥。

---

## 附：仅在当前终端临时试用（关掉窗口即失效）

如果你只想在某一个终端窗口临时测试，不改全局文件：

- **macOS / Linux（zsh / bash）**：
  ```bash
  export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
  export ANTHROPIC_AUTH_TOKEN="sk-你的密钥"
  export ANTHROPIC_MODEL="claude-sonnet-5-5"
  claude
  ```
- **Windows PowerShell**：
  ```powershell
  $env:ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
  $env:ANTHROPIC_AUTH_TOKEN="sk-你的密钥"
  $env:ANTHROPIC_MODEL="claude-sonnet-5-5"
  claude
  ```

---

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 启动后依然要求 `/login` 登录官方账号 | 旧终端未重开，或 `settings.json` 文件名多了 `.txt` / JSON 格式有语法错误 | 新开一个终端运行 `env \| grep ANTHROPIC_` 检查变量，或直接用 [CC Switch](/guide/clients/cc-switch/) 一键修复 |
| 界面右上角显示红色 `offline` 字样 | Claude Code 探测 Google/官方遥测域名的状态提示，**不影响实际 API 对话** | 只要发消息能正常流式回复即可忽略；加上 `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1"` 可减少无关请求 |
| 报错 `404 Not Found` | `ANTHROPIC_BASE_URL` 末尾多写了 `/v1` | 改成 `{{CONSOLE_URL}}` 保存并重开 `claude` |
| 报错 `401` 无效密钥 | Token 前面多写了 `Bearer `，或复制漏了字符 | 只填 `sk-...` 本身 |
| 报错 `503 无可用渠道` 或 `model not found` | `ANTHROPIC_MODEL` 与 [模型广场]({{CONSOLE_URL}}/pricing) 里的名称不一致，或当前密钥分组不支持 `/v1/messages` | 从模型广场一键复制完整 ID，或在会话内输入 `/model 模型完整ID` 切换 |

## 相关

- [CC Switch（新手一键配置）](/guide/clients/cc-switch/)
- [Anthropic Messages 原生接口](/api/anthropic-messages/)
- [错误码](/guide/errors/)
