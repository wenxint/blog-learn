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

# 目标平台：默认交叉编译到 Linux，这样在 macOS 上也能产出可直接上服务器的二进制。
# 在服务器本机构建时它会自动是 linux/amd64，无需显式指定。
TARGET_OS="${TARGET_OS:-linux}"
TARGET_ARCH="${TARGET_ARCH:-amd64}"

# 国内网络下 Go 下载 toolchain 时，sumdb 校验地址会超时，这里给个可用默认值。
# 用 :- 写法，外部已设置的环境变量优先。
export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
export GOSUMDB="${GOSUMDB:-sum.golang.org}"

# 没装 pnpm 时退回 npx，避免「一键构建」在干净机器上直接失败。
if command -v pnpm >/dev/null 2>&1; then
  PNPM=(pnpm)
else
  echo "==> 未检测到 pnpm，改用 npx pnpm@10"
  PNPM=(npx --yes pnpm@10)
fi

echo "==> 项目根目录: $ROOT"

echo
echo "==> [1/3] 后端静态检查"
(cd "$ROOT/backend" && go vet ./...)

echo
echo "==> [2/3] 后端编译 (目标 ${TARGET_OS}/${TARGET_ARCH})"
# CGO_ENABLED=0：产出纯静态二进制，不依赖服务器上的 glibc 版本，
# 避免「本机编译、服务器报 GLIBC_2.xx not found」这类问题。
(cd "$ROOT/backend" && \
  CGO_ENABLED=0 GOOS="$TARGET_OS" GOARCH="$TARGET_ARCH" make build)

echo
echo "==> [3/3] 前端构建"
(cd "$ROOT/frontend" && "${PNPM[@]}" install --frozen-lockfile && "${PNPM[@]}" build)

echo
echo "==> 完成"
echo "    后端: $ROOT/backend/bin/blog-server"
echo "    前端: $ROOT/frontend/dist"
du -sh "$ROOT/backend/bin/blog-server" "$ROOT/frontend/dist" 2>/dev/null || true
