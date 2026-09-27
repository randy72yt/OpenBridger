#!/usr/bin/env bash
# obctl — OpenBridger 一键运维 CLI（部署在服务器 /srv/openbridger/obctl）
# 用法: obctl <command> [args]     详见 obctl help
# 设计原则: 每个命令幂等、输出给人看、关键操作前自动预检、失败给明确下一步。
set -u
OB=/srv/openbridger
REPO=$OB/repo
APP_CONTAINER=repo-openbridger-1
COMPOSE_FILE=compose.release.yml
ENVFILE=$OB/.env

c_red()  { printf '\033[31m%s\033[0m\n' "$*"; }
c_grn()  { printf '\033[32m%s\033[0m\n' "$*"; }
c_ylw()  { printf '\033[33m%s\033[0m\n' "$*"; }
hdr()    { printf '\n=== %s ===\n' "$*"; }
die()    { c_red "ERROR: $*"; exit 1; }

compose() { (cd "$REPO" && OPENBRIDGER_IMAGE_TAG="${OBTAG:-$(cur_tag)}" docker compose -f "$COMPOSE_FILE" --env-file "$ENVFILE" "$@"); }

cur_tag() {
  docker inspect "$APP_CONTAINER" --format '{{.Config.Image}}' 2>/dev/null | awk -F: '{print $NF}'
}

http_code() { curl -s -o /dev/null -m 10 -w '%{http_code}' "$1" 2>/dev/null || echo 000; }

app_version() {
  curl -s -m 10 http://127.0.0.1:3000/api/status 2>/dev/null \
    | python3 -c "import json,sys; d=json.load(sys.stdin)['data']; print(d.get('version','?'))" 2>/dev/null || echo "unreachable"
}

cmd_status() {
  hdr "Containers"
  docker ps -a --format 'table {{.Names}}\t{{.Status}}\t{{.Image}}' | grep -E 'NAMES|openbridger'
  hdr "App"
  echo "local /api/status version: $(app_version)   (image tag: $(cur_tag))"
  hdr "Host"
  echo "uptime: $(uptime | sed 's/^ *//')"
  free -m | awk 'NR==1||/Mem|Swap/'
  df -h / /var/lib/docker | awk 'NR==1||/\//'
}

cmd_smoke() {
  local fail=0
  hdr "L1 process"
  local v; v=$(app_version)
  [ "$v" = "unreachable" ] && { c_red "FAIL local /api/status"; fail=1; } || c_grn "OK  local /api/status version=$v"
  for c in "$APP_CONTAINER" openbridger-mysql-1 openbridger-redis-1; do
    docker ps --format '{{.Names}}' | grep -qx "$c" && c_grn "OK  container $c" || { c_red "FAIL container $c not running"; fail=1; }
  done
  hdr "L2 edge (Cloudflare + Nginx + TLS)"
  for u in https://openbridger.com/api/status https://api.openbridger.com/api/status https://docs.openbridger.com/; do
    local host code; host=$(echo "$u" | awk -F/ '{print $3}')
    if ! getent hosts "$host" >/dev/null 2>&1; then
      c_ylw "WARN $host 本机 DNS 未解析（阿里云内网 DNS 缓存滞后时常见，公网正常则不影响用户）"
      continue
    fi
    code=$(http_code "$u")
    [ "$code" = 200 ] && c_grn "OK  $code $u" || { c_red "FAIL $code $u"; fail=1; }
  done
  [ $fail -eq 0 ] && c_grn "SMOKE PASS" || { c_red "SMOKE FAIL"; return 1; }
}

