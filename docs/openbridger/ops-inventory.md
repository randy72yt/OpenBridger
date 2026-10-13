# 生产环境台账

最后更新：2026-10-10（§3 注册/支付开关按当日生产实测刷新；§5 发布记录补 09-26/27 批次并记录版本核对缺口）

**本文件不记录公网 IP、密钥、密码、Token。** 主机连接信息保存在运维本机的 `~/.ssh/config`（Host 别名 `ob-prod`），应用运行时变量保存在服务器 `/srv/openbridger/.env`（权限 600）。这里只登记非敏感资产与「去哪儿找」的指针，供交接和回滚使用。

## 1. 资产登记

| 类别 | 项 | 值 | 备注 |
| --- | --- | --- | --- |
| 主机 | 提供商 | 阿里云轻量应用服务器 | S01 已建 |
| 主机 | 规格 / 地域 / 系统 | 2 vCPU / 1.6Gi 内存 / 系统盘 40G（可用 34G）；Ubuntu 24.04.2 LTS | 2026-09-21 实测；内存偏紧，已启用 2G swap（swappiness=10） |
| 主机 | 重要约束 | **不在服务器上编译**：内存不足以支撑 Go 与前端构建 | 镜像在本机或 CI 构建后推送（见下方「构建策略」） |
| 主机 | 部署目录 | `/srv/openbridger` | 内含 `repo/`（源码）、`.env`、备份目录 |
| 主机 | 开放端口 | 22（ufw limit 防爆破，全网可达——2026-10-10 前台账误记「限来源 IP」，实为全网；如需锁源 IP 见 ops-handbook §5.7 建议）、80/443（**仅 Cloudflare IP 段**，2026-10-10 起） | 3000 仅映射宿主 `127.0.0.1:3000`，不对外 |
| 接入 | SSH 配置位置 | 运维本机 `~/.ssh/config`，Host `ob-prod` | 不在仓库 |
| 接入 | 连接命令 | `ssh ob-prod` | 需要 IP、端口、用户、私钥四项，由运维保管 |
| 域名 | 主站 / 控制台 | `openbridger.com` | Cloudflare，2026-09-21 已解析（橙云代理）；`www` 为 CNAME |
| 域名 | 文档站 | `docs.openbridger.com` | Cloudflare，2026-09-21 已解析（橙云代理） |
| TLS | 证书 | Let's Encrypt（certbot 签发，覆盖 `@` / `www` / `docs`） | 2026-09-21 签发，2026-12-20 到期，已配自动续期；Nginx 强制 HTTP→HTTPS（301），流式配置 `proxy_buffering off` 已保留 |
| 数据库 | 选型 | 本机容器 **MySQL 8.0**，库 `openbridger` | 2026-09-21 部署；`innodb-buffer-pool-size=64M`、`max-connections=80`、`performance_schema=OFF`（内存占用 432MiB → 108MiB） |
| Redis | 选型 | 本机容器 **Redis 7-alpine** | `maxmemory 64mb`、`noeviction`、开启 appendonly |
| 运行时变量 | 位置 | 服务器 `/srv/openbridger/.env`（600） | `SQL_DSN`、`REDIS_CONN_STRING`、`SESSION_SECRET`、`CRYPTO_SECRET`、`TRUSTED_PROXIES`、`OPENBRIDGER_IMAGE_TAG` |
| 本地开发变量 | 位置 | 仓库根 `.env`（已被 `.gitignore` 忽略） | 只服务本地，不参与部署 |
| 备份 | 位置 | `/srv/openbridger/backup` + 异地对象存储 | 数据卷 tar + 数据库 dump |
| 镜像 | 命名 | `openbridger:<git-short-sha>` | 不可变 tag，回滚用上一个 sha |

## 2. 端口与网络

| 服务 | 绑定 | 可达范围 |
| --- | --- | --- |
| Nginx | 0.0.0.0:80 / :443 | **仅 Cloudflare IP 段**（ufw `cf-edge` 规则 22 条，2026-10-10 上线；直连源站实测被拒）。certbot HTTP-01 验证经 CF 代理回源，`--dry-run` 已实证续期不受影响 |
| OpenBridger 应用容器 | `127.0.0.1:3000` | 仅宿主，不对外 |
| MySQL 容器 | `172.17.0.1:3306` | 仅本机 docker0 网关，公网不可达 |
| Redis 容器 | `172.17.0.1:6379` | 同上 |

数据库与 Redis **不能**绑 `127.0.0.1`——应用容器经 docker0 网关访问宿主，绑 127.0.0.1 会 `connection refused`（2026-09-21 实际踩到）。

## 3. 已完成的安全开关（2026-09-21 首测；注册/支付两行 2026-10-10 复核）

