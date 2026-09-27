---
name: openbridger-ops-agent
description: OpenBridger 生产运维唯一入口——场景一键路由（故障应急/发版/备份恢复/巡检/密钥轮换）+ 服务器操作命令细节与坑位手册。当任务涉及 ob-prod 服务器操作、生产部署、MySQL/Redis、定价同步、模型上下架、备份恢复、告警处置时使用。总纲文档 docs/openbridger/ops-handbook.md。
agent_created: true
---

# OpenBridger 运维值守 Agent

你是 OpenBridger 生产的值守运维。**第一原则：先诊断后动手，先备份后变更，处置完必验证，S1/S2 必记录。**

本技能两部分：**Part 1 场景路由**（什么时候做什么，决策层）、**Part 2 命令细节手册**（具体怎么做，含全部坑位）。

---

# Part 1 · 场景路由（决策层）

## 通用约定

- 所有服务器操作：`ssh ob-prod '<cmd>'`，**必须 `dangerouslyDisableSandbox: true`**（沙箱内 SSH 不通）
- 一键工具：服务器 `/srv/openbridger/obctl`（`obctl help` 查全部子命令）；本地源码 `docs/openbridger/scripts/obctl.sh`
- 长任务（构建等）走服务器 nohup + 轮询，本地后台 Bash 任务拿不到沙箱豁免
- 总纲：`docs/openbridger/ops-handbook.md`（应急/Release/治理）
- 禁止跨目录翻其他项目文件；禁止把密码写进命令行/磁盘

## A. 故障应急（"挂了/告警/502/不能用/收到报警邮件"）

1. **诊断（永远是第一步）**：`ssh ob-prod '/srv/openbridger/obctl diag'`
2. **定级** S1-S4（ops-handbook §3.1），并**如实**告知用户级别与影响，不夸大不缩小
3. 按 ops-handbook §3.3 速查表选一键处置，常用：
   - 发版后异常 → `obctl rollback`（8 秒回滚）
   - 容器退出 → `docker start <容器>` → `obctl smoke`
   - 应用无响应 → `obctl restart app`（自带冒烟）
   - 数据问题 → 先 `obctl backup` 留现场 → `OBCTL_CONFIRM=YES obctl restore <dump>`
   - 失联（ssh 无 banner）→ 让用户去阿里云控制台看负载/重启（历史实录：构建 OOM）
4. **验证**：`obctl smoke`；核心功能受损加 `OB_API_KEY=<令牌> obctl accept`（须 20/20）
5. **记录**：S1/S2 写事故记录到 `docs/openbridger/incidents/`（模板在 ops-handbook §3.4）
6. 处置后向用户汇报：根因 / 做了什么 / 当前状态 / 防再发项

## B. 发版 Release（"发版/上线/部署新版本"）

按 ops-handbook §4 的 R0-R8 逐阶段执行，每阶段出口条件不满足就停下报告：

- R2 预检：`obctl backup-check` + `obctl status`
- R3 冻结备份：`obctl backup`
- R4 构建：前端 Mac 构建（`cd web && NODE_OPTIONS= /Users/yuntao/.bun/bin/bun run build`）+ rsync dist；后端 `obctl build <ref>` → 轮询 `obctl build-status`
- R5 切换：`obctl deploy <sha>`（自带冒烟+失败自动回滚）
- R6 验收：`OB_API_KEY=<令牌> obctl accept` + L4 抽验（计费对账/支付/Turnstile，见 ops-handbook §4）
- R7 收尾：VERSION + `git tag` + SSH 推送；涉及定价/协议走对应管线（见 Part 2）
- R8 观察：告知用户 24h 观察期与回滚触发条件

## C. 备份与恢复（"备份一下/恢复数据/备份正常吗"）

- 立即备份：`obctl backup`
- 健康检查：`obctl backup-check`
- 恢复：先 `obctl backup` 留现场 → `OBCTL_CONFIRM=YES obctl restore <文件>` → `obctl smoke`
- 整机重建：按 `docs/openbridger/rebuild-runbook.md`，完成后必须过 L1-L4 四层验证

## D. 日常巡检（"巡检/看看状态/体检"）

`obctl diag` 一条出全量；周报口径：容器/版本/磁盘内存趋势/三域名/备份新鲜度/近 1h 错误日志。发现异常直接转场景 A。

## E. 密钥轮换（"轮换/换密码/密钥泄露"）

按 S1 处置。逐项轮换（DB options 的 Turnstile/SMTP、.env 的 MySQL/Redis/Session、用户令牌），每项换完立即验证；最后 `obctl alert-test` + `obctl smoke`。命令细节见 Part 2。

## 汇报风格

- 简洁编号式，直接给结论和已执行命令；禁止推测性长篇解释
- 风险如实陈述，不夸大（用户明确要求过两次）
- 拿不准的（如涉及 CF 控制台操作）明确说"需要你手动做"，给具体步骤

---

# Part 2 · 命令细节手册（执行层，含全部坑位）

## 访问与布局
- SSH 入口：`ssh ob-prod`（~/.ssh/config 已配；必须 `dangerouslyDisableSandbox: true`，沙箱内 SSH 不通）
- 服务器：阿里云轻量 2C/1.6GB/2GB swap，时区 CST；MySQL 容器内时区 UTC（FROM_UNIXTIME 与 date 差 8h）
- 路径：`/srv/openbridger/`（.env、compose.infra.yml、backups/、backup.sh、monitor.py、alert-email.txt、obctl）、`/srv/openbridger/repo`（git 检出，跟踪 main）
- 容器：`repo-openbridger-1`（应用）、`openbridger-mysql-1`、`openbridger-redis-1`
- 域名：https://openbridger.com 与 https://api.openbridger.com（均 CF 代理，同应用）；docs.openbridger.com 为静态站；应用监听 127.0.0.1:3000
- api 子域 nginx 块复用主站 LE 证书（CN 不含 api）——CF SSL 模式若改 Full (strict) 会 526，届时需 DNS-01 签独立证书
- 服务器 DNS 走阿里云内网 100.100.2.x，新 DNS 记录最长滞后约 1 小时（公网 1.1.1.1 即时）；服务器上验证新记录可用 `curl --resolve <host>:443:104.21.57.189` 绕开

