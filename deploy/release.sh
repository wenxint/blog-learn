#!/usr/bin/env bash
#
# 日常增量发布脚本（区别于首次部署的 remote-deploy.sh）
#
# 用法：
#   ./deploy/release.sh root@47.119.155.176
#   ./deploy/release.sh root@47.119.155.176 --with-articles
#   ./deploy/release.sh root@47.119.155.176 --skip-backend
#   ./deploy/release.sh root@47.119.155.176 --dry-run
#
# 只做「日常发布」该做的事：
#   1. 本地构建（前端 dist / 后端 Linux 二进制）
#   2. 发布前备份（数据库 + 旧二进制 + 旧 dist）
#   3. 先传二进制 .new，再原子替换；前端 rsync 后修权限
#   4. 重启 systemd 并做健康检查
#   5. 可选：导入 docs/articles/*.md 到线上数据库
#
# 它刻意不做：
#   - 不装 nginx/mysql、不建用户/库（那是 server-init.sh 的事）
#   - 不重跑 migrations（避免把已经清理掉的种子数据重新插回来）
#   - 不覆盖 .env.prod（生产密钥只在服务器上维护）
#
# 首次部署或换服务器请用：deploy/remote-deploy.sh

set -euo pipefail

usage() {
  # 只打印文件头注释，遇到 set -euo 就停，避免把代码也打出来
  awk 'NR>1 && /^set -euo/{exit} NR>1{sub(/^# ?/, ""); print}' "$0"
}

TARGET="${1:-}"
if [[ -z "$TARGET" ]]; then
  usage
  exit 1
fi
if [[ "$TARGET" == "-h" || "$TARGET" == "--help" ]]; then
  usage
  exit 0
fi
shift

PORT=8080
WITH_ARTICLES=0
PASSWORD="${BLOG_ADMIN_PASSWORD:-}"
BACKUP=1
SKIP_FRONTEND=0
SKIP_BACKEND=0
DRY_RUN=0
TARGET_OS="${TARGET_OS:-linux}"
TARGET_ARCH="${TARGET_ARCH:-amd64}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --port) PORT="${2:?缺少端口}"; shift 2 ;;
    --with-articles) WITH_ARTICLES=1; shift ;;
    --password) PASSWORD="${2:?缺少密码}"; shift 2 ;;
    --no-backup) BACKUP=0; shift ;;
    --skip-frontend) SKIP_FRONTEND=1; shift ;;
    --skip-backend) SKIP_BACKEND=1; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "未知参数: $1（用 --help 查看用法）" >&2; exit 1 ;;
  esac
done

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOST="${TARGET#*@}"
BASE_URL="${BLOG_BASE_URL:-http://$HOST:$PORT}"
TS="$(date +%Y%m%d-%H%M%S)"
BACKUP_DIR="/backup/release-$TS"

say() { printf '\n\033[1;36m==>\033[0m %s\n' "$*"; }
ok()  { printf '    \033[32m✓\033[0m %s\n' "$*"; }
warn(){ printf '    \033[33m!\033[0m %s\n' "$*"; }

run() {
  if [[ "$DRY_RUN" == 1 ]]; then
    printf '    [dry-run] %s\n' "$*"
  else
    "$@"
  fi
}

# ---------- 0. 前置检查 ----------
say "[0/6] 检查环境"
command -v ssh >/dev/null || { echo "缺少 ssh" >&2; exit 1; }
command -v rsync >/dev/null || { echo "缺少 rsync" >&2; exit 1; }
ssh -o BatchMode=yes -o ConnectTimeout=8 "$TARGET" 'echo ok' >/dev/null 2>&1 \
  || { echo "SSH 免密登录失败：请先 ssh-copy-id $TARGET" >&2; exit 1; }
ok "SSH 免密可用：$TARGET"
ok "目标站点：$BASE_URL"

if [[ "$WITH_ARTICLES" == 1 && -z "$PASSWORD" && "$DRY_RUN" == 0 ]]; then
  if [[ -t 0 ]]; then
    read -rsp "后台管理员密码（用于导入文章）: " PASSWORD
    echo
  else
    echo "非交互环境需要 --password 或 BLOG_ADMIN_PASSWORD 才能导入文章" >&2
    exit 1
  fi
fi

# ---------- 1. 发布前备份 ----------
if [[ "$BACKUP" == 1 ]]; then
  say "[1/6] 发布前备份 → $BACKUP_DIR"
  if [[ "$DRY_RUN" == 1 ]]; then
    printf '    [dry-run] mysqldump + 旧二进制 + 旧 dist → %s\n' "$BACKUP_DIR"
  else
    ssh "$TARGET" "set -e
      mkdir -p '$BACKUP_DIR'
      export MYSQL_PWD=\$(grep '^DB_PASSWORD=' /opt/blog/backend/.env.prod | cut -d= -f2-)
      mysqldump -ublog --default-character-set=utf8mb4 blog > '$BACKUP_DIR/blog.sql' 2>/dev/null || true
      cp -a /opt/blog/backend/bin/blog-server '$BACKUP_DIR/' 2>/dev/null || true
      cp -a /var/www/blog/dist '$BACKUP_DIR/dist' 2>/dev/null || true
      # 只保留最近 10 次发布备份，避免磁盘被历史备份吃满
      ls -1d /backup/release-* 2>/dev/null | sort | head -n -10 | xargs -r rm -rf
      echo '$BACKUP_DIR'"
    ok "备份完成"
  fi
