# 在 Claude Code 中使用 OpenBridger

Claude Code 使用 Anthropic Messages 协议。网关对应的路径是 `POST /v1/messages`。密钥可以放在 `x-api-key`（Claude Code 会这么做），网关在这条路径上接受它。

模型名必须是 [模型广场]({{CONSOLE_URL}}/pricing) 里的完整 ID，而且当前分组要有能处理 Messages 请求的线路。只在广场里看得到名字还不够，需要实际发一条短请求并用 [使用日志]({{CONSOLE_URL}}/usage-logs) 确认。

## 开始之前

- 已安装 `claude` 命令。
- 一把名为 `claude-code` 的 API Key。
- 知道你要使用的模型 ID。

## 当前终端临时生效

```bash
export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
export ANTHROPIC_AUTH_TOKEN="sk-你的密钥"

cd /path/to/your/project
claude
```

`ANTHROPIC_BASE_URL` 是站点根，**不要**写成 `{{API_BASE_URL}}`。Claude Code 会自己请求 `{BASE}/v1/messages`。写成带 `/v1` 的地址会变成 `/v1/v1/messages`。

关掉这个终端后变量就没了。恢复官方服务：

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_AUTH_TOKEN
```

然后新开终端再运行 `claude`。

## 写进 Claude Code 的配置

`~/.claude/settings.json`：

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "{{CONSOLE_URL}}",
    "ANTHROPIC_AUTH_TOKEN": "sk-你的密钥"
  }
}
```

这个文件里有密钥，不要提交，不要截图。项目目录里的 `.claude/settings.json` 同样不要写入真实密钥。

如果客户端允许指定模型，把值设成广场上的完整 ID。Claude Code 默认发出的模型名如果和广场不一致，会得到 `503`「无可用渠道」或上游的 model not found。以日志里记录的模型 ID 为准。

检查当前终端实际生效的变量：

```bash
env | grep ANTHROPIC_
```

只确认变量在不在，不要把输出贴到聊天或工单。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 仍要求登录官方账号 | 新开终端，确认两个 `ANTHROPIC_` 变量已导出，再启动 `claude` |
| 401 | 重新复制密钥。不要在 token 前面再加 `Bearer` |
| 503 无可用渠道 | 广场中的模型 ID 与 Claude Code 发出的名字不一致，或该分组没有 Messages 线路 |
| 请求没进使用日志 | Base URL 还指着官方主机。用 `env | grep ANTHROPIC_BASE_URL` 核对 |

<div class="callout warning"><strong>协议要对上</strong><p>Chat Completions 可用，不等于 Claude Code 可用。需要模型、分组和 <code>/v1/messages</code> 同时成立。</p></div>

## 相关

- [CC Switch](/guide/clients/cc-switch/)
- [错误码](/guide/errors/)
