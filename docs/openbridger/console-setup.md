# 控制台操作指引（S15–S19）

最后更新：2026-09-21

本文件只覆盖**必须由人在浏览器/控制台里完成**的步骤——SSH 侧能做的都已在服务器上执行完毕。按编号顺序做，每节给出入口 URL、填写值和验证方法。

已完成的前置：服务已在 `https://openbridger.com` 上线，HTTPS 正常，管理员已创建，公开注册与在线支付已关闭。

---

## S15. 管理员开启 MFA / Passkey

**入口**：登录 `https://openbridger.com` → 右上角头像 → 菜单里的 **Security**（或直接访问 `https://openbridger.com/security`）

**做什么**：

1. 页面里找到 **Passkey / 两步验证** 卡片，点击添加。
2. 优先用 **Passkey**（Face ID / Touch ID / 密码管理器），比 TOTP 抗钓鱼。
3. 若只有 TOTP 选项：用 1Password / Authy / Google Authenticator 扫码，**把恢复码抄下来存到密码管理器**，不要只留在浏览器里。

**验证**：登出后用无痕窗口重新登录，确认被要求第二步验证。

**当前状态（2026-09-21）**：✅ 已完成。数据库确认 `two_fas.is_enabled=1`（管理员 `xinlingwong`，role=100），已生成 4 条 TOTP 备份码。请确认备份码已离线保存。

### 15.1 Passkey 注册失败与修复（2026-09-21）

**现象**：管理员在 `/security` 页面点击注册 Passkey，前端报错，服务端返回 500。

**日志证据**：

```
[ERR] auth session internal error (POST /api/user/passkey/register/begin):
      Passkey 不允许使用不安全的 Origin: http://localhost:3000
[GIN] 500 | POST /api/user/passkey/register/begin
```

**根因（代码缺陷）**：

1. `setting/system_setting/system_setting_old.go:3` 把站点地址硬编码为 `var ServerAddress = "http://localhost:3000"`
2. `setting/system_setting/passkey.go:46-48`：Passkey 的 `Origins` 为空时会拿 `ServerAddress` 兜底
3. `service/passkey/service.go:83-88`：Origin 以 `http://` 开头且 `AllowInsecureOrigin=false` 时直接拒绝

结果：**任何生产部署，只要管理员没手动配置过站点地址，Passkey 就必然注册失败**，且错误信息完全不提示真正的修复方向。代码里其实已有更好的 `autoDetect` 分支（从请求 Host + scheme 推导，见 `service/passkey/service.go:94` 起），但因为 `ServerAddress` 非空而永远走不到。

**修复（已完成）**：写入 options 表，`ServerAddress = https://openbridger.com`。`/api/status` 的 `server_address` 已验证同步为该值。

```sql
INSERT INTO options (`key`, value) VALUES ("ServerAddress","https://openbridger.com")
  ON DUPLICATE KEY UPDATE value="https://openbridger.com";
```

**这个配置不止影响 Passkey**：`ServerAddress` 还用于 OAuth 回调 URI（`oauth/oidc.go:57` 等）、密码重置邮件链接（`controller/misc.go:256`）、支付回调地址。生产环境必须设置为真实域名。

**待办**：上报上游修正 `GetPasskeySettings()` 的兜底策略——`Origins` 为空时应优先从请求推导，而不是用硬编码的 localhost。

**为什么必须做**：这是唯一一个管理员账号，泄露即全站失守（可改价格、删渠道、看他人 Key）。

---

## S16. Cloudflare：SSL 模式与缓存规则

**入口**：Cloudflare Dashboard → 选中 `openbridger.com`

### 16.1 SSL/TLS 加密模式

左侧 **SSL/TLS** → **概述（Overview）** → 把模式改为 **Full (strict)**。

- 现在是源站已有 Let's Encrypt 真证书，但 CF 默认可能是「灵活（Flexible）」——那样 CF 到源站这一段是明文。改成 Full (strict) 后整条链路加密。
- 改完立刻验证：`https://openbridger.com/api/status` 应正常返回 200。若变 525/526，说明源站证书链有问题，回退到 Flexible 再排查。

### 16.2 边缘证书（可选但建议）

左侧 **SSL/TLS** → **Edge Certificates**：

- **Always Use HTTPS**：开
- **HTTP Strict Transport Security (HSTS)**：先只在 `openbridger.com` 开，勾 `includeSubdomains` 前确认 `docs.` 已 HTTPS（已签证书，可以勾）

### 16.3 API 与流式路径绕过缓存（必须做）

左侧 **Caching** → **Cache Rules** → **Create rule**：

- 规则名：`Bypass API and relay`
- 表达式（Custom filter expression），选 **URI Path**：

```
(http.request.uri.path wildcard "/v1/*") or (http.request.uri.path wildcard "/api/*")
```

- 设置：**Cache eligibility → Bypass cache**（或 Eligible for cache 关闭）
- 部署

**为什么**：`/v1/chat/completions` 是 SSE 流式响应，`/api/pricing` 是实时报价。被边缘缓存会让用户拿到上一个用户的响应或过期价格。当前实测 `/api/status` 返回 `cf-cache-status: DYNAMIC`（默认没缓存），但这是 CF 的推断行为，不可依赖。

