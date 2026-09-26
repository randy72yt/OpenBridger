# OpenBridger 上线执行手册（step-by-step）

最后更新：2026-09-21

本手册是[上线计划](./launch-plan.md)的可执行版本：把「上线之前」拆成 34 个按依赖排序的步骤，每步给出操作、通过标准和证据位置。验收口径以[生产验收](./production-acceptance.md)中的用例为准（OPS/AUTH/API/RELAY/BILL/PAY/LEGAL/UI/DOC/OBS/E2E）。

## 0. 使用方式

1. **严格按顺序执行**，前一步未通过不进入下一步；被阻断时记进「阻断」列，不跳步。
2. 每步完成后，在 `production-acceptance.md` 第 12 节执行总表填写：执行日期、环境、结果、证据位置（截图或命令输出存放处）。
3. 所有秘密只出现在服务器上的 `/srv/openbridger/.env`（权限 600），不进版本库、不进文档、不进聊天记录。
4. 每步都写明回退方式；回滚靠「上一个镜像 tag + 数据卷快照」，不靠现场改配置。

## 1. 环境基线（已确认）

| 项 | 值 |
| --- | --- |
| 域名 | `openbridger.com`（主站与控制台）、`docs.openbridger.com`（文档站）；注册与 DNS 均在 Cloudflare，尚未解析 |
| 主机 | 阿里云轻量应用服务器 |
| 监听 | 容器内 3000，映射到宿主 `127.0.0.1:3000`，由宿主 Nginx 终止 TLS |
| 部署物 | `compose.release.yml` + `Dockerfile.dev`（源码构建，需先生成 `web/dist`） |
| 镜像 tag | `openbridger:<git-short-sha>`，不可变 |
| API Base | `https://openbridger.com/v1` |
| 工作目录 | `/srv/openbridger`（`.env`、`compose.release.yml`、证书、备份） |

`compose.release.yml` 中 `SESSION_COOKIE_TRUSTED_URL` 与 `FRONTEND_BASE_URL` 已写死 `https://openbridger.com`，与实际域名一致，**本轮无需改动**（后续建议改环境变量）。

---

## 段 1：主机与网络安全（S01–S05）

### S01 创建并加固轻量服务器

- 操作：阿里云轻量新建实例（Debian 12 或 Ubuntu 22.04；建议 2C4G 起、系统盘 60G+）；SSH 改用密钥登录并禁用密码登录；创建初始快照。
- 安全组：仅放行 22（限制来源 IP）、80、443。**不要把 3000 暴露到公网。**
- 通过标准：只能密钥登录；`curl http://<公网IP>` 从外部访问 3000 端口不通。
- 证据：安全组规则截图、快照 ID。
- 回退：用初始快照重建。
- 登记：把规格、地域、系统版本填进[生产环境台账](./ops-inventory.md)（**只填非敏感项，IP 和密钥不进仓库**）。

### S02 安装 Docker 与 Compose 插件

```bash
curl -fsSL https://get.docker.com | sh
apt-get install -y docker-compose-plugin
docker run --rm hello-world        # 必须成功
```

如需加速：在阿里云容器镜像服务申请加速器地址，写入 `/etc/docker/daemon.json` 后 `systemctl restart docker`。
通过标准：`docker version` 与 `docker compose version` 正常。

### S03 安装并配置 Nginx 反向代理

`/etc/nginx/sites-available/openbridger.conf` 要点（流式调用必须关缓冲）：

```nginx
server {
    listen 443 ssl http2;
    server_name openbridger.com;
    ssl_certificate     /etc/nginx/certs/openbridger.com.pem;
    ssl_certificate_key /etc/nginx/certs/openbridger.com.key;

    client_max_body_size 25m;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;          # SSE/流式必需
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }
}
server {
    listen 80;
    server_name openbridger.com;
    return 301 https://$host$request_uri;
}
```

通过标准：`nginx -t` 通过；`systemctl reload nginx` 无报错。

### S04 生成生产秘密

```bash
mkdir -p /srv/openbridger && cd /srv/openbridger
openssl rand -hex 32   # SESSION_SECRET
openssl rand -hex 32   # CRYPTO_SECRET
openssl rand -hex 24   # 数据库密码
```

