# OpenAI Codex 完整配置教程（桌面版 App / VS Code 插件 / CLI）

OpenAI Codex 目前有三种使用形态：**Codex 桌面客户端（macOS / Windows）**、**VS Code / Cursor 内的 Codex 扩展**，以及 **终端命令行 `codex` CLI**。

好消息是：**这三种形态共用同一个本地配置文件夹（`.codex`）**。只要配好一次，桌面版、VS Code 插件和命令行就同时打通了。下面提供两种配置方式：**新手强烈推荐「方式一（图形界面一键配）」**，熟悉文件的同学也可以用「方式二（手动新建两个配置文件）」。

---

## 开始之前（必做 2 步）

1. 打开 [API 密钥]({{CONSOLE_URL}}/keys)，点击「创建密钥」，名称填 `codex`，分组选择支持 Responses 协议的分组（如 `default` 或 `codex`），复制生成的 `sk-...` 密钥。
2. 打开 [模型广场]({{CONSOLE_URL}}/pricing)，复制你要使用的完整模型 ID（例如 `gpt-6.2-astra` 或 `gpt-6.2-sol`）。

---

## 方式一：用 CC Switch 图形界面配置（零代码，最推荐新手）

不用找隐藏文件夹，也不用担心文件后缀改错：

1. 下载并打开 **CC Switch**（完整步骤见 [CC Switch 新手教程](/guide/clients/cc-switch/)）。
2. 在 CC Switch 顶部切换到 **`Codex`** -> 点击右上角 **`+`** -> 选 **「自定义配置」**。
3. 依次填入：
   - **名称**：`OpenBridger`
   - **Base URL**：`{{API_BASE_URL}}`（带 `/v1`）
   - **API Key**：粘贴 `sk-...`
   - **协议（Wire API）**：`responses`
   - **模型（Model）**：`gpt-6.2-astra`（或从模型广场复制的模型 ID）
4. 点击 **「保存」**，再点击卡片上的 **「启用」**。
5. **彻底退出并重新打开 Codex 桌面版 App**（或重载 VS Code），直接开始对话即可！

---

## 方式二：手动修改 `.codex` 配置文件（桌面版 / VS Code / CLI 通用）

如果你不想装辅助工具，按下面 **1 -> 2 -> 3 -> 4 -> 5** 五步手工创建两个小文件即可。

<div class="callout warning"><strong>为什么桌面版不能只配终端 export 环境变量？</strong><p>在 macOS / Windows 上直接双击图标打开 <strong>Codex 桌面版 App</strong> 时，系统桌面进程默认不会读取你在终端 <code>~/.zshrc</code> 里写的 <code>export</code> 变量。直接把密钥写进下方的个人 <code>auth.json</code> 文件，桌面版、VS Code 插件和 CLI 三端都能稳定读取。</p></div>

