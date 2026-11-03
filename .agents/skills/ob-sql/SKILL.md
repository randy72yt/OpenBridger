---
name: ob-sql
description: >-
  Run SQL against OpenBridger production MySQL safely. Use when the user asks
  to 查数据库/改配置/写库/看表, or any task needing direct prod DB access.
  Enforces the UTF-8 file workflow (inline Chinese via SSH gets mangled),
  read-only by default, and mandatory read-back verification for writes.
---

# OpenBridger 生产 SQL

## 只读查询

```bash
printf 'SELECT ...\n' | ssh ob-prod '/srv/openbridger/obctl sql -'
```

## 写入（必须走 UTF-8 文件，禁止内联中文）

直接在 SSH 命令里内联中文会乱码（mysql 客户端连接字符集问题，2026-11-02 套餐标题曾中招）：

```bash
# 1. 本地写 UTF-8 SQL 文件
scp /tmp/change.sql ob-prod:/tmp/change.sql
# 2. 以 utf8mb4 连接执行
ssh ob-prod 'docker exec -i openbridger-mysql-1 mysql -uroot -p"$MYSQL_ROOT_PASSWORD" --default-character-set=utf8mb4 openbridger < /tmp/change.sql'
# 3. 回读核对（必须）
```

## 纪律

- 写库前确认是否需要先备份（见 ob-backup skill）；大范围 UPDATE/DELETE 先 SELECT 计数。
- options 表键如 `Price`、`USDExchangeRate`、`PayMethods` 改动立即生效面广，改前和用户确认。
- 表名/列名反引号：PostgreSQL 才用 `"col"`，这里是 MySQL，用 `` `col` ``。
- 不把密钥类行（EpayKey 等）的值读进对话。
