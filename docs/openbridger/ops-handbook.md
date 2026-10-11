# OpenBridger 运维总纲（Ops Handbook）

版本：v1.0 · 2026-09-27 · 适用部署：单机 ob-prod（地址见运维机 `~/.ssh/config`，2C/1.6G，阿里云轻量）  
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
| 无请求错误率/延迟监控  | 上游 A 劣化只能从用户反馈发现 | 应用 logs 表已有用量记录；需要时写 SQL 查近 1h 错误占比                                                   |
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
| 上游 A 故障         | 非本站故障，渠道层 503；等上游恢复，可公告用户                                                          | dr-plan L3                             |
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
| 上游安全公告 | 每月        | `docs/openbridger/scripts/upstream-sync-check.sh`——exit 0 直接结束；exit 2 按技能 `openbridger-upstream-sync` 整合（安全修复无条件摘）。2026-10-10 首查：落后 9 个 release（rc.35–rc.43），无 CVE 级命中 |
| 备份恢复演练 | 每季度       | 临时容器灌库（dr-plan §7）                        |
| 回滚演练   | 每次发版      | deploy/rollback 本身即演练                     |
| 整机重建演练 | 每半年或重大变更后 | rebuild-runbook 全流程（**尚未做过端到端，建议公测前做一次**） |
| 密钥审查   | 每季度       | 谁还持有令牌/PAT；失效的删                           |

**机密纪律**（不可违反）：.env 只存服务器 600 权限；密码不出现在命令行/文件（SQL 一律 `obctl sql file`）；PAT 用完即 NULL；令牌泄露按 S1 处置。

## 5.5 运营配置（试运行，2026-09-27 落地）

| 项 | 现状 | 怎么改 |
| --- | --- | --- |
| 新用户注册送额度 | `QuotaForNewUser=500000`（$1；1 USD = 500000 quota） | 后台或 `PUT /api/option/`，改完注册测试号验证 |
| 兑换码 | 已开通：Beta 批次两档（面额与张数明细见后台兑换码管理页，不入册）；用户控制台「充值与支付」自助兑换 | 后台批量生成（单次上限 100 张）；发额度 = 发个码，不再改库 |
| **自助在线支付充值** | 已开通（`OPENBRIDGER_ONLINE_PAYMENT_ENABLED=true`），支付方式仅保留支付宝/微信（最低 ¥10；$10 → ¥73，按 1 USD=¥7.30）。订单创建、签名、回调地址均正常；**支付完成后跳 `/wallet?pay=success`**（v0.1.4 修复，原为 `/usage-logs`） | ⚠️ **当前不可付款**：epay 网关返回「该域名不可发起支付：域名没过白」，需商户在柠檬支付后台把 `openbridger.com` 加入支付域名白名单 |
| **低余额自动提醒** | 内置：消费后余额 < `QuotaRemindThreshold=250000`（$0.5）自动发邮件，含钱包充值链接 | 改 `QuotaRemindThreshold`；用户级阈值用 `QuotaWarningThreshold` |
| 定价页 | `/pricing` 已开放（`HeaderNavModules.pricing.enabled=true`，无需登录）；数据来自 `/api/pricing`（**33 个模型**，2026-10-11 按 E4 下架 5 个上游故障模型后；v4 峰谷表达式已核对保留） | 要下线：把 pricing.enabled 改 false |
| 支付合规确认 | `payment_setting.compliance_*` 已确认（user 1，v1） | 该字段禁止走通用 option 接口，只能在后台确认 |
| 支持渠道 | 页脚挂 `support@openbridger.com` + 状态页链接；docs 站新增「联系与故障申报」页（中英） | 页脚改 `Footer` option（支持 HTML）；docs 改 `docs-site/src/pages*/guide--support.md` 后 `node scripts/build.mjs` + rsync 到 `/srv/openbridger/docs-site/` |
| status 子域 | 用 UptimeRobot 默认地址（免费计划不支持自定义域名） | 想要 `status.openbridger.com` 需在 CF 配 Redirect Rule |

**注意**：Email Obfuscation 会把页面里的 mailto 改写成 `/cdn-cgi/l/email-protection#...`，验证时不要 grep 原始邮箱字符串。

