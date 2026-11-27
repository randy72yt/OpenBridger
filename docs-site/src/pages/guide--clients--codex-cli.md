# Codex 配置教程（桌面版 App / VS Code 插件 / CLI）

Codex 的 **桌面版客户端**、**VS Code 插件**与 **命令行 CLI** 共用同一个本地 `.codex` 目录，配好一次三端同时生效。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `codex` 的密钥（分组选支持 Responses 的分组，如 `default` / `codex`），复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID（如 `gpt-6.2-astra`、`gpt-6.2-sol`）。

---

## 方式一：用 CC Switch 图形界面一键配（零代码推荐）

1. 打开 **CC Switch**（未安装先看 [CC Switch 教程](/guide/clients/cc-switch/)）。
2. 顶部点 **`Codex`** -> 右上角点 **`+`** -> 选 **「自定义配置」**。
3. 填 `Base URL` = `{{API_BASE_URL}}`、`API Key` = `sk-...`、`Wire API` = `responses`、`Model` = `gpt-6.2-astra`。
4. 点 **「保存」** -> 点卡片上的 **「启用」** -> **彻底退出重开 Codex 桌面版**（或重载 VS Code）即完成。

---

## 方式二：手动新建 `.codex` 下的两个文件（不装辅助工具）

<div class="callout warning"><strong>桌面版必看</strong><p>双击启动 Codex 桌面版 App 不会读取终端 <code>~/.zshrc</code> 里的环境变量。按下面第 3、4 步写入 <code>auth.json</code> 与 <code>config.toml</code>，桌面版、VS Code 和 CLI 都能直接生效。</p></div>

<div class="steps">
  <div class="step"><strong>第 1 步：安装你需要的客户端（任选）</strong><p>• <strong>桌面版 App</strong>：安装官方 Codex 客户端（首次若弹登录框先关掉软件，不用登录）。<br>• <strong>VS Code 插件</strong>：扩展商店搜 <code>OpenAI Codex</code>（<code>openai.chatgpt</code>）安装。<br>• <strong>CLI 命令行</strong>：需 Node.js 22+，终端执行 <code>npm install -g @openai/codex</code>。</p></div>
  <div class="step"><strong>第 2 步：打开 <code>.codex</code> 隐藏文件夹</strong><p>• <strong>macOS</strong>：打开访达（Finder），按 <code>Command + Shift + G</code> 输入 <code>~/</code> 回车，再按 <code>Command + Shift + .</code> 显示隐藏文件，新建或打开 <code>.codex</code> 文件夹。<br>• <strong>Windows</strong>：打开文件资源管理器，顶部地址栏输入 <code>%USERPROFILE%</code> 回车，在顶部「查看」勾选<strong>「文件扩展名」</strong>，新建或打开 <code>.codex</code> 文件夹。</p></div>
  <div class="step"><strong>第 3 步：新建 <code>auth.json</code> 填入密钥</strong><p>在 <code>.codex</code> 内新建文件 <strong><code>auth.json</code></strong>（确认后缀不是 <code>.json.txt</code>），粘贴下方 JSON 并替换你的密钥。</p></div>
  <div class="step"><strong>第 4 步：新建 <code>config.toml</code> 填入地址与模型</strong><p>在 <code>.codex</code> 内再新建 <strong><code>config.toml</code></strong>，粘贴下方配置并保存。</p></div>
  <div class="step"><strong>第 5 步：彻底重启软件并验证</strong><p>桌面版按 <code>Cmd + Q</code>（Windows 右下角托盘退出）重开；VS Code 按 <code>Cmd/Ctrl + Shift + P</code> 运行 <code>Developer: Reload Window</code>；CLI 直接运行 <code>codex</code>。新建对话发一句测试，在 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 看到记录即成功。</p></div>
</div>

### `auth.json` 内容

```json
{
  "OPENAI_API_KEY": "sk-替换成你在控制台复制的密钥"
}
```

### `config.toml` 内容

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

> `model_reasoning_effort` 可改 `"medium"`（更快）或 `"high"`（适合复杂重构）；`auth.json` 含私密密钥，切勿上传 Git。

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 打开桌面版仍弹官方网页登录 | 确认 `config.toml` 含 `preferred_auth_method = "apikey"`、文件名没有变成 `.toml.txt`，彻底退出桌面版重开 |
| 客户端更新后界面下拉框找不到自定义模型 | 不要点界面内置模型下拉框，直接改 `config.toml`（或 CC Switch）里的 `model` 后重开并**新建对话** |
| 报 `401` / `503` | `401` 重复制 `sk-...`；`503` 去 [模型广场]({{CONSOLE_URL}}/pricing) 核对完整模型 ID 与密钥分组 |

## 相关

- [CC Switch（新手一键配置）](/guide/clients/cc-switch/)
- [Responses 接口](/api/openai-responses/)
- [错误码](/guide/errors/)
