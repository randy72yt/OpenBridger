---
name: ob-deploy
description: >-
  Release a new version of OpenBridger to production (ob-prod). Use when the
  user names a git ref / version and asks to ship, deploy, release, or 上线/发版.
  Covers the full loop: frozen backup precheck, off-box GHCR image build,
  obctl pull/deploy, verification, and rollback paths. Never deploy without an
  explicitly named ref from the user.
---

# OpenBridger 发版闭环

## 硬规则（先读）

1. 没有用户明确给出的 git ref + 发版指令 → 只做巡检/备份/文档，禁止 deploy。
2. 密钥永不进对话/git；生产 .env 不动。
3. 发版优先离机构建（GHCR），服务器本地编译仅作备选——2026-11-03 服务器编译曾把整机 OOM 拖死。

## 流程

1. **前置检查**：`ssh ob-prod '/srv/openbridger/obctl backup-check'`（备份必须新鲜）；`obctl status`（磁盘 <70%）。
2. **冻结备份**：`ssh ob-prod '/srv/openbridger/obctl backup'`。
3. **确认目标代码已推送**：本地 `~/code/OpenBridger` 提交并推送 main（工作流文件改动用 SSH 远端 `git push git@github.com:randy72yt/OpenBridger.git main`，OAuth 凭据无 workflow scope）。
4. **离机构建**：推 main 后 `.github/workflows/ghcr-release.yml` 自动构建（本地 `gh run list -R randy72yt/OpenBridger` / `gh run watch -R ... <id>` 看进度，约 5 分钟），产物 `ghcr.io/randy72yt/openbridger:<短sha>`。
5. **拉取**：`ssh ob-prod '/srv/openbridger/obctl pull <短sha>'`（GHCR 包需 public，404 时先让用户在 GitHub Packages 设 public 或服务器 docker login ghcr.io）。
6. **部署**：`ssh ob-prod '/srv/openbridger/obctl deploy <短sha>'`。L1 本地冒烟失败才自动回滚；L2（经 CF 公网自测）失败只是假失败告警，**不要**接受回滚。
7. **外网验证**：从本机 curl `https://openbridger.com/api/status`、`https://api.openbridger.com/api/status`、`https://docs.openbridger.com/` 均 200；`docker inspect repo-openbridger-1` 确认镜像 tag 与 healthy。
8. **回滚**（异常时）：`ssh ob-prod '/srv/openbridger/obctl rollback'`。

## 备选：服务器本地构建（仅 GHCR 不可用）

`obctl build <git-ref>` → `obctl build-status` 轮询至 BUILD OK → deploy。
若 SSH banner 超时/站点失联：多半是构建 OOM，停手，让用户在阿里云轻量控制台强制重启，恢复后重查 build-status。

## 完成后

汇报：结论 → 已执行 → 验证结果（三域名状态码）→ 下一步。必要时更新 `~/Downloads/OpenBridger-AI-交接包.md` 的镜像基线。
