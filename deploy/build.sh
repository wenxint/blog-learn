#!/usr/bin/env bash
#
# 一键构建前后端。
#
# 用法：
#   ./deploy/build.sh
#
# 产物：
#   backend/bin/blog-server     单文件可执行（含全部依赖，无需 Go 环境）
#   frontend/dist/              静态文件，交给 Nginx

set -euo pipefail

# 脚本所在目录的上一级 = 项目根目录（不管从哪里调用都能定位）
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "==> 项目根目录: $ROOT"

echo
echo "==> [1/3] 后端静态检查"
(cd "$ROOT/backend" && go vet ./...)

echo
echo "==> [2/3] 后端编译"
(cd "$ROOT/backend" && make build)

echo
echo "==> [3/3] 前端构建"
(cd "$ROOT/frontend" && pnpm install --frozen-lockfile && pnpm build)

echo
echo "==> 完成"
echo "    后端: $ROOT/backend/bin/blog-server"
echo "    前端: $ROOT/frontend/dist"
du -sh "$ROOT/backend/bin/blog-server" "$ROOT/frontend/dist" 2>/dev/null || true
