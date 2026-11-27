# 在 OpenCode / Aider 中使用 OpenBridger

OpenCode 与 Aider 是两款轻量快速的开源终端 AI 编程工具，支持在同一个项目里按需混用 GPT-6.2、Claude 5.5、DeepSeek V4.1 等不同厂商模型（例如用推理模型做架构规划，用高性价比模型执行代码修改）。

## 开始之前

- 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建一把终端专用密钥。
- 从 [模型广场]({{CONSOLE_URL}}/pricing) 复制要使用的完整模型 ID。

---

## 一、OpenCode 配置方式

1. 安装 OpenCode：

```bash
npm install -g opencode-ai
```

2. 编辑配置文件 `~/.config/opencode/opencode.json`，添加自定义 OpenAI 兼容供应商（通过环境变量读取密钥，避免明文落盘）：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "provider": {
    "openbridger": {
      "npm": "@ai-sdk/openai-compatible",
      "name": "OpenBridger",
      "options": {
        "baseURL": "{{API_BASE_URL}}",
        "apiKey": "{env:OPENBRIDGER_API_KEY}"
      },
      "models": {
        "gpt-6.2-astra": {
          "name": "GPT-6.2 Astra"
        },
        "claude-sonnet-5-5": {
          "name": "Claude Sonnet 5.5"
        },
        "deepseek-v4.1": {
          "name": "DeepSeek V4.1"
        }
      }
    }
  }
}
```

3. 在终端导出密钥并运行 `opencode`：

```bash
export OPENBRIDGER_API_KEY="sk-你的密钥"
opencode
```

---

## 二、Aider 配置方式

Aider 基于 LiteLLM 路由，给模型名加上 `openai/` 前缀即可将任意模型走 OpenAI 兼容协议发往 OpenBridger：

```bash
# 安装（推荐使用 uv 或 pipx）
pipx install aider-chat

# 导出网关地址（带 /v1）与密钥
export OPENAI_API_BASE="{{API_BASE_URL}}"
export OPENAI_API_KEY="sk-你的密钥"

# 指定模型启动（保留 openai/ 前缀以强制使用 OpenAI 兼容协议）
aider --model openai/claude-sonnet-5-5
```

若使用 Architect 双模型模式（一个模型负责推理方案，一个模型负责落盘编辑）：

```bash
aider --architect \
  --model openai/gpt-6.2-astra \
  --editor-model openai/deepseek-v4.1-flash
```

## 常见问题

| 现象 | 处理 |
| --- | --- |
| Aider 提示缺少 `ANTHROPIC_API_KEY` | 模型名前面漏写了 `openai/` 前缀，加上 `openai/` 即可复用同一把 `OPENAI_API_KEY` |
| `503` 无可用渠道 | 检查 `models` 里填写的模型 ID 是否在 [模型广场]({{CONSOLE_URL}}/pricing) 当前分组下可用 |

## 相关

- [OpenAI Codex](/guide/clients/codex-cli/)
- [Claude Code](/guide/clients/claude-code/)
- [Base URL 填写速查](/guide/base-urls/)
