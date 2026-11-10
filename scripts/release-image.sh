#!/usr/bin/env bash
# Mac 侧发版：构建 linux/amd64 镜像并 docker load 到 ob-prod。
# 不执行 obctl deploy，不打印密钥。
# 生产机 ob-prod（2 vCPU / ~1.6GB）禁止编译。2026-11-10 服务器上的 obctl build 打挂整机。
# 用法: scripts/release-image.sh [git-ref]
set -euo pipefail

REMOTE_HOST="${OB_PROD_HOST:-ob-prod}"
DISK_LIMIT=70

die() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
用法: scripts/release-image.sh [git-ref]

在运维 Mac（Intel x86_64，Docker Desktop）上构建生产镜像并传入 ob-prod。
不在生产机上编译，不执行 obctl deploy，不打印密钥。

  无参数     工作区必须干净，构建 HEAD
  git-ref    工作区必须干净，且 HEAD 就是该提交（本脚本不切换分支）

脚本依次做：
  1. cd web && NODE_OPTIONS= bun run build
  2. TAG=$(git rev-parse --short=9 HEAD)
     docker buildx build --platform linux/amd64 -f Dockerfile.dev -t openbridger:$TAG --load .
  3. 确认 ob-prod 磁盘 <70% 后：
     docker save openbridger:$TAG | gzip | ssh -o BatchMode=yes ob-prod 'gunzip | docker load'

前端打进镜像，运行容器只挂载 /data。不需要 rsync web/dist。
结束后打印 tag。下一步由人执行：
  ssh ob-prod '/srv/openbridger/obctl backup'
  ssh ob-prod "/srv/openbridger/obctl deploy $TAG"

可选环境变量 OB_PROD_HOST（默认 ob-prod）只改 ssh 主机名。
EOF
}

# 工作区干净，并且（若传入 ref）HEAD 就是该提交。stdout 打印至少 9 位的短 sha。
gate_git_ref() {
  local ref="${1:-}"
  git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "不在 git 工作区"
  if [ -n "$ref" ]; then
    git rev-parse --verify --quiet "${ref}^{commit}" >/dev/null \
      || die "无法解析 git ref: $ref"
    local want have
    want=$(git rev-parse "${ref}^{commit}") || die "无法解析 git ref: $ref"
    have=$(git rev-parse HEAD) || die "无法解析 HEAD"
    if [ "$want" != "$have" ]; then
      die "HEAD ($(git rev-parse --short=9 HEAD)) 不是 $ref ($(git rev-parse --short=9 "$want"))。先在干净工作区 checkout 该 ref。本脚本不切换分支，也不在 ob-prod 上编译。"
    fi
  fi
  if [ -n "$(git status --porcelain --untracked-files=normal)" ]; then
    if [ -n "$ref" ]; then
      die "已指定 ref，但工作区不干净。未提交的改动会被 docker build 打进镜像。请先提交或 stash。"
    fi
    die "工作区不干净。请提交或 stash，或 checkout 目标提交后执行: scripts/release-image.sh <git-ref>"
  fi
  local tag
  tag=$(git rev-parse --short=9 HEAD) || die "无法得到短 sha"
  case "$tag" in
    [0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]*) ;;
    *) die "无法得到至少 9 位的短 sha（得到 $tag）" ;;
  esac
  printf '%s\n' "$tag"
}

# 读 df -P 文本（stdin）。每个数据行的使用率必须 < limit。stdout 打印「挂载点 百分比」。
disk_within_limit() {
  local limit="$1"
  local line pct mount saw=0
  while IFS= read -r line || [ -n "$line" ]; do
    [ -n "$line" ] || continue
    case "$line" in
      Filesystem*) continue ;;
    esac
    pct=$(printf '%s\n' "$line" | awk '{print $5}')
    mount=$(printf '%s\n' "$line" | awk '{print $6}')
    pct=${pct%%%}
    case "$pct" in
      ''|*[!0-9]*)
        printf 'ERROR: 无法解析 df 行: %s\n' "$line" >&2
        return 1
        ;;
    esac
    if [ -z "$mount" ]; then
      printf 'ERROR: df 行缺少挂载点: %s\n' "$line" >&2
      return 1
    fi
    printf '%s %s\n' "$mount" "$pct"
    saw=1
    if [ "$pct" -ge "$limit" ]; then
      printf 'ERROR: %s 已用 %s%%，须 <%s%% 才允许 docker load\n' "$mount" "$pct" "$limit" >&2
      return 1
    fi
  done
  if [ "$saw" -ne 1 ]; then
    printf 'ERROR: df 没有数据行\n' >&2
    return 1
  fi
}

