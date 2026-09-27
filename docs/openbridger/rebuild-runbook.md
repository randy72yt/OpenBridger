# OpenBridger 完整重建手册（从零到生产可用）

版本：v1.0 · 2026-09-27 · 配套文档：`dr-plan.md`（灾备总体方案）
适用场景：服务器整机丢失后全新重建，也适用于迁移到新机。**每一步都可直接复制执行**。

---

## 0. 前置检查（重建前确认手头有什么）

| 需要的资产 | 在哪 | 没有怎么办 |
| --- | --- | --- |
| 数据备份 `latest-mysql.sql.gz` | Mac `~/backups/openbridger/` | 没有则数据丢失，只能空库重来 |
| 配置归档 `latest-config.tar.gz`（含 .env/nginx/crontab/ufw） | 同上 | 没有则需按 §6 清单逐项人工重建配置 |
| GitHub fork 推送权限（SSH key） | 本机 `~/.ssh/` | 用 GitHub 网页重新加 deploy key |
| Cloudflare 控制台 | 账号密码 + 2FA | 无法恢复则换域名 |
| 阿里云控制台 | 账号 | 开新机用 |
| 你的 API 令牌 `sk-a4d4...` | 你的密码管理器 | 恢复后重建令牌即可 |

**全程预计 1.5–2.5 小时**（构建占大头）。按顺序做，不要跳步。

---

## 1. 新开服务器（约 15 分钟）

1. 阿里云轻量应用服务器：**≥2C4G**、系统 **Ubuntu 24.04**、地域随意（建议同原地域）
2. 防火墙/安全组：放行 **22 / 80 / 443**
3. 把本机 SSH 公钥加进新机 `/root/.ssh/authorized_keys`，并在本机 `~/.ssh/config` 把 `ob-prod` 的 HostName 改成新 IP
4. 基础环境（SSH 上机后）：

```bash
apt update && apt upgrade -y
apt install -y docker.io docker-compose-v2 nginx certbot python3-certbot-nginx ufw git curl python3
systemctl enable --now docker nginx
ufw allow OpenSSH && ufw allow "Nginx Full" && ufw --force enable
mkdir -p /srv/openbridger/backups
```

---

## 2. 恢复配置文件（约 5 分钟）

在 **Mac** 上解包并上传（归档内路径已含目录结构）：

```bash
mkdir -p /tmp/ob-restore && tar xzf ~/backups/openbridger/latest-config.tar.gz -C /tmp/ob-restore
ls /tmp/ob-restore/srv/openbridger/   # 应看到 .env compose.infra.yml monitor.py alert-email.txt
ls /tmp/ob-restore/etc/nginx/         # 应看到 sites-available/ conf.d/ snippets/

# 应用配置（含全部运行密钥，权限必须 600）
scp /tmp/ob-restore/srv/openbridger/.env ob-prod:/srv/openbridger/.env
scp /tmp/ob-restore/srv/openbridger/compose.infra.yml ob-prod:/srv/openbridger/
scp /tmp/ob-restore/srv/openbridger/monitor.py /tmp/ob-restore/srv/openbridger/alert-email.txt ob-prod:/srv/openbridger/

# Nginx 配置（先放好，§5 签完证书再启用）
scp /tmp/ob-restore/etc/nginx/sites-available/openbridger ob-prod:/etc/nginx/sites-available/
scp /tmp/ob-restore/etc/nginx/conf.d/ob-ratelimit.conf ob-prod:/etc/nginx/conf.d/
scp /tmp/ob-restore/etc/nginx/snippets/ob-proxy.conf /tmp/ob-restore/etc/nginx/snippets/ob-block-scan.conf ob-prod:/etc/nginx/snippets/
```

服务器上收尾：

```bash
chmod 600 /srv/openbridger/.env
chmod +x /srv/openbridger/monitor.py
```

---

## 3. 拉代码与构建（约 30–60 分钟，全程最耗时）

