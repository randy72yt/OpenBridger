# 在 Codex 中使用 OpenBridger

Codex 通过自定义模型供应商调用 Responses 协议。网关路径是 `POST /v1/responses`。密钥放在环境变量，配置文件只写变量名。

## 开始之前

- 已安装 Codex CLI。
- 一把名为 `codex` 的 API Key。
- 模型广场中的完整模型 ID。该模型需要能被 Responses 路径选到线路，否则会 `503`「无可用渠道」。

## 配置

```bash
export OPENBRIDGER_API_KEY="sk-你的密钥"
```

`~/.codex/config.toml`：

```toml
model = "MODEL_ID_FROM_CONSOLE"
model_provider = "openbridger"

[model_providers.openbridger]
name = "OpenBridger"
base_url = "{{API_BASE_URL}}"
env_key = "OPENBRIDGER_API_KEY"
wire_api = "responses"
```

`base_url` 含 `/v1`。`wire_api` 必须是 `responses`，这样客户端才会走 `/v1/responses`，而不是 Chat Completions。

不要把 `sk-` 密钥写进 `config.toml`。

## 启动

```bash
cd /path/to/your/project
codex
```

到 [使用日志]({{CONSOLE_URL}}/usage-logs) 核对模型和密钥名 `codex`。

恢复官方供应商时，把 `model_provider` 改回原来的值，或删掉这段供应商，然后重启 Codex。不需要退出 ChatGPT 账号。

## 检查

- 界面上的模型名与 `config.toml` 的 `model` 一致。
- `echo "$OPENBRIDGER_API_KEY" >/dev/null` 只用来看变量是否为空。不要把密钥打到屏幕上再复制。
- 变量写在个人 shell 配置里可以，不要写进项目仓库。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 401 | 环境变量为空，或密钥已禁用 |
| 503 无可用渠道 | 模型 ID 不是广场上的完整 ID，或当前分组没有 Responses 线路 |
| 请求去了别的供应商 | `model_provider` 没有改成 `openbridger`，或旧进程还在 |

## 相关

- [CC Switch](/guide/clients/cc-switch/)
- [OpenAI 兼容调用](/guide/clients/openai-compatible/)
