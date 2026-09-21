# 生产环境台账

最后更新：2026-09-21

**本文件不记录公网 IP、密钥、密码、Token。** 主机连接信息保存在运维本机的 `~/.ssh/config`（Host 别名 `ob-prod`），应用运行时变量保存在服务器 `/srv/openbridger/.env`（权限 600）。这里只登记非敏感资产与「去哪儿找」的指针，供交接和回滚使用。

## 1. 资产登记

| 类别 | 项 | 值 | 备注 |
| --- | --- | --- | --- |
| 主机 | 提供商 | 阿里云轻量应用服务器 | S01 已建 |
| 主机 | 规格 / 地域 / 系统 | 2 vCPU / 1.6Gi 内存 / 系统盘 40G（可用 34G）；Ubuntu 24.04.2 LTS | 2026-09-21 实测；内存偏紧，已启用 2G swap（swappiness=10） |
| 主机 | 重要约束 | **不在服务器上编译**：内存不足以支撑 Go 与前端构建 | 镜像在本机或 CI 构建后推送（见下方「构建策略」） |
| 主机 | 部署目录 | `/srv/openbridger` | 内含 `repo/`（源码）、`.env`、备份目录 |
| 主机 | 开放端口 | 22（限来源 IP）、80、443 | 3000 仅映射宿主 `127.0.0.1:3000`，不对外 |
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
| Nginx | 0.0.0.0:80 / :443 | 公网（经 ufw 放行） |
| OpenBridger 应用容器 | `127.0.0.1:3000` | 仅宿主，不对外 |
| MySQL 容器 | `172.17.0.1:3306` | 仅本机 docker0 网关，公网不可达 |
| Redis 容器 | `172.17.0.1:6379` | 同上 |

数据库与 Redis **不能**绑 `127.0.0.1`——应用容器经 docker0 网关访问宿主，绑 127.0.0.1 会 `connection refused`（2026-09-21 实际踩到）。

## 3. 已完成的安全开关（2026-09-21 生产实测）

| 项 | 值 | 验证方式 |
| --- | --- | --- |
| 公开注册 | `false` | `/api/status` 的 `register_enabled=false`（写入 `options` 表 `RegisterEnabled`） |
| 密码注册 | `false` | `password_register_enabled=false`（`PasswordRegisterEnabled`） |
| 在线支付 | 关闭 | 容器 env `OPENBRIDGER_ONLINE_PAYMENT_ENABLED=false`；所有支付接口挂 `middleware.RequireOnlinePayment()` |
| Cookie | `SESSION_COOKIE_SECURE=true`、`SESSION_COOKIE_TRUSTED_URL=https://openbridger.com` | `docker inspect` 已核对 |
| 可信代理 | `TRUSTED_PROXIES=172.17.0.1` | 同上 |
| CF 缓存 | `/api/status` 返回 `cf-cache-status: DYNAMIC` | 暂未被缓存，仍建议显式加 Bypass 规则 |

未完成、需运维在后台/控制台处理：管理员 MFA/Passkey、SMTP 与 Turnstile、CF SSL 模式 Full(strict) 与 Bypass 缓存规则。

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

## 6. 维护规则

1. IP、密钥、密码、API Key 一律不写进本仓库任何文件；需要交接时单独走受控渠道。
2. 服务器上的 `.env` 变更不通过仓库分发，直接在主机上改并留变更记录。
3. 每次发布必须能指名回滚目标（上一个镜像 tag），否则不允许发布。
4. 资产变化（换机、换库、换域名）先更新本表，再执行。