### 3.1 拉代码（服务器上）

```bash
git clone git@github.com:randy72yt/OpenBridger.git /srv/openbridger/repo
cd /srv/openbridger/repo
cat /srv/openbridger/backups/config-*.repo-head 2>/dev/null || true   # Mac 归档里有 repo-head 记录
git checkout v0.1.0     # 或归档里记录的更新的 tag/SHA
```

如果新机没配 GitHub SSH：用 `https://github.com/randy72yt/OpenBridger.git` 只读 clone 也行（fork 是私有的话需要 token）。

### 3.2 前端构建（**在 Mac 上构建，不要在服务器上构建**）

> 原因：vite 构建吃内存，2C4G 小机会 OOM（生产事故实录）。

```bash
# Mac 上（本地仓库与服务器同一代码版本）
cd ~/code/OpenBridger
git fetch --tags && git checkout v0.1.0
cd web && NODE_OPTIONS= /Users/yuntao/.bun/bin/bun install
NODE_OPTIONS= /Users/yuntao/.bun/bin/bun run build
rsync -az --delete dist/ ob-prod:/srv/openbridger/repo/web/dist/
ssh ob-prod 'ls /srv/openbridger/repo/web/dist/index.html'   # 必须存在
```

### 3.3 后端镜像构建（服务器上）

> **2C4G 关键补丁**：Go 默认并行编译会把小机打崩（实录：sshd 都被压垮）。构建前打串行补丁，构建后还原。

```bash
cd /srv/openbridger/repo
cp Dockerfile.dev /tmp/Dockerfile.dev.orig
sed -i "s|^RUN go build|ENV GOFLAGS=-p=1 GOMAXPROCS=1\nRUN go build|" Dockerfile.dev
grep -A1 "GOFLAGS" Dockerfile.dev    # 确认补丁生效

OPENBRIDGER_IMAGE_TAG=$(git rev-parse --short HEAD) \
  docker compose -f compose.release.yml --env-file /srv/openbridger/.env build

cp /tmp/Dockerfile.dev.orig Dockerfile.dev   # 还原
docker images | grep openbridger             # 确认镜像生成
```

---

## 4. CI/CD 说明（重建时要不要跑 CI？）

**结论：重建场景不需要跑 CI，直接用已验证的 tag 即可。** 完整规则：

| 场景 | 要不要跑 CI | 怎么跑 |
| --- | --- | --- |
| **灾难重建/迁移（本手册）** | **不用** | 部署的是已打过 tag、CI 通过过的版本，无代码变更 |
| 改了代码、合并前 | **必须** | `gh workflow run "CI" --repo randy72yt/OpenBridger --ref <分支>` |
| 日常兜底 | 自动 | 每天 01:23（北京）定时跑 main |

fork 仓库的特殊性（已踩坑确认）：**push/PR 事件不触发 Actions**，只能手动 `gh workflow run` 或定时。workflow 文件（.github/workflows/）的推送必须用 **SSH**（`git push git@github.com:...`），OAuth token 缺 workflow scope 会被拒。

CI 三个 job：backend（Go 测试）、database matrix（MySQL 定价+安全用例）、frontend（vitest）。前端本地复现命令：`cd web && NODE_OPTIONS= node_modules/.bin/vitest run <文件>`（NODE_OPTIONS 必须清空，否则 worker 起不来）。

---

## 5. 启动服务与数据恢复（约 20 分钟）

### 5.1 启动基础设施（MySQL + Redis）

```bash
cd /srv/openbridger
docker compose -f compose.infra.yml --env-file .env up -d
docker ps    # 等 openbridger-mysql-1 显示 healthy（约 30–60 秒）
```

### 5.2 灌入数据

