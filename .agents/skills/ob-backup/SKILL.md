---
name: ob-backup
description: >-
  Run and verify OpenBridger backups, including off-site copy to the Mac. Use
  when the user asks to 备份/异地备份/检查备份, before any risky operation
  (restore, migration, manual SQL writes), or when obctl deploy precheck
  reports a stale backup.
---

# OpenBridger 备份闭环

## 机上备份 + 新鲜度

```bash
ssh ob-prod '/srv/openbridger/obctl backup'          # 产出 latest-mysql.sql.gz / latest-config.tar.gz
ssh ob-prod '/srv/openbridger/obctl backup-check'    # 新鲜度检查（约 26h 窗口）
```

## 异地（Mac）

```bash
rsync -aL ob-prod:/srv/openbridger/backups/latest-mysql.sql.gz ~/backups/openbridger/
rsync -aL ob-prod:/srv/openbridger/backups/latest-config.tar.gz ~/backups/openbridger/
```

可另存日期戳副本，本地保留约 14 天。

## 纪律

- 发版（R3）、手工改库、恢复演练前必须有新鲜冻结备份。
- 备份文件不含密钥讨论内容；不要把 dump 内容贴进对话。