写入 `/srv/openbridger/.env`，`chmod 600 .env`。
通过标准：`.env` 仅 root 可读；值不出现在任何文档或提交中。

### S05 确定数据库与 Redis

- 二选一：① 本机容器化 PostgreSQL 17 / MySQL 8.0（**必须与三库测试矩阵验证过的版本一致**）；② 阿里云 RDS + Redis。
- 无论哪种：开启自动备份，数据库与 Redis 不对外网开放。
- 通过标准：拿到可连接的 `SQL_DSN` 与 `REDIS_CONN_STRING`，应用能连上。
- 记录：选型结论与 DSN 主机（不记密码）。

---

## 段 2：域名与 TLS（S06–S09）

### S06 确定并写入版本号

`VERSION` 文件当前为空，构建镜像时会被写入二进制。发布前写入（如 `0.1.0-rc1`），并记入发布记录。
通过标准：`cat VERSION` 非空；`/api/status` 返回该版本。

### S07 Cloudflare DNS

| 记录 | 类型 | 值 | 代理 |
| --- | --- | --- | --- |
| `@` | A | 轻量服务器公网 IP | 开启（橙云） |
| `www` | CNAME | `openbridger.com` | 开启 |
| `docs` | A 或 CNAME | 同 IP / `openbridger.com` | 开启 |

通过标准：`dig +short openbridger.com` 返回 CF 边缘 IP；本地 `dig` 与服务器 IP 一致。

### S08 Cloudflare 证书与 HTTPS

- SSL/TLS 概述设为 **Full (strict)**；开启 Always Use HTTPS 与 HSTS。
- 在 CF → SSL/TLS → Origin Server 生成 **Origin CA 证书**（15 年），保存为 `/etc/nginx/certs/openbridger.com.pem` 与 `.key`，`chmod 600`。
- 通过标准：浏览器访问 `https://openbridger.com` 无证书告警；`curl -I https://openbridger.com/api/status` 返回 200。

### S09 Cloudflare 缓存规则（重要）

API 与流式响应不能被边缘缓存。在 CF → Caching → Cache Rules 新增规则：

- 匹配：`openbridger.com/v1/*` 与 `openbridger.com/api/*` → **Bypass cache**
- 静态资源与 `docs` 子域可正常缓存。

通过标准：连续两次 `curl -s https://openbridger.com/api/status` 响应头无 `cf-cache-status: HIT`；流式接口返回逐块输出而非一次性下发。

---

## 段 3：构建与首次部署（S10–S14）

### S10 拉取确定版本源码

```bash
cd /srv/openbridger
git clone https://github.com/randy72yt/OpenBridger.git repo
cd repo && git checkout <发布 commit 的 sha>
ls relaykit/go.mod                 # 必须存在，否则镜像构建会失败
```

通过标准：`git rev-parse --short HEAD` 得到将要用作镜像 tag 的 sha；`relaykit` 目录存在。

### S11 生成前端产物

```bash
cd /srv/openbridger/repo/web
bun install --frozen-lockfile      # 无 bun 先安装：curl -fsSL https://bun.sh/install | bash
bun run build
test -f dist/index.html && echo OK
```

通过标准：`web/dist/index.html` 存在（`Dockerfile.dev` 会校验这个文件，缺失则构建失败）。

### S12 构建发布镜像

```bash
cd /srv/openbridger/repo
export OPENBRIDGER_IMAGE_TAG=$(git rev-parse --short HEAD)
docker compose -f compose.release.yml build
docker images | grep openbridger
```

通过标准：镜像 `openbridger:<sha>` 生成成功。

### S13 补全生产环境变量

`.env` 至少包含：`OPENBRIDGER_IMAGE_TAG`、`SQL_DSN`、`REDIS_CONN_STRING`、`SESSION_SECRET`、`CRYPTO_SECRET`、`TRUSTED_PROXIES`。

`TRUSTED_PROXIES` 填**容器看到的反向代理地址**：Nginx 在宿主时通常是 docker 网桥网关，用 `ip addr show docker0` 确认（常见 `172.17.0.1`），不要填 `none` 也不要填 `0.0.0.0/0`。

