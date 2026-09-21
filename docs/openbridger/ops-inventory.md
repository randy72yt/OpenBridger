# 生产环境台账

最后更新：2026-09-21

**本文件不记录公网 IP、密钥、密码、Token。** 主机连接信息保存在运维本机的 `~/.ssh/config`（Host 别名 `ob-prod`），应用运行时变量保存在服务器 `/srv/openbridger/.env`（权限 600）。这里只登记非敏感资产与「去哪儿找」的指针，供交接和回滚使用。

## 1. 资产登记

| 类别 | 项 | 值 | 备注 |
| --- | --- | --- | --- |
| 主机 | 提供商 | 阿里云轻量应用服务器 | S01 已建 |
| 主机 | 规格 / 地域 / 系统 | 待登记 | 建实例后补（建议 2C4G 起，Debian 12 或 Ubuntu 22.04） |
| 主机 | 部署目录 | `/srv/openbridger` | 内含 `repo/`（源码）、`.env`、备份目录 |
| 主机 | 开放端口 | 22（限来源 IP）、80、443 | 3000 仅映射宿主 `127.0.0.1:3000`，不对外 |
| 接入 | SSH 配置位置 | 运维本机 `~/.ssh/config`，Host `ob-prod` | 不在仓库 |
| 接入 | 连接命令 | `ssh ob-prod` | 需要 IP、端口、用户、私钥四项，由运维保管 |
| 域名 | 主站 / 控制台 | `openbridger.com` | Cloudflare，待解析 |
| 域名 | 文档站 | `docs.openbridger.com` | Cloudflare，待解析 |
| 数据库 | 选型 | 待定 | 本机容器 PG17/MySQL8.0 或阿里云 RDS；须与三库测试矩阵一致 |
| Redis | 选型 | 待定 | 本机容器或阿里云 Redis |
| 运行时变量 | 位置 | 服务器 `/srv/openbridger/.env`（600） | `SQL_DSN`、`REDIS_CONN_STRING`、`SESSION_SECRET`、`CRYPTO_SECRET`、`TRUSTED_PROXIES`、`OPENBRIDGER_IMAGE_TAG` |
| 本地开发变量 | 位置 | 仓库根 `.env`（已被 `.gitignore` 忽略） | 只服务本地，不参与部署 |
| 备份 | 位置 | `/srv/openbridger/backup` + 异地对象存储 | 数据卷 tar + 数据库 dump |
| 镜像 | 命名 | `openbridger:<git-short-sha>` | 不可变 tag，回滚用上一个 sha |

## 2. 发布记录

每次生产部署后追加一行：

| 日期 | commit sha | 镜像 tag | 变更摘要 | 执行人 | 结果 | 回滚目标 |
| --- | --- | --- | --- | --- | --- | --- |
| — | — | — | 尚未首次部署 | — | — | — |

## 3. 维护规则

1. IP、密钥、密码、API Key 一律不写进本仓库任何文件；需要交接时单独走受控渠道。
2. 服务器上的 `.env` 变更不通过仓库分发，直接在主机上改并留变更记录。
3. 每次发布必须能指名回滚目标（上一个镜像 tag），否则不允许发布。
4. 资产变化（换机、换库、换域名）先更新本表，再执行。
