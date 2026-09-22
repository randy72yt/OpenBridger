#!/usr/bin/env bash
# OpenBridger 生产验收回归脚本（无上游依赖部分）
#
# 用途：每次生产发版后回归核心安全与可用性边界。脚本不含任何秘密。
#
# 用法：
#   export OB_BASE=https://openbridger.com          # 默认就是这个
#   export OB_API_KEY=sk-xxxx                        # 可选；给了才跑 group C
#   ./docs/openbridger/scripts/prod-acceptance.sh
#
# 退出码：0 = 无 FAIL；1 = 存在 FAIL

set -uo pipefail
BASE="${OB_BASE:-https://openbridger.com}"
API_KEY="${OB_API_KEY:-}"
UA="OpenBridger-Acceptance/1.0"

PASS=0; FAIL=0; NA=0; INFO=0

# 颜色（非交互时自动降级）
if [ -t 1 ]; then
  G=$'\033[32m'; R=$'\033[31m'; Y=$'\033[33m'; N=$'\033[0m'
else
  G=""; R=""; Y=""; N=""
fi

# check <编号> <描述> <期望码,支持 | 多选> <请求...>
check() {
  local id="$1" desc="$2" expect="$3"; shift 3
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" -A "$UA" --max-time 25 "$@" 2>/dev/null)
  if [ -z "$code" ]; then
    printf "  ${R}FAIL${N}  %-6s %-42s 无响应\n" "$id" "$desc"; FAIL=$((FAIL+1)); return
  fi
  local ok=0
  IFS='|' read -ra opts <<< "$expect"
  for o in "${opts[@]}"; do [ "$code" = "$o" ] && ok=1; done
  if [ "$ok" = 1 ]; then
    printf "  ${G}PASS${N}  %-6s %-42s %s\n" "$id" "$desc" "$code"; PASS=$((PASS+1))
  else
    printf "  ${R}FAIL${N}  %-6s %-42s 期望 %s 实得 %s\n" "$id" "$desc" "$expect" "$code"; FAIL=$((FAIL+1))
  fi
}

# json_field <url> <python 表达式，作用于 data>
# 信息项：取值仅供参考，不计通过 / 失败。第 5 参数为 info 时启用。
json_check() {
  local id="$1" desc="$2" expr="$3" url="$4" expected="${5:-}"
  local body
  body=$(curl -s -A "$UA" --max-time 25 "$url" 2>/dev/null)
  local val
  val=$(echo "$body" | python3 -c "
import json,sys
try:
    d=json.load(sys.stdin)
    print($expr)
except Exception as e:
    print('PARSE_ERR')
" 2>/dev/null)
  if [ "$expected" = "info" ]; then
    printf "  ${Y}INFO${N}  %-6s %-42s %s\n" "$id" "$desc" "$val"; INFO=$((INFO+1)); return
  fi
  case "$val" in
    True)  printf "  ${G}PASS${N}  %-6s %-42s %s\n" "$id" "$desc" "$val"; PASS=$((PASS+1));;
    False) printf "  ${R}FAIL${N}  %-6s %-42s %s\n" "$id" "$desc" "$val"; FAIL=$((FAIL+1));;
    *)     printf "  ${Y}NA${N}    %-6s %-42s %s\n" "$id" "$desc" "$val"; NA=$((NA+1));;
  esac
}

echo "OpenBridger 生产验收回归"
echo "目标：$BASE   时间：$(date '+%F %T %Z')"
echo "────────────────────────────────────────────────────────────────"

echo
echo "[A] 公开面与静态资源"
check A1 "状态接口可达" "200" "$BASE/api/status"
check A2 "根路径返回前端" "200" "$BASE/"
json_check A3 "公开注册开关（依运营策略）" "'register_enabled='+str(d.get('data',{}).get('register_enabled'))" "$BASE/api/status" info
json_check A4 "Turnstile 人机校验开关" "'turnstile_check='+str(d.get('data',{}).get('turnstile_check'))" "$BASE/api/status" info
json_check A5 "站点地址（须为生产域名）" "'server_address='+str(d.get('data',{}).get('server_address'))" "$BASE/api/status" info
json_check A6 "在线支付开关（须为 False）" "'online_payment_enabled='+str(d.get('data',{}).get('online_payment_enabled'))" "$BASE/api/status" info

