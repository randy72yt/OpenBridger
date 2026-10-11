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
# 计费阵地(GUARD):上游所有权的、我们计费/定价路径依赖的文件与目录。两类来源:
#   (a) import billingexpr 的上游 seam 文件(接线所在);
#   (b) 上游 pricing/billing 家族文件(我们未改,但定价控制面与其交互/依赖其语义,含对账用的 billing.go)。
# 同包编译级依赖(service/model/common 等)由闸门 go build 兜底,不在此列。
# 命中即最严口径:禁止整文件 checkout,逐块 diff 人工审查,PR 附结论(见 SKILL.md 阵地规则)。
GUARD_FILES=(
  pkg/billingexpr
  model/pricing.go
  model/option.go
  relaykit/dto/pricing.go
  relay/helper/price.go
  relay/relay_task.go
  relay/common/relay_info.go
  service/quota.go
  service/text_quota.go
  service/task_polling.go
  service/log_info_generate.go
  controller/pricing.go
  controller/model_pricing_config.go
  controller/billing.go
  controller/channel-test.go
  model/task.go
)
SECURITY_RE='security|cve|vulnerab|xss|csrf|inject|bypass|提权|越权|泄露'
# release notes 是行文 prose,宽词表会命中分区标题/功能名(如 "Security &"、host-injectable、bypass threshold);
# notes 扫描只用窄词表:真正的漏洞语言
SECURITY_NOTES_RE='cve|vulnerab|exploit|advisory|privilege escalation|安全漏洞|漏洞修复|提权|越权'
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

# 3) 有变化 → 拉取 tag 并出报告(网络抖动时若本地已有所需对象则降级继续)
if ! git fetch "$UPSTREAM_URL" --tags --quiet 2>/dev/null; then
  echo "WARN: git fetch 上游失败(网络?),尝试用本地已有对象出报告" >&2
fi
git cat-file -e "${baseline_point}^{commit}" 2>/dev/null || fail "本地缺少基线提交 $baseline_point"
git cat-file -e "${latest}^{commit}" 2>/dev/null || fail "本地缺少 tag $latest(fetch 失败且本地无此对象)"
range="$baseline_point..$latest"

echo "==== 上游有新 release: $baseline_release → $latest ===="
echo
RELEASES="$(git tag -l | grep -E "$TAG_RE" | sort -V \
           | awk -v b="$baseline_release" -v l="$latest" '$0==b{f=1;next} $0==l{print;exit} f{print}')"
echo "-- 区间内 release(整合时逐个读 release notes):"
for t in $RELEASES; do
  echo "  $t  $(git log -1 --format=%cd --date=short "$t" 2>/dev/null)"
done
echo
sec_hit=0
echo "-- 安全相关提交(取舍规则:无条件摘;test 基建类不计入,防误报):"
sec="$(git log --format='%h %s' "$range" | grep -iE "$SECURITY_RE" | grep -viE '^[0-9a-f]+ test|_test\.go' || true)"
if [ -n "$sec" ]; then echo "$sec"; sec_hit=1; else echo "  (无关键词命中)"; fi
if command -v gh >/dev/null 2>&1; then
  echo "  release notes 安全标记:"
  for t in $RELEASES; do
    body="$(gh api "repos/QuantumNous/new-api/releases/tags/$t" --jq .body 2>/dev/null || true)"
    m="$(printf '%s' "$body" | grep -iE "$SECURITY_NOTES_RE" | head -1 || true)"
    if [ -n "$m" ]; then echo "  🔴 $t: $(printf '%s' "$m" | cut -c1-90)"; sec_hit=1; fi
  done
  [ "$sec_hit" -eq 0 ] && echo "  (各 release notes 无安全关键词)"
else
  echo "  (无 gh,跳过 release notes 扫描——结论仅依据提交信息)"
fi
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
echo "-- 计费阵地(GUARD)命中(我们计费路径依赖的上游文件;命中即最严口径,逐块人工审查,禁整文件 checkout):"
hit=0
for f in "${GUARD_FILES[@]}"; do
  n="$(git log --oneline "$range" -- "$f" | wc -l | tr -d ' ')"
  if [ "$n" -gt 0 ]; then echo "  $f  (${n} commits)"; hit=1; fi
done
if [ "$hit" -eq 0 ]; then echo "  (阵地零命中)"; fi
echo
echo "-- 摘取候选规模参考:"
stat="$(git diff --stat "$range" -- relay/channel | tail -1 || true)"
echo "  relay/channel(新适配器): ${stat:-无变化}"
stat="$(git diff --stat "$range" -- relay | tail -1 || true)"
echo "  relay/ 总量: ${stat:-无变化}"
echo
if [ "$sec_hit" -eq 1 ]; then
  echo "结论: 🔴 存在安全相关变更——触发入口①,尽快按 SKILL.md 整合(至少摘安全部分)。"
  echo "NEXT: 按 docs/openbridger/skills/openbridger-upstream-sync/SKILL.md 流程整合(取舍规则见 ops-handbook §5.6)。"
else
  echo "结论: ⚪ 无安全命中——本次可整段跳过,在 state.md 记一行「检查-跳过」即可。"
  echo "落后不是负债(§5.6 定策:不追 latest)。整合只由四个入口触发:①安全(本行)②需求(要上某新模型/渠道)③事故(疑似上游已修的 bug)④迁移(上游计费实现满足三条件)。"
fi
exit 2