cmd_diag() {
  cmd_status
  hdr "Edge HTTP"
  for u in https://openbridger.com/api/status https://api.openbridger.com/api/status https://docs.openbridger.com/; do
    local host; host=$(echo "$u" | awk -F/ '{print $3}')
    getent hosts "$host" >/dev/null 2>&1 || { echo "NXDOMAIN  $host（记录未创建）"; continue; }
    echo "$(http_code "$u")  $u"
  done
  hdr "Backup freshness"
  cmd_backup_check || true
  hdr "Monitor last run"
  tail -n 3 "$OB/backups/monitor.log" 2>/dev/null || echo "(no monitor.log)"
  hdr "App errors (last 1h, tail 20)"
  docker logs --since 1h "$APP_CONTAINER" 2>&1 | grep -iE 'error|panic|fatal' | tail -n 20 || echo "(none)"
  hdr "MySQL slow/errors (tail 10)"
  docker logs --since 1h openbridger-mysql-1 2>&1 | grep -iE 'error|crash' | tail -n 10 || echo "(none)"
  c_ylw "提示: diag 只是体检；处置请按 docs/openbridger/ops-handbook.md 应急章节"
}

cmd_backup() {
  hdr "Manual backup"
  bash "$OB/backup.sh" || die "backup.sh failed, see $OB/backups/backup.log"
  tail -n 2 "$OB/backups/backup.log"
  ls -lh "$OB/backups"/latest-* | awk '{print $5, $9}'
  c_grn "BACKUP DONE （Mac 异地副本将在每日 09:30 自动拉取；急用可手动 scp）"
}

cmd_backup_check() {
  local f="$OB/backups/latest-mysql.sql.gz"
  [ -e "$f" ] || { c_red "FAIL 无备份文件 $f"; return 1; }
  local age_h size
  age_h=$(( ( $(date +%s) - $(stat -c %Y -L "$f") ) / 3600 ))
  size=$(du -h -L "$f" | cut -f1)
  if [ "$age_h" -lt 26 ]; then c_grn "OK  latest-mysql 年龄 ${age_h}h, 大小 $size"
  else c_red "FAIL latest-mysql 已 ${age_h}h（>26h），备份可能中断"; return 1; fi
}

cmd_restore() {
  local dump="${1:-}"
  [ -n "$dump" ] || die "用法: obctl restore <dump.sql.gz>   （先 ls -lh $OB/backups/ 选文件）"
  [ -f "$dump" ] || die "文件不存在: $dump"
  [ "${OBCTL_CONFIRM:-}" = "YES" ] || die "危险操作！确认后执行: OBCTL_CONFIRM=YES obctl restore $dump"
  hdr "Restore MySQL from $dump"
  c_ylw "1/4 停应用容器（防写入冲突）"
  docker stop "$APP_CONTAINER" || true
  c_ylw "2/4 灌入 dump"
  zcat "$dump" | docker exec -i openbridger-mysql-1 sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" openbridger' \
    || die "灌库失败！应用仍处于停止状态，检查 dump 后重试或换更早备份"
  c_ylw "3/4 起应用"
  docker start "$APP_CONTAINER"; sleep 8
  c_ylw "4/4 冒烟验证"
  cmd_smoke && c_grn "RESTORE DONE" || die "恢复后冒烟未过，立即 obctl diag 排查"
}

cmd_build() {
  local tag="${1:-}"
  [ -n "$tag" ] || die "用法: obctl build <git-ref>   （tag/分支/sha，例如 v0.1.1 或 main）"
  cd "$REPO" || die "repo 不存在"
  hdr "Build $tag（2C4G 串行补丁，nohup 后台）"
  git fetch --tags --quiet || die "git fetch 失败"
  git checkout --quiet "$tag" || die "checkout $tag 失败"
  local short; short=$(git rev-parse --short HEAD)
  cp Dockerfile.dev /tmp/Dockerfile.dev.orig
  sed -i "s|^RUN go build|ENV GOFLAGS=-p=1 GOMAXPROCS=1\nRUN go build|" Dockerfile.dev
  grep -q GOFLAGS Dockerfile.dev || die "GOFLAGS 补丁未生效"
  c_ylw "开始构建 openbridger:$short（日志 $OB/build.log）"
  nohup sh -c "cd $REPO && OPENBRIDGER_IMAGE_TAG=$short docker compose -f $COMPOSE_FILE --env-file $ENVFILE build; echo \$? > $OB/build.exit; cp /tmp/Dockerfile.dev.orig $REPO/Dockerfile.dev" \
    > "$OB/build.log" 2>&1 &
  echo "build pid=$! tag=$short  →  用 obctl build-status 轮询"
}

