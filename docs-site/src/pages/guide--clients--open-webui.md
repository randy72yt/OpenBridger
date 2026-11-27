# Open WebUI 保姆级配置教程（课题组共享 / 自部署面板）

Open WebUI 是目前高校课题组、实验室和个人服务器最常部署的开源类 ChatGPT 网页端。管理员在后台配好一次 OpenBridger 连接，即可给课题组成员分配使用指定大模型。

<div class="callout warning"><strong>多人共用时的额度安全提醒</strong><p>管理员在「管理员面板」添加的连接会被该 Open WebUI 实例上有权限的成员共同调用，消耗统一记在这把 API Key 名下。建议在控制台 <a href="{{CONSOLE_URL}}/keys">API 密钥</a> 为该实例<strong>单独创建一把密钥并设置额度上限</strong>。</p></div>

---

## 方式一：在网页管理后台可视化配置（推荐，无需重启容器）

需要使用拥有 **管理员（Admin）** 权限的账号登录 Open WebUI，按下面 **1 -> 2 -> 3 -> 4 -> 5** 五步操作：

<div class="steps">
  <div class="step">
    <strong>第 1 步：进入「管理员面板 -> 外部连接」</strong>
    <p>登录 Open WebUI 网页端，点击<strong>左下角（或右上角）的个人头像</strong> -> 选择 <strong>「管理员面板（Admin Panel）」</strong> -> 点击顶部导航栏的 <strong>「设置（Settings）」</strong> -> 切换到左侧的 <strong>「外部连接（Connections）」</strong>。</p>
  </div>
  <div class="step">
    <strong>第 2 步：新增或编辑一条 OpenAI API 连接</strong>
    <p>在最上方的 <strong><code>OpenAI API</code></strong> 区域（确保右侧总开关为开启状态），点击右侧的 <strong><code>+</code>（添加连接 / Add Connection）</strong> 按钮打开配置弹窗。</p>
  </div>
  <div class="step">
    <strong>第 3 步：填写 URL、密钥与模型白名单（关键！）</strong>
    <p>在弹窗里依次填写：<br>
    • <strong>URL / 接口地址</strong>：填 <strong><code>{{API_BASE_URL}}</code></strong>（末尾带 <code>/v1</code>，不要带斜杠）<br>
    • <strong>密匙 / API Key</strong>：粘贴你在控制台创建的 <code>sk-...</code> 密钥<br>
    • <strong>模型 ID（Model IDs，强烈建议填）</strong>：在下方模型输入框粘贴你想在界面上展示的几个常用模型完整 ID（例如 <code>gpt-6.2-astra</code>、<code>claude-sonnet-5-5</code>、<code>deepseek-v4.1</code>），每输完一个点右侧 <code>+</code> 号加入。这样首页模型下拉框干净清爽，不会一次拉出上百个不常用的模型，网页加载也更快！</p>
  </div>
  <div class="step">
    <strong>第 4 步：点旋转箭头验证并点击「保存」</strong>
    <p>点击 URL 输入框右侧的 <strong>🔄（验证连接 / Verify Connection）</strong> 图标，界面右上角弹出绿色「已验证服务器连接」提示后，点击弹窗右下角 <strong>「保存」</strong>，<strong>再拉到外部连接页面最底部点击一次总「保存」按钮</strong>（很多新手漏点最底下这个保存！）。</p>
  </div>
  <div class="step">
    <strong>第 5 步：新建对话选择模型测试</strong>
    <p>点击左上角 <strong>「新对话（New Chat）」</strong> 返回聊天页，在顶部模型选择框选中刚才添加的模型发送一句消息，随后在控制台 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 核对记录即可。</p>
  </div>
</div>

### 表单填写速查卡

<div class="config-card">
  <div><span>后台路径</span>头像 -> <code>管理员面板</code> -> <code>设置</code> -> <code>外部连接</code> -> <code>OpenAI API</code></div>
  <div><span>URL</span><code>{{API_BASE_URL}}</code></div>
  <div><span>API Key</span>单独创建的 <code>sk-...</code> 密钥（建议设额度上限）</div>
  <div><span>模型 ID 白名单</span>按需填入 <code>gpt-6.2-astra</code>、<code>claude-sonnet-5-5</code> 等并点 <code>+</code></div>
</div>

---

## 方式二：Docker 启动时直接通过环境变量注入

如果你刚准备用 Docker 部署 Open WebUI（若不使用本地 Ollama，建议顺手关闭 Ollama 探测避免后台卡顿）：

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

> 修改 `docker-compose.yml` 环境变量后，必须运行 `docker compose up -d` 重新创建容器才会生效，仅刷新浏览器网页不会重读环境变量。

---

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 打开新对话页面转圈加载很慢（卡十几秒） | ① 默认开启了本地不存在的 `Ollama API` 超时等待；② 未填模型白名单导致每次拉全量列表 | 在「外部连接」里把不用的 **Ollama API 开关关掉**，并在 OpenAI 连接里填入常用模型 ID 白名单 |
| 普通成员账号登录后提示「没有可用模型」 | Open WebUI 新版默认将新接入模型的权限设为仅管理员可见 | 进入「管理员面板 -> 工作空间 -> 模型」，把对应模型的可见权限设为 Public 或指定用户组 |
| 弹窗点了保存但刷新又没了 | 漏点了「外部连接」页面最右下角的全局 **「保存」** 按钮 | 弹窗保存后，务必再点页面底部的总保存按钮 |

## 相关

- [创建 API Key](/guide/api-key/)
- [Dify / n8n / Coze 工作流](/guide/clients/dify-workflows/)
- [用量与日志](/guide/account/usage/)