### S14 启动并做健康检查

```bash
docker compose -f compose.release.yml up -d
docker compose -f compose.release.yml logs -f --tail=100
curl -s http://127.0.0.1:3000/api/status | head -c 300
```

通过标准：`/api/status` 返回 `"success":true`；容器 healthcheck 变为 healthy；日志无 panic、无数据库连接错误。
回退：`docker compose down`，修正 `.env` 后重来。

---

## 段 4：初始化与安全配置（S15–S19）

### S15 首次初始化与管理员加固

浏览器打开 `https://openbridger.com`，创建 root 账号（强密码，不复用任何测试密码），随后：

- 开启 MFA / Passkey，保存恢复码到离线位置；
- 确认普通账号无法越权访问管理接口（AUTH-005）。

通过标准：root 可登录且二次验证生效；非管理员访问管理接口被拒。

### S16 Cookie 与可信来源

确认生效值：`SESSION_COOKIE_SECURE=true`、`SESSION_COOKIE_TRUSTED_URL=https://openbridger.com`（Compose 已写死，与域名一致）、`TRUSTED_REDIRECT_DOMAINS=openbridger.com`。
通过标准：登录后 Cookie 带 `Secure`、`HttpOnly`、`SameSite`；退出登录与 refresh 正常（OPS-002）。

### S17 关闭公开注册与在线支付

- 后台关闭公开注册，改为邀请制/管理员建号；
- `OPENBRIDGER_ONLINE_PAYMENT_ENABLED=false`（发布 Compose 已默认 false）；
- 验证：钱包页无充值/套餐入口；直接调用订单创建接口被服务端拦截；`PAY-001～003` 记录「首版不适用」。

通过标准：页面与接口双侧均关闭，不能只关页面。

### S18 配置 SMTP 与发件域名

准备 `support@openbridger.com` 邮箱，配置 SMTP 后：

- 在 DNS 加 SPF、DKIM、DMARC（CF 侧添加 TXT 记录）；
- 发送测试邮件；验证邮箱验证与找回密码（AUTH-003、AUTH-004）。

通过标准：邮件真实送达且不在垃圾箱；找回密码全流程可完成。

### S19 反滥用与告警

- 注册/登录启用 Cloudflare Turnstile；
- 检查登录失败锁定、限流；
- 配置日志轮转与基础告警（ERROR_LOG_ENABLED 已 true），确认日志中不出现密钥、完整 Key、密码（OBS-001）。

**边缘层限流（2026-09-22 已上线）**：源站 Nginx `/etc/nginx/conf.d/ob-ratelimit.conf` 定义了 `ob_mail`(10r/m, burst 6) 与 `ob_auth`(30r/m, burst 20)，命中路由为 `/api/verification`、`/api/reset_password`、`/api/user/register`、`/api/user/login`。运维要点：

- **误伤排查**：`grep 'limiting requests' /var/log/nginx/error.log`（驳回时每条都记，`zone=` 字段标明是 ob_mail 还是 ob_auth）。
- **临时放宽**：注释掉配置里对应的 `limit_req` 行 → `nginx -t` → `systemctl reload nginx`（reload 平滑，不断连接）。
- **回滚**：`/etc/nginx/sites-available/openbridger.bak.<时间戳>` 是当次改动的备份。
- **禁止**改动 `map $http_cf_connecting_ip $ob_client_ip` 与 `limit_req_log_level error`，原因见 `prelaunch-test-plan.md` 1.5。

通过标准：自动化注册被拦截；日志无敏感明文。

---

## 段 5：上游渠道、模型与定价（S20–S24）

### S20 建立上游渠道

渠道 Base 填 `https://<上游域名>/v1`——**不要再拼一层 `/v1`**（`.env` 中的 `OPENBRIDGER_UPSTREAM_BASE` 已含 `/v1`，重复拼接会 `404 Invalid URL`，此前被误判为 403）。
通过标准：渠道测试连通成功。

### S21 上架首批模型并逐个实测

