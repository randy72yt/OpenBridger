# 在 Gemini CLI 中使用 OpenBridger

Google 官方开源的 `gemini-cli` 支持通过环境变量指定自定义网关地址，直连网关上的 `/v1beta` 原生 Gemini 接口完成代码分析、百万级长上下文检索与终端操作。

## 开始之前

- 本机已安装 Node.js 20 或更高版本。
- 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建一把名为 `gemini-cli` 的密钥。
- 在 [模型广场]({{CONSOLE_URL}}/pricing) 确认当前分组下可用的 Gemini 模型 ID（如 `gemini-4-neon`、`gemini-4-carbon`、`gemini-4-hydrogen`）。

## 1. 安装 Gemini CLI

```bash
npm install -g @google/gemini-cli
```

## 2. 配置环境变量并启动

`GOOGLE_GEMINI_BASE_URL` 填站点根地址 `{{CONSOLE_URL}}`（**不要**带 `/v1` 或 `/v1beta`，CLI 会自动补齐路径）：

```bash
export GEMINI_API_KEY="sk-你的密钥"
export GOOGLE_GEMINI_BASE_URL="{{CONSOLE_URL}}"

cd /path/to/your/project
gemini -m gemini-4-neon
```

启动时如果询问认证方式，选择 **Use Gemini API Key**。建议显式带上 `-m 模型ID`，确保发出的模型名称与 [模型广场]({{CONSOLE_URL}}/pricing) 中的在售模型完全一致。

如需写入长期个人环境，可将两行 `export` 放入 `~/.zshrc` 或 `~/.bashrc`（切勿写入项目代码仓库）。

## 3. 非交互单次调用示例

在脚本或 Git Hook 中直接传入提示词执行：

```bash
gemini -m gemini-4-hydrogen -p "总结最近一次 git diff 的改动要点"
```

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 提示登录 Google 账号 | 终端未读取到 `GEMINI_API_KEY`，运行 `env \| grep GEMINI` 检查是否已导出 |
| `404` 报错 | `GOOGLE_GEMINI_BASE_URL` 多写了 `/v1` 或 `/v1beta`，应改为 `{{CONSOLE_URL}}` |
| `503` 无可用渠道 | 未加 `-m` 参数导致 CLI 默认请求的模型不在你的当前分组内，使用 `-m` 指定广场中的完整模型 ID |

## 相关

- [Google Gemini 原生接口](/api/gemini-native/)
- [OpenAI Codex](/guide/clients/codex-cli/)
- [Claude Code](/guide/clients/claude-code/)
