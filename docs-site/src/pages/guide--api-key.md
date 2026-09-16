# 创建和管理 API Key

API Key 是工具或程序调用 OpenBridger 的凭证。建议为不同工具、项目和环境分别创建密钥。

<div class="steps"><div class="step"><strong>进入 API 密钥</strong><p>登录控制台并打开 API 密钥页面。</p></div><div class="step"><strong>新建密钥</strong><p>使用便于识别的名称，例如 Codex、Claude Code 或项目名称。</p></div><div class="step"><strong>限制范围</strong><p>根据页面提供的能力设置分组、模型范围、额度或有效期。</p></div><div class="step"><strong>复制并验证</strong><p>将密钥保存到密码管理器或环境变量中，再发起短请求验证。</p></div></div>

<div class="grid"><div class="card"><h3>按用途隔离</h3><p>不同工具使用不同 Key，便于查看用量和单独停用。</p></div><div class="card"><h3>及时轮换</h3><p>发现泄露、人员变动或项目结束时，立即删除或禁用旧 Key。</p></div><div class="card"><h3>避免前端暴露</h3><p>不要把真实 Key 写入网页源码、公开仓库、截图或客户端日志。</p></div></div>

<div class="callout warning"><strong>密钥安全</strong><p>OpenBridger 的支持人员不会要求你提供密码或完整可用的 API Key。</p></div>
