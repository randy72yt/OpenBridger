# 在 Cherry Studio 中使用 OpenBridger

Cherry Studio 支持自定义 OpenAI 兼容服务商。本文只说明配置方法，不要求通过第三方推广链接访问。

<div class="steps"><div class="step"><strong>新增自定义服务商</strong><p>进入模型服务设置，添加 OpenAI 类型的自定义服务商，名称填写 OpenBridger。</p></div><div class="step"><strong>填写密钥和地址</strong><p>API 地址填写 <code>{{API_BASE_URL}}</code>，使用为该客户端单独创建的 Key。</p></div><div class="step"><strong>添加模型</strong><p>从控制台复制完整模型 ID，不要自行缩写版本号。</p></div><div class="step"><strong>启用并验证</strong><p>选择新服务商和模型发送短消息，同时检查使用日志。</p></div></div>

Base URL 不要重复追加 `/v1` 或 `/chat/completions`。

## 推荐配置

<div class="config-card"><div><span>服务商名称</span><code>OpenBridger</code></div><div><span>服务商类型</span><code>OpenAI</code></div><div><span>API 地址</span><code>{{API_BASE_URL}}</code></div><div><span>API Key</span><code>单独创建的密钥</code></div><div><span>模型 ID</span><code>控制台中的完整模型名</code></div></div>

## 常见问题

| 现象 | 处理方式 |
| --- | --- |
| 检查密钥失败 | 核对 API 地址、Key 状态、账户额度和网络 |
| 模型列表为空 | 从控制台手动添加完整模型 ID |
| 404 / model not found | 检查大小写、短横线、版本后缀和分组权限 |
| 没有流式输出 | 检查 Stream 选项，并用非流式请求对比 |

无需退出 Cherry Studio 或其他服务商账号。每次对话按当前选择的服务商和模型发送请求。
