# 事故：在 1.6G 生产机上跑镜像构建，打崩 sshd 与站点

日期 / 定级：2026-11-04（本地时间约 20:2x–21:57）/ S1（生产不可用）/ 状态：已恢复，流程已改

## 影响面

- 生产站 `openbridger.com`、直连 `47.80.28.213:80/443` 约 20 分钟无响应（`/api/status` 返回 000）。
- 22 端口 TCP 可建连但 **sshd 发不出 banner**（`Connection timed out during banner exchange`），无法登录处置。
- 数据面未受损：`compose up` 从未执行，容器一直跑旧镜像 `openbridger:805c9226a`，**没有半截部署**；MySQL/Redis 随实例重启后正常。
- 试运营期外部流量近零，无计费差错。

## 时间线

- 19:38 部署 `openbridger:805c9226a`（校准日志分页修复），容器健康。
- 19:5x 手动触发 `pricing_calibration` 任务，任务以 `failed / task lease expired` 收尾（当日已定位为 handler 未调 `FinishSystemTask`，属另一缺陷，已在 `e503db5c6` 修）。
- 20:0x 提交 `e503db5c6`（按模型取校准窗口 + 任务收尾）。准备部署。
- 20:1x 发现服务器 `repo` 停在 `54c7d555e` 且工作区有历史补丁残留 → `git reset --hard origin/main`；发现 `web/dist` 未纳入 git 且为旧产物 → 从本机 rsync 覆盖。
- 20:2x 以 `obctl build main` 在**服务器上**触发镜像构建；先前一次过期的构建仍在，出现**两个构建并发**。
- 20:3x 手工 `kill` 构建进程（含共享 buildkit executor），两个构建同时中断；随后重新触发第三次构建。
- 20:4x `go build` 编译 `modernc.org/sqlite`，RSS 约 594MB，系统已用 1497/1613MB。
- 20:4x–21:57 sshd 失联、80/443 超时。本机反复 ssh 重试十余次（加重负载）。
- 21:57 用户在阿里云控制台重启实例，服务恢复（`up` 时间与容器 `Up 57 minutes` 吻合）。

## 根因

1. **在 2 vCPU / 1.6GB 生产机上编译 Go**。`go build` 单进程 RSS 约 600MB，与 MySQL/Redis/应用容器叠加后打满内存，sshd 无法 fork 出会话，Web 层一起失联。
2. **`obctl build` 每轮成功后 `docker builder prune -f`**，模块与编译缓存全丢。`modernc.org/sqlite` 每次从头重编，在 2C 机器上单轮 30 分钟不结束，把「一次构建」放大成「长期高内存占用」。
3. **并发构建**：上一次构建没结束就再起一个，两个 buildkit 争内存；手工 kill 又同时废掉两个，形成"永远构建不完"的循环。
4. **流程执行问题**：`openbridger-release` 技能已写明「不在服务器上构建前端（1.6G 内存会 OOM，历史上有过 sshd 假死）」、`release-traps.md` 已写明「构建卡住 + ssh 无 banner = 内存耗尽打崩 sshd，用控制台重启，不要重复 ssh」，但本次未加载技能、未跑 R2 预检（`MemAvailable ≥ 300MB` 闸门）、未要显式 git ref、自行选了 `main`、绕过 `build-image.sh` 手工 `obctl build`。

## 处置

- 未做任何"半截"操作：构建失败即停，未执行 `compose up`，线上保持 `805c9226a`。
- 用户重启实例后：站点 200、三容器 `healthy`、`PRECHECK PASS`（备份 11h、磁盘 35%、MemAvailable 955MB）。
- 实例已加 2GB swap（`Swap: 2047MB`），缓解但**不替代**下面的流程修改。

## 防再发（已完成）

1. **镜像改在本机构建**：`openbridger-release` R4b 重写为「Mac `docker build --platform linux/amd64` → `docker save | gzip | ssh | docker load`」，脚本 `scripts/build-image.sh` 同步重写；服务器只允许 `docker load` 与 `obctl deploy/rollback/smoke`。
2. **硬规则进技能**：新增「服务器永不编译（禁止 `obctl build`/`go build`/`docker build`/`compose build`）」与「同一时刻只允许一个构建（`/tmp/openbridger-image-build.lock`）」，`IMAGE_BUILD_BUSY` 不允许换终端重跑。
3. **预检闸门语义更新**：服务器不再承担编译，`MemAvailable ≥ 300MB` 改为保证容器切换余量；新增传输前服务器 docker 盘 ≥ 2GB 检查。
4. **前端新鲜度判据修正**：改按「`web/` 下是否有文件比 `web/dist` 新」判断，不再用提交时间（先构建后提交会误杀）。
5. **传输后校验镜像 ID**：两端 `docker image inspect --format '{{.Id}}'` 必须一致，不以 `Loaded image` 字样当成功。
6. 本事故记录写入 `docs/openbridger/incidents/`，`release-traps.md` 增加「历史：服务器构建的两个致命坑（已禁用该路径）」。

## 遗留

- [ ] 若要恢复"服务器自助构建"（例如脱离本机发版），必须先换 4GB 以上机型，并在 compose 构建里加 `--mount=type=cache` 与 `GOFLAGS=-p=1`；当前结论是**不做**。
- [ ] 本机构建首次需下 Go 模块，依赖本机代理；CI 化时在构建机上常驻模块缓存。

## 后续（2026-11-10）

上面「防再发」写了服务器不再编译，但 `obctl build` 仍会在 ob-prod 上启动 `docker compose build`。当日再次执行把服务器打挂，只能强制重启。仓库里的 `obctl build` 现已默认拒绝（`OBCTL_ALLOW_SERVER_BUILD=YES` 仅 emergency-only），Mac 侧入口是 `scripts/release-image.sh`。前端在镜像内，运行容器只挂载 `/data`，rsync `web/dist` 不改变正在跑的镜像。