<div class="steps">
  <div class="step">
    <strong>第 1 步：安装 Codex（三选一或全装均可）</strong>
    <p><strong>• 桌面版 App</strong>：安装官方 OpenAI Codex 桌面客户端（首次打开如果弹出登录 ChatGPT 账号界面，先直接关掉软件，不用登录官方账号）。<br>
    <strong>• VS Code 插件</strong>：在 VS Code 左侧扩展商店搜索 <code>OpenAI Codex</code>（扩展 ID <code>openai.chatgpt</code>）点击安装。<br>
    <strong>• 命令行 CLI</strong>：电脑需先装 Node.js 22+，在终端运行 <code>npm install -g @openai/codex</code>。</p>
  </div>
  <div class="step">
    <strong>第 2 步：打开电脑上的 <code>.codex</code> 配置文件夹</strong>
    <p><strong>• macOS 打开方法</strong>：打开「访达（Finder）」，按下快捷键 <code>Command + Shift + G</code>，输入 <code>~/.codex</code> 回车。如果提示找不到文件夹，就输入 <code>~/</code> 回车进入个人主目录，按 <code>Command + Shift + .</code>（句号）显示隐藏文件，右键新建一个名为 <code>.codex</code> 的文件夹并点进去。<br>
    <strong>• Windows 打开方法</strong>：打开「此电脑 / 文件资源管理器」，点击顶部地址栏输入 <code>%USERPROFILE%</code> 回车（例如进入 <code>C:\Users\你的用户名</code>），在里面新建一个名为 <code>.codex</code> 的文件夹并双击进入。</p>
  </div>
  <div class="step">
    <strong>第 3 步：新建密钥文件 <code>auth.json</code></strong>
    <p>在 <code>.codex</code> 文件夹里新建一个纯文本文件，命名为 <strong><code>auth.json</code></strong>（Windows 用户请在文件资源管理器顶部「查看」里勾选「文件扩展名」，确保文件名不是 <code>auth.json.txt</code>），把下面内容复制进去并替换成你的密钥后保存：</p>
  </div>
  <div class="step">
    <strong>第 4 步：新建路由与模型文件 <code>config.toml</code></strong>
    <p>在同一个 <code>.codex</code> 文件夹里再新建一个文件，命名为 <strong><code>config.toml</code></strong>（同样确认不是 <code>config.toml.txt</code>），把下方的 <code>config.toml</code> 配置完整复制进去并保存。</p>
  </div>
  <div class="step">
    <strong>第 5 步：重启桌面版 / 重载 VS Code 并验证</strong>
    <p><strong>• Codex 桌面版</strong>：如果之前开着，macOS 按 <code>Command + Q</code>（Windows 在右下角托盘图标右键点退出）彻底关闭，再重新打开 Codex 桌面端，新建一个对话。<br>
    <strong>• VS Code 插件</strong>：按下 <code>Command + Shift + P</code>（Windows 按 <code>Ctrl + Shift + P</code>），输入 <code>Developer: Reload Window</code> 回车重载窗口，打开左侧 Codex 面板新建会话。<br>
    <strong>• CLI 命令行</strong>：在项目文件夹下直接输入 <code>codex</code> 回车。<br>
    发送一句测试消息后，打开控制台 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a>，能看到刚才这把密钥的请求记录即大功告成！</p>
  </div>
</div>

### 第 3 步要复制的 `auth.json` 内容

```json
{
  "OPENAI_API_KEY": "sk-把这里替换成你在控制台复制的密钥"
}
```

### 第 4 步要复制的 `config.toml` 内容

```toml
model_provider = "openbridger"
model = "gpt-6.2-astra"
model_reasoning_effort = "high"
disable_response_storage = true
preferred_auth_method = "apikey"

[model_providers.openbridger]
name = "OpenBridger"
base_url = "{{API_BASE_URL}}"
wire_api = "responses"
```

- `model`：填你在 [模型广场]({{CONSOLE_URL}}/pricing) 复制的完整模型 ID（例如 `gpt-6.2-astra`、`gpt-6.2-sol`）。
- `model_reasoning_effort`：可选 `"low"`、`"medium"`、`"high"`（日常小改动用 `"medium"` 更快，复杂架构与排查 Bug 用 `"high"`）。
- `wire_api`：**必须保持 `"responses"`**，不能改成 `"chat"`。
- **安全提醒**：`~/.codex/auth.json` 里包含你的私密密钥，切勿上传到 GitHub 或发给他人。

---

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 打开 Codex 桌面版仍然强制弹出官方网页登录框 | ① `config.toml` 漏写了 `preferred_auth_method = "apikey"`；② Windows 下文件名实际变成了 `config.toml.txt` | 开启系统「显示文件扩展名」去掉多余的 `.txt`，确认文件在 `~/.codex/` 下，彻底退出桌面版重开 |
| Codex 自动更新版本后，下拉框找不到自定义模型 | 新版界面只展示官方内置列表，但底层的 `config.toml` 依然生效 | 不要在界面下拉框乱切官方模型，直接在 `config.toml`（或 CC Switch）里改 `model = "..."` 后重启客户端并**新建对话** |
| 报错 `401` 无效的令牌 | `auth.json` 里的 `sk-...` 前后多了空格，或密钥额度已用尽 | 重新从控制台复制粘贴 `sk-...`，不要加 `Bearer ` |
| 报错 `503` 分组下无可用渠道 | 当前密钥所属分组不支持 Responses 协议，或模型 ID 拼写不一致 | 去 [模型广场]({{CONSOLE_URL}}/pricing) 核对模型完整 ID 及可用分组 |
| 换了新模型但在旧对话里还是报错 | 旧会话绑定了之前的供应商状态与推理上下文 | 点击客户端左上角 **新建对话（New Thread / New Chat）** 再发消息 |

## 相关

- [CC Switch（新手一键配置）](/guide/clients/cc-switch/)
- [Responses 接口（Codex / 推理）](/api/openai-responses/)
- [错误码](/guide/errors/)