**验证结果（2026-09-21）**：`https://openbridger.com/api/status` 与 `/v1/models` 响应头均为 `cf-cache-status: DYNAMIC`。**DYNAMIC 与 BYPASS 等价——都不缓存**，安全目标已达成。CF 的动态内容判定优先于 Cache Rule，所以头里不会显示 BYPASS，这是正常现象，不必纠结；规则保留着即可兜住未来新增的静态化路径。

---

## S17. 邮件发送（SMTP）—— 方案：Resend

**选型结论：用你已经有的 Resend。** 理由：

- 免费额度 3,000 封/月、100 封/天，远超当前邀请制试运行的量，零成本
- 自带 DKIM/SPF 配置界面，不用自己拼 TXT 记录
- 提供标准 SMTP 接口，本项目的邮件模块（`common/email.go`）直接兼容，无需改代码
- 比自建 Postfix / 阿里云邮件推送省事，且不在国内备案约束内

（备选 Zoho / Cloudflare Email Routing + Gmail 不作为首选：Zoho 免费版已取消 SMTP，CF Email Routing 只收不发。）

### 17.1 Resend 侧（你操作）

1. 登录 Resend → 左侧 **Domains** → **Add Domain** → 输入 `openbridger.com` → 选区域（默认即可）
2. Resend 会给出一组 DNS 记录：**MX**（接收退信）、**TXT（SPF）**、**2~3 条 CNAME（DKIM）**
3. 逐条加到 Cloudflare DNS。注意：
   - Resend 给的主机名是 `resend._domainkey` 这种形式，CF 里填 Name 时**不要**再补 `.openbridger.com`
   - DKIM/SPF 记录设为 **DNS only（灰云）**，不要走 CF 代理
   - 已有的 `openbridger.com` MX 记录不冲突就保留
4. 回 Resend 点 **Verify DNS Records**，状态变 **Verified** 才算完成（通常几分钟，偶尔要等 30 分钟）
5. 左侧 **API Keys** → **Create API Key**：
   - 权限选 **Sending access**（不要 Full access）
   - 域名范围限定 `openbridger.com`
   - 复制生成的 `re_xxxxxxxx` 密钥（只显示一次）

### 17.2 OpenBridger 侧（你填写，我可代填）

**入口**：`https://openbridger.com/system-settings/operations/email`

| 字段 | 填什么 |
| --- | --- |
| SMTP Server | `smtp.resend.com` |
| SMTP Port | `465` |
| 安全模式 | **SSL/TLS**（隐式 TLS；不要选 STARTTLS，465 端口下会失败） |
| SMTP Account（用户名） | `resend`（固定字符串，不是你的邮箱） |
| SMTP Token（密码） | 上一步复制的 `re_xxxxxxxx` API Key |
| SMTP From（发件人） | `support@openbridger.com` |
| Insecure Skip Verify | 关闭 |
| Force Auth Login | 关闭 |

**填写后找我验证**：我会触发一封测试邮件，并在服务器侧确认连接与认证结果。

**兼容性已确认**：`common/email.go:45` 在 `SMTPSSLEnabled=true` 时走 `tls.Dial`，`AutoSMTPAuth` 默认选 PLAIN——Resend 两者都支持。

**注意**：发件地址必须落在**已验证域名**下（`support@openbridger.com` 可以）。用未验证域名会被 Resend 拒绝。

---

## S18. Turnstile 人机校验（反自动化）

### 18.1 Cloudflare 侧（你操作）

1. CF Dashboard → 左侧 **Turnstile** → **Add site**
2. 表单填写：
   - **Site name**：`OpenBridger`
   - **Domain**：`openbridger.com`（如需本地联调再加 `localhost`）
   - **Widget Mode**：选 **Managed**（推荐；CF 自动决定是否弹验证，无感率最高）
   - **Pre-clearance**：可选，暂时不开
3. 创建后会显示两个值：
   - **Site Key**（公开，前端用，形如 `0x4AAAAAAA...`）
   - **Secret Key**（保密，服务端用）
   - 两个都复制下来

### 18.2 OpenBridger 侧（你填写，我可代填）

**入口**：`https://openbridger.com/system-settings/auth/bot-protection`

| 字段 | 填什么 |
| --- | --- |
| TurnstileCheckEnabled | 打开 |
| TurnstileSiteKey | 上一步的 Site Key |
| TurnstileSecretKey | 上一步的 Secret Key |

**代码约束**：`controller/option.go:226` 会校验——先填 Key 再开开关，否则保存失败并提示「请先填入 Turnstile 校验相关配置信息」。**顺序：先填两个 Key，保存，再打开开关。**

**验证**：开启后登录/注册页会出现 Turnstile 挂件；我这边会跑一次匿名请求确认校验链路通。

---

## S19. 待办收尾

完成上面四项后，剩下的收尾项：

| 项 | 状态 |
| --- | --- |
| 协议正文写入 legal 配置段 | 待协议 11 项待确认定稿（见 `legal-drafts/`） |
| 告警（监控通知） | 依赖 S17 SMTP 配通后才能发邮件告警 |
| 备份定时任务 | 待配（`/srv/openbridger/backup`） |

---

## 附：需要你返回给我的东西

做完 S17.1 和 S18.1 后，把这两样给我（走受控渠道，**不要贴在仓库文档里**）：

1. Resend API Key（`re_...`）
2. Turnstile Site Key + Secret Key

我来完成服务端填写、发信验证和 Turnstile 链路验证。如果你更愿意自己填后台，那就照上面的表填，填完告诉我，我只做验证。
