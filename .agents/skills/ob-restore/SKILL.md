---
name: ob-restore
description: >-
  Restore OpenBridger production database/config from a backup dump. DANGEROUS,
  destructive operation. Use only when the user explicitly asks to 恢复/restore/
  回档 with a clear reason. Always preserve the current state (fresh backup)
  before overwriting, and require explicit user confirmation of the target
  dump.
---

# OpenBridger 恢复（危险操作）

## 确认门（缺一不可）

1. 用户明确要求恢复，并知道会覆盖现库。
2. 已定位目标 dump 文件名与时间。
3. 已做「保留现场」备份（恢复前先 `obctl backup`）。

## 流程

```bash
ssh ob-prod '/srv/openbridger/obctl backup'                      # 1. 保留现场
OBCTL_CONFIRM=YES ssh ob-prod '/srv/openbridger/obctl restore <dump.sql.gz>'   # 2. 灌库
ssh ob-prod '/srv/openbridger/obctl smoke'                       # 3. 冒烟（L2 假失败注意）
```

## 验证

- L1 本机 `/api/status` 200；关键数据抽查（用户数、余额、最近日志）。
- 外网三域名 200。
- L2 smoke 失败大概率是 Cloudflare 403 假失败，用外网结果判断。

## 纪律

- 恢复失败且应用停止时：检查 dump 完整性后重试，或换更早备份；不要带着空库对外服务。
- 完整灾备/重建流程见仓库 `docs/openbridger/dr-plan.md` 与 `rebuild-runbook.md`。