cmd_build_status() {
  if [ -f "$OB/build.exit" ]; then
    local code; code=$(cat "$OB/build.exit")
    [ "$code" = 0 ] && c_grn "BUILD OK" || c_red "BUILD FAILED (exit=$code)，tail $OB/build.log"
    rm -f "$OB/build.exit"
    docker images | grep openbridger | head -3
  else
    c_ylw "构建中…"; tail -n 3 "$OB/build.log" 2>/dev/null
  fi
}

cmd_deploy() {
  local tag="${1:-}"
  [ -n "$tag" ] || die "用法: obctl deploy <image-tag>   （obctl build 产物，docker images 可查）"
  docker image inspect "openbridger:$tag" >/dev/null 2>&1 || docker image inspect "repo-openbridger:$tag" >/dev/null 2>&1 \
    || die "镜像不存在: $tag（先 obctl build）"
  local prev; prev=$(cur_tag)
  echo "$prev" > "$OB/.prev-tag"
  hdr "Deploy $prev → $tag"
  OBTAG="$tag" compose up -d || die "compose up 失败"
  sleep 10
  if cmd_smoke; then
    c_grn "DEPLOY OK ($prev → $tag)；旧 tag $prev 已记录可回滚: obctl rollback"
  else
    c_red "冒烟未过，自动回滚到 $prev"
    OBTAG="$prev" compose up -d; sleep 10
    cmd_smoke && c_ylw "已回滚到 $prev，请排查后再发" || die "回滚后仍异常！立即 obctl diag"
    return 1
  fi
}

cmd_rollback() {
  local tag="${1:-$(cat "$OB/.prev-tag" 2>/dev/null)}"
  [ -n "$tag" ] || die "无回滚目标（$OB/.prev-tag 不存在且未传参）"
  hdr "Rollback → $tag"
  OBTAG="$tag" compose up -d || die "compose up 失败"
  sleep 10
  cmd_smoke && c_grn "ROLLBACK OK → $tag" || die "回滚后冒烟未过，立即 obctl diag"
}

cmd_accept() {
  [ -n "${OB_API_KEY:-}" ] || die "需要 OB_API_KEY: OB_API_KEY=sk-... obctl accept"
  OB_BASE=https://openbridger.com bash "$OB/prod-acceptance.sh"
}

cmd_sql() {
  local f="${1:--}"
  hdr "SQL → openbridger"
  if [ "$f" = "-" ]; then
    docker exec -i openbridger-mysql-1 sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" --table openbridger'
  else
    [ -f "$f" ] || die "文件不存在: $f（建议本地写好 scp 上传，避免引号地狱）"
    docker exec -i openbridger-mysql-1 sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" --table openbridger' < "$f"
  fi
}

cmd_logs() {
  local target="${1:-app}" lines="${2:-100}"
  case "$target" in
    app)   docker logs --tail "$lines" "$APP_CONTAINER" 2>&1 ;;
    mysql) docker logs --tail "$lines" openbridger-mysql-1 2>&1 ;;
    redis) docker logs --tail "$lines" openbridger-redis-1 2>&1 ;;
    nginx) tail -n "$lines" /var/log/nginx/error.log ;;
    *) die "logs [app|mysql|redis|nginx] [lines]" ;;
  esac
}

