# 错误码速查

网关拒绝请求时会返回如下 JSON，括号内的 `request id`（同响应头 `X-Oneapi-Request-Id`）可用于向客服定位问题：

```json
{
  "error": {
    "message": "错误说明 (request id: 2026...)",
    "type": "new_api_error",
    "code": "错误代码"
  }
}
```

## 1. 密钥与权限错误（401 / 403）

| 状态码 | `code` | 常见提示文案 | 解决方法 |
| --- | --- | --- | --- |
| `401` | 空 | `无效的令牌` | 密钥填错、已禁用、已过期，或这把密钥设定的独立额度已用尽（即使钱包有钱也会报此错）。回 [API 密钥]({{CONSOLE_URL}}/keys) 检查或重复制。 |
| `401` | 空 | `未提供令牌` | 请求漏传了 API Key，或在代码里漏了 `Authorization: Bearer sk-...` 头。 |
| `403` | 空 | `无权访问 ... 分组` / `分组已被弃用` | 把密钥分组改回默认分组或 `auto`。 |
| `403` | `access_denied` | IP 不在允许列表 | 清空或更新该密钥的 IP 白名单。 |
| `403` | 空 | 账号被封禁 | 用注册邮箱发信至 [support@openbridger.com](mailto:support@openbridger.com) 核实。 |

## 2. 余额与套餐错误（403）

| 状态码 | `code` | 常见提示文案 | 解决方法 |
| --- | --- | --- | --- |
| `403` | `insufficient_user_quota` | `用户额度不足` / `预扣费额度失败` | 钱包余额不足以支付本次预扣。前往 [钱包]({{CONSOLE_URL}}/wallet) 充值或兑换。 |
| `403` | `insufficient_user_quota` | `订阅额度不足或未配置订阅` | 扣费偏好设成了「仅用订阅」或套餐当期额度已用完。可在钱包切换为「优先订阅 / 优先钱包」或充值钱包。 |

## 3. 模型、地址与限流错误（400 / 404 / 429 / 503）

| 状态码 | `code` | 常见提示文案 | 解决方法 |
| --- | --- | --- | --- |
| `400` | 请求参数错误 | JSON 解析失败、`max_tokens` 超限、纯文本模型传了图片 | 检查请求参数或换支持多模态的模型。 |
| `404` | 路径不存在 | `404 page not found` | 客户端地址多拼或少拼了 `/v1`，查 [Base URL 填写速查](/guide/base-urls/)。 |
| `429` | 限流 | `您已达到请求数限制` 或上游 `rate_limit_exceeded` | 降低并发请求速率，做退避等待（如间隔 2～5 秒）后重试。 |
| `503` | `model_not_found` | `分组 ... 下模型 ... 无可用渠道` | 从 [模型广场]({{CONSOLE_URL}}/pricing) 重新复制完整模型 ID，并检查密钥分组。 |
| `501` | `api_not_implemented` | `API not implemented` | 调用了暂未开放的冷门端点（如 `/v1/fine-tunes`、`/v1/files`）。 |
| `500` | 内部或上游瞬时错误 | 服务端异常 | 记录 `request id` 稍后重试或换备用模型；持续出现请联系支持。 |

## 相关

- [故障排查](/troubleshooting/)
- [Base URL 填写速查](/guide/base-urls/)
- [联系与故障申报](/guide/support/)