| 项 | 值 | 验证方式 |
| --- | --- | --- |
| 公开注册 | `true`（2026-09-22 运营策略：注册常开；2026-09-27 上线注册送额度 `QuotaForNewUser=500000` 即 $1） | 2026-10-10 实测 `/api/status` `register_enabled=true` |
| 密码注册 | `true`（邮箱验证 `email_verification=true` + Turnstile 保护） | 2026-10-10 实测 `password_register_enabled=true` |
| 在线支付 | **已开启**（2026-09-26 用户拍板，`OPENBRIDGER_ONLINE_PAYMENT_ENABLED=true`；epay 支付宝/微信，最低 ¥10） | 订单创建/签名正常；**epay 网关域名白名单未过、实际不可付款**（柠檬支付商户后台待办，见运维总纲 §6 缺口 #3；PAY-001~003 未验收） |
| Turnstile | **已启用**（`TurnstileCheckEnabled=true`，Site/Secret Key 已入库） | 未带 token 的登录请求返回 `{"message":"Turnstile token 为空"}`；Secret Key 经 CF siteverify 端点确认为有效（`invalid-input-response` 而非 `invalid-input-secret`） |
| Cookie | `SESSION_COOKIE_SECURE=true`、`SESSION_COOKIE_TRUSTED_URL=https://openbridger.com` | `docker inspect` 已核对 |
| 可信代理 | `TRUSTED_PROXIES=172.17.0.1` | 同上 |
| CF 缓存 | `/api/status`、`/v1/models` 返回 `cf-cache-status: DYNAMIC` | **未缓存**，目标达成；Cache Rule 已配但边缘仍标记为 DYNAMIC（CF 默认动态判定优先），两者等价不缓存 |
| HSTS | `strict-transport-security: max-age=2592000; includeSubDomains`，`x-content-type-options: nosniff` | 2026-09-21 已开；max-age 先设 1 个月，**Preload 未开**（不可逆） |
| 管理员双因子 | TOTP `is_enabled=1` + Passkey 已注册（2026-09-21） | `two_fas`、`passkey_credentials` 均有记录；备份码 4 条需离线保管。缺陷与修复见 [控制台指引 15.1](./console-setup.md#151-passkey-注册失败与修复2026-09-21) |

未完成、需运维处理（2026-10-13 收敛）：~~根域 SPF 缺失~~（已解决：dig 实测 SPF/DKIM/DMARC 三件套齐，SPF 含 amazonses.com 即 Resend 现行发信底层，_spf.resend.com 已不存在勿用）；Turnstile Secret 曾两次明文泄露，**放量前必须轮换**（执行计划 B2）；CF SSL 若升级 Full(strict)，`api` 子域需先签含该主机名的证书（现源站证书 CN 不含 api，否则 526）。已完成项：MFA/Passkey（09-21）、SMTP/Turnstile 启用（09-21/22）、MySQL root 轮换（09-27，B1）。

### Nginx 配置资产（2026-09-22 上线）

Web 层配置**不在仓库**，全部落在主机：

| 文件 | 作用 |
| --- | --- |
| `/etc/nginx/sites-available/openbridger` | 主站 + 文档站 server 块，含 4 个限流 location；最近一次改动备份为同目录 `openbridger.bak.<时间戳>` |
| `/etc/nginx/conf.d/ob-ratelimit.conf` | `$ob_client_ip` map + `ob_mail`(10r/m, burst 6) / `ob_auth`(30r/m, burst 20) 两个 zone |
| `/etc/nginx/snippets/ob-proxy.conf` | 反代头部公共片段（含 `proxy_buffering off`，流式必须保留） |
| `/etc/nginx/snippets/ob-block-scan.conf` | 扫描路径（`/wp-*`、`/.env`、`/.git` 等）返回 404 |

- 限流驳回记录：`grep 'limiting requests' /var/log/nginx/error.log`（`zone=` 字段标明来源）。
- 两个不可随意改动的点：计数键必须取 `CF-Connecting-IP`（否则全世界共享一个桶）；`limit_req_log_level` 必须保持 `error`（设 warn/info 会被默认 `error_log` 级别过滤掉）。详见[上线前测试计划 1.5 节](./prelaunch-test-plan.md)。
- 改配置的标准动作：编辑 → `nginx -t` → `systemctl reload nginx`（平滑，不断连接）。

### 文档站（docs.openbridger.com）

2026-09-22 已部署，此前因静态目录为空返回 403。产物来自仓库 `docs-site/dist`（54 个文件 / 约 460K），上传至主机 `/srv/openbridger/docs-site`；`/`、`/en/`、`/zh/`、`/app.js`、`/styles.css` 均 200。

更新方式：本机在 `docs-site/` 下 `npm run build`（即 `node scripts/build.mjs`），然后把 `dist/` 覆盖上传即可，**不要在服务器上构建**（内存不足）。

## 4. 构建策略

服务器 1.6Gi 内存，**不具备**编译 Go 与前端的能力。镜像在运维本机（x86_64 + Docker + bun）构建后传入：

```bash
# 运维本机
cd <repo> && bun install --frozen-lockfile && bun run build      # 生成 web/dist
docker build -f Dockerfile.dev -t openbridger:$(git rev-parse --short HEAD) .
docker save openbridger:<tag> | gzip -1 | ssh ob-prod 'gunzip | docker load'
# 服务器
cd /srv/openbridger/repo && docker compose --env-file /srv/openbridger/.env -f compose.release.yml up -d
```

注意：`docker compose -f compose.release.yml build` 会因缺少运行时变量而报错（该文件的 `environment` 段带 `${VAR:?}`），**构建请直接用 `docker build`**，运行时变量只在部署时注入。

后续若要自动化：启用 GitHub Actions → 构建推送 GHCR → 服务器 `docker pull`，可省掉本机上传。

## 5. 发布记录

每次生产部署后追加一行：

| 日期 | commit sha | 镜像 tag | 变更摘要 | 执行人 | 结果 | 回滚目标 |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-09-21 | `1e93ef211` | `openbridger:1e93ef211` | 生产首次部署：MySQL/Redis/应用容器启动，`/api/status` 200 且容器 healthy | 运维 | 已启动，**尚未开放对外访问**（域名未解析、未初始化管理员） | 无上一个版本；异常时可 `docker compose down` 并用 systemctl 快照回滚 |
| 2026-09-22 | `63fb5ea55` | `openbridger:63fb5ea55` | 邮件配额保护：发信入口 `common.SendEmail()` 加 Redis 每日上限（默认 80 封/天），防匿名接口耗尽 Resend 配额导致当天邮件静默失效 | 运维 | 已部署并 healthy；回归脚本 17 项全通过。配额拦截的端到端行为待首次发信后确认计数 | `openbridger:1e93ef211` |
| 2026-09-26 | `59c0971a9`（tag `v0.1.0`） | `openbridger:59c0971a9` | 执行计划 A 阶段：v0.1.0 基线上线；协议 v1.0 写入 `legal` 段；生产回归 20/20 | 运维 | 已部署 healthy | `openbridger:63fb5ea55` |
| 2026-09-27 | `ab26333c39`…`5609c71f5`（tag `v0.1.4` 在 `9c8b9789b`） | `openbridger:9c8b9789b`（已核对） | 试运营批次：注册送额度 $1、兑换码、支持渠道露出、epay 在线支付开启、低余额提醒、支付完成跳转修复、支付下单合规校验、兑换码日志脱敏；协议升 v1.1；UptimeRobot 外部拨测 | 运维 | 各功能生产可用（运维总纲 §5.5） | `openbridger:59c0971a9` |
| 2026-10-11 | `426476ddb`（main，含 PR #3/#4/#5） | `openbridger:426476ddb` | v0.1.5 实际部署上产（此前生产停在 0.1.4）；obctl build 单分支克隆坑修复；验收 20/20（OB_TEST_MODEL=gpt-5.6-luna）；5 个上游故障模型按 E4 下架（38→33）；上游 A 同期大范围 503（见 incidents/2026-10-11） | 运维 | SMOKE PASS，`/api/status` version=0.1.5；B3 测试令牌确认早已删除 | `openbridger:9c8b9789b` |

> **版本核对缺口（已闭环）**：2026-10-10 发现生产报 0.1.0 而 VERSION 已 0.1.5；2026-10-11 核对生产实际运行 `openbridger:9c8b9789b`（= v0.1.4，VERSION 文件当时未同步），并于当日部署 `426476ddb`，现生产 `/api/status` 报 **0.1.5**，与仓库 VERSION 一致。遗留教训：v0.1.4 tag 的 VERSION 文件仍为 0.1.0——**发版时 VERSION 与 tag 必须同commit**（R7 已有此要求，0.1.4 执行时漏了）。

### 与主版本相关的运行时变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `EMAIL_DAILY_LIMIT_ENABLE` | `true` | 邮件每日上限开关 |
| `EMAIL_DAILY_LIMIT` | `80` | 每日上限。Resend 免费版为 100/天，留出 20 封余量调阈值时要兼顾告警邮件 |

## 6. 维护规则

1. IP、密钥、密码、API Key 一律不写进本仓库任何文件；需要交接时单独走受控渠道。
2. 服务器上的 `.env` 变更不通过仓库分发，直接在主机上改并留变更记录。
3. 每次发布必须能指名回滚目标（上一个镜像 tag），否则不允许发布。
4. 资产变化（换机、换库、换域名）先更新本表，再执行。
