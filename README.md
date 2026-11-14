# 我的博客（Go + Vue3 前后端分离）

一个边学边写的博客系统：后端从零用 Go 写，前端用 Vue3 做。

- 后端：Go 1.26 · Gin · GORM · MySQL 8.4 · JWT（34 个接口、24 个单测）
- 前端：Vue 3 · Vite · Pinia · Vue Router · Element Plus（前台 6 页 + 后台 5 页）
- 文档总入口：[docs/README.md](docs/README.md)
- 建站系列（13 篇，后端 6 + 前端 2 + 部署 3 + 总览/复盘 2）：[docs/articles/](docs/articles/)
  - 从整体架构开始：[01-整体架构](docs/articles/01-整体架构.md)
  - 部署重点：[10-从本地到线上](docs/articles/10-从本地到线上.md)、[11-生产配置与安全清单](docs/articles/11-生产配置与安全清单.md)、[12-上线实录](docs/articles/12-上线实录.md)
- 我学到了什么：[docs/08-学习档案.md](docs/08-学习档案.md)
- 架构总览（as-built）：[docs/00-架构总览.md](docs/00-架构总览.md)
- 接口文档：[docs/02-接口文档.md](docs/02-接口文档.md)
- 学习路线（20 课进度与要点）：[docs/01-学习路线图.md](docs/01-学习路线图.md)
- 部署指南：[docs/03-部署.md](docs/03-部署.md)
- 上线清单（youzhichong.xyz）：[docs/04-上线清单-youzhichong.xyz.md](docs/04-上线清单-youzhichong.xyz.md)
- 上线实录完整版（2026-11-14）：[docs/05-上线实录-2026-11-14.md](docs/05-上线实录-2026-11-14.md)
- 自助部署手册（无 AI 辅助）：[docs/06-自助部署手册.md](docs/06-自助部署手册.md)
- 前端视觉重构（设计 token、rem 适配，番外）：[docs/07-前端视觉重构.md](docs/07-前端视觉重构.md)
- 笔记目录：[docs/notes/README.md](docs/notes/README.md)
  - [踩坑记录](docs/notes/踩坑记录.md) · [后端速查](docs/notes/后端速查.md) · [前端速查](docs/notes/前端速查.md) · [自测题](docs/notes/自测题.md)

---

## 快速开始

### 0. 前置条件

```bash
go version     # 需要 go 1.24+
node -v        # 需要 node 20+
pnpm -v
mysql --version  # 需要 MySQL 8.0+
```

### 1. 准备数据库

```bash
# 启动 MySQL（macOS + Homebrew）
brew services start mysql@8.4

# 建库 + 专用账号（只需执行一次）
mysql -uroot -p <<'SQL'
ALTER USER 'root'@'localhost' IDENTIFIED BY '你的root密码';
CREATE DATABASE IF NOT EXISTS blog DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER IF NOT EXISTS 'blog'@'localhost' IDENTIFIED BY 'blog123456';
CREATE USER IF NOT EXISTS 'blog'@'127.0.0.1' IDENTIFIED BY 'blog123456';
GRANT ALL PRIVILEGES ON blog.* TO 'blog'@'localhost';
GRANT ALL PRIVILEGES ON blog.* TO 'blog'@'127.0.0.1';
FLUSH PRIVILEGES;
SQL
```

### 2. 初始化表结构与种子数据

```bash
cd backend
cp .env.example .env      # 然后按本机情况改 DB_PASSWORD / JWT_SECRET
# JWT_SECRET 用 openssl rand -hex 32 生成

# 必须加 --default-character-set=utf8mb4：
# LANG 为空的环境（很多服务器、容器、CI 都是）里 mysql 客户端默认用 latin1，
# UTF-8 的种子文件会被双重编码存进库，页面显示成 "Go ä»Žé›¶å¼€å§‹" 这种乱码。
mysql -ublog -p --default-character-set=utf8mb4 blog < migrations/001_init.sql    # 建表
mysql -ublog -p --default-character-set=utf8mb4 blog < migrations/002_seed.sql    # 文章/分类/标签种子
mysql -ublog -p --default-character-set=utf8mb4 blog < migrations/003_add_article_tags_and_comments.sql
mysql -ublog -p --default-character-set=utf8mb4 blog < migrations/004_seed_article_tags.sql
mysql -ublog -p --default-character-set=utf8mb4 blog < migrations/005_seed_comments.sql
```

### 3. 启动后端（:8090）

```bash
cd backend
make run          # 等价于 go run ./cmd/server
```

验证：

```bash
curl http://127.0.0.1:8090/api/v1/ping
curl http://127.0.0.1:8090/api/v1/articles
```

### 4. 启动前端（:5173）

