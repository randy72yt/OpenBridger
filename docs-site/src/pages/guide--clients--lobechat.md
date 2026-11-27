# LobeChat 配置教程（网页端 / 桌面端 / Docker）

通过 LobeChat 内置的 `OpenAI`（或自定义服务商）接入 OpenBridger。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `lobechat` 的密钥，复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID（如 `gpt-6.2-astra`、`claude-sonnet-5-5`）。

## 配置 5 步

<div class="steps">
  <div class="step"><strong>第 1 步：进入 AI 服务商设置</strong><p>点左上角<strong>个人头像</strong> -> <strong>⚙️ 应用设置</strong> -> 左侧点 <strong>「AI 服务商（Language Model）」</strong>。</p></div>
  <div class="step"><strong>第 2 步：开启 OpenAI（或新建自定义服务商）</strong><p>点击列表中的 <strong><code>OpenAI</code></strong> 并打开右上角启用开关（也可拉到底部点 <code>+ 添加自定义服务商</code>，接口格式选 <code>OpenAI</code>）。</p></div>
  <div class="step"><strong>第 3 步：填密钥与代理地址</strong>
    <div class="config-card">
      <div><span>API Key</span>粘贴 <code>sk-...</code> 完整密钥</div>
      <div><span>API 代理地址（Proxy URL）</span><code>{{API_BASE_URL}}</code>（带 <code>/v1</code>）</div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：添加模型并点「检查」</strong><p>在下方模型列表点 <strong><code>+ 添加模型</code></strong> 填入完整模型 ID；把上方「连通性检查」下拉框选为刚添加的模型，点击 <strong>「检查」</strong> 显示通过。</p></div>
  <div class="step"><strong>第 5 步：新建会话选模型对话</strong><p>回聊天页新建会话，点输入框上方的模型图标切换到刚添加的模型即可。</p></div>
</div>

### 附：Docker 自部署启动命令

```bash
docker run -d -p 3210:3210 \
  -e OPENAI_API_KEY="sk-你的密钥" \
  -e OPENAI_PROXY_URL="{{API_BASE_URL}}" \
  -e ACCESS_CODE="自定义访问密码" \
  lobehub/lobe-chat
```

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 点「检查」报模型不存在 | 先把「检查」按钮左侧的测试模型下拉框改成你刚添加的模型 ID |
| 报 `404` | 代理地址改为 `{{API_BASE_URL}}` |
| 网页端报跨域 / Network Error | 在该服务商设置里关掉「使用客户端请求模式（Client-Side Fetching）」 |

## 相关

- [Cherry Studio](/guide/clients/cherry-studio/)
- [Chatbox](/guide/clients/chatbox/)
- [故障排查](/troubleshooting/)
