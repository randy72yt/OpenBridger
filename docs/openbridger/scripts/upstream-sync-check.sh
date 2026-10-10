#!/usr/bin/env bash
# upstream-sync-check.sh — 上游(new-api)release 跟踪检查器
#
# 用法:  docs/openbridger/scripts/upstream-sync-check.sh
#   exit 0  上游最新 release 与基线一致,本次无需整合(月检直接结束)
#   exit 2  上游有新 release,输出整合报告,按 SKILL.md 流程处理
#   exit 1  环境错误(网络/gh/状态文件)
#
# 状态文件: skills/openbridger-upstream-sync/state.md(本脚本只读;整合完成后由执行者更新)
# 策略依据: ops-handbook.md §5.6 —— 不追 upstream/main,按 release 粒度选择性同步

set -euo pipefail

UPSTREAM_URL="https://github.com/QuantumNous/new-api.git"
STATE_FILE="$(cd "$(dirname "$0")/.." && pwd)/skills/openbridger-upstream-sync/state.md"

# 热区:我们对上游文件的就地修改面(与 ops-handbook §5.6 登记一致;新增就地修改须同步此清单)
HOT_FILES=(
  common/constants.go common/email.go common/env.go common/init.go
  controller/audit.go controller/topup.go controller/user.go controller/payment_webhook_availability.go
  model/main.go model/model_pricing_config.go model/system_task.go
  router/api-router.go router/web-router.go
  setting/operation_setting/general_setting.go
  controller/security_account_test.go controller/security_enrollment_test.go
  service/task_billing_test.go router/plugin_router_test.go
)
SECURITY_RE='security|cve|vulnerab|xss|csrf|inject|bypass|提权|越权|泄露'
TAG_RE='^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$'

fail() { echo "ERROR: $*" >&2; exit 1; }

[ -f "$STATE_FILE" ] || fail "状态文件不存在: $STATE_FILE"
baseline_release="$(awk -F': *' '/^baseline_release:/{print $2}' "$STATE_FILE" | tr -d '[:space:]')"
baseline_point="$(awk -F': *' '/^baseline_point:/{print $2}' "$STATE_FILE" | tr -d '[:space:]')"
[ -n "$baseline_release" ] && [ -n "$baseline_point" ] || fail "state.md 缺 baseline_release / baseline_point 字段"

# 1) 上游最新 release(gh 优先,失败退回 ls-remote + 版本序)
latest=""
if command -v gh >/dev/null 2>&1; then
  latest="$(gh api repos/QuantumNous/new-api/releases/latest --jq .tag_name 2>/dev/null || true)"
fi
if [ -z "$latest" ]; then
  latest="$(git ls-remote --tags "$UPSTREAM_URL" 2>/dev/null \
    | awk -F'refs/tags/' '$2 ~ /^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$/ {print $2}' \
    | sort -V | tail -1)"
fi
[ -n "$latest" ] || fail "无法获取上游最新 release(检查网络或 gh 登录)"

# 2) 无变化 → 跳过
if [ "$latest" = "$baseline_release" ]; then
  echo "SKIP: 上游最新 release 仍为 $baseline_release,基线未变,本次无需整合。"
  exit 0
fi

# 3) 有变化 → 拉取 tag 并出报告
git fetch "$UPSTREAM_URL" --tags --quiet 2>/dev/null || fail "git fetch 上游失败"
git cat-file -e "${baseline_point}^{commit}" 2>/dev/null || fail "本地缺少基线提交 $baseline_point"
git cat-file -e "${latest}^{commit}" 2>/dev/null || fail "本地缺少 tag $latest(fetch 后仍无,确认 tag 名)"
range="$baseline_point..$latest"

echo "==== 上游有新 release: $baseline_release → $latest ===="
echo
echo "-- 区间内 release(整合时逐个读 release notes):"
for t in $(git tag -l | grep -E "$TAG_RE" | sort -V \
           | awk -v b="$baseline_release" -v l="$latest" '$0==b{f=1;next} $0==l{print;exit} f{print}'); do
  echo "  $t  $(git log -1 --format=%cd --date=short "$t" 2>/dev/null)"
done
echo
echo "-- 安全相关提交(取舍规则:无条件摘):"
sec="$(git log --format='  %h %s' "$range" | grep -iE "$SECURITY_RE" || true)"
if [ -n "$sec" ]; then echo "$sec"; else echo "  (无关键词命中)"; fi
echo
echo "-- 热区文件命中(我们就地修改过的上游文件,整合需人工合并):"
hit=0
for f in "${HOT_FILES[@]}"; do
  n="$(git log --oneline "$range" -- "$f" | wc -l | tr -d ' ')"
  if [ "$n" -gt 0 ]; then echo "  $f  (${n} commits)"; hit=1; fi
done
if [ "$hit" -eq 0 ]; then echo "  (热区零命中——本次大概率纯附加摘取,冲突面小)"; fi
stat="$(git diff --stat "$range" -- "${HOT_FILES[@]}" | tail -1 || true)"
echo "  热区差异量: ${stat:-0 files changed}"
echo
echo "-- 摘取候选规模参考:"
stat="$(git diff --stat "$range" -- relay/channel | tail -1 || true)"
echo "  relay/channel(新适配器): ${stat:-无变化}"
stat="$(git diff --stat "$range" -- relay | tail -1 || true)"
echo "  relay/ 总量: ${stat:-无变化}"
echo
echo "NEXT: 按 docs/openbridger/skills/openbridger-upstream-sync/SKILL.md 流程整合(取舍规则见 ops-handbook §5.6)。"
exit 2