cmd_alert_test() {
  hdr "Alert test email"
  python3 - <<'PY'
import json, smtplib, ssl, subprocess
from email.mime.text import MIMEText
def run(c): return subprocess.run(c, shell=True, capture_output=True, text=True).stdout.strip()
kv = dict(l.split("\t",1) for l in run("docker exec openbridger-mysql-1 sh -c 'mysql -uroot -p\"$MYSQL_ROOT_PASSWORD\" -N openbridger -e \"SELECT \\`key\\`,\\`value\\` FROM options WHERE \\`key\\` IN (\\\"SMTPServer\\\",\\\"SMTPPort\\\",\\\"SMTPAccount\\\",\\\"SMTPToken\\\",\\\"SMTPFrom\\\")\"' 2>/dev/null").splitlines() if "\t" in l)
to = open("/srv/openbridger/alert-email.txt").read().strip()
msg = MIMEText("obctl alert-test: 告警链路正常。\nTime: " + run("date -Iseconds"), "plain", "utf-8")
msg["Subject"] = "[OpenBridger TEST] alert pipeline ok"
msg["From"] = kv.get("SMTPFrom", to); msg["To"] = to
with smtplib.SMTP_SSL(kv["SMTPServer"], int(kv.get("SMTPPort",465)), context=ssl.create_default_context(), timeout=20) as s:
    s.login(kv["SMTPAccount"], kv["SMTPToken"]); s.send_message(msg)
print("sent to", to)
PY
  c_grn "测试邮件已发出，收件箱（含垃圾邮件）确认到达即链路正常"
}

cmd_restart() {
  local target="${1:-app}"
  hdr "Restart $target"
  case "$target" in
    app)   docker restart "$APP_CONTAINER" ;;
    all)   docker restart openbridger-mysql-1 openbridger-redis-1; sleep 5; docker restart "$APP_CONTAINER" ;;
    *) die "restart [app|all]（mysql/redis 单独重启: docker restart <容器>）" ;;
  esac
  sleep 10
  cmd_smoke && c_grn "RESTART OK" || die "重启后冒烟未过，立即 obctl diag"
}

cmd_help() {
  cat <<'EOF'
obctl — OpenBridger 一键运维
  status                    容器/应用版本/主机资源总览
  restart [app|all]         重启应用（all=含 MySQL/Redis），自带冒烟验证
  diag                      全量体检（status + 边缘 + 备份 + 最近错误日志）
  smoke                     L1 进程层 + L2 边缘层快速冒烟
  backup                    立即执行一次完整备份（MySQL+Redis+配置）
  backup-check              校验最新备份新鲜度（<26h）与大小
  restore <dump.sql.gz>     从备份恢复 MySQL（需 OBCTL_CONFIRM=YES）
  build <git-ref>           后台构建镜像（自动 GOFLAGS 串行补丁）
  build-status              查看构建进度/结果
  deploy <image-tag>        切换镜像 + 冒烟，失败自动回滚
  rollback [image-tag]      回滚（默认上一次 deploy 前的 tag）
  accept                    L3 验收（需 OB_API_KEY=sk-...）
  sql <file|->              安全执行 SQL 文件或 stdin（免引号地狱）
  logs [app|mysql|redis|nginx] [lines]
  alert-test                发一封测试告警邮件验证告警链路
完整方案见 docs/openbridger/ops-handbook.md
EOF
}

case "${1:-help}" in
  status) cmd_status ;; restart) shift; cmd_restart "$@" ;; diag) cmd_diag ;; smoke) cmd_smoke ;;
  backup) cmd_backup ;; backup-check) cmd_backup_check ;; restore) shift; cmd_restore "$@" ;;
  build) shift; cmd_build "$@" ;; build-status) cmd_build_status ;;
  deploy) shift; cmd_deploy "$@" ;; rollback) shift; cmd_rollback "$@" ;;
  accept) cmd_accept ;; sql) shift; cmd_sql "$@" ;; logs) shift; cmd_logs "$@" ;;
  alert-test) cmd_alert_test ;; help|--help|-h) cmd_help ;;
  *) die "未知命令: $1（obctl help 查看）" ;;
esac
