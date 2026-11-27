# 故障排查（30 秒定位问题）

遇到报错先看弹窗里的 **三位数字状态码**，95% 的问题按下面第一张表 30 秒就能解决。

## 一、按报错数字对号入座

| 看到的报错 | 真正原因 | 一步解决 |
| --- | --- | --- |
| **`401` 无效的令牌** | ① `sk-...` 复制不全或前后多了空格；② 在密码框里多写了 `Bearer`；③ 这把密钥设了额度上限且已用完 | 回 [API 密钥]({{CONSOLE_URL}}/keys) 点复制重贴，确认密钥未设限或未禁用 |
| **`403` 用户额度不足** | 账户余额为 ¥0 或不够本次预扣 | 去 [钱包]({{CONSOLE_URL}}/wallet) 充值或兑换后再试 |
| **`404` Not Found** | Base URL 多写了 `/v1`（变成 `/v1/v1/...`）或漏写了 `/v1` | 对照 [Base URL 填写速查](/guide/base-urls/) 修正地址 |
| **`503` 无可用渠道** | **地址和密钥都已经通了！** 是模型名字手打错了，或密钥所选分组没有这个模型 | 去 [模型广场]({{CONSOLE_URL}}/pricing) 点击复制完整模型 ID；检查密钥分组（可改为 `auto`） |
| **`429` 请求过多 / 限流** | 并发太猛（如整本 PDF 翻译瞬间发几十条）或上游繁忙 | 调低翻译插件每秒并发数，等几秒重试或换同级别其他模型 |
| **改完配置还是提示官方登录 / 没记录** | 客户端没重启，或者还在旧的聊天窗口里发消息 | **彻底退出重开软件**，并点击 **「+ 新建对话」** 再选模型提问 |

---

## 二、核对你的软件地址有没有填反

<div class="config-card">
  <div><span>Codex / Cursor / Cherry Studio / Chatbox / LobeChat / Open WebUI</span><code>{{API_BASE_URL}}</code>（带 <code>/v1</code>）</div>
  <div><span>Claude Code / Gemini CLI / Anthropic SDK</span><code>{{CONSOLE_URL}}</code>（<strong>不带 /v1</strong>）</div>
  <div><span>沉浸式翻译 / Coze HTTP 节点</span><code>{{API_BASE_URL}}/chat/completions</code>（完整路径）</div>
</div>

---

## 三、开发者终端自检（可选）

如果不确定是客户端软件 Bug 还是密钥本身问题，在终端执行下面命令：

```bash
curl -i "{{API_BASE_URL}}/models" \
  -H "Authorization: Bearer sk-你的密钥"
```

- 返回 `HTTP/2 200` 及模型 JSON：说明网络与密钥完全正常，只需回头检查客户端软件设置。
- 仍无法解决时，复制报错里的 `request id`（或日志截图），参考 [联系与故障申报](/guide/support/) 发邮件处理。
