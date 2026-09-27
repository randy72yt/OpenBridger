# OpenBridger 灾备方案（Disaster Recovery Plan）

版本：v1.0 · 2026-09-27 · 适用部署：单机阿里云轻量（ob-prod, 47.80.28.213）

## 1. 目标

| 指标 | 目标 | 依据 |
| --- | --- | --- |
| RPO（数据丢失上限） | ≤ 24 小时 | 每日 03:30 自动备份；配置变更随每日备份归档 |
| RTO（服务恢复上限） | ≤ 2 小时 | 新机重建手册（本文 §5）已验证过的恢复流程 |
| 可用性目标 | 个人/小团队使用，不承诺 SLA | 单机上游客观限制 |

## 2. 资产与单点清单

| 资产 | 位置 | 丢失后果 | 保护方式 |
| --- | --- | --- | --- |
| MySQL 数据（用户/额度/日志/定价） | 容器卷（宿主机 docker volume） | 最严重：账目与用户数据丢失 | 每日 mysqldump（本地 7 天 + 周日留存 4 周 + Mac 异地 14 天） |
| Redis（会话/限流计数） | 容器卷 | 轻：会话失效重登，限流计数重置 | 每日 AOF 归档（本地 7 天） |
| `/srv/openbridger/.env`（全部运行密钥） | 宿主机 | 严重：无法重建同配置服务 | 每日 config 归档（600 权限）+ Mac 异地 |
| 源码与镜像 tag | GitHub fork + 服务器 docker | 中：GitHub 可重新拉取构建 | main 分支 + v0.1.0 tag；旧镜像 tag 保留本地 |
| 域名/DNS/证书 | Cloudflare + 服务器 Nginx | 低：CF 托管不受服务器故障影响 | TLS 自动续签；DNS 即改即生效 |
| Cloudflare 配置（Turnstile/Email Routing/缓存规则） | CF 控制台 | 中：需人工重建 | 本文 §6 清单记录 |

## 3. 备份现状（自动运行中）

| 项 | 频率 | 留存 | 位置 |
| --- | --- | --- | --- |
| mysqldump（`--single-transaction --routines --triggers`） | 每日 03:30 | 日备 7 天 + 周备 4 周 | 服务器 `/srv/openbridger/backups/` |
| Redis AOF | 每日 03:30 | 7 天 | 同上 |
| 配置归档（.env + compose.infra.yml + monitor.py + repo HEAD） | 每日 03:30 | 7 天，600 权限 | 同上 |
| **异地拉取**（latest-mysql + latest-config + 日期副本） | 每日 09:30（WorkBuddy 自动化） | Mac 上 14 天 | `~/backups/openbridger/` |

注意：Mac 异地拉取依赖 Mac 开机且 WorkBuddy 运行；连续多日未开机时异地副本会滞后，开机后自动补拉最新。

## 4. 故障场景与处置

| 级别 | 场景 | 处置 | 预计恢复 |
| --- | --- | --- | --- |
| L0 | 容器崩溃/进程退出 | docker `unless-stopped` 自动重启；5 分钟巡检发现异常发邮件 | 分钟级，自动 |
| L1 | 数据误删/损坏（如误操作删表） | 用最近 dump 按 §5.3 恢复 | 30 分钟 |
| L2 | 服务器整机丢失/不可恢复 | 按 §5 全新重建（需异地备份可用） | ≤ 2 小时 |
| L3 | 上游 dddai 不可用 | 非本站故障；渠道层超时返回 503，监控会反映为错误率上升；待上游恢复 | 不可控 |
| L4 | Cloudflare 账号问题 | 人工登录 CF 处理；DNS/证书为托管服务，历史无单点故障 | 不可控 |

## 5. 整机重建手册（L2，RTO ≤ 2h）

前置：Mac 上 `~/backups/openbridger/` 有 latest-mysql.sql.gz 与 latest-config.tar.gz；GitHub fork 可访问；CF 控制台可登录。

### 5.1 新机准备（约 20 分钟）
1. 阿里云轻量新开实例（≥2C4G，Ubuntu/Debian），安全组放行 22/80/443
2. 安装 docker + compose plugin；配置 SSH（建议沿用原 key）
3. `mkdir -p /srv/openbridger/backups`

