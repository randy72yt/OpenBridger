---
name: ob-inspect
description: >-
  Inspect OpenBridger production health (ob-prod). Use when the user asks to
  巡检/检查线上/看看服务器, or when the site seems down, slow, or after any
  incident. Covers obctl status/diag/smoke, known false-failure traps
  (Cloudflare 403 on self-tests), container restart, logs, and common
  failure triage including the server-wedged (OOM) signature.
---

# OpenBridger 线上巡检

## 巡检命令

```bash
ssh ob-prod '/srv/openbridger/obctl status'        # 容器/版本/内存/磁盘
ssh ob-prod '/srv/openbridger/obctl diag'           # 深度诊断（可加 backup-check）
```

## 验证口径（重要）

- **本机 L1**：`curl http://127.0.0.1:3000/api/status`（服务器上）。
- **外网验证必须从非源站发起**（本机 Mac curl 三域名）：源站自己经 Cloudflare 访问公网域名常被 Bot Fight 拦 **403**，`obctl smoke` 的 L2 因此会假失败——不是故障。

## 常见处置

| 症状 | 动作 |
|---|---|
| 应用无响应（L1 挂） | `obctl restart app`，再 L1 |
| 看日志 | `obctl logs app 200`（或 mysql/redis/nginx） |
| ping 通但 SSH banner 超时、全站 000 | **整机卡死**（2C/1.6G，曾因构建 OOM + 换页风暴）。让用户在阿里云轻量控制台强制重启；恢复后查 `sudo journalctl -b -1 | grep -i "memory pressure"` 定因 |
| 磁盘涨 | `docker system df`；`docker builder prune -f`；旧镜像只留 3 个 |
| 拨测 401 | 先查 channel token，勿盲目轮换 |

## 完成后

简短汇报：容器状态、版本、内存/磁盘、三域名外网状态码；异常项给出下一步建议。