else
  say "[1/6] 跳过备份（--no-backup）"
fi

# ---------- 2. 本地构建 ----------
say "[2/6] 本地构建"
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
export GOSUMDB="${GOSUMDB:-sum.golang.org}"

if [[ "$SKIP_FRONTEND" == 0 ]]; then
  if [[ "$DRY_RUN" == 1 ]]; then
    printf '    [dry-run] 前端构建：frontend -> dist/\n'
  else
    if [[ -x "$ROOT/frontend/node_modules/.bin/vite" ]]; then
      (cd "$ROOT/frontend" && ./node_modules/.bin/vite build >/dev/null)
    elif command -v pnpm >/dev/null 2>&1; then
      (cd "$ROOT/frontend" && pnpm install --frozen-lockfile && pnpm build >/dev/null)
    else
      (cd "$ROOT/frontend" && npx --yes pnpm@10 install --frozen-lockfile && npx --yes pnpm@10 build >/dev/null)
    fi
    ok "前端构建完成：$(du -sh "$ROOT/frontend/dist" | cut -f1)"
  fi
fi

if [[ "$SKIP_BACKEND" == 0 ]]; then
  if [[ "$DRY_RUN" == 1 ]]; then
    printf '    [dry-run] 后端构建：go vet + %s/%s 交叉编译\n' "$TARGET_OS" "$TARGET_ARCH"
  else
    (cd "$ROOT/backend" && go vet ./... && \
      CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" make build >/dev/null)
    ok "后端构建完成：$(du -sh "$ROOT/backend/bin/blog-server" | cut -f1)"
  fi
fi

# ---------- 3. 上传 ----------
say "[3/6] 上传产物"
if [[ "$SKIP_FRONTEND" == 0 ]]; then
  run rsync -az --delete "$ROOT/frontend/dist/" "$TARGET:/var/www/blog/dist/"
  # 本地 umask 偏严时 SVG/PNG 会是 600，Nginx 读不到会 403
  run ssh "$TARGET" 'chown -R blog:blog /var/www/blog/dist && find /var/www/blog/dist -type d -exec chmod 755 {} + && find /var/www/blog/dist -type f -exec chmod 644 {} +'
  ok "前端已上传并修正权限"
fi

if [[ "$SKIP_BACKEND" == 0 ]]; then
  # 先传 .new，校验通过后再 mv，避免覆盖正在运行的二进制
  run scp -q "$ROOT/backend/bin/blog-server" "$TARGET:/opt/blog/backend/bin/blog-server.new"
  run ssh "$TARGET" 'chown blog:blog /opt/blog/backend/bin/blog-server.new && chmod 755 /opt/blog/backend/bin/blog-server.new'
  ok "后端二进制已上传为 .new"
fi

# ---------- 4. 重启 ----------
say "[4/6] 切换并重启"
if [[ "$SKIP_BACKEND" == 0 ]]; then
  run ssh "$TARGET" 'set -e
    mv /opt/blog/backend/bin/blog-server.new /opt/blog/backend/bin/blog-server
    systemctl restart blog-server
    sleep 1
    systemctl is-active blog-server >/dev/null'
  ok "blog-server 已重启"
else
  warn "跳过后端发布，不重启服务"
fi

# ---------- 5. 文章导入（可选） ----------
say "[5/6] 文章内容"
if [[ "$WITH_ARTICLES" == 1 ]]; then
  if ! command -v python3 >/dev/null; then
    echo "缺少 python3，无法导入文章" >&2; exit 1
  fi
  if [[ "$DRY_RUN" == 1 ]]; then
    printf '    [dry-run] python3 scripts/import_articles.py --base-url %s\n' "$BASE_URL"
  else
    (cd "$ROOT" && python3 scripts/import_articles.py \
      --base-url "$BASE_URL" \
      --password "$PASSWORD")
  fi
  ok "文章已导入（按标题幂等：有则更新，无则创建）"
else
  warn "未导入文章（需要时加 --with-articles）"
fi

# ---------- 6. 健康检查 ----------
say "[6/6] 健康检查"
if [[ "$DRY_RUN" == 1 ]]; then
  printf '    [dry-run] 检查 /api/v1/ping、首页、systemd 状态\n'
  printf '\n\033[33m[dry-run] 未上传、未重启、未导入。\033[0m\n'
  exit 0
fi

ssh "$TARGET" "set -e
  echo -n '    后端 ping: '
  curl -fsS --max-time 5 http://127.0.0.1:8090/api/v1/ping >/dev/null && echo OK
  echo -n '    首页 HTTP: '
  curl -s -o /dev/null -w '%{http_code}\n' --max-time 5 http://127.0.0.1:$PORT/
  echo -n '    服务状态: '
  systemctl is-active blog-server nginx | tr '\n' ' '; echo"

say "发布完成"
echo "    站点：$BASE_URL"
echo "    备份：$BACKUP_DIR"
echo "    回滚：用备份里的 blog.sql / blog-server / dist 覆盖，然后 systemctl restart blog-server"