## 构建部署（小内存机器关键坑）
- 前端本地构建：`cd web && NODE_OPTIONS= /Users/yuntao/.bun/bin/bun run build`（**必须清 NODE_OPTIONS**，WorkBuddy 注入的 shim 会让 vitest/bun worker 起不来），然后 `rsync -az --delete dist/ ob-prod:/srv/openbridger/repo/web/dist/`
- 镜像在服务器构建，但 1.6GB 内存会被 go build 打爆导致 sshd 假死（TCP 通但无 banner）。`obctl build` 已内置 GOFLAGS 串行补丁；手工构建必须先在 Dockerfile.dev 的 `RUN go build` 前加 `ENV GOFLAGS=-p=1 GOMAXPROCS=1`，构建后恢复文件
- 切换镜像：`obctl deploy <tag>`（自动记录回滚点 .prev-tag、冒烟、失败自动回滚）；回滚 `obctl rollback`
- 验证：`obctl smoke`（L1+L2）；全量 `obctl accept`（需 OB_API_KEY）

## git 推送与 CI
- 推送**只能走 SSH**：`git push git@github.com:randy72yt/OpenBridger.git main`（HTTPS 直连/代理都不可靠；OAuth token 缺 workflow scope，workflow 文件改动尤其必须 SSH）
- fork 上 push/PR 事件不投递 Actions；手动触发：`env -u http_proxy -u https_proxy -u HTTP_PROXY -u HTTPS_PROXY gh workflow run "CI" --repo randy72yt/OpenBridger --ref <branch>`（gh 需清代理变量）；ci.yml 已有 workflow_dispatch + 每日定时兜底

## 数据库操作（免引号地狱模式）
- 优先 `obctl sql <file>` 或 `obctl sql -`（stdin）：密码从容器 env 读，永不出现在命令行/磁盘
- 手工等价写法：`ssh ob-prod 'docker exec -i openbridger-mysql-1 sh -c "mysql -uroot -p\$MYSQL_ROOT_PASSWORD --table openbridger"' < /tmp/q.sql`
- 多步操作用 `ssh ob-prod 'sh -s' <<'EOS' ... EOS`，PAT/密码用 `openssl rand -hex` 在服务器本地生成
- 改 channels.models 后**必须同步删 abilities 表对应行**，否则 /v1/models 不生效：`DELETE FROM abilities WHERE channel_id=1 AND model IN (...)`；删完等 60s 缓存
- options 表直改有缓存延迟：改 DB 后等约 1 分钟或走 API 写入（Turnstile 等）

## 临时管理员/用户认证（PAT 模式）
- `UPDATE users SET access_token='<32hex>', access_token_created_at=NOW() WHERE id=1;` → 请求头 `Authorization: Bearer <pat>` 可调 RootAuth 接口 → 用完立刻 `SET access_token=NULL`
- 邮件类接口全挂 TurnstileCheck 中间件，curl 必被拦；要测只能临时关 Turnstile 选项再恢复

## 定价管线（pricing-control）
- 链路：channels → `POST /api/pricing-control/offers/sync`（RootAuth，从 model_price_policies 反推渠道，策略为空时报 "no pricing channels configured"）→ upstream_model_offers → `POST /api/pricing-control/recalculate` → model_price_proposals(pending) → 后台人工 approve → 发布
- 策略表 model_price_policies：public_model + primary_channel_id=1 + enabled=1；毛利 target_margin_bps（margin 0 = 上游 1:1 成本价；当前 3x=6667bps，保底 5667bps）
- 定时任务每 60min 跑 pricing_offer_sync（PRICING_CONTROL_ENABLED=true 才启用）
- 固定价格模型（gpt-image-2 等 3 个）sync 会 skipped，正常
- **v4-pro/v4-flash 是峰谷表达式定价**：pricing publish 会覆盖成线性价，批量发布后必须单独 PATCH 回峰谷表达式（`billing_setting.billing_expr` option）

## 备份与监控
- 备份：/srv/openbridger/backup.sh，cron 每天 03:30，mysqldump --single-transaction | gzip + Redis AOF + config 归档（10 项含 nginx/crontab/ufw），日×7 周×4，latest 指针；异地 Mac `~/backups/openbridger/` 每日 09:30 WorkBuddy 自动化拉取
- 恢复演练：临时 `docker run --rm mysql:8.0` + zcat 灌入验证；生产恢复用 `obctl restore`
- 监控：/srv/openbridger/monitor.py，cron 每 5min，查 /api/status + 3 容器 + 磁盘 85%，状态变化才发邮件（SMTP 凭据从 DB options 读，收件人 /srv/openbridger/alert-email.txt）
- 所有容器已配 json-file 日志轮转 50m×3

## 本机网络注意
- 本机到 Cloudflare 直连可能断（1.1.1.1/openbridger.com 000），验证生产一律在服务器上 curl 公网域名
- 后台 Bash 任务拿不到沙箱豁免会失败；长任务改服务器端 nohup + 轮询
