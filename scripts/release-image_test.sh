#!/usr/bin/env bash
# 回归：obctl build 默认拒绝；release-image.sh 的 git/磁盘闸门。不连生产，不构建镜像。
set -euo pipefail

ROOT=$(cd "$(dirname "$0")/.." && pwd)
OBCTL=$ROOT/docs/openbridger/scripts/obctl.sh
REL=$ROOT/scripts/release-image.sh

fail() {
  printf 'FAIL: %s\n' "$*" >&2
  exit 1
}

assert_grep() {
  local haystack="$1" needle="$2"
  printf '%s\n' "$haystack" | grep -q -- "$needle" || fail "缺少「$needle」。输出: $haystack"
}

assert_absent() {
  local haystack="$1" needle="$2"
  if printf '%s\n' "$haystack" | grep -q -- "$needle"; then
    fail "不应出现「$needle」。输出: $haystack"
  fi
}

out=$(bash "$OBCTL" build main 2>&1) && fail "obctl build 默认应非 0"
assert_grep "$out" "生产机禁止编译"
assert_grep "$out" "scripts/release-image.sh"
assert_grep "$out" "OBCTL_ALLOW_SERVER_BUILD=YES"
assert_absent "$out" "build pid="

out=$(OBCTL_ALLOW_SERVER_BUILD=yes bash "$OBCTL" build main 2>&1) && fail "小写 yes 仍应拒绝"
assert_grep "$out" "生产机禁止编译"
assert_absent "$out" "build pid="

out=$(OBCTL_ALLOW_SERVER_BUILD=1 bash "$OBCTL" build main 2>&1) && fail "OBCTL_ALLOW_SERVER_BUILD=1 仍应拒绝"
assert_grep "$out" "生产机禁止编译"
assert_absent "$out" "build pid="

out=$(OBCTL_ALLOW_SERVER_BUILD=YES bash "$OBCTL" build 2>&1) && fail "紧急开口无参数应走用法错误"
assert_grep "$out" "用法: obctl build"
assert_grep "$out" "EMERGENCY"
assert_absent "$out" "生产机禁止编译"
assert_absent "$out" "build pid="

out=$(bash "$OBCTL" help 2>&1)
assert_grep "$out" "OBCTL_ALLOW_SERVER_BUILD=YES"
assert_grep "$out" "scripts/release-image.sh"
assert_grep "$out" "build-status"
assert_grep "$out" "不会开始构建"

out=$(bash "$OBCTL" build-status 2>&1 || true)
assert_absent "$out" "生产机禁止编译"
assert_absent "$out" "build pid="
assert_absent "$out" "docker build"
assert_absent "$out" "go build"

out=$(bash "$REL" --help 2>&1)
assert_grep "$out" "linux/amd64"
assert_grep "$out" "70%"
assert_grep "$out" "不执行 obctl deploy"
assert_absent "$out" "SESSION_SECRET"
assert_absent "$out" "SQL_DSN"

out=$(bash "$REL" a b 2>&1) && fail "多余参数应拒绝"
assert_grep "$out" "用法:"

out=$(bash "$REL" --nope 2>&1) && fail "未知参数应拒绝"
assert_grep "$out" "未知参数"

# Path is absolute so the test can be run from any cwd. shellcheck cannot follow it.
# shellcheck disable=SC1090,SC1091
source "$REL"

sample=$(printf '%s\n' \
  'Filesystem 1024-blocks Used Available Capacity Mounted on' \
  '/dev/vda1 100 69 31 69% /' \
  '/dev/vdb 100 10 90 10% /var/lib/docker')
disk_out=$(printf '%s\n' "$sample" | disk_within_limit 70)
assert_grep "$disk_out" "/ 69"
assert_grep "$disk_out" "/var/lib/docker 10"

over=$(printf '%s\n' \
  'Filesystem 1024-blocks Used Available Capacity Mounted on' \
  '/dev/vda1 100 70 30 70% /')
err=$(printf '%s\n' "$over" | disk_within_limit 70 2>&1) && fail "70% 应拒绝"
assert_grep "$err" "70%"

err=$(printf '%s\n' 'Filesystem 1024-blocks Used Available Capacity Mounted on' | disk_within_limit 70 2>&1) && fail "空 df 应拒绝"
assert_grep "$err" "没有数据行"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
git -C "$tmp" init -q -b main
git -C "$tmp" config user.email test@example.com
git -C "$tmp" config user.name test
printf 'a\n' > "$tmp/file"
git -C "$tmp" add file
git -C "$tmp" commit -q -m init
first=$(git -C "$tmp" rev-parse HEAD)
printf 'b\n' >> "$tmp/file"
git -C "$tmp" add file
git -C "$tmp" commit -q -m second

tag=$(cd "$tmp" && gate_git_ref)
case "$tag" in
  [0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f][0-9a-f]*) ;;
  *) fail "干净仓库的 tag 非法: $tag" ;;
esac
tag2=$(cd "$tmp" && gate_git_ref HEAD)
[ "$tag" = "$tag2" ] || fail "显式 HEAD 应得到同一 tag"

err=$(cd "$tmp" && gate_git_ref "$first" 2>&1) && fail "HEAD 与 ref 不一致应拒绝"
assert_grep "$err" "不是"

printf 'dirty\n' >> "$tmp/file"
err=$(cd "$tmp" && gate_git_ref 2>&1) && fail "脏工作区无 ref 应拒绝"
assert_grep "$err" "不干净"
err=$(cd "$tmp" && gate_git_ref HEAD 2>&1) && fail "脏工作区加 ref 应拒绝"
assert_grep "$err" "不干净"

remote_calls=$(grep -n 'remote_ssh' "$REL")
assert_grep "$remote_calls" "docker load"
assert_absent "$remote_calls" "deploy"
assert_absent "$remote_calls" "compose"

printf 'ok\n'
