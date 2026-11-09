# 在 Open WebUI 中使用 OpenBridger

添加全局 OpenAI 连接需要 Open WebUI 的管理员。普通用户只能使用管理员已经启用的模型。

## 在管理界面添加

<div class="steps"><div class="step"><strong>打开连接设置</strong><p>Admin Panel 或 Settings 里的 OpenAI API Connections。</p></div><div class="step"><strong>新增一条</strong><p>名称 OpenBridger，URL 填 <code>{{API_BASE_URL}}</code>，密钥用单独创建的 API Key。</p></div><div class="step"><strong>刷新模型</strong><p>保存后刷新。自动发现失败时，按 Open WebUI 的模型过滤规则手工填入模型广场中的 ID。</p></div><div class="step"><strong>用短消息核对</strong><p>在 OpenBridger <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 里应出现这把密钥。</p></div></div>

<div class="callout warning"><strong>一把密钥会被实例上的用户共用</strong><p>管理员保存的连接对多个 Open WebUI 用户生效时，用量都记在这把密钥上。按实例或团队分开建密钥，并在密钥上设置额度上限。</p></div>

## 用环境变量

```bash
OPENAI_API_BASE_URL={{API_BASE_URL}}
OPENAI_API_KEY=sk-你的密钥
```

改的是容器环境，不是浏览器。修改 Docker Compose 后要重新创建或重启 Open WebUI 容器。只刷新网页不会读到新值。

若你使用的 Open WebUI 版本把变量名写成 `OPENAI_API_BASE_URLS`（带 S）或要求 URL 列表，以该版本文档为准，值仍然是 `{{API_BASE_URL}}`。

## 相关

- [创建 API Key](/guide/api-key/)
- [用量与日志](/guide/account/usage/)
