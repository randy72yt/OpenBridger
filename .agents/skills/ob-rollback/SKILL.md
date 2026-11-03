---
name: ob-rollback
description: >-
  Roll back OpenBridger production to a previous image or recover from a bad
  release / wrong auto-rollback. Use when the user asks to 回滚/切回上一版, or
  after a deploy failure. Includes the fast manual switch command that avoids
  rebuilding.
---

# OpenBridger 回滚

## 标准回滚

```bash
ssh ob-prod '/srv/openbridger/obctl rollback'            # 回到 .prev-tag 记录的上一版
ssh ob-prod '/srv/openbridger/obctl rollback <tag>'      # 回到指定镜像 tag
```

## 被误回滚后秒级切回（不重建）

deploy 的 L2 假失败曾触发误回滚（11-03 起已改为只告警）。若发现正常版本被回滚：

```bash
ssh ob-prod 'cd /srv/openbridger/repo && OPENBRIDGER_IMAGE_TAG=<sha> docker compose -f compose.release.yml --env-file /srv/openbridger/.env up -d'
```

## 验证

`docker inspect repo-openbridger-1 --format '{{.Config.Image}} {{.State.Status}}'` + 本机 curl 三域名 200。

## 纪律

- 回滚前先 `obctl rollback` 不带参数查看语义；目标 tag 必须已存在于 `docker images`（被清理则先 `obctl pull <sha>`）。
- 回滚也是变更：完成后外网验证并汇报版本。
