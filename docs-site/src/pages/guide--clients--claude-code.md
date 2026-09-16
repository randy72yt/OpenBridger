# 在 Claude Code 中使用 OpenBridger

支持 Anthropic 兼容协议的线路可以通过环境变量连接 Claude Code。请先在控制台确认目标模型和线路可用。

## 临时配置

```bash
export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
export ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"

cd /path/to/your/project
claude
```

关闭当前终端后，这组临时变量会失效。要恢复官方服务，运行：

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_AUTH_TOKEN
```

## 常见问题

| 现象 | 检查方式 |
| --- | --- |
| 仍然要求登录 | 重新打开终端，确认两个 `ANTHROPIC_` 环境变量已生效 |
| 401 / unauthorized | 重新复制 API Key，避免空格、换行和重复的 Bearer 前缀 |
| model not found | 确认当前分组拥有 Claude Code 请求模型的权限 |
| 请求立即失败 | 在 AI 体验台测试模型，并检查余额与使用日志 |

<div class="callout warning"><strong>兼容性说明</strong><p>平台支持的模型不等于都支持 Claude Code。模型、线路和 Anthropic 协议三者需要同时匹配。</p></div>

## 持久化配置

也可以把环境变量写入 `~/.claude/settings.json`：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{CONSOLE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "YOUR_API_KEY"
  }
}
```

该文件包含 API Key，不要提交到仓库或分享截图。团队项目不要把真实密钥写入项目级配置。可使用 `env | grep ANTHROPIC` 检查当前终端实际生效的变量。
