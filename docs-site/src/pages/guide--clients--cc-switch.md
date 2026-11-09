# 用 CC Switch 切换 OpenBridger

CC Switch 用来在本机保存多套供应商，再写进 Codex 或 Claude Code。OpenBridger 没有「浏览器一键导入」的专用链接。请在 CC Switch 里手工新建供应商。

## 两套地址不要混用

<div class="config-card"><div><span>Claude Code 的 Base URL</span><code>{{CONSOLE_URL}}</code></div><div><span>Codex / OpenAI 兼容的 Base URL</span><code>{{API_BASE_URL}}</code></div><div><span>密钥</span><code>sk-...</code>，建议单独一把，名为 <code>cc-switch</code></div><div><span>模型</span><code>模型广场中的完整 ID</code></div></div>

## 步骤

<div class="steps"><div class="step"><strong>在控制台建密钥</strong><p>见 <a href="/guide/api-key/">创建 API Key</a>。</p></div><div class="step"><strong>在 CC Switch 里选目标应用</strong><p>Claude Code 用 Anthropic 兼容，Base URL 为站点根。Codex 用 Responses，Base URL 含 <code>/v1</code>。</p></div><div class="step"><strong>填写模型 ID</strong><p>从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制。</p></div><div class="step"><strong>启用并重启客户端</strong><p>保存后完全退出 Codex 或 Claude Code 再打开。只切界面、不重启时，旧环境变量还会生效。</p></div></div>

同时只保留一套 Base URL。终端里残留的 `ANTHROPIC_BASE_URL` 或 `OPENAI_BASE_URL` 会盖过刚刚切换的配置。

## 等价的环境变量

Claude Code：

```bash
export ANTHROPIC_BASE_URL="{{CONSOLE_URL}}"
export ANTHROPIC_AUTH_TOKEN="sk-你的密钥"
```

OpenAI 兼容：

```bash
export OPENAI_BASE_URL="{{API_BASE_URL}}"
export OPENAI_API_KEY="sk-你的密钥"
```

字段说明分别见 [Claude Code](/guide/clients/claude-code/) 和 [Codex](/guide/clients/codex-cli/)。

## 常见问题

| 现象 | 处理 |
| --- | --- |
| 切换后仍是旧供应商 | 新开终端，检查 `OPENAI_` 与 `ANTHROPIC_` 变量 |
| 401 | 重新复制密钥，去掉空格 |
| 503 无可用渠道 | 模型 ID 或协议和分组不匹配 |
| 很快失败且日志写额度不足 | 看钱包和密钥自身额度 |

## 相关

- [故障排查](/troubleshooting/)
