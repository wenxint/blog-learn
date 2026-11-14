#!/usr/bin/env bash
#
# 本地一键部署到远程服务器（需已配置 SSH 免密登录）。
#
# 用法：
#   ./deploy/remote-deploy.sh root@47.119.155.176          # 默认用 8080 端口（备案前）
#   ./deploy/remote-deploy.sh root@47.119.155.176 80       # 备案通过后用 80
#
# 它会依次：构建 → 上传产物 → 远端初始化 → 跑迁移 → 起服务 → 自检。
#
# 注意：本脚本没有在真实服务器上跑过，第一次执行请留意每段输出。

set -euo pipefail

TARGET="${1:?用法: $0 user@host [端口，默认 8080]}"
PORT="${2:-8080}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> 目标: $TARGET  端口: $PORT"

echo
echo "==> [1/6] 本地构建"
"$ROOT/deploy/build.sh"

echo
echo "==> [2/6] 检查 SSH 免密"
if ! ssh -o BatchMode=yes -o ConnectTimeout=8 "$TARGET" 'echo ok' >/dev/null 2>&1; then
  echo "SSH 免密登录失败。请先配置密钥：" >&2
  echo "  ssh-copy-id $TARGET" >&2
  exit 1
fi
echo "    OK"

echo
echo "==> [3/6] 上传产物"
ssh "$TARGET" 'mkdir -p /opt/blog/backend/bin /opt/blog/backend/uploads /var/www/blog/dist /opt/blog/deploy'
rsync -az --delete "$ROOT/frontend/dist/"        "$TARGET:/var/www/blog/dist/"
rsync -az "$ROOT/backend/bin/blog-server"        "$TARGET:/opt/blog/backend/bin/"
rsync -az "$ROOT/backend/migrations/"            "$TARGET:/opt/blog/backend/migrations/"
rsync -az "$ROOT/backend/.env.prod"              "$TARGET:/opt/blog/backend/.env.prod"
rsync -az "$ROOT/deploy/"                        "$TARGET:/opt/blog/deploy/"
echo "    完成"

echo
echo "==> [4/6] 远端初始化（装 nginx/mysql、建库、装 systemd）"
# 密码不从本地命令行传，改为在远端从 .env.prod 读取，减少泄漏面
ssh "$TARGET" 'PW=$(grep "^DB_PASSWORD=" /opt/blog/backend/.env.prod | cut -d= -f2-) && bash /opt/blog/deploy/server-init.sh "$PW" '"$PORT"''

echo
echo "==> [5/6] 数据库迁移"
ssh "$TARGET" 'cd /opt/blog/backend && export MYSQL_PWD=$(grep "^DB_PASSWORD=" .env.prod | cut -d= -f2-) && for f in migrations/00*.sql; do echo "  ==> $f"; mysql -ublog --default-character-set=utf8mb4 blog < "$f"; done'

echo
echo "==> [6/6] 权限、启动与自检"
ssh "$TARGET" 'set -e
  chown blog:blog /opt/blog/backend/.env.prod && chmod 600 /opt/blog/backend/.env.prod
  chown -R blog:blog /opt/blog/backend
  systemctl restart blog-server
  sleep 2
  echo "--- 后端 ping ---"
  curl -s --max-time 5 http://127.0.0.1:8090/api/v1/ping || echo "(后端未响应)"
  echo
  echo "--- 本机 8080 页面 ---"
  curl -s -o /dev/null -w "HTTP %{http_code}\n" --max-time 5 http://127.0.0.1:'"$PORT"'/ || true
  echo "--- 服务状态 ---"
  systemctl is-active blog-server nginx
'

echo
echo "==> 完成"
echo "本机验证就绪后，从外部访问需要在阿里云安全组放行 ${PORT} 端口。"
echo "然后用浏览器打开 http://47.119.155.176:${PORT}"