```bash
cd frontend
pnpm install
pnpm dev
```

浏览器打开 <http://127.0.0.1:5173>

> 开发期前端通过 Vite 的 proxy 把 `/api` 和 `/uploads` 转发到 `127.0.0.1:8090`，
> 所以浏览器看起来是同源的，不需要处理跨域。

---

## 目录结构

```
blog/
├── backend/                 # Go 后端
│   ├── cmd/server/          # 程序入口（配置 → 连库 → 路由 → 优雅关闭）
│   ├── internal/
│   │   ├── config/          # 读 .env，配置校验
│   │   ├── database/        # GORM 连接与连接池
│   │   ├── model/           # GORM 实体（对应 6 张表）
│   │   ├── repository/      # 数据访问层
│   │   ├── service/         # 业务层（校验、事务、状态流转）
│   │   ├── handler/         # HTTP 层（参数、响应、错误翻译）
│   │   ├── middleware/      # RequestID / 日志 / Recover / CORS / JWT 鉴权
│   │   ├── router/          # 路由与依赖装配
│   │   └── pkg/             # response / jwt / hash / pagination / authctx
│   ├── migrations/          # 按编号递增的 SQL 迁移脚本
│   ├── uploads/             # 上传的图片（已 gitignore）
│   └── Makefile             # make run / build / test / migrate / seed / db
├── frontend/
│   └── src/
│       ├── api/             # axios 实例与接口封装
│       ├── router/          # 路由表
│       ├── layouts/         # 布局
│       ├── views/           # 页面
│       ├── utils/           # 工具函数
│       └── styles/          # 全局样式
└── docs/                    # 架构、路线图、踩坑记录
```

---

## 接口一览（`/api/v1`）

### 公开

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/ping` | 健康检查 |
| GET | `/articles` | 文章列表（分页、关键词、分类、标签筛选） |
| GET | `/articles/:id` | 文章详情（阅读数 +1） |
| GET | `/categories` | 分类列表（含文章数） |
| GET | `/tags` | 标签列表（含文章数） |
| GET | `/archives` | 归档（按月统计） |
| GET | `/articles/:id/comments` | 评论树 |
| POST | `/articles/:id/comments` | 发表评论（进入待审核） |

### 认证

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/auth/register` | 注册（默认 author 角色） |
| POST | `/auth/login` | 登录，返回 JWT |
| GET | `/auth/profile` | 当前用户（需 token） |
| PUT | `/auth/password` | 修改密码（需 token） |

### 后台（需 `Authorization: Bearer <token>`）

| 方法 | 路径 | 权限 |
|---|---|---|
| POST / PUT | `/admin/articles`、`/admin/articles/:id` | 登录 |
| DELETE | `/admin/articles/:id` | admin |
| POST / PUT | `/admin/categories`、`/admin/categories/:id` | admin |
| DELETE | `/admin/categories/:id` | admin |
| POST / PUT | `/admin/tags`、`/admin/tags/:id` | 登录 |
| DELETE | `/admin/tags/:id` | admin |
| GET | `/admin/comments` | 登录 |
| PUT | `/admin/comments/:id/status` | admin |
| DELETE | `/admin/comments/:id` | admin |
| POST | `/admin/upload` | 登录 |

### 诊断（仅 dev）

`/debug/db` · `/debug/models` · `/debug/tx-rollback` · `/debug/panic` · `/debug/slow`

---

## 种子账号

| 用户名 | 密码 | 角色 |
|---|---|---|
| admin | 123456 | admin |

> 仅本地开发使用。生产环境务必修改密码和 `JWT_SECRET`。

---

## 常用命令

```bash
# 后端
cd backend
make run        # 启动
make build      # 编译到 bin/blog-server
make test       # 单元测试
make vet        # 静态检查
make fmt        # 格式化
make migrate    # 重跑建表脚本
make seed       # 重灌种子数据
make db         # 进 MySQL 终端

# 前端
cd frontend
pnpm dev        # 开发服务器
pnpm build      # 生产构建到 dist/
pnpm preview    # 预览构建产物
```

---

## 常见问题

**Q: 后端启动报 `listen tcp :8090: bind: address already in use`**

上次的进程没退。`lsof -nP -iTCP:8090 -sTCP:LISTEN` 找到 PID 后 `kill` 掉。

**Q: 前端页面空白 / 请求 500**

先确认后端在 8090 上活着：`curl http://127.0.0.1:8090/api/v1/ping`。
再确认 `vite.config.js` 里的 proxy target 是 `http://127.0.0.1:8090`。

**Q: 中文变乱码**

检查三层字符集：服务端默认、库/表 `utf8mb4`、以及**客户端连接**（`~/.my.cnf` 里加 `default-character-set=utf8mb4`）。