首批可用 9 个：`gpt-5.6-luna`、`claude-haiku-4-5`、`gemini-3.8-flash`、`gpt-5.6-terra`、`claude-sonnet-5`、`deepseek-v4-pro`、`gpt-5.6-sol`、`claude-opus-5`、`gemini-3.1-pro-preview`。
注意：**`deepseek-flash` 在上游不存在**，必须用 `deepseek-v4-flash`。

逐个执行非流式与流式调用，记录请求 ID、token 数、耗时与返回状态（RELAY-001、API-003、API-004）。未通过的模型先隐藏，不用别名顶替。

### S22 开启动态定价同步

`PRICING_CONTROL_ENABLED=true`（发布 Compose 已开启）。确认：

- 定时同步任务确实执行；
- 同步结果展示 warnings，被跳过的模型运营可见（事项 rhU308）；
- 本地倍率与上游 `group_ratio` 一致——倍率几天内会漂移（deepseek 6.8→1、glm 6→0.9），手工填值必然失真（事项 r6UjdF）。

通过标准：连续两次同步后倍率与上游一致；报价不过期（有效期 2 小时）。

### S23 与上游账单对账

对同一把上游 Key，调用前后取 `GET /v1/dashboard/billing/usage`，`quota = Δtotal_usage × 5000`，与本地扣费比对（BILL-001，事项 rpOYB6）。
通过标准：偏差在可接受范围内且原因可解释；异常值记录后重测，不通过口头确认关闭。

### S24 设置毛利与人工审批

设置目标毛利与最低价 → 生成定价方案 → **人工审核后发布**（任务不会自动发布售价）。
通过标准：方案可生成、可审核、可发布；发布后前台售价与方案一致；BILL-003、BILL-004 通过。

---

## 段 6：法律文本与文档站（S25–S27）

### S25 协议定稿

草稿在仓库根目录 `legal-drafts/`（用户协议、隐私政策）。发布前必须：

1. 拍板 `launch-plan.md` 2026-09-21 章节列出的 11 项待确认；
2. **首版不开放在线支付**，协议五/六、隐私七的支付相关表述据此收窄；
3. 删除全部「待确认／待公布」字样，补上版本号与生效日期。

通过标准：正文无「待确认」；主体为 OpenBridger；联系邮箱真实可收信。

**状态：已完成（2026-09-26）**

- 运营主体 **OpenBridger**、联系邮箱 **support@openbridger.com**（已验证可收信）已确认。
- 定稿正文在 `legal-drafts/published/{user-agreement,privacy-policy}.zh-CN.md`，v1.0 / 2026-09-26。
- **退款条款按用户决定保留并加注，未删除**：正文写明「当前未开放在线支付，暂无适用退款情形；开放付费前随协议更新公布」。原建议（整段删除）已作废。
- 其余 11 项待确认的处理方式：能确认的写死（主体、邮箱、支付关闭、退款标注）；不能确认的一律改为中性表述或权利保留条款（不指定管辖地、不承诺具体保留期限、不作「零日志」承诺），**不留下「待确认/待公布」字样**。`legal-drafts/*.md` 保留审阅留痕稿。

### S26 写入 legal 配置段

把定稿正文写入 `legal` 配置段的 `user_agreement` / `privacy_policy`（后台系统设置或数据库对应记录），**无需发版**。
通过标准：`GET /api/user-agreement` 与 `GET /api/privacy-policy` 返回非空正文；`/api/status` 中 `user_agreement_enabled`、`privacy_policy_enabled` 均为 `true`（LEGAL-001）。

**状态：已完成（2026-09-26）**。写入方式：`PUT /api/option/` 提交 `legal.user_agreement` / `legal.privacy_policy`（点号键由 `handleConfigUpdate` 处理，同时更新内存，无需重启容器）。落库 6834 / 6604 字节；匿名 `GET /api/user-agreement` 返回 200；`/api/status` 两个开关均为 `true`。

> 更新正文复用同一接口即可。不要用 `source /srv/openbridger/.env` 取数据库密码（含括号会报语法错并回显密码），改用 `grep "^SQL_DSN=" ... | sed -E 's|^SQL_DSN=[^:]+:([^@]+)@.*$|\1|'`。

### S27 部署文档站

