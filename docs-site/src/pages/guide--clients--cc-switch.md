# 通过 CC Switch 管理 OpenBridger 配置

目前采用手动新建供应商配置的方式。参考资料中的浏览器一键唤起和自动导入属于其他平台能力，OpenBridger 尚未声称支持。

<div class="config-card"><div><span>Profile Name</span><code>OpenBridger</code></div><div><span>Base URL</span><code>{{CONSOLE_URL}}</code></div><div><span>Auth Token</span><code>YOUR_API_KEY</code></div><div><span>Model</span><code>控制台中的完整模型 ID</code></div></div>

## 配置步骤

<div class="steps"><div class="step"><strong>创建独立密钥</strong><p>在 OpenBridger 控制台创建一个仅供 CC Switch 使用的 API Key。</p></div><div class="step"><strong>选择目标应用</strong><p>在 CC Switch 中为 Codex 或 Claude Code 新建自定义供应商。</p></div><div class="step"><strong>填写地址与模型</strong><p>根据目标应用选择 OpenAI Responses 或 Anthropic 兼容配置，并填写完整模型 ID。</p></div><div class="step"><strong>启用并重启</strong><p>保存、切换到新配置，然后重启对应客户端。</p></div></div>

<div class="callout warning"><strong>排障建议</strong><p>同时只保留一套生效的 Base URL 和 API Key。旧环境变量可能覆盖刚切换的配置。</p></div>

## 环境变量配置

Claude Code：

```bash
export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
export ANTHROPIC_AUTH_TOKEN="YOUR_API_KEY"
```

OpenAI Compatible：

```bash
export OPENAI_BASE_URL="{{API_BASE_URL}}"
export OPENAI_API_KEY="YOUR_API_KEY"
```

## 常见异常

| 现象 | 可能原因 | 处理方式 |
| --- | --- | --- |
| 切换后仍走旧供应商 | 终端仍保留旧环境变量 | 重开终端并检查 `OPENAI`、`ANTHROPIC` 变量 |
| unauthorized | Key 错误、空格或已禁用 | 重新复制或创建 CC Switch 专用 Key |
| model not found | 当前分组没有模型权限 | 从控制台复制完整模型 ID |
| connection refused / timeout | 网络、代理或 DNS 问题 | 检查端点连通性并排除冲突代理 |
| 一直使用默认模型 | Profile 模型字段未生效 | 保存后重启目标客户端 |
| 请求很快失败 | 额度不足或预扣失败 | 检查钱包、套餐和使用日志 |
