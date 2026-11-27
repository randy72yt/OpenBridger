# Gemini CLI 配置教程

Google 官方开源的 `gemini-cli` 走原生 `/v1beta` 接口，适合百万 Token 超长代码库分析与终端操作。

## 开始前准备

1. 本机已装 Node.js 20+（终端执行 `npm install -g @google/gemini-cli` 安装）。
2. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `gemini-cli` 的密钥，复制 `sk-...`。
3. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制 Gemini 模型完整 ID（如 `gemini-4-neon`、`gemini-4-carbon`、`gemini-4-hydrogen`）。

<div class="callout warning"><strong>地址不要带 /v1 或 /v1beta</strong><p><code>GOOGLE_GEMINI_BASE_URL</code> 只填站点根地址 <strong><code>{{CONSOLE_URL}}</code></strong>，CLI 会自动拼接 <code>/v1beta/...</code>。</p></div>

---

## 方式一：用 CC Switch 图形界面配置（支持 Gemini CLI）

打开 **CC Switch**（见 [CC Switch 教程](/guide/clients/cc-switch/)）-> 顶部切到 **`Gemini`** -> 点右上角 **`+`** 新建自定义配置 -> Base URL 填 `{{CONSOLE_URL}}`、API Key 填 `sk-...` -> 保存并点 **「启用」**，新开终端运行 `gemini -m gemini-4-neon` 即可。

---

## 方式二：终端命令或 `.gemini/.env` 配置

### 1. 长期生效（推荐）：写入 `~/.gemini/.env`

在个人主目录下的 `.gemini` 文件夹新建 `.env` 文件（macOS/Linux 为 `~/.gemini/.env`，Windows 为 `%USERPROFILE%\.gemini\.env`）：

```ini
GEMINI_API_KEY=sk-替换成你的密钥
GOOGLE_GEMINI_BASE_URL={{CONSOLE_URL}}
GEMINI_MODEL=gemini-4-neon
```

保存后在任意代码目录运行 `gemini`（首次启动若询问登录方式，选 **Use Gemini API Key**）。

### 2. 当前终端单次启动

- **macOS / Linux**：
  ```bash
  export GEMINI_API_KEY="sk-你的密钥"
  export GOOGLE_GEMINI_BASE_URL="{{CONSOLE_URL}}"
  gemini -m gemini-4-neon
  ```
- **Windows PowerShell**：
  ```powershell
  $env:GEMINI_API_KEY="sk-你的密钥"
  $env:GOOGLE_GEMINI_BASE_URL="{{CONSOLE_URL}}"
  gemini -m gemini-4-neon
  ```

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 启动弹浏览器要求登录 Google 账号 | 首次认证菜单选 **Use Gemini API Key**，确认 `GEMINI_API_KEY` 已写入 |
| 报 `404` | `GOOGLE_GEMINI_BASE_URL` 改为 `{{CONSOLE_URL}}`（删掉末尾的 `/v1` 或 `/v1beta`） |
| 报 `503 无可用渠道` | 启动时加上 `-m 广场里的完整模型ID`（避免默认请求未开放的旧模型名） |

## 相关

- [CC Switch（新手一键配置）](/guide/clients/cc-switch/)
- [Google Gemini 原生接口](/api/gemini-native/)
- [错误码](/guide/errors/)