**坑：`PayMethods` 的值必须是 `[]map[string]string`**——即**所有字段值都要是字符串**。写成 `{"min_topup":10}`（数字）会导致 `json: cannot unmarshal number into Go value of type string`，配置同步每分钟报一条错且该项不生效。正确写法：`{"min_topup":"10"}`。同理，用 `PUT /api/option/` 写数组型配置会被存成 Go 的 `[map[...]]` 格式（非法 JSON），数组型配置**一律用 SQL 直接写**。

## 5.6 上游同步策略（fork 维护，2026-10-10 定策）

基座 fork 自 `QuantumNous/new-api`（fork 点 `6298b0f32`，2026-09-07；AGPL-3.0，公开仓库即网络条款的合规履行，配置/密钥不入库的纪律不变）。**不追 latest、不做整分支 merge**：上游自身处于 rc 阶段高频发版，「稳健」来自我们自己的验证闸门（三库矩阵 + `obctl accept` 20/20 + 8 秒回滚），不来自上游版本号。落后上游多少提交属预期，按需摘取：

| 上游变更类型 | 动作 |
| --- | --- |
| 安全修复（auth / relay / webhook） | 无条件摘，冲突手动解 |
| 新渠道、新模型适配器（`relay/channel/` 新目录） | 商业需要才摘；纯附加，零冲突 |
| 上游重构、目录搬家、接口改形 | 主动跳过，除非阻塞了必须摘的修复 |
| 与定价控制面重叠的计费类新功能 | 跳过（自研替代已在生产运行） |

操作口径：分支上 `git checkout upstream/main -- <路径>` 搬文件 → 三库矩阵 + 前端构建 + `obctl accept`（须 20/20）→ **PR 留痕**（流程模板：PR #3）。

**何时整合（2026-10-10 定，消灭开放式判断）**：只有四个入口，**没有「到时间了就整合」**——

| 入口 | 信号 | 谁发现 |
| --- | --- | --- |
| ① 安全 | 检查脚本结论 🔴（提交宽词表 + release notes 窄词表，test 基建类过滤防误报） | 月检自动 |
| ② 需求 | 要上某新模型 / 新渠道 | 用户/运营自然知道 |
| ③ 事故 | 我们踩到的 bug 疑似上游已修 | incident 处置时反查 |
| ④ 迁移 | 上游计费实现满足三条件（见下文阵地段） | 年度评估 |

除四个入口外一律不整合；脚本结论 ⚪ = 在 state.md 记一行「检查-跳过」即完事，落后 release 数不是负债、不需要清零。

就地修改面登记：`git diff 6298b0f32 HEAD --name-status` 中的 M 文件即对上游文件的就地修改。2026-10-10 基线：新增（A）115 个文件、就地修改（M）69 个——其中**后端核心就地修改仅约 15 个**（`common/email.go`、`controller/topup.go`/`user.go`/`audit.go`、`model/main.go` 及定价相关、`router/*.go`、`setting/operation_setting`），其余为 web 品牌文案、i18n、构建配置，冲突机械好解。原则：**能放新文件就不改上游文件；新增就地修改保持最小，并在下次盘点时更新本节数字**。

工具化（2026-10-10）：检查器 `scripts/upstream-sync-check.sh`（比对 `releases/latest` 与 `skills/openbridger-upstream-sync/state.md` 基线——无变化 exit 0 跳过，有变化 exit 2 输出报告：区间 release 列表 / 安全关键词命中 / 热区文件命中 / 计费阵地命中 / relay 规模）；整合流程封装为技能 `skills/openbridger-upstream-sync/`（自包含，任何 AI 工具可执行）。热区与阵地清单以脚本内 `HOT_FILES`/`GUARD_FILES` 数组与本节登记为准，同步维护。

计费阵地（GUARD，2026-10-10 增）：**billingexpr 表达式引擎是上游自带**（`pkg/billingexpr/`，fork 点即存在、我们零改动），我们的定价控制面是叠加层，计费接线经过一批**上游拥有**的 seam 文件（`relay/helper/price.go`、`service/quota.go`/`text_quota.go`、`relay/relay_task.go`、`model/task.go` 等，全清单见脚本 `GUARD_FILES`）。阵地命中即最严口径：**禁止整文件 checkout，逐块 diff 人工审查，PR 附结论，闸门必跑 billingexpr 与定价控制面用例**。上游同期也在发展表达式化定价（rc.35–rc.43 已见 `convert legacy prices into expression drafts`、`model_pricing_conversion.go` 等；首查实测 10 个阵地文件全部被上游改动），凡属其计费「新功能」一律④桶跳过；仅当上游实现覆盖毛利+审批流+表达式版本语义、脱离 rc 连发、且存量策略数据迁移成本可接受时，才评估整体迁移。

