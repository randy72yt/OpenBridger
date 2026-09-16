# 在 Codex 中使用 OpenBridger

Codex 可以通过自定义模型供应商连接 OpenBridger。API Key 使用环境变量保存，配置文件只引用环境变量名。

## 配置 API Key

```bash
export OPENBRIDGER_API_KEY="YOUR_API_KEY"
```

## 配置自定义供应商

编辑 `~/.codex/config.toml`：

```toml
model = "MODEL_ID_FROM_CONSOLE"
model_provider = "openbridger"

[model_providers.openbridger]
name = "OpenBridger"
base_url = "{{API_BASE_URL}}"
env_key = "OPENBRIDGER_API_KEY"
wire_api = "responses"
```

<div class="callout"><strong>模型与协议</strong><p>模型名称必须与控制台展示的完整 ID 一致。Codex 使用 Responses 协议，因此所选模型和线路也必须支持该协议。</p></div>

## 启动与验证

```bash
cd /path/to/your/project
codex
```

- `401`：检查环境变量是否为空、Key 是否有效。
- `model not found`：从控制台重新复制完整模型 ID。
- 请求无法完成：先在 AI 体验台测试同一模型，再查看使用日志。

<div class="callout warning"><strong>不要把密钥写进项目</strong><p>不要把真实 API Key 写进 config.toml 或代码仓库。需要恢复官方服务时，切回原模型供应商并重新启动 Codex。</p></div>

## 配置检查

- 启动后确认界面显示的模型与 `config.toml` 一致。
- 使用 `echo $OPENBRIDGER_API_KEY` 只检查变量是否为空，不要把输出复制到聊天或工单。
- 长期使用时可在个人 shell 配置中设置变量，但不要写进项目目录。
- 不需要退出 ChatGPT 账号；切换 `model_provider` 后重启 Codex 即可切换服务。