### 5.2 配置与代码（约 10 分钟）
1. Mac 上解包 config：`tar xzf ~/backups/openbridger/latest-config.tar.gz -C /tmp/ob-config`
2. `scp /tmp/ob-config/.env /tmp/ob-config/compose.infra.yml 新机:/srv/openbridger/`，权限 600
3. `git clone git@github.com:randy72yt/OpenBridger.git /srv/openbridger/repo`，`git checkout <config归档里的repo-head或v0.1.0>`
4. 构建镜像（小内存机需先给 Dockerfile.dev 加 `ENV GOFLAGS=-p=1 GOMAXPROCS=1`，构建后还原）或按 compose.release.yml 构建；前端 dist 从本地构建产物 rsync 至 `web/dist/`

### 5.3 数据恢复（约 15 分钟）
1. 启动 infra：`cd /srv/openbridger && docker compose -f compose.infra.yml --env-file .env up -d`
2. 等 MySQL healthy 后：`scp ~/backups/openbridger/latest-mysql.sql.gz 新机:/srv/openbridger/backups/`
3. `zcat /srv/openbridger/backups/latest-mysql.sql.gz | docker exec -i openbridger-mysql-1 sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" openbridger'`
4. 抽查：`users/options/channels/tokens` 行数与备份前一致

### 5.4 切换流量（约 10 分钟 + DNS 传播）
1. 启动应用：`cd /srv/openbridger/repo && OPENBRIDGER_IMAGE_TAG=<tag> docker compose -f compose.release.yml --env-file /srv/openbridger/.env up -d`
2. CF DNS 把 `openbridger.com` / `api.openbridger.com` A 记录指向新 IP（TTL 300，分钟级生效）
3. Nginx + TLS：按原方式签证书（Let's Encrypt 重新签发即可）
4. 跑回归：`OB_BASE=https://openbridger.com OB_API_KEY=<恢复后有效令牌> bash prod-acceptance.sh`，20/20 即恢复完成

### 5.5 重建后必查
- Turnstile siteverify 正常；邮件发送正常（SMTP 走 Resend，与服务器无关）
- 定价 sync 定时任务恢复（system_tasks 表随数据恢复）
- 监控脚本与 cron 重新安装（backup.sh / monitor.py / crontab，参考 §3 与运维台账）

## 6. Cloudflare 侧配置清单（L4 重建依据）

- DNS：A `@`/`api` → 服务器 IP（Proxied）；TXT SPF `v=spf1 include:_spf.mx.cloudflare.net include:amazonses.com ~all`；TXT `_dmarc` `v=DMARC1; p=none; rua=mailto:support@openbridger.com`；MX route1/2/3.mx.cloudflare.net（Email Routing 自动）
- Email Routing：`support@` → 用户私人邮箱
- Turnstile：widget 绑 openbridger.com；site key/secret 存于 DB options（随数据恢复）
- 缓存规则：`/v1/*`、`/api/*` Bypass cache
- 边缘限流规则（见 docs/openbridger/ 运维台账）

## 7. 演练制度

| 演练 | 频率 | 上次执行 |
| --- | --- | --- |
| 备份恢复演练（临时容器灌库验证） | 每季度 | 2026-09-26 ✅ |
| 回滚演练（切旧镜像再切回） | 每次发版 | 2026-09-26 ✅ |
| 整机重建演练（按 §5 全流程） | 每半年或重大变更后 | 未做（建议首次公测前做一次） |

## 8. 已知遗留风险（如实陈述）

1. **异地备份依赖 Mac 开机**：Mac 长期关机时 RPO 退化为仅存服务器本地（整机丢失即丢数据）。如需硬保障，后续可加对象存储（COS/OSS，约 ¥1/月）。
2. **Redis 只本地留存**：会话数据不做异地，可接受（损失=全员重登）。
3. **重建手册未完整演练过**：§5 各步骤都单独验证过，但端到端整机重建未实操，首次真实故障时 RTO 可能超出 2h。
4. **CF 配置无自动备份**：靠 §6 文字清单人工重建，约 15 分钟。