```bash
# Mac 上传备份
scp ~/backups/openbridger/latest-mysql.sql.gz ob-prod:/srv/openbridger/backups/restore.sql.gz

# 服务器上恢复
zcat /srv/openbridger/backups/restore.sql.gz | docker exec -i openbridger-mysql-1 \
  sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" openbridger'

# 抽查（与事故前行数对比；参考值：users=2, channels=1, model_price_policies=42）
docker exec -i openbridger-mysql-1 sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" -N openbridger' <<'SQL'
SELECT CONCAT('users=', COUNT(*)) FROM users;
SELECT CONCAT('options=', COUNT(*)) FROM options;
SELECT CONCAT('channels=', COUNT(*)) FROM channels;
SELECT CONCAT('tokens=', COUNT(*)) FROM tokens;
SELECT CONCAT('policies=', COUNT(*)) FROM model_price_policies;
SQL
```

### 5.3 启动应用

```bash
cd /srv/openbridger/repo
OPENBRIDGER_IMAGE_TAG=$(git rev-parse --short HEAD) \
  docker compose -f compose.release.yml --env-file /srv/openbridger/.env up -d
sleep 10
docker ps    # 三容器都应 healthy/running
curl -s http://127.0.0.1:3000/api/status | head -c 200   # 本地 200 才算应用活了
```

### 5.4 TLS 证书与 Nginx

```bash
# §2 已把站点配置放到 sites-available，先建软链但不急着重载
ln -sf /etc/nginx/sites-available/openbridger /etc/nginx/sites-enabled/openbridger

# 证书：certbot 会自动改写 nginx 配置挂上 ssl 段（DNS 此时必须先指向新机，见 §6 先做 6.1）
certbot --nginx -d openbridger.com -d www.openbridger.com -d docs.openbridger.com --agree-tos -m support@openbridger.com --redirect

nginx -t && systemctl reload nginx
systemctl enable certbot.timer   # 自动续期
```

### 5.5 恢复自动化（备份 + 监控 cron）

```bash
# Mac 上归档里有 crontab 内容
scp /tmp/ob-restore/tmp/ob-crontab.txt ob-prod:/tmp/
ssh ob-prod 'crontab /tmp/ob-crontab.txt && rm /tmp/ob-crontab.txt && crontab -l'
# 应看到：30 3 * * * backup.sh 和 */5 * * * * monitor.py
# backup.sh 本身不在归档里——从 repo 或本文 §附录 重新放置（见附录 A）
```

---

## 6. DNS 切换（Cloudflare，约 5 分钟 + 传播）

### 6.1 改记录（**要在 §5.4 签证书之前做**，certbot 要验证域名指向）

CF 控制台 → openbridger.com → DNS → Records：

| Type | Name | Content | Proxy |
| --- | --- | --- | --- |
| A | `@` | 新服务器 IP | Proxied（橙云） |
| A | `api` | 新服务器 IP | Proxied |
| A | `docs` | 新服务器 IP | Proxied |

以下记录**不用动**（与 IP 无关）：TXT SPF、TXT `_dmarc`、MX（route1/2/3.mx.cloudflare.net）。

### 6.2 确认 CF 侧其他配置仍在（无需重建，仅核对）

- 缓存规则：`/v1/*`、`/api/*` → Bypass cache
- Email Routing：`support@` → 你的邮箱（Enabled）
- Turnstile：widget 绑 `openbridger.com`（绑域名不绑 IP，换服务器不受影响；密钥存在 DB options 里，随数据恢复）

---

## 7. 启动后验证（**必须做，"启动成功" ≠ "恢复完成"**）

四层验证，全过才算恢复完成。**缺少这一步等于没做灾备**——配置错位、密钥失效、缓存规则丢失这类问题，进程活着也发现不了。

### L1 进程层（1 分钟）
```bash
docker ps --format "{{.Names}} {{.Status}}"     # 3 容器 healthy
curl -s http://127.0.0.1:3000/api/status | python3 -c "import json,sys; d=json.load(sys.stdin)['data']; print(d['version'], d['turnstile_check'], d['online_payment_enabled'] if 'online_payment_enabled' in d else '')"
```

