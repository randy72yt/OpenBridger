# Open WebUI 配置教程（课题组共享 / 自部署面板）

管理员在 Open WebUI 后台添加一次 OpenBridger 连接，即可开放给实例成员使用。

<div class="callout warning"><strong>共享密钥提醒</strong><p>实例成员共用同一连接扣费，建议在 <a href="{{CONSOLE_URL}}/keys">API 密钥</a> 为该实例单独创建密钥并设好额度上限。</p></div>

## 方式一：管理员后台可视化配置（推荐）

<div class="steps">
  <div class="step"><strong>第 1 步：进入外部连接</strong><p>用管理员账号登录 -> 点个人头像 -> <strong>「管理员面板（Admin Panel）」</strong> -> 顶部 <strong>「设置」</strong> -> 左侧 <strong>「外部连接（Connections）」</strong>。</p></div>
  <div class="step"><strong>第 2 步：新增 OpenAI 连接</strong><p>在 <strong><code>OpenAI API</code></strong> 行右侧点击 <strong><code>+</code></strong> 按钮（如果不用本地 Ollama，顺手把下方 <code>Ollama API</code> 开关关掉避免页面卡顿）。</p></div>
  <div class="step"><strong>第 3 步：填写地址、密钥与模型白名单</strong>
    <div class="config-card">
      <div><span>URL</span><code>{{API_BASE_URL}}</code>（带 <code>/v1</code>）</div>
      <div><span>API Key</span>单独创建的 <code>sk-...</code> 密钥</div>
      <div><span>模型 ID（推荐填）</span>输入常用完整 ID（如 <code>gpt-6.2-astra</code>、<code>claude-sonnet-5-5</code>）点 <code>+</code> 加入</div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：验证并点两次「保存」</strong><p>点 URL 右侧 <strong>🔄 验证</strong> 提示通过 -> 点弹窗右下角 <strong>「保存」</strong> -> <strong>再拉到外部连接页面最底部点一次总「保存」按钮</strong>。</p></div>
  <div class="step"><strong>第 5 步：开放成员可见权限（多人实例）</strong><p>若普通成员看不到新模型，去「管理员面板 -> 工作空间 -> 模型」把模型权限设为 Public 或指定用户组。</p></div>
</div>

## 方式二：Docker 启动命令

```bash
docker run -d -p 3000:8080 \
  -e OPENAI_API_BASE_URL="{{API_BASE_URL}}" \
  -e OPENAI_API_KEY="sk-你的密钥" \
  -e ENABLE_OLLAMA_API="false" \
  -v open-webui:/app/backend/data \
  --name open-webui \
  --restart always \
  ghcr.io/open-webui/open-webui:main
```

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 打开新对话转圈十几秒 | 关掉不用的 `Ollama API` 开关，并在 OpenAI 连接里填常用模型 ID 白名单 |
| 刷新页面后配置消失 | 弹窗点保存后，漏点了外部连接页面最底部的总 **「保存」** 按钮 |

## 相关

- [创建 API Key](/guide/api-key/)
- [Dify / n8n / Coze 工作流](/guide/clients/dify-workflows/)
- [用量与日志](/guide/account/usage/)