## 5.7 运营文档迁出 runbook（备而不发，2026-10-10 制定）

背景：`docs/openbridger/`（运维手册、skills、台账）与 `.agents/skills/` 是自写运营知识，**不属于 AGPL 必须公开的软件源码**，放在公开 fork 只是历史顺带。2026-10-10 已完成文档脱敏（服务器 IP→`~/.ssh/config` 指针、上游供应商→「上游 A/B」、供应商域名移除、兑换码批次明细→后台指针）；本 runbook 在触发条件到达时执行整体迁出，平时不动。

**触发条件（任一）**：出现竞争模仿迹象；文档需写入更敏感内容（具体毛利策略数值、供应商合同条款）；公开仓库被频繁 fork/clone 监控。

**步骤**：

1. 新建私有仓库（如 `openbridger-ops`），用 `git filter-repo --path docs/openbridger --path .agents` 从本仓库抽取完整历史导入；
2. 公开仓库以**普通删除提交**移除同路径——不 rewrite 公开历史（生产镜像按 git sha 命名，rewrite 会作废整个回滚指认体系）；
3. 修正引用：AGENTS.md/CLAUDE.md 与代码注释中指向 `docs/openbridger/` 的路径；`docs-site/`（对外文档站源码）**保留**在公开仓库；
4. 迁移本机与服务器上的副本路径；更新 `skills/README.md` 分发说明；
5. 历史已公开内容（2026-10-10 前的 IP、供应商名等）视为已泄露，依靠「ufw 仅放行 Cloudflare 段」对冲源站直连风险；如需彻底清除历史再立专项评估。

迁出后 AGPL 义务不变：软件源码仍在公开 fork，持续满足网络条款。

## 6. 已知缺口清单（按优先级）

| # | 缺口                                                                                                                                                                                                                                           | 需要谁                   | 动作                                      |
| - | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------- | --------------------------------------- |
| 1 | ~~`api.openbridger.com` DNS 记录不存在~~ ✅ 2026-09-27 已修复：CF A 记录 + nginx server 块（复用主站 LE 证书）+ 全链路 200 验证。**注意**：origin 证书 CN 不含 api 主机名，CF SSL 模式若改为 Full (strict) 会 526，届时需给 api 签独立证书（DNS-01 需 CF API token）；另需确认 CF 缓存 Bypass 规则对 api 子域同样生效 | —                     | 已完成                                     |
| 2 | ~~无外部拨测~~ ✅ 2026-09-27 已上线：UptimeRobot 4 监控点 + 公开状态页 https://stats.uptimerobot.com/aHV66DcfpB（免费计划不支持自定义域名；status.openbridger.com 如需可做 CF Redirect Rule） | — | 已完成 |
| 3 | **在线支付实际不可付款**（订单能建、网关拒绝）：epay 商户后台未授权支付域名，`submit.php` 返回「该域名不可发起支付」；支付→回调→到账闭环因此未经验证 | 用户（柠檬支付商户后台） | 把 `openbridger.com` 加入支付域名白名单，然后小额实测（¥10）验证回调与到账 |
| 4 | 整机重建未端到端演练                                                                                                                                                                                                                                   | Agent + 用户开临时新机       | rebuild-runbook 全流程走一遍                  |
| 5 | 异地备份依赖 Mac 开机                                                                                                                                                                                                                                | 可选                    | 加对象存储（COS/OSS 约 ¥1/月）双写                 |
| 6 | Redis 不异地、CF 配置靠文字清单                                                                                                                                                                                                                         | 已接受                   | dr-plan §8                              |

## 附录：命令速查

```bash
ssh ob-prod '/srv/openbridger/obctl diag'         # 出事先跑这条
ssh ob-prod '/srv/openbridger/obctl smoke'        # 处置后验证
ssh ob-prod 'OB_API_KEY=sk-... /srv/openbridger/obctl accept'   # 全量验收 20/20
ssh ob-prod '/srv/openbridger/obctl backup'       # 发版/危险操作前
ssh ob-prod '/srv/openbridger/obctl rollback'     # 8 秒回滚
ssh ob-prod '/srv/openbridger/obctl logs app 200' # 看应用日志
```
