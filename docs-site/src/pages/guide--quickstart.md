# 完成第一次调用

这条路径只做四件事：注册并登录、确认钱包里有额度、创建一把 API Key、用 curl 打通 `POST /v1/chat/completions`。

## 开始之前

- 一个可以收信的邮箱，用于注册和以后找回账号。
- 能打开控制台 [{{CONSOLE_URL}}]({{CONSOLE_URL}})。
- 本机有 `curl`。Python 或 Node 示例在 [OpenAI 兼容调用](/guide/clients/openai-compatible/)。

注册页提供的方式以页面为准。程序默认打开的是用户名或邮箱加密码的注册；如果页面要求邮箱验证码或人机验证，按页面做完再继续。GitHub 等第三方登录只有在登录页实际出现时才可用。

<div class="callout"><strong>截图待补</strong><p>注册成功后的控制台首页，标出「钱包」「API 密钥」「模型广场」「使用日志」四个入口。由运营用当前界面替换本段。</p></div>

## 步骤

<div class="steps"><div class="step"><strong>注册并登录</strong><p>打开 <a href="{{CONSOLE_URL}}">{{CONSOLE_URL}}</a>，完成注册后进入控制台。登录失败次数过多会被临时限制，等一会儿再试，不要连续撞密码。</p></div><div class="step"><strong>看钱包余额</strong><p>打开 <a href="{{CONSOLE_URL}}/wallet">钱包</a>。余额以人民币显示。注册不赠送额度，新用户余额是 0。要先用页面上的在线充值或兑换码，见 <a href="/guide/account/redeem/">充值与兑换</a>，余额够一次短对话再继续。</p></div><div class="step"><strong>复制模型 ID</strong><p>打开 <a href="{{CONSOLE_URL}}/pricing">模型广场</a>，选一个你的分组能用的对话模型，复制完整模型 ID。不要自己缩短版本号。</p></div><div class="step"><strong>创建 API Key</strong><p>打开 <a href="{{CONSOLE_URL}}/keys">API 密钥</a>，新建一把名为 <code>curl-test</code> 的密钥。复制出来的值以 <code>sk-</code> 开头。细节见 <a href="/guide/api-key/">创建 API Key</a>。</p></div><div class="step"><strong>发送一条短请求</strong><p>把下面命令里的密钥和模型 ID 换成你自己的，然后执行。</p></div></div>

```bash
export OPENBRIDGER_API_KEY="sk-把控制台复制的密钥贴在这里"
export OPENBRIDGER_MODEL="模型广场里的完整模型 ID"

curl "{{API_BASE_URL}}/chat/completions" \
  -H "Authorization: Bearer ${OPENBRIDGER_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"${OPENBRIDGER_MODEL}\",
    \"messages\": [{\"role\": \"user\", \"content\": \"用一句话回复：pong\"}]
  }"
```

成功时响应是一段 JSON，里面有 `choices[0].message.content`。同时响应头里有 `X-Oneapi-Request-Id`，报错时把它和正文一起留下来。

## 确认请求真的经过网关

1. 打开 [使用日志]({{CONSOLE_URL}}/usage-logs)。
2. 按刚才的时间和密钥名 `curl-test` 筛选。
3. 核对模型 ID、状态和消耗。日志里有这一条，才说明客户端没有把请求发到别的地址。

也可以先在 [AI 体验台]({{CONSOLE_URL}}/playground) 用同一模型发一句短消息。体验台走的是登录会话，不是 API Key；它能验证模型和额度，不能代替上面的 curl。

## 这一步常见错误

| 现象 | 处理 |
| --- | --- |
| `401`，文案「无效的令牌」 | 密钥复制错了、已禁用、已过期，或这把密钥自己的额度用尽。重新复制，不要手写第二遍 `Bearer` 或 `sk-`。 |
| `403`，文案含「用户额度不足」 | 钱包或当前扣费来源不够这次预扣。先充值或换一个更小的请求。见 [额度与计费](/guide/billing/)。 |
| `503`，文案含「无可用渠道」 | 模型 ID 写错，或当前分组没有能服务它的线路。从模型广场重新复制。 |
| 连接失败、超时 | 先确认地址是 `{{API_BASE_URL}}/chat/completions`，没有写成 `/v1/v1/`。再看 [故障排查](/troubleshooting/)。 |

完整状态码表在 [错误码](/guide/errors/)。

## 下一步

- 接到编辑器或聊天客户端：[Cursor](/guide/clients/cursor/)、[Claude Code](/guide/clients/claude-code/)、[Cherry Studio](/guide/clients/cherry-studio/)、[Chatbox](/guide/clients/chatbox/)、[LobeChat](/guide/clients/lobechat/)。
- 写进自己的程序：[OpenAI 兼容调用](/guide/clients/openai-compatible/) 和 [SDK 与 LangChain](/guide/clients/sdk/)。
