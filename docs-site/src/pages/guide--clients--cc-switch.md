# CC Switch 保姆级配置教程（新手首选）

如果你不想自己找隐藏文件夹、不想手写 `config.toml` 或 `settings.json` 代码，**CC Switch** 是最省心的图形化配置工具。你在界面上填好地址和密钥点一下「启用」，它会自动帮你把 **Codex 桌面版 / VS Code 插件 / CLI** 或 **Claude Code** 的底层配置文件全部写好。

---

## 一、开始前准备（2 分钟）

1. **创建专属 API Key**
   - 打开控制台 [API 密钥]({{CONSOLE_URL}}/keys) 页面，点击右上角 **「创建密钥 / 添加密钥」**。
   - **名称**填 `cc-switch`，**分组**选择你要使用的模型所在分组（例如 `default` 或 `codex`），点击确认保存。
   - 在密钥列表里点击右侧 **「复制」** 图标，拿到完整的 `sk-...` 密钥备用。
2. **复制完整模型名称**
   - 打开 [模型广场]({{CONSOLE_URL}}/pricing)，找到你想用的模型（例如 Codex 常用 `gpt-6.2-astra`、`gpt-6.2-sol`；Claude 常用 `claude-opus-5-5`、`claude-sonnet-5-5`），点击模型名称一键复制完整 ID，不要自己手打缩写。
3. **下载并安装 CC Switch 桌面软件**
   - 前往 CC Switch 官网 `https://ccswitch.io`（或 GitHub `farion1231/cc-switch` 的 Releases 页面）下载对应系统安装包：
   - **macOS**：下载 `.dmg` 后拖入「应用程序（Applications）」。首次打开若提示「无法验证开发者」，打开系统「设置 -> 隐私与安全性」，拉到页面底部点击 **「仍要打开」**。
   - **Windows**：下载 `.msi` 或 `.exe` 安装包双击安装，完成后打开桌面上的 **CC Switch**。

<div class="callout warning"><strong>两套地址千万别填反</strong><p>配置 <strong>Codex</strong> 时，地址末尾<strong>必须带 <code>/v1</code></strong>（即 <code>{{API_BASE_URL}}</code>）；配置 <strong>Claude Code</strong> 时，地址末尾<strong>绝对不要带 <code>/v1</code></strong>（只填 <code>{{CONSOLE_URL}}</code>）。</p></div>

---

## 二、用 CC Switch 配置 Codex（桌面版 App / VS Code / CLI 通用）

Codex 桌面客户端、VS Code 扩展和命令行共用同一套底层配置，在 CC Switch 里配一次，三端同时生效：

<div class="steps">
  <div class="step"><strong>第 1 步：切换到 Codex 页面</strong><p>打开 CC Switch 软件，在顶部（或左侧）应用栏点击切换到 <strong><code>Codex</code></strong> 标签页。</p></div>
  <div class="step"><strong>第 2 步：新建自定义供应商</strong><p>点击界面右上角的 <strong><code>+</code>（添加供应商 / Add Provider）</strong> 按钮；如果弹出预设模板列表，选择 <strong>「自定义配置（Custom）」</strong>。</p></div>
  <div class="step"><strong>第 3 步：逐项填写配置表单</strong><p>严格对照下方卡片填入每一栏，不要多加空格或斜杠：</p></div>
  <div class="step"><strong>第 4 步：保存并点击「启用」</strong><p>点击弹窗右下角的 <strong>「保存 / 添加」</strong>。回到供应商列表，找到刚创建的 <code>OpenBridger</code> 卡片，点击卡片上的 <strong>「启用（Enable）」</strong> 按钮，确认卡片亮起「已启用 / 当前使用」标记。</p></div>
  <div class="step"><strong>第 5 步：重启 Codex 客户端让配置生效</strong><p>如果 Codex 桌面版或 VS Code 刚才已经开着，<strong>必须彻底退出后重开</strong>（macOS 按 <code>Cmd + Q</code> 完全退出 Codex 桌面端再打开；VS Code 按 <code>Cmd + Shift + P</code> 输入 <code>Reload Window</code> 回车）。新开一个会话发一句 <code>你好</code> 即可使用。</p></div>
</div>

### Codex 表单逐项填写对照表

