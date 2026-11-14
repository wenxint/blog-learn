#!/usr/bin/env bash
#
# 生成生产环境 .env
#
# 用法：
#   ./deploy/gen-env.sh --db-password '你的数据库密码'
#   ./deploy/gen-env.sh --db-password 'xxx' --out /etc/blog/blog.env
#   ./deploy/gen-env.sh                      # 不给密码则自动生成随机强密码并打印出来
#
# 说明：
#   - 后端只认「当前工作目录下的 .env」（代码里 godotenv.Load(".env") 写死文件名），
#     所以 --out 默认就是 /opt/blog/backend/.env。
#   - 也可以生成到 /etc/blog/blog.env，再配合 systemd 的 EnvironmentFile 使用，
#     这样配置和代码目录分离，回滚代码不会丢密钥。
#   - 生成的密码都是十六进制，不含引号/空格/特殊字符，避免 .env 解析踩坑。

set -euo pipefail

OUT="/opt/blog/backend/.env"
DB_HOST="127.0.0.1"
DB_PORT="3306"
DB_USER="blog"
DB_NAME="blog"
DB_PASSWORD=""
UPLOAD_DIR="/opt/blog/backend/uploads"
APP_PORT="8090"
COMMENT_RATE_LIMIT="3"
UPLOAD_MAX_MB="5"

usage() {
  sed -n '2,18p' "$0" | sed 's/^# \{0,1\}//'
  exit "${1:-0}"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out)            OUT="$2"; shift 2 ;;
    --db-host)        DB_HOST="$2"; shift 2 ;;
    --db-port)        DB_PORT="$2"; shift 2 ;;
    --db-user)        DB_USER="$2"; shift 2 ;;
    --db-name)        DB_NAME="$2"; shift 2 ;;
    --db-password)    DB_PASSWORD="$2"; shift 2 ;;
    --upload-dir)     UPLOAD_DIR="$2"; shift 2 ;;
    --port)           APP_PORT="$2"; shift 2 ;;
    -h|--help)        usage 0 ;;
    *) echo "未知参数: $1" >&2; usage 1 ;;
  esac
done

command -v openssl >/dev/null 2>&1 || { echo "缺少 openssl，无法生成密钥" >&2; exit 1; }

GENERATED_PASSWORD=0
if [[ -z "$DB_PASSWORD" ]]; then
  # 32 位十六进制，强度足够且不会有特殊字符
  DB_PASSWORD="$(openssl rand -hex 16)"
  GENERATED_PASSWORD=1
fi

JWT_SECRET="$(openssl rand -hex 32)"

mkdir -p "$(dirname "$OUT")"

cat > "$OUT" <<EOF
# 由 deploy/gen-env.sh 生成于 $(date '+%Y-%m-%d %H:%M:%S')
# 不要提交到 Git。权限应为 600。

APP_ENV=prod
APP_PORT=${APP_PORT}

DB_HOST=${DB_HOST}
DB_PORT=${DB_PORT}
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}
DB_CHARSET=utf8mb4
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10

JWT_SECRET=${JWT_SECRET}
JWT_EXPIRE_HOURS=24

UPLOAD_DIR=${UPLOAD_DIR}
UPLOAD_MAX_MB=${UPLOAD_MAX_MB}
DB_CONN_MAX_LIFETIME_MIN=30

COMMENT_RATE_LIMIT=${COMMENT_RATE_LIMIT}
COMMENT_RATE_WINDOW_SEC=60

# 部署在 Nginx 后面必须填 127.0.0.1，否则评论 IP 全是 127.0.0.1
TRUSTED_PROXIES=127.0.0.1
EOF

chmod 600 "$OUT"

echo "==> 已生成: $OUT"
echo "    APP_ENV=prod  TRUSTED_PROXIES=127.0.0.1  UPLOAD_DIR=$UPLOAD_DIR"
echo "    JWT_SECRET 已随机生成（64 位十六进制）"

if [[ "$GENERATED_PASSWORD" -eq 1 ]]; then
  echo
  echo "    数据库密码（自动生成，请记下并用于建库）:"
  echo "    $DB_PASSWORD"
  echo
  echo "    建库命令示例："
  echo "      sudo mysql -e \"CREATE DATABASE IF NOT EXISTS $DB_NAME DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;\""
  echo "      sudo mysql -e \"CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASSWORD';\""
  echo "      sudo mysql -e \"GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'localhost'; FLUSH PRIVILEGES;\""
fi

echo
echo "    校验：只应看到文件名和权限 600，不打印内容"
ls -l "$OUT"