echo
echo "[B] 权限边界（未认证必须被拒）"
check B1 "用户信息 /api/user/self" "401|403" "$BASE/api/user/self"
check B2 "令牌列表 /api/token/" "401|403" "$BASE/api/token/"
check B3 "渠道列表 /api/channel/" "401|403" "$BASE/api/channel/"
check B4 "模型清单 /v1/models" "401|403" "$BASE/v1/models"
check B5 "聊天补全 /v1/chat/completions" "401|403|400" -X POST "$BASE/v1/chat/completions" -H "Content-Type: application/json" -d '{"model":"probe","messages":[]}'
check B6 "支付接口 /api/user/pay" "401|403" -X POST "$BASE/api/user/pay"

echo
echo "[C] 认证后闭环（需要 OB_API_KEY）"
if [ -n "$API_KEY" ]; then
  check C1 "用令牌取模型清单" "200" "$BASE/v1/models" -H "Authorization: Bearer $API_KEY"
  check C2 "用令牌发起对话（无渠道应为 4xx）" "200|400|402|403|404|429" -X POST "$BASE/v1/chat/completions" \
       -H "Authorization: Bearer $API_KEY" -H "Content-Type: application/json" \
       -d '{"model":"probe-not-exist","messages":[{"role":"user","content":"hi"}],"max_tokens":1}'
else
  printf "  ${Y}跳过${N}  C1-C2 未提供 OB_API_KEY\n"; NA=$((NA+2))
fi

echo
echo "[D] 在线支付必须封死"
check D1 "支付接口 POST 被拒" "401|403" -X POST "$BASE/api/user/pay"
check D2 "支付回调异常 ingress 被拒" "401|403|404" "$BASE/api/user/pay_callback"

echo
echo "[E] Cloudflare 层"
# 各种 SDK 的 User-Agent 必须不被 CF 拦；1010 = Browser Integrity Check 误伤
for pair in "OpenAI/Python 1.55.3:E1" "python-httpx/0.28.1:E2" "axios/1.7.9:E3" "PostmanRuntime/7.40.0:E4" "okhttp/4.12.0:E5" "Go-http-client/2.0:E6"; do
  ua="${pair%:*}"; id="${pair##*:}"
  check "$id" "UA 兼容：${ua%% *}" "200" -A "$ua" "$BASE/api/status"
done
# 流式路径不得被边缘缓存
cache_status=$(curl -s -D- -o /dev/null -A "$UA" --max-time 25 "$BASE/v1/models" 2>/dev/null | tr -d '\r' | awk -F': ' '/^[Cc]f-[Cc]ache-[Ss]tatus/{print $2}')
case "$cache_status" in
  DYNAMIC|BYPASS|MISS|EXPIRED) printf "  ${G}PASS${N}  %-6s %-42s %s\n" "E7" "API 路径不落缓存" "$cache_status"; PASS=$((PASS+1));;
  "")  printf "  ${Y}NA${N}    %-6s %-42s 未取得 cf-cache-status\n" "E7" "API 路径不落缓存"; NA=$((NA+1));;
  *)   printf "  ${R}FAIL${N}  %-6s %-42s %s（命中缓存，流式会串数据）\n" "E7" "API 路径不落缓存" "$cache_status"; FAIL=$((FAIL+1));;
esac

echo
echo "────────────────────────────────────────────────────────────────"
printf "结果：%s 通过 / %s 失败 / %s 跳过 / %s 信息项\n" "$PASS" "$FAIL" "$NA" "$INFO"
[ "$FAIL" -eq 0 ] || exit 1
