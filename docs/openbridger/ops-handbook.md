# OpenBridger 运维总纲（Ops Handbook）

版本：v1.0 · 2026-09-27 · 适用部署：单机 ob-prod（47.80.28.213，2C/1.6G，阿里云轻量）  
地位：**运维入口文档**。出事先看本文 §3 应急速查表，再按引用文档深入。

## 0. 体系总览与文档地图

```
发现 ──→ 定级 ──→ 诊断 ──→ 处置 ──→ 验证 ──→ 记录
monitor.py   §3.1    obctl diag  §3.2 速查表   obctl smoke   §3.4 postmortem
告警邮件     S1-S4                /回滚/重建    /accept       incidents/
```

| 主题       | 文档                                 | 工具                             |
| -------- | ---------------------------------- | ------------------------------ |
| 监控告警     | 本文 §1                              | monitor.py（5min 巡检）、obctl      |
| 备份灾备     | `dr-plan.md`（RPO≤24h / RTO≤2h）     | backup.sh、obctl backup/restore |
| 应急响应     | 本文 §3                              | obctl diag/smoke/rollback      |
| 恢复重建     | `rebuild-runbook.md`（L1-L4 四层验证）   | obctl restore/deploy           |
| 发布       | 本文 §4（R0-R8 阶段）                    | obctl build/deploy/accept      |
| 日常操作命令与坑 | WorkBuddy 技能 openbridger-ops-agent Part 2  | —                              |
| 一键执行路由   | WorkBuddy 技能 openbridger-ops-agent | —                              |
| 非敏台账     | `ops-inventory.md`                 | —                              |

**obctl**：服务器 `/srv/openbridger/obctl`，`obctl help` 查看全部子命令。源码在 repo `docs/openbridger/scripts/obctl.sh`，改动后 scp 覆盖服务器副本。

## 1. 监控体系openbridger-ops-agent

### 1.1 监控矩阵（现状，全部自动运行）

| 指标     | 工具                                       | 频率      | 阈值/判据            | 告警        |
| ------ | ---------------------------------------- | ------- | ---------------- | --------- |
| 应用存活   | monitor.py → `127.0.0.1:3000/api/status` | 5 min   | 非 200 或超时        | 邮件（状态变化时） |
| 容器存活   | monitor.py → docker ps                   | 5 min   | 3 容器缺一           | 同上        |
| 磁盘     | monitor.py → df / 与 /var/lib/docker      | 5 min   | ≥85%             | 同上        |
| 备份新鲜度  | obctl backup-check（人工/周检）                | 按需      | latest dump >26h | 人工        |
| 日志体量   | docker json-file 轮转                      | 持续      | 50m×3 自动截断       | —         |
| 告警链路本身 | obctl alert-test                         | 每月/改配置后 | 测试邮件到达           | 人工        |
| 外部拨测   | UptimeRobot 云看板（4 监控点：主域/主域 api/api 子域/docs） | 5 min | 非 2xx/3xx | 邮件 xinlingwong668@gmail.com（宕+恢复） |

告警邮件：`support@openbridger.com`（CF Email Routing → 私人邮箱），SMTP 凭据从 DB options 读，SMTP 改密码后告警链路易断——**改完必须跑一次 `obctl alert-test`**。

外部看板（2026-09-27 上线）：UptimeRobot 免费计划，4 个监控点（主域首页、主域 /api/status、api 子域 /api/status、docs），5 分钟间隔，邮件通知 Up+Down。**公开状态页：https://stats.uptimerobot.com/aHV66DcfpB**。注意：免费计划不支持自定义域名（status.openbridger.com 如需可用 CF Redirect Rule 跳转）；v2 API 新建监控已对新账户关闭（报 plan 错误），须用 v3（Bearer + JSON，`timeout` 必填 ≤60）；短信联系人需后台激活。

### 1.2 已知监控盲区（如实陈述，按需补齐）

| 盲区           | 影响                   | 补齐方式                                                                                  |
| ------------ | -------------------- | ------------------------------------------------------------------------------------- |
| 无外部拨测        | 服务器能自查但 CF/出口断时收不到告警 | ✅ 已解决（2026-09-27）：UptimeRobot 云看板上线，见 §1.1 |
| 无请求错误率/延迟监控  | 上游 dddai 劣化只能从用户反馈发现 | 应用 logs 表已有用量记录；需要时写 SQL 查近 1h 错误占比                                                   |
| 无 TLS 证书到期监控 | certbot 自动续签失败时无感知   | UptimeRobot 拨测 HTTPS 自带证书告警                                                           |
| 内存无告警        | OOM 前无预警             | monitor.py 加一条 free 检查（P2）                                                            |

## 2. 数据备份与灾备

完整方案见 `dr-plan.md`，此处只给日常操作：

| 动作            | 命令                                                                                |
| ------------- | --------------------------------------------------------------------------------- |
| 立即备份一次（发版前必做） | `obctl backup`                                                                    |
| 检查备份健康        | `obctl backup-check`                                                              |
| 恢复 MySQL      | `obctl restore /srv/openbridger/backups/mysql-<ts>.sql.gz`（需 `OBCTL_CONFIRM=YES`） |
| 整机重建          | 按 `rebuild-runbook.md` 执行                                                         |