<div class="config-card">
  <div><span>供应商名称（Name）</span><code>OpenBridger</code>（仅作本地显示，可任意取名）</div>
  <div><span>API 地址 / Base URL</span><code>{{API_BASE_URL}}</code>（<strong>必须以 /v1 结尾</strong>）</div>
  <div><span>API 密钥（API Key）</span>粘贴你在控制台复制的 <code>sk-...</code> 完整密钥</div>
  <div><span>接口协议（Wire API）</span>选择 <code>responses</code>（若默认已是则保持不动）</div>
  <div><span>默认模型（Model）</span><code>gpt-6.2-astra</code> 或 <code>gpt-6.2-sol</code>（从模型广场复制）</div>
  <div><span>推理强度（Reasoning Effort，可选）</span>日常写代码选 <code>medium</code> 或 <code>high</code></div>
</div>

---

## 三、用 CC Switch 配置 Claude Code（桌面端 / CLI 通用）

<div class="steps">
  <div class="step"><strong>第 1 步：切换到 Claude Code 页面</strong><p>在 CC Switch 顶部（或左侧）应用栏点击切换到 <strong><code>Claude Code</code></strong> 标签页。</p></div>
  <div class="step"><strong>第 2 步：新建自定义供应商</strong><p>点击右上角 <strong><code>+</code>（添加供应商）</strong>，在预设下拉菜单中选择 <strong>「自定义配置（Custom）」</strong>。</p></div>
  <div class="step"><strong>第 3 步：填写地址、密钥与模型</strong><p>注意这里的地址<strong>不带 <code>/v1</code></strong>（详见下方对照表）。</p></div>
  <div class="step"><strong>第 4 步：保存并点击「启用」</strong><p>点击 <strong>「保存」</strong>，然后在列表里点击 <code>OpenBridger Claude</code> 卡片上的 <strong>「启用（Enable）」</strong> 按钮。</p></div>
  <div class="step"><strong>第 5 步：重新打开终端或桌面端验证</strong><p>关掉旧的终端窗口，新开一个终端输入 <code>claude</code>（或重启 Claude 桌面端），不需要登录 Anthropic 官方账号即可直接提问。</p></div>
</div>

### Claude Code 表单逐项填写对照表

<div class="config-card">
  <div><span>供应商名称（Name）</span><code>OpenBridger Claude</code></div>
  <div><span>请求地址 / Base URL</span><code>{{CONSOLE_URL}}</code>（<strong>不要加 /v1！</strong>）</div>
  <div><span>API Key / Auth Token</span>粘贴你在控制台复制的 <code>sk-...</code> 完整密钥</div>
  <div><span>主模型 / 模型映射（Model，可选）</span>填入广场中的完整 ID，如 <code>claude-sonnet-5-5</code> 或 <code>claude-opus-5-5</code></div>
</div>

---

## 四、进阶技巧：在 Claude Code 或 Codex 里跨厂商换模型

只要你的密钥分组里有对应模型且支持该协议，你可以在 CC Switch 里随时新建多张卡片一键切换：

- **在 Claude Code 里用 DeepSeek / GPT**：新建一张 Claude Code 自定义卡片，Base URL 仍填 `{{CONSOLE_URL}}`，把模型字段改成广场里的 `deepseek-v4.1` 或 `gpt-6.2-astra`，点「启用」后新开会话即可。
- **日常省额度技巧**：复杂重构或 Debug 时在 CC Switch 点启用旗舰模型卡片；写普通注释或小脚本时一键切到轻量模型卡片，每次切换后**新开一个对话窗口**避免旧上下文缓存冲突。

## 五、新手最常踩的 4 个坑

| 遇到的问题 | 原因 | 10 秒解决方法 |
| --- | --- | --- |
| 在 CC Switch 里保存了，但打开软件还是提示登录官方账号 | ① 忘了点卡片上的 **「启用」** 按钮；② 软件没彻底重启 | 回 CC Switch 确认卡片显示「已启用」，然后按 `Cmd+Q` / 任务栏右键**彻底退出**客户端再重开 |
| 发送消息报 `404` | Codex 漏写了 `/v1`，或者 Claude Code 多写了 `/v1` | Codex 填 `{{API_BASE_URL}}`；Claude Code 填 `{{CONSOLE_URL}}` |
| 发送消息报 `503 无可用渠道` | 填的模型名字手打错了，或者建密钥时选的分组不支持该模型 | 去 [模型广场]({{CONSOLE_URL}}/pricing) 重新点复制模型名，检查密钥分组 |
| 发送消息报 `403 用户额度不足` | 新注册账号余额为 0，尚未充值或兑换额度 | 先到 [钱包]({{CONSOLE_URL}}/wallet) 充值或兑换后再试 |

## 相关

- [Codex（桌面版 / 插件 / CLI）](/guide/clients/codex-cli/)
- [Claude Code（桌面端 / CLI）](/guide/clients/claude-code/)
- [用量与日志](/guide/account/usage/)
- [错误码](/guide/errors/)
