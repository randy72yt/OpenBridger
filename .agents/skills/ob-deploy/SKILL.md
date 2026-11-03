---
name: ob-deploy
description: >-
  Release a new version of OpenBridger to production (ob-prod). Use when the
  user names a git ref / version and asks to ship, deploy, release, or 上线/发版.
  Covers the full loop: frozen backup precheck, Mac-side amd64 image build
  piped to the server, obctl deploy, verification, and rollback paths. Never
  deploy without an explicitly named ref from the user. CI runs are kept
  minimal by policy — do not push to main just to trigger builds.
---

# OpenBridger 发版闭环

## 硬规则（先读）

1. 没有用户明确给出的 git ref + 发版指令 → 只做巡检/备份/文档，禁止 deploy。
2. 密钥永不进对话/git；生产 .env 不动。
3. **尽量减少 CI 触发次数**（用户明确要求）：不在 Mac 上构建的情况下才用 CI（workflow_dispatch 手动触发；ghcr-release.yml 已无 push 触发）。不要为了构建而推送空提交。
4. 生产机（2C/1.6G）**永不本地编译**（2026-11-03 曾整机 OOM 拖死）。

## 流程（主路径：Mac 构建 → 管道传镜像）

1. **前置检查**：`ssh ob-prod '/srv/openbridger/obctl backup-check && /srv/openbridger/obctl status'`（备份新鲜、磁盘 <70%）。
2. **冻结备份**：`ssh ob-prod '/srv/openbridger/obctl backup'`。
3. **确认目标代码**：本地 `~/code/OpenBridger` 在目标 ref；前端已构建（`cd web && bun run build`，dist 与代码一致）。
4. **Mac 构建镜像并直传服务器**（无需镜像仓库；Mac Docker 为 linux/amd64）：
   ```bash
   SHA=$(git rev-parse --short HEAD)
   docker buildx build --platform linux/amd64 -t openbridger:$SHA -f Dockerfile.dev --load .
   docker save openbridger:$SHA | gzip | ssh ob-prod 'gunzip | docker load'
   ```
5. **部署**：`ssh ob-prod "/srv/openbridger/obctl deploy $SHA"`。L1 本地冒烟失败才自动回滚；L2（经 CF 公网自测）失败只是假失败告警，**不要**接受回滚。
6. **外网验证**：本机 curl 三域名（openbridger.com / api.openbridger.com / docs.openbridger.com）均 200；`docker inspect repo-openbridger-1` 确认镜像 tag 与 healthy。
7. **回滚**（异常时）：`ssh ob-prod '/srv/openbridger/obctl rollback'`。

## 备选：CI 构建（仅 Mac 不可用时，手动触发）

GitHub → Actions → Release image to GHCR → Run workflow（或 `gh workflow run ghcr-release.yml -R randy72yt/OpenBridger`）→ 产物 `ghcr.io/randy72yt/openbridger:<短sha>`（**私有包**，fork 仓库无法改 public）→ 服务器需先 `docker login ghcr.io`（read:packages token）→ `obctl pull <短sha>` → deploy。

## 完成后

汇报：结论 → 已执行 → 验证结果（三域名状态码）→ 下一步。必要时更新 `~/Downloads/OpenBridger-AI-交接包.md` 的镜像基线。