### L2 边缘层（1 分钟，走 CF+Nginx+TLS 全链路）
```bash
curl -s -o /dev/null -w '%{http_code}\n' https://openbridger.com/api/status        # 200
curl -s -o /dev/null -w '%{http_code}\n' https://api.openbridger.com/api/status    # 200
curl -s -o /dev/null -w '%{http_code}\n' https://docs.openbridger.com/             # 200（docs 站需 §附录 B 恢复）
```

### L3 功能层（3 分钟，自动化验收脚本）
```bash
OB_BASE=https://openbridger.com OB_API_KEY=<你的令牌> bash /srv/openbridger/prod-acceptance.sh
# 必须 20 通过 / 0 失败（脚本在 repo docs/openbridger/scripts/，scp 到服务器跑）
```
覆盖：注册/登录安全、安全头、限流生效、缓存 bypass、法律接口、真实模型调用。

### L4 业务层（5 分钟，自动化测不了的）
1. **计费抽验**：真实调用一次 deepseek-v4-flash，查 `logs` 表 quota = `prompt×0.66 + completion×1.98`（3x 谷值；峰值时段 ×2 系数）
2. **邮件链路**：`python3 /srv/openbridger/monitor.py` 应输出 OK；手工触发一次找回密码邮件确认 SMTP 正常
3. **Turnstile**：浏览器开注册页，人机验证组件正常加载且能通过
4. **支付**：钱包页显示"在线充值"，创建一笔 10 元订单能拿到支付链接（不支付也行）
5. **备份闭环**：手动跑一次 `/srv/openbridger/backup.sh`，然后 Mac 上重跑异地拉取确认拿到新备份

### 验证 Checklist（打印打勾）

- [ ] L1 三容器 healthy、/api/status 版本正确
- [ ] L2 三个域名全 200
- [ ] L3 prod-acceptance 20/20
- [ ] L4 计费数值精确吻合
- [ ] L4 邮件收发正常
- [ ] L4 Turnstile 注册页可过
- [ ] L4 支付下单链接正常
- [ ] L4 备份+异地拉取闭环
- [ ] crontab 两条任务在列
- [ ] 监控 monitor.py 正常输出

---

## 附录 A：backup.sh 放置

备份脚本不属于 repo（含服务器特定路径），重建时从 Mac 归档所在会话或找 WorkBuddy 重新生成（最新版逻辑：mysqldump + Redis AOF + config 归档 + 日 7/周 4 轮转 + latest 指针）。恢复后**第一件事就是把它放回 `/srv/openbridger/backup.sh` 并 `chmod +x`**，否则当晚就断备份。

## 附录 B：docs 文档站恢复

`/srv/openbridger/docs-site/` 是静态文件（不在备份内）。从 repo 重新生成或直接 rsync：

```bash
# Mac 上（如果本地有副本）或重新从 repo 的 docs 源生成后：
rsync -az <docs-site目录>/ ob-prod:/srv/openbridger/docs-site/
```

## 附录 C：已知坑位（全部实录）

1. **2C4G 构建 OOM**：Go 并行编译打崩整机（sshd 无响应需控制台重启）→ 必须 GOFLAGS 串行补丁（§3.3）
2. **前端在服务器构建必 OOM** → Mac 构建 + rsync（§3.2）
3. **nested heredoc 远程执行**：`ssh 'sh -s' <<EOF` 内层 `$VAR` 会被吃掉 → 脚本一律 scp 上传后执行
4. **options 表直改有缓存延迟**：改 DB 后等约 1 分钟或走 API 写入（Turnstile 等）
5. **model_price_policies 为空会导致每小时定价 sync 全败**：恢复数据后确认 42 条策略在（§5.2 抽查已含）
6. **v4-pro/v4-flash 是峰谷表达式定价**：pricing publish 会覆盖成线性价，批量发布后必须单独 PATCH 回峰谷表达式（见 openbridger-ops-agent 技能 Part 2 定价章节）