remote_ssh() {
  ssh -o BatchMode=yes -o ConnectTimeout=20 "$REMOTE_HOST" "$@"
}

main() {
  local ref=""
  if [ "$#" -gt 1 ]; then
    die "用法: scripts/release-image.sh [git-ref]    （--help 查看）"
  fi
  case "${1:-}" in
    -h|--help|help) usage; return 0 ;;
    "") ;;
    -*) die "未知参数: $1（scripts/release-image.sh --help）" ;;
    *) ref=$1 ;;
  esac

  local root
  root=$(git rev-parse --show-toplevel 2>/dev/null) || die "不在 git 仓库内。请在 OpenBridger 检出目录运行。"
  cd "$root" || die "无法进入 $root"

  local tag
  tag=$(gate_git_ref "$ref") || exit 1

  command -v bun >/dev/null 2>&1 || die "找不到 bun。在运维 Mac 上安装 bun 后重试。"
  command -v docker >/dev/null 2>&1 || die "找不到 docker。"
  command -v gzip >/dev/null 2>&1 || die "找不到 gzip。"
  command -v ssh >/dev/null 2>&1 || die "找不到 ssh。"

  docker info >/dev/null 2>&1 || die "Docker daemon 未运行。打开 Docker Desktop 后再试。"
  docker buildx version >/dev/null 2>&1 || die "docker buildx 不可用。"
  if ! docker buildx ls 2>/dev/null | grep -q 'linux/amd64'; then
    die "docker buildx 未提供 linux/amd64。生产镜像必须是 linux/amd64，已中止。"
  fi

  printf '检查 %s 磁盘（须 <%s%%，构建前）\n' "$REMOTE_HOST" "$DISK_LIMIT"
  local report
  report=$(remote_ssh 'df -P / /var/lib/docker 2>/dev/null || df -P /') \
    || die "无法读取 ${REMOTE_HOST} 磁盘。未开始构建，未执行 docker load。"
  printf '%s\n' "$report" | disk_within_limit "$DISK_LIMIT" \
    || die "服务器磁盘未低于 ${DISK_LIMIT}%。未开始构建，未执行 docker load。"

  printf '前端构建 %s/web（NODE_OPTIONS 已清空）\n' "$root"
  # NODE_OPTIONS='' is the verified `NODE_OPTIONS= bun run build`: clear WorkBuddy's shim.
  ( cd "$root/web" && NODE_OPTIONS='' bun run build )
  [ -f "$root/web/dist/index.html" ] || die "缺少 web/dist/index.html，前端构建没有产出入口。"

  printf '构建 openbridger:%s（platform linux/amd64）\n' "$tag"
  docker buildx build --platform linux/amd64 -f "$root/Dockerfile.dev" -t "openbridger:$tag" --load "$root"

  local platform
  platform=$(docker image inspect "openbridger:$tag" --format '{{.Os}}/{{.Architecture}}') \
    || die "无法检查本地镜像 openbridger:${tag} 的平台。未向服务器传输。"
  [ "$platform" = "linux/amd64" ] || die "本地镜像平台是 ${platform}，必须是 linux/amd64。未向服务器传输。"

  printf '再次检查 %s 磁盘（docker load 之前）\n' "$REMOTE_HOST"
  report=$(remote_ssh 'df -P / /var/lib/docker 2>/dev/null || df -P /') \
    || die "无法读取 ${REMOTE_HOST} 磁盘。未执行 docker load。"
  printf '%s\n' "$report" | disk_within_limit "$DISK_LIMIT" \
    || die "服务器磁盘未低于 ${DISK_LIMIT}%。未执行 docker load。"

  printf '传入 %s：docker load openbridger:%s（不执行 deploy）\n' "$REMOTE_HOST" "$tag"
  docker save "openbridger:$tag" | gzip | remote_ssh 'gunzip | docker load'

  local remote_platform
  remote_platform=$(remote_ssh "docker image inspect 'openbridger:${tag}' --format '{{.Os}}/{{.Architecture}}'") \
    || die "docker load 之后在 ${REMOTE_HOST} 上看不到 openbridger:${tag}"
  [ "$remote_platform" = "linux/amd64" ] || die "服务器镜像平台是 ${remote_platform}，必须是 linux/amd64。"

  cat <<EOF
已加载 openbridger:${tag}（linux/amd64）
前端在镜像内，运行容器只挂载 /data。不需要 rsync web/dist。
本脚本没有执行 deploy，也没有打印密钥。下一步：
  ssh ${REMOTE_HOST} '/srv/openbridger/obctl backup'
  ssh ${REMOTE_HOST} '/srv/openbridger/obctl deploy ${tag}'
EOF
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  main "$@"
fi