备份链路：服务器本地（日×7+周×4）→ Mac 异地（每日 09:30 WorkBuddy 自动化拉取，存 14 天）。**Mac 长期关机则异地副本滞后**（dr-plan §8 风险 1）。

## 3. 应急响应（Incident Response）

### 3.1 定级

| 级别 | 定义                 | 响应                       | 例              |
| -- | ------------------ | ------------------------ | -------------- |
| S1 | 全站不可用 / 数据丢失或泄露    | 立即处置，事后 24h 内 postmortem | 三容器全挂、误删表、密钥泄露 |
| S2 | 核心链路受损（推理调用/支付/登录） | 当日处置                     | 上游超时率飙升、支付下单失败 |
| S3 | 非核心功能受损，有绕行        | 3 日内                     | docs 站挂、邮件延迟   |
| S4 | 轻微/体验问题            | 下次发版顺带                   | UI 错位          |

### 3.2 标准流程

1. **诊断**：`ssh ob-prod '/srv/openbridger/obctl diag'`（一条命令出：容器/版本/资源/三域名/备份/监控/近 1h 错误日志）
2. **处置**：按 §3.3 速查表执行
3. **验证**：`obctl smoke`；涉及功能变更再 `OB_API_KEY=sk-... obctl accept`（须 20/20）
4. **记录**：S1/S2 必须写事故记录（§3.4 模板，存 `docs/openbridger/incidents/`）

### 3.3 场景速查表（一键处置）

| 症状                  | 一键处置                                                                               | 说明                                     |
| ------------------- | ---------------------------------------------------------------------------------- | -------------------------------------- |
| 单容器退出               | `docker start <容器>`，再 `obctl smoke`                                                | unless-stopped 通常已自动拉起                 |
| 应用 502/无响应          | `obctl logs app 200` 看 panic；`docker restart repo-openbridger-1`                   | 反复崩溃看内存（1.6G 小机）                       |
| 发版后异常               | `obctl rollback`                                                                   | ~8 秒切回旧镜像，再排查                          |
| 数据误删/损坏             | `obctl restore <最近dump>`                                                           | 会停应用约 1-3 分钟；先 `obctl backup` 留现场      |
| 磁盘 ≥85%             | `docker system prune -f`；`journalctl --vacuum-size=200M`                           | 日志已轮转，多为镜像/构建残留                        |
| 内存耗尽/OOM            | `obctl status` 看 free；重启应用容器；**禁止在服务器上构建前端**                                       | 2C4G 构建必须走 GOFLAGS 补丁（obctl build 已内置） |
| 邮件发不出               | `obctl alert-test` 定位；查 DB options SMTP 凭据                                         | 改 SMTP 密码后最常见                          |
| 定价 sync 失败          | 查 `model_price_policies` 是否非空（42 条）；v4 峰谷表达式是否被发布压平                                | 见 openbridger-ops-agent 技能 Part 2 定价章节   |
| 模型下架仍可见             | DELETE abilities 行 + 等 60s 缓存                                                      | 同上                                     |
| 服务器失联（ssh 无 banner） | 阿里云控制台看负载；大概率 OOM → 控制台重启                                                          | 历史实录：docker build 打崩 sshd              |
| 整机丢失（S1）            | `rebuild-runbook.md` 全流程                                                           | RTO ≤2h，前提是 Mac 异地备份在手                 |
| 上游 dddai 故障         | 非本站故障，渠道层 503；等上游恢复，可公告用户                                                          | dr-plan L3                             |
| 密钥泄露（S1）            | 立即轮换：DB options（Turnstile/SMTP）、.env（MySQL/Redis/Session）、用户令牌；全程命令见 openbridger-ops-agent 技能 Part 2 | 轮换后 `obctl alert-test` + `obctl smoke` |

### 3.4 事故记录模板（S1/S2 必填）

存为 `docs/openbridger/incidents/YYYY-MM-DD-<slug>.md`：

```markdown
# 事故：<标题>
日期 / 定级：S? / 状态：已恢复
## 时间线（发现→定级→处置→恢复，各时间点）
## 影响面（谁、多久、损失）
## 根因（技术原因，如实，不归咎人）
## 处置过程（执行过的命令）
## 防再发（action items，带 owner 和期限）
```

## 4. Release 方案（R0-R8）

原则：**每次发布都可 8 秒回滚；发布即演练回滚**。任何阶段不满足出口条件就停，不带病进入下一阶段。

