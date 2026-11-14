#!/usr/bin/env bash
#
# 服务器端初始化：把一台干净的 Linux 变成能跑这个博客的环境。
#
# 用法（在服务器上以 root 执行）：
#   sudo bash server-init.sh '你的数据库密码' 8080
#
# 参数：
#   $1  数据库密码（必填，用于创建 blog 账号；会和 backend/.env.prod 里的 DB_PASSWORD 一致）
#   $2  Nginx 监听端口（默认 8080；备案通过后改成 80）
#
# 做这几件事（全部幂等，可重复执行）：
#   1. 建 blog 低权限用户与目录
#   2. 装 nginx / mysql-server（按 apt 或 dnf 自动选择）
#   3. 建库 + 建 blog 账号
#   4. 装 systemd 单元并 enable
#   5. 装 nginx 站点配置（端口可指定）
#
# 前置：产物已经 rsync 到 /opt/blog 和 /var/www/blog/dist（见 remote-deploy.sh）。
#
# 注意：本脚本没有在真实服务器上跑过，第一次执行请逐段看输出。

set -euo pipefail

DB_PASSWORD="${1:-}"
PORT="${2:-8080}"

if [[ -z "$DB_PASSWORD" ]]; then
  echo "用法: sudo bash $0 '数据库密码' [端口，默认 8080]" >&2
  exit 1
fi
if [[ "$(id -u)" -ne 0 ]]; then
  echo "请用 root 执行（sudo bash $0 ...）" >&2
  exit 1
fi

echo "==> [1/6] 用户与目录"
id blog >/dev/null 2>&1 || useradd -r -m -s /bin/bash blog
mkdir -p /opt/blog/backend/bin /opt/blog/backend/uploads /var/www/blog /backup
chown -R blog:blog /opt/blog /var/www/blog

echo "==> [2/6] 安装依赖"
if command -v apt-get >/dev/null 2>&1; then
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -qq
  apt-get install -y -qq nginx mysql-server curl rsync >/dev/null
  MYSQL_SVC=mysql
elif command -v dnf >/dev/null 2>&1; then
  dnf install -y -q nginx mysql-server curl rsync >/dev/null
  MYSQL_SVC=mysqld
elif command -v yum >/dev/null 2>&1; then
  yum install -y -q nginx mysql-server curl rsync >/dev/null
  MYSQL_SVC=mysqld
else
  echo "未识别的包管理器，请手动安装 nginx / mysql / rsync 后重跑" >&2
  exit 1
fi

echo "==> [3/6] 启动 MySQL"
systemctl enable --now "$MYSQL_SVC" >/dev/null 2>&1 || true
for i in $(seq 1 30); do
  mysqladmin ping >/dev/null 2>&1 && break
  sleep 1
done
mysqladmin ping >/dev/null 2>&1 || { echo "MySQL 没起来，看 journalctl -u $MYSQL_SVC" >&2; exit 1; }

echo "==> [4/6] 建库与账号"
# 单引号转义，避免密码里的 ' 破坏 SQL
ESC_PW=$(printf '%s' "$DB_PASSWORD" | sed "s/'/''/g")
mysql <<SQL
CREATE DATABASE IF NOT EXISTS blog DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER IF NOT EXISTS 'blog'@'localhost' IDENTIFIED BY '${ESC_PW}';
GRANT ALL PRIVILEGES ON blog.* TO 'blog'@'localhost';
FLUSH PRIVILEGES;
SQL
echo "    库 blog 就绪"

echo "==> [5/6] systemd"
if [[ -f /opt/blog/deploy/blog-server.service ]]; then
  cp /opt/blog/deploy/blog-server.service /etc/systemd/system/blog-server.service
  systemctl daemon-reload
  systemctl enable blog-server >/dev/null 2>&1 || true
  echo "    已安装（进程等产物上传后再启动）"
else
  echo "    跳过：/opt/blog/deploy/blog-server.service 不存在，请先上传 deploy/ 目录"
fi

echo "==> [6/6] Nginx（监听 ${PORT}）"
if [[ -f /opt/blog/deploy/nginx-youzhichong.xyz.conf ]]; then
  sed "s/listen 80;/listen ${PORT};/" \
    /opt/blog/deploy/nginx-youzhichong.xyz.conf > /etc/nginx/conf.d/blog.conf
  rm -f /etc/nginx/sites-enabled/default
  nginx -t
  systemctl enable --now nginx >/dev/null 2>&1 || true
  systemctl reload nginx
  echo "    站点已启用，端口 ${PORT}"
else
  echo "    跳过：nginx 配置不存在，请先上传 deploy/ 目录"
fi

echo
echo "==> 环境就绪"
echo "下一步："
echo "  1. 上传产物（本地执行 deploy/remote-deploy.sh）"
echo "  2. 数据库迁移：cd /opt/blog/backend && for f in migrations/00*.sql; do mysql -ublog -p --default-character-set=utf8mb4 blog < \$f; done"
echo "  3. 启动：systemctl start blog-server && systemctl status blog-server"
echo "  4. 本机验证：curl http://127.0.0.1:${PORT}/ 和 curl http://127.0.0.1:8090/api/v1/ping"
echo
echo "别忘了在阿里云控制台放行安全组的 ${PORT} 端口，否则外面连不进来。"
