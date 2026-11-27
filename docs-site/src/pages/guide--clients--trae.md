# Trae 配置教程

在 Trae IDE 的 **Chat** 与 **Builder / Agent** 面板中添加 OpenAI 兼容自定义模型（无需退出你原有的 Trae 账号）。

## 开始前准备

1. 在 [API 密钥]({{CONSOLE_URL}}/keys) 创建名为 `trae` 的密钥，复制 `sk-...`。
2. 在 [模型广场]({{CONSOLE_URL}}/pricing) 复制完整模型 ID（如 `gpt-6.2-astra`、`claude-sonnet-5-5`）。

## 配置 5 步

<div class="steps">
  <div class="step"><strong>第 1 步：打开模型设置</strong><p>点击右侧 AI 面板右上角 <strong>⚙️ 齿轮图标</strong> -> <strong><code>Models</code></strong>（或点对话输入框左下角模型下拉菜单底部的 <strong><code>+ 添加模型</code></strong>）。</p></div>
  <div class="step"><strong>第 2 步：选 OpenAI Compatible</strong><p>在服务商（Provider）下拉框选 <strong><code>OpenAI Compatible</code></strong>（旧版若无此项请先升级 Trae，或在编辑器内安装 <a href="/guide/clients/cline/">Cline / Roo Code</a>）。</p></div>
  <div class="step"><strong>第 3 步：填写参数</strong>
    <div class="config-card">
      <div><span>Base URL</span><code>{{API_BASE_URL}}</code>（带 <code>/v1</code>）</div>
      <div><span>Model ID</span>粘贴模型广场完整 ID（如 <code>gpt-6.2-astra</code>）</div>
      <div><span>API Key</span>粘贴 <code>sk-...</code> 完整密钥</div>
    </div>
  </div>
  <div class="step"><strong>第 4 步：保存添加</strong><p>点击 <strong>「添加 / 保存」</strong>（如需同时用 Claude 和 GPT，重复添加不同模型 ID 即可）。</p></div>
  <div class="step"><strong>第 5 步：分别在 Chat 与 Builder 切换模型</strong><p>新建会话，在输入框下方选中你刚添加的模型（<strong>Chat</strong> 和 <strong>Builder</strong> 模式各自独立记住所选模型，切换模式后留意核对）。</p></div>
</div>

## 常见问题

| 现象 | 解决方法 |
| --- | --- |
| 行内代码补全没出现在使用日志 | 行内 Tab 补全走 Trae 自带服务，属正常现象 |
| 报 `404` / `503` | `404` 检查 Base URL 为 `{{API_BASE_URL}}`；`503` 去 [模型广场]({{CONSOLE_URL}}/pricing) 重新复制完整模型 ID |

## 相关

- [Cursor](/guide/clients/cursor/)
- [Cline / Roo Code](/guide/clients/cline/)
- [故障排查](/troubleshooting/)
