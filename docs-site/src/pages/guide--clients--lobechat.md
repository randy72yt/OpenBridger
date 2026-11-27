# LobeChat 保姆级配置教程（网页端 / 桌面端 / Docker 自部署）

LobeChat（LobeHub）支持丰富的助手角色市场、插件与多模态对话。无需额外安装插件，直接通过内置的 OpenAI 服务商或自定义服务商按 **1 -> 2 -> 3 -> 4 -> 5** 五步即可接入 OpenBridger。

---

## 开始之前

1. 打开控制台 [API 密钥]({{CONSOLE_URL}}/keys)，创建一把名为 `lobechat` 的密钥并复制 `sk-...`。
2. 打开 [模型广场]({{CONSOLE_URL}}/pricing)，复制你想使用的完整模型 ID（例如 `gpt-6.2-astra`、`claude-sonnet-5-5`、`gemini-4-neon`）。

---

## 详细配置步骤（1-2-3-4-5）

<div class="steps">
  <div class="step">
    <strong>第 1 步：打开「应用设置 -> AI 服务商」</strong>
    <p>打开 LobeChat，点击界面<strong>左上角的个人头像</strong>，在弹出菜单中点击 <strong>⚙️「应用设置（Settings）」</strong>，再点击左侧菜单的 <strong>「AI 服务商（Language Model / AI Provider）」</strong>。</p>
  </div>
  <div class="step">
    <strong>第 2 步：选择 OpenAI（或新建自定义服务商）</strong>
    <p>在服务商列表中点击第一项 <strong><code>OpenAI</code></strong>，确认右上角的<strong>启用开关处于打开状态</strong>。<br>
    <em>（如果你想保留原本的 OpenAI 官方配置不冲突，也可以点击列表最底部的 <code>+ 添加自定义服务商</code>，接口格式选 <code>OpenAI</code>，名称填 <code>OpenBridger</code>。）</em></p>
  </div>
  <div class="step">
    <strong>第 3 步：填写 API Key 与 API 代理地址</strong>
    <p>在右侧表单填入：<br>
    • <strong>API Key</strong>：粘贴你在控制台复制的 <code>sk-...</code> 密钥。<br>
    • <strong>API 代理地址（API Proxy URL / Endpoint）</strong>：填入 <strong><code>{{API_BASE_URL}}</code></strong>（末尾带 <code>/v1</code>，不要多加 <code>/chat/completions</code>）。<br>
    <em>（若使用本地或服务器浏览器端访问遇到跨域提示，将下方的「客户端请求模式 / Client Fetch」开关关闭，让请求经服务端转发即可。）</em></p>
  </div>
  <div class="step">
    <strong>第 4 步：添加模型并点「检查」验证连通性</strong>
    <p>向下滚动到「模型列表」区域：<br>
    1. 点击 <strong><code>+ 添加模型</code></strong>（或点「获取模型列表」自动拉取），在「模型 ID」栏粘贴从 <a href="{{CONSOLE_URL}}/pricing">模型广场</a> 复制的完整 ID（如 <code>gpt-6.2-astra</code>），按需勾选「视觉」「工具调用」开关后保存。<br>
    2. 在上方的「连通性检查（Connectivity Check）」下拉框里选中你刚添加的这个模型，点击右侧 <strong>「检查」</strong> 按钮，显示绿色勾号「检查通过」即配置成功！</p>
  </div>
  <div class="step">
    <strong>第 5 步：回到聊天页新建会话使用</strong>
    <p>返回 LobeChat 聊天主界面，点击左上角新建一个助手会话（如「随便聊聊」），点击聊天框上方的<strong>模型图标</strong>切换到刚才启用的模型，发送消息并在控制台 <a href="{{CONSOLE_URL}}/usage-logs">使用日志</a> 核对记录。</p>
  </div>
</div>

### 表单填写速查卡

<div class="config-card">
  <div><span>设置入口</span>左上角头像 -> <code>应用设置</code> -> <code>AI 服务商</code> -> <code>OpenAI</code></div>
  <div><span>API Key</span><code>sk-...</code></div>
  <div><span>API 代理地址</span><code>{{API_BASE_URL}}</code></div>
  <div><span>模型 ID</span>模型广场中的完整 ID（如 <code>gpt-6.2-astra</code>）</div>
</div>

---

## 附：Docker 自部署环境变量配置

如果你在自己的服务器或电脑上通过 Docker 部署 LobeChat，可在启动参数中直接传入：

```bash
docker run -d -p 3210:3210 \
  -e OPENAI_API_KEY="sk-你的密钥" \
  -e OPENAI_PROXY_URL="{{API_BASE_URL}}" \
  -e ACCESS_CODE="你自定义的网页访问密码" \
  lobehub/lobe-chat
```

## 常见问题与排障

| 遇到的问题 | 原因 | 解决方法 |
| --- | --- | --- |
| 点「检查」报错，提示某默认模型不存在 | 连通性检查下拉框默认选在 `gpt-4o-mini` 等你未添加或密钥分组不含的模型上 | 先把检查按钮左侧的模型下拉框改成你刚从模型广场添加的模型，再点「检查」 |
| 报错 `404` | 代理地址漏写了 `/v1` 或重复写了 `/v1/v1` | 改为 `{{API_BASE_URL}}` |
| 网页版报跨域（CORS / Network Error） | 浏览器直接发起请求被本地扩展拦截 | 在该服务商设置里关掉「使用客户端请求模式（Use Client-Side Fetching）」 |
| 请求仍然打到了官方 `api.openai.com` | 会话里选的还是其他服务商下的同名模型 | 在模型选择菜单认准你配置的 `OpenAI` / `OpenBridger` 分组标题 |

## 相关

- [Cherry Studio](/guide/clients/cherry-studio/)
- [Chatbox](/guide/clients/chatbox/)
- [故障排查](/troubleshooting/)