| 阶段 | 名称   | 动作                                                                                                                                                                    | 出口条件                               |
| -- | ---- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- |
| R0 | 开发完成 | 本地：`go test ./...`；`cd web && NODE_OPTIONS= bun run build && vitest run`                                                                                              | 本地构建+测试全绿                          |
| R1 | CI   | `env -u http_proxy -u https_proxy gh workflow run "CI" --repo randy72yt/OpenBridger --ref <分支>`（fork 不投递 push 事件，必须手动）                                                | 三 job（backend/database/frontend）全绿 |
| R2 | 预检   | `obctl backup-check`；`obctl status` 看磁盘/内存；写变更说明（改了什么、影响面、回滚点）                                                                                                        | 备份新鲜、磁盘 <70%、变更说明成文                |
| R3 | 冻结备份 | `obctl backup`                                                                                                                                                        | 新 dump 生成（发版专用现场）                  |
| R4 | 构建   | 前端：Mac 构建 + `rsync -az --delete dist/ ob-prod:/srv/openbridger/repo/web/dist/`；后端：`obctl build <git-ref>` → `obctl build-status` 轮询                                   | 镜像 `openbridger:<sha>` 生成          |
| R5 | 切换   | `obctl deploy <sha>`（自动记录回滚点、冒烟、失败自动回滚）                                                                                                                               | SMOKE PASS，版本号正确                   |
| R6 | 验收   | `OB_API_KEY=sk-... obctl accept`（须 20/20）+ L4 抽验：计费一次真实调用对账、支付下单链接、Turnstile 注册页                                                                                      | 全部通过                               |
| R7 | 收尾   | 更新 VERSION；`git tag vX.Y.Z && git push git@github.com:randy72yt/OpenBridger.git main --tags`（必须 SSH）；涉及定价则走定价管线重算+审批+发布+v4 表达式 PATCH；涉及协议则更新 legal 并 PUT /api/option/ | tag 落库、生产版本=tag                    |
| R8 | 观察期  | 24h 内盯告警邮件；`obctl diag` 抽查两次                                                                                                                                          | 无 S2+ 异常                           |

**回滚触发条件**（R5-R8 任一时刻命中即执行 `obctl rollback`）：

- prod-acceptance 失败 ≥1 项且 10 分钟内无法定位
- 真实调用 5xx 或计费数值不符
- 容器反复重启 / 内存耗尽

**发布节奏**：当前单人维护，不设固定发版窗；建议累计若干变更集中发一次，避免频繁停机（每次切换 ~8s）。

## 5. 治理与节奏

| 事项     | 频率        | 动作                                        |
| ------ | --------- | ----------------------------------------- |
| 告警邮件   | 实时        | 收到即 `obctl diag`，按 §3 处置                  |
| 周检     | 每周        | `obctl diag` + `obctl backup-check`，5 分钟  |
| 月检     | 每月        | `obctl alert-test`；检查磁盘趋势；抽查一次 dump 可解压   |
| 备份恢复演练 | 每季度       | 临时容器灌库（dr-plan §7）                        |
| 回滚演练   | 每次发版      | deploy/rollback 本身即演练                     |
| 整机重建演练 | 每半年或重大变更后 | rebuild-runbook 全流程（**尚未做过端到端，建议公测前做一次**） |
| 密钥审查   | 每季度       | 谁还持有令牌/PAT；失效的删                           |

**机密纪律**（不可违反）：.env 只存服务器 600 权限；密码不出现在命令行/文件（SQL 一律 `obctl sql file`）；PAT 用完即 NULL；令牌泄露按 S1 处置。

## 6. 已知缺口清单（按优先级）

| # | 缺口                                                                                                                                                                                                                                           | 需要谁                   | 动作                                      |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------- | --------------------------------------- |
| 1 | ~~`api.openbridger.com` DNS 记录不存在~~ ✅ 2026-09-27 已修复：CF A 记录 + nginx server 块（复用主站 LE 证书）+ 全链路 200 验证。**注意**：origin 证书 CN 不含 api 主机名，CF SSL 模式若改为 Full (strict) 会 526，届时需给 api 签独立证书（DNS-01 需 CF API token）；另需确认 CF 缓存 Bypass 规则对 api 子域同样生效 | —                     | 已完成                                     |
| 2 | ~~无外部拨测~~ ✅ 2026-09-27 已上线：UptimeRobot 4 监控点 + 公开状态页 https://stats.uptimerobot.com/aHV66DcfpB（免费计划不支持自定义域名；status.openbridger.com 如需可做 CF Redirect Rule） | — | 已完成 |
| 3 | 整机重建未端到端演练                                                                                                                                                                                                                                   | Agent + 用户开临时新机       | rebuild-runbook 全流程走一遍                  |
| 4 | 异地备份依赖 Mac 开机                                                                                                                                                                                                                                | 可选                    | 加对象存储（COS/OSS 约 ¥1/月）双写                 |
| 5 | Redis 不异地、CF 配置靠文字清单                                                                                                                                                                                                                         | 已接受                   | dr-plan §8                              |

## 附录：命令速查

```bash
ssh ob-prod '/srv/openbridger/obctl diag'         # 出事先跑这条
ssh ob-prod '/srv/openbridger/obctl smoke'        # 处置后验证
ssh ob-prod 'OB_API_KEY=sk-... /srv/openbridger/obctl accept'   # 全量验收 20/20
ssh ob-prod '/srv/openbridger/obctl backup'       # 发版/危险操作前
ssh ob-prod '/srv/openbridger/obctl rollback'     # 8 秒回滚
ssh ob-prod '/srv/openbridger/obctl logs app 200' # 看应用日志
```
