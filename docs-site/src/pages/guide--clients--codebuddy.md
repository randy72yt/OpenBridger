# CodeBuddy 配置教程（独立 IDE / VS Code & JetBrains 插件）

在腾讯云代码助手 **CodeBuddy** 中添加 OpenAI 兼容自定义模型用于 **Chat** 与 **Craft / Agent** 编码（保留原账号登录即可）。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `codebuddy` 的密钥，复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID（如 `gpt-6.2-astra`、`claude-sonnet-5-5`）。

## 配置 5 步

<div class="steps">
  <div class="step"><strong>第 1 步：打开自定义模型管理</strong><p>在 CodeBuddy 面板点右上角 <strong>⚙️ 设置 -> 模型管理</strong>（或点对话输入框下方的模型下拉菜单 -> <strong><code>+ 添加模型</code></strong>）。</p></div>
  <div class="step"><strong>第 2 步：选 OpenAI Compatible</strong><p>点击添加自定义模型，协议类型选 <strong><code>OpenAI Compatible</code></strong>。</p></div>
  <div class="step"><strong>第 3 步：填写参数</strong>
    <div class="config-card">
      <div><span>显示名称（若有）</span>如 <code>OB - GPT-6.2</code></div>
      <div><span>Model ID（模型标识）</span>粘贴模型广场完整 ID（如 <code>gpt-6.2-astra</code>）</div>
      <div><span>Base URL</span><code>{{API_BASE_URL}}</code>（带 <code>/v1</code>）</div>
      <div><span>API Key</span>粘贴 <code>sk-...</code> 完整密钥</div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：测试并保存</strong><p>点击测试连接（如有）后点 <strong>「保存」</strong>。</p></div>
  <div class="step"><strong>第 5 步：在 Chat / Craft 切换模型</strong><p>新建对话，在输入框底部下拉框选中刚添加的模型提问（<strong>Chat</strong> 与 <strong>Craft</strong> 模式需分别选一次）。</p></div>
</div>

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 找不到自定义模型菜单 | 升级 CodeBuddy 到最新版，或在 VS Code 搭配 [Cline / Roo Code](/guide/clients/cline/) |
| 报 `404` / `503` | `404` 确认 Base URL 为 `{{API_BASE_URL}}`；`503` 确认 `Model ID` 没和显示名称填反 |

## 相关

- [Cursor](/guide/clients/cursor/)
- [Cline / Roo Code](/guide/clients/cline/)
- [故障排查](/troubleshooting/)