```bash
cd /srv/openbridger/repo/docs-site
npm install && npm run build       # 产物在 dist/
```

把 `dist/` 部署为 `docs.openbridger.com` 静态站点（Nginx server_name docs.openbridger.com，root 指向 dist）。
随后把 `docs-site/src/pages/legal--*.md` 的占位页改为直链主站正式协议页。
通过标准：文档站可访问；按 DOC-001 实际走一遍快速开始教程能完成。

---

## 段 7：验收、备份与发布（S28–S34）

### S28 确认 CI 真的跑起来

本仓库是 `QuantumNous/new-api` 的 fork，至今 `total_count = 0`：需要登录 Actions 页面手动启用工作流。推一个 commit 触发后确认：

- 出现真实运行记录（**在出现记录前不能把远端检查计为通过**）；
- 三库矩阵日志中真的出现 `mysql` 与 `postgres` 两个子测试的 PASS 行（缺 DSN 时该测试会静默跳过且不计失败，事项 ryIHwN）。

### S29 OPS-001 配置隔离

确认生产未使用示例镜像 `new-api:latest`、无示例密码、秘密仅由环境注入；`/api/status` 不泄露敏感配置。

### S30 OPS-002 HTTPS、Cookie 与跨域

参见 S08、S16 的通过标准；确认 HTTP 全量跳转 HTTPS，跨域仅允许本站域名。

### S31 OPS-003 备份与恢复演练

```bash
# 数据卷
docker run --rm -v openbridger_data:/data -v /srv/openbridger/backup:/backup \
  alpine tar czf /backup/ob-data-$(date +%F).tgz /data
# 数据库
pg_dump ... > /srv/openbridger/backup/db-$(date +%F).sql   # 或 mysqldump
```

把备份传到异地（OSS/对象存储），并在**临时实例**上恢复一次验证可用。
通过标准：恢复后服务可启动、数据完整；备份脚本纳入定时任务。

### S32 OPS-004 回滚演练

记录当前 tag → 部署上一个 tag → 确认服务恢复 → 再切回。
通过标准：回滚在可接受时间内完成，数据无损坏。

### S33 完整 P0 验收

按 `production-acceptance.md` 第 0 节主线（阶段 A–H）与 E2E-001 在预发布环境跑一遍，结果填入第 12 节执行总表：

- AUTH-001～005、API-001～005、RELAY-001、BILL-001～005、LEGAL-001、UI-001、DOC-001、OBS-001、OPS-001～004、E2E-001；
- PAY-001～003 记录「首版不适用」并附支付入口/接口已阻断的证据；
- RELAY-002～003 若未启用备用路由则记「不适用」，且页面不得宣传自动故障转移。

失败项修复后**重新执行原用例**，不用口头确认关闭。

### S34 生产发布与邀请制试运行

- 先开放少量受邀用户，由管理员发放固定试用额度并设置单用户上限；
- 每日巡检：上游成功率、余额、毛利、错误率、用户反馈；
- 人工审核所有价格变更；
- 连续观察 3～7 天后再决定是否开放注册或扩大规模。

---

## 2. 当前状态一览

| 段 | 状态 |
| --- | --- |
| 段 1 主机与网络 | 待实施（S01 起） |
| 段 2 域名与 TLS | 域名已确认 `openbridger.com`，CF 未解析（S07 起） |
| 段 3 构建部署 | 本地已验证 Compose 解析与镜像构建；生产未执行 |
| 段 4 初始化与安全 | 支付关闭、钱包隐藏已在本地源码容器验证；生产未配置 |
| 段 5 上游与定价 | 凭证有效、9/10 候选模型可用、倍率已漂移；生产未上架 |
| 段 6 法律与文档 | 草稿已定位，待拍板与写入配置 |
| 段 7 验收发布 | CI 尚无运行记录；P0 用例未执行 |

## 3. 开始执行前需要确认的三件事

1. 轻量服务器已创建并可 SSH（给我规格与系统版本即可，不需要 IP 写进文档）。
2. 数据库/Redis 选型：本机容器还是阿里云 RDS/Redis。
3. 协议是否按「首版不开放在线支付」定稿——这会直接减少待确认项。
