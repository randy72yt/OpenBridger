# 故障排查

无法调用或消耗异常时，按照下面的顺序检查，通常比反复更换配置更快。

<div class="steps"><div class="step"><strong>检查 Base URL</strong><p>OpenAI 兼容客户端填写 <code>{{API_BASE_URL}}</code>，不要重复追加 <code>/v1</code>。</p></div><div class="step"><strong>检查 API Key</strong><p>确认 Key 未禁用，复制时没有空格、换行或重复的 Bearer 前缀。</p></div><div class="step"><strong>检查模型与权限</strong><p>从控制台复制完整模型 ID，并确认当前分组可以使用。</p></div><div class="step"><strong>检查用量</strong><p>查看余额、调用日志、状态码和请求时间。</p></div></div>

| 异常 | 常见原因 | 处理方式 |
| --- | --- | --- |
| 401 | Key 错误、禁用或格式错误 | 重新复制或创建测试 Key |
| 403 | 分组、模型或线路权限不匹配 | 检查模型广场和账户分组 |
| 404 / model not found | 模型名称错误 | 复制完整模型 ID，注意大小写和版本号 |
| 429 | 频率、并发或上游限流 | 降低并发，稍后重试 |
| 请求超时 | 网络、上游响应慢或上下文过长 | 用短问题测试，降低输出长度 |
| 流式输出中断 | 网络或客户端兼容性 | 暂时关闭 stream，对比同一模型 |
| 消耗偏高 | 上下文重复、Agent 工具调用 | 清理历史并在日志中核对实际模型 |
| 一直显示旧配置 | 客户端缓存或环境变量未刷新 | 重启客户端与终端，检查当前 Provider 和环境变量 |
| 模型列表无法刷新 | 客户端没有调用模型列表或地址格式不兼容 | 手动填写模型 ID，并核对客户端需要根地址还是 `/v1` |

## 快速自检

```bash
curl {{API_BASE_URL}}/models \
  -H "Authorization: Bearer YOUR_API_KEY"
```

能返回模型列表，说明网络和 Key 基本正常；客户端仍失败时，继续检查客户端地址格式、模型名和本地缓存。
