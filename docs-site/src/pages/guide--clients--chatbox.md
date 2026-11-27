# Chatbox 配置教程（电脑端 / 手机 App）

在 Chatbox 中添加「OpenAI API 兼容」自定义提供方即可跨端使用（不要点登录 Chatbox 官方订阅）。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `chatbox` 的密钥，复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID（如 `gpt-6.2-astra`、`claude-sonnet-5-5`）。

## 配置 5 步

<div class="steps">
  <div class="step"><strong>第 1 步：打开设置</strong><p>点击左下角 <strong>⚙️ 设置（Settings）</strong> -> 切到 <strong>「模型提供方（Model Provider）」</strong>。</p></div>
  <div class="step"><strong>第 2 步：添加自定义提供方</strong><p>拉到左侧列表最底部点 <strong><code>+ 添加自定义提供方</code></strong>，名称填 <code>OpenBridger</code>，API 模式选 <strong><code>OpenAI API 兼容</code></strong>（旧版无此按钮则直接选 <code>OpenAI API</code>）。</p></div>
  <div class="step"><strong>第 3 步：填写域名与密钥</strong>
    <div class="config-card">
      <div><span>API 域名（API Host）</span><code>{{API_BASE_URL}}</code>（带 <code>/v1</code>）</div>
      <div><span>API 路径（若有）</span>保持默认 <code>/chat/completions</code></div>
      <div><span>API 密钥</span>粘贴 <code>sk-...</code> 完整密钥</div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：新建模型并保存</strong><p>在下方模型区点 <strong><code>+ 新建</code></strong>，粘贴完整模型 ID（支持看图则勾选「视觉」），点右下角 <strong>「保存」</strong>。</p></div>
  <div class="step"><strong>第 5 步：新建对话切换模型</strong><p>回主界面点 <strong><code>+ 新对话</code></strong>（旧对话仍绑着旧模型），在输入框下方选中 <code>OpenBridger</code> 的模型发消息测试。</p></div>
</div>

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 报 `404`（路径含 `/v1/v1`） | 确认 API 域名为 `{{API_BASE_URL}}` 且 API 路径为 `/chat/completions` |
| 报 `401` | 重新复制粘贴 `sk-...`，去掉首尾空格 |
| 使用日志无记录 | 点击 **`+ 新对话`**，在输入框下方重新点选自定义模型 |

## 相关

- [Cherry Studio](/guide/clients/cherry-studio/)
- [LobeChat](/guide/clients/lobechat/)
- [Base URL 填写速查](/guide/base-urls/)
