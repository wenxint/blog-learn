# 上线清单：把博客部署到 youzhichong.xyz

> 场景：已有阿里云/腾讯云**国内**服务器，域名 `youzhichong.xyz` 托管在**阿里云云解析**。
> 目标形态沿用 [deploy/nginx.conf](deploy/nginx.conf)：Nginx 对外 + Go 在 `127.0.0.1:8090` + MySQL 本机。

---

## 当前阻塞项（2026-11-14 实测）

| 检查项 | 实测结果 | 结论 |
|---|---|---|
| 阿里云云解析记录 | 权威 NS `dns29.hichina.com` 上 `youzhichong.xyz` 和 `www` 均返回 `47.119.155.176` | ✅ 记录是对的 |
| 域名 NS | RDAP 显示 `DNS29.HICHINA.COM` / `DNS30.HICHINA.COM` | ✅ 委派目标正确 |
| **域名注册局状态** | 2026-11-14 复查仍为 `clientHold`；`last changed` 停在 `2026-07-01T08:21:29Z`，实名认证通过后未更新 | ❌ 注册局未解除，全球递归仍是 NXDOMAIN |
| 服务器 `47.119.155.176` | 2026-11-14 已部署：Alibaba Cloud Linux 4 + MySQL 8.0.46 + nginx 1.30.4 + systemd，服务以 `blog` 用户运行，`8080` 内部全通 | ✅ 已部署 |
| 阿里云安全组 | 外网 `8080` 不通（服务器内 `curl 127.0.0.1:8080` 正常） | ❌ 需在控制台放行 |
| 备案 | 未确认 | ⚠️ 待查 |

**结论：DNS 面板没问题，卡在域名被注册商 `clientHold`。** 这状态下改任何解析记录都不会生效——递归解析器在 `.xyz` 注册局那一层就拿到 NXDOMAIN，根本走不到阿里云。

排查 clientHold（按可能性排序）：

1. **实名认证未通过/未完成**——国内注册域名最常见的 clientHold 原因，`clientRenewProhibited` 也常与之同时出现。阿里云控制台 →「域名」→ `youzhichong.xyz` → 看「实名认证」一栏。
2. **注册邮箱未验证**——查阿里云账号邮箱有没有 ICANN 的验证邮件（未点链接验证，宽限期后会被 hold）。
3. **投诉/滥用或欠费**——`.xyz` 被滥用投诉较多，注册局可能挂起。
4. 前三条都排查完仍不明原因，直接提工单，附上 `clientHold` 这个状态词，让客服给具体原因。

> 时间线佐证：域名 `2026-06-02` 注册，RDAP 里 `last changed` 是 `2026-07-01`，即约一个月后被改了状态，符合「实名认证超期未通过」的典型节奏。

---

## 第 0 步：先确认备案（这一步不通，后面全白做）

国内服务器的 **80/443 端口对未备案域名会被拦截**，表现是浏览器打不开、curl 超时或返回云厂商的备案提示页。技术配置再对也没用。

**怎么查自己有没有备案**

1. 登录阿里云控制台 → 搜索「备案」→ 进入 ICP 备案控制台，看有没有已备案的订单。
2. 或用工信部公共查询：<https://beian.miit.gov.cn> → 「ICP备案查询」，输入 `youzhichong.xyz`。

**如果没备案**，需要准备：

| 条件 | 说明 |
|---|---|
| 服务器 | 包年包月 ECS 才有免费「备案服务号」；按量付费/EIP 单独购买的情况要先确认 |
| 域名 | `.xyz` 可以备案，但域名必须已完成**实名认证**（阿里云域名实名一般 1~2 天） |
| 主体 | 身份证、手机号、应急联系方式；个人备案网站名称不能含「博客」以外的经营性字样 |
| 时长 | 阿里云初审 1~2 个工作日 + 管局审核 7~20 个工作日 |

**备案期间怎么先验证项目**：用 `服务器公网IP:8080` 直接访问（见第 5 步的临时方案），不要指望域名。

> 如果你的服务器在**香港/新加坡**等境外区域，跳过本节，域名配好解析即可上线。

---

## 第 1 步：配置 DNS 解析

阿里云云解析 → 域名列表 → `youzhichong.xyz` → 解析设置 → 添加记录：

| 记录类型 | 主机记录 | 记录值 | TTL |
|---|---|---|---|
| A | `@` | 你的服务器公网 IP | 600 |
| A | `www` | 你的服务器公网 IP | 600 |

验证（在**本地**执行）：

```bash
dig +short youzhichong.xyz A
dig +short www.youzhichong.xyz A
# 两条都应返回你的服务器 IP
```

> 我本地查过：当前 `youzhichong.xyz` **没有任何 A 记录**，所以要先把这两条加上。
> 解析生效通常 1~10 分钟，TTL 600 的话改错也能较快回滚。

---

## 第 2 步：服务器初始化

以下命令在服务器上以有 sudo 权限的账号执行（Ubuntu/Debian 系）。

```bash
# 1. 专用低权限用户
sudo useradd -r -m -s /bin/bash blog

# 2. 目录
sudo mkdir -p /opt/blog /var/www/blog /opt/blog/backend/bin /opt/blog/backend/uploads /backup
sudo chown -R blog:blog /opt/blog /var/www/blog

# 3. 装依赖
sudo apt update
sudo apt install -y nginx mysql-server curl rsync

# 4. 建库与账号（密码换成强密码）
sudo mysql <<'SQL'
CREATE DATABASE IF NOT EXISTS blog DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
CREATE USER IF NOT EXISTS 'blog'@'localhost' IDENTIFIED BY '换成强密码';
GRANT ALL PRIVILEGES ON blog.* TO 'blog'@'localhost';
FLUSH PRIVILEGES;
SQL
```

**放行端口**（两处都要）：

- 阿里云控制台 → 安全组 → 入方向放行 `80`、`443`（`22` 通常已放行）
- 服务器本机：`sudo ufw allow 22,80,443/tcp`（若启用了 ufw）

---

## 第 3 步：构建

**方式 A：本地构建（推荐，服务器不用装 Go/Node）**

```bash
# 在 Mac 上、项目根目录执行
./deploy/build.sh
```

`deploy/build.sh` 已改为**交叉编译到 linux/amd64**（`CGO_ENABLED=0`，纯静态二进制），产物：

- `backend/bin/blog-server`（22MB，可直传 Linux）
- `frontend/dist/`（1.0MB 静态文件）

> **实测结果**：在你这台 Mac（arm64）上跑 `./deploy/build.sh` 全流程通过，
> 产物为 `ELF 64-bit LSB executable, x86-64, statically linked`，
> 即**可直接在 x86_64 Linux 上运行**，不依赖服务器上的 glibc。

> 前置：本地要能跑通 `go build`（`go.mod` 要求 Go 1.26.4）和 `pnpm build`。
>
> 这台 Mac 本机是 **go 1.24.5**，低于 `go.mod` 要求的 1.26.4，首次构建会自动下载 toolchain。
> 而默认的 sumdb 校验地址 `sum.golang.google.cn` 会连接超时（`dial tcp ... i/o timeout`），
> 只设 `GOPROXY` 不够，**必须再设 `GOSUMDB=sum.golang.org`** 让代理代答校验：
>
> ```bash
> export GOPROXY=https://goproxy.cn,direct
> export GOSUMDB=sum.golang.org   # 关键；设成 off 反而会报 "checksum database disabled"
> go version                      # 首次会下载 go1.26.4，约 1~2 分钟
> ```
>
> 前端还需要 pnpm（node 已有 v25.2.1，pnpm 未装）：
>
> ```bash
> npm i -g pnpm          # 或临时用：npx pnpm@10 ...
> ```
>
> 实在折腾不动就走方式 B（服务器上构建）。

**方式 B：服务器上构建**

服务器装 Go 1.26+ 和 Node 20+，把仓库拉上去（或 rsync 源码）后：

```bash
cd /opt/blog && ./deploy/build.sh
# 服务器上 GOOS 自动是 linux，无需交叉编译
```

---

## 第 4 步：上传产物

```bash
# 在本地项目根目录执行，server 换成你的服务器地址
rsync -avz backend/bin/blog-server   blog@server:/opt/blog/backend/bin/
rsync -avz backend/migrations/       blog@server:/opt/blog/backend/
rsync -avz frontend/dist/            blog@server:/var/www/blog/dist/
# 部署脚本与配置模板（第 5、7、8 步要用到）
rsync -avz deploy/                   blog@server:/opt/blog/deploy/
```

> 升级时二进制**先传临时名再 mv**，避免替换正在运行的文件被截断：
> `rsync ... blog-server.new` → `mv blog-server.new blog-server` → `systemctl restart blog-server`

---

## 第 5 步：配置生产环境（.env.prod）

项目里现在有两套配置，**通过 `ENV_FILE` 切换，互不干扰**：

| 文件 | 口径 | 用途 | 是否入库 |
|---|---|---|---|
| `backend/.env` | `APP_ENV=dev`、相对 `UPLOAD_DIR`、`TRUSTED_PROXIES` 空 | 本地开发 | ❌ 被 `.gitignore` 挡住 |
| `backend/.env.prod` | `APP_ENV=prod`、绝对路径、`TRUSTED_PROXIES=127.0.0.1`、新 JWT | 线上 | ❌ 同样被挡住 |
| `backend/.env.example` | 模板 | 告诉你要配哪些项 | ✅ 入库 |

后端读哪个文件由 `ENV_FILE` 决定（[config.go](../backend/internal/config/config.go)）：

```go
envFile, explicit := os.LookupEnv("ENV_FILE")
if !explicit {
    envFile = ".env"          // 不设就用默认的 .env
}
```

- 不设 `ENV_FILE` → 读 `backend/.env`（本地开发）
- `ENV_FILE=.env.prod` → 读 `backend/.env.prod`（线上）
- **设了却找不到文件 → 直接报错退出**，不会静默退回默认值（否则很容易"以为走了生产配置，其实用的是默认值"）

### 5.1 上传生产配置

```bash
rsync -avz backend/.env.prod blog@server:/opt/blog/backend/.env.prod
ssh blog@server 'chown blog:blog /opt/blog/backend/.env.prod && chmod 600 /opt/blog/backend/.env.prod'
```

### 5.2 让 systemd 用它

[deploy/blog-server.service](../deploy/blog-server.service) 里已经配好：

```ini
Environment=ENV_FILE=/opt/blog/backend/.env.prod
```

`systemctl restart blog-server` 之后走的就是生产配置。

### 5.3 如果服务器上还没有 .env.prod

在服务器上直接生成（自动随机 `JWT_SECRET`、写死生产口径）：

```bash
# 用第 2 步建库时设的密码
sudo -u blog /opt/blog/deploy/gen-env.sh --out /opt/blog/backend/.env.prod --db-password '你的数据库密码'

# 或让脚本生成随机密码，它会打印对应的建库命令
sudo -u blog /opt/blog/deploy/gen-env.sh --out /opt/blog/backend/.env.prod
```

脚本已经处理好这几个最容易漏的项：

| 变量 | 值 | 漏了会怎样 |
|---|---|---|
| `APP_ENV` | `prod` | `/debug/*` 路由和 SQL 打印不关 |
| `TRUSTED_PROXIES` | `127.0.0.1` | 评论 IP 全是 127.0.0.1，限流误伤所有人 |
| `UPLOAD_DIR` | `/opt/blog/backend/uploads` | 相对路径依赖工作目录，图片可能丢失 |
| `JWT_SECRET` | 新随机值 | 本地值已随文件外传，必须换 |

> 已验证：`ENV_FILE` 指向 prod 配置时程序确实读了该文件——用不同 `DB_PORT` 区分（默认读 3306、指定后读 3307）；指向不存在的文件会报 `加载配置失败: ENV_FILE 指定的配置文件不存在`。

### 5.4 另一种选择：配置放 /etc

不想让配置留在代码目录（回滚代码、重传产物时更安全）：

```bash
sudo mkdir -p /etc/blog
sudo /opt/blog/deploy/gen-env.sh --out /etc/blog/blog.env --db-password '你的密码'
sudo chown root:root /etc/blog/blog.env && sudo chmod 600 /etc/blog/blog.env
# 编辑 blog-server.service：注释掉 Environment=ENV_FILE=...，改成
#   EnvironmentFile=/etc/blog/blog.env
sudo systemctl daemon-reload && sudo systemctl restart blog-server
```

两条路**选一条**即可。同时用的话，`EnvironmentFile` 注入的环境变量优先级更高，会覆盖 `.env.prod` 里的同名项。


---

## 第 6 步：执行数据库迁移

**按编号顺序、只执行一次**（`migrations/` 里的文件执行过就不要改）：

```bash
cd /opt/blog/backend
for f in migrations/00*.sql; do
  echo "==> $f"
  # --default-character-set=utf8mb4 不能省：服务器上 LANG 一般是空的，
  # mysql 客户端默认 latin1，中文会被双重编码存进库，页面显示成乱码。
  mysql -ublog -p --default-character-set=utf8mb4 blog < "$f"
done
```

> 迁移要求 MySQL 8.0+（用了 `utf8mb4_0900_ai_ci`）。

---

## 第 7 步：启动后端

```bash
sudo cp /opt/blog/deploy/blog-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now blog-server

systemctl status blog-server
journalctl -u blog-server -f     # 看日志

# 本机自检
curl -s http://127.0.0.1:8090/api/v1/ping
```

---

## 第 8 步：配置 Nginx

```bash
sudo cp /opt/blog/deploy/nginx-youzhichong.xyz.conf /etc/nginx/conf.d/blog.conf
sudo rm -f /etc/nginx/sites-enabled/default      # 避免默认站点抢 80
sudo nginx -t && sudo systemctl reload nginx
```

---

## 第 9 步：上 HTTPS

**必须在备案通过、且第 8 步 `http://youzhichong.xyz` 能打开之后**再执行（certbot 要通过 80 端口验证域名）。

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d youzhichong.xyz -d www.youzhichong.xyz
```

certbot 会自动改写 Nginx 配置、加 443 和 80→443 跳转。续期验证：

```bash
sudo certbot renew --dry-run
```

> 未备案时先用 IP 验证：把 Nginx 的 `listen 80` 临时改成 `listen 8080`，然后访问 `http://服务器IP:8080`。
> HTTPS 在备案前无法签发（80 验证走不通），这是备案的另一个硬约束。

---

## 第 10 步：验收

```bash
# 1. 服务活着
curl -s http://127.0.0.1:8090/api/v1/ping

# 2. 数据库通
curl -s http://127.0.0.1:8090/api/v1/articles | head -c 200

# 3. debug 路由已关闭（prod 应为 404）
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8090/api/v1/debug/db

# 4. 后台接口未登录被拒（应 401）
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8090/api/v1/admin/articles

# 5. 域名访问前端
curl -s -o /dev/null -w "%{http_code}\n" https://youzhichong.xyz/

# 6. SPA 回退生效（刷新详情页不 404）
curl -s -o /dev/null -w "%{http_code}\n" https://youzhichong.xyz/article/1

# 7. 真实 IP 透传：发一条评论后查库，ip 不应是 127.0.0.1
mysql -ublog -p blog -e "SELECT id,nickname,ip FROM comments ORDER BY id DESC LIMIT 1;"
```

浏览器再手动过一遍：首页列表 → 文章详情（代码高亮）→ 评论提交 → 登录 → 后台发文 → 上传图片。

---

## 常见坑（按出现概率排序）

| 现象 | 原因 | 处理 |
|---|---|---|
| 域名打不开，IP 能开 | 未备案被拦 | 先备案 |
| 502 Bad Gateway | Go 没起或没监听 8090 | `systemctl status blog-server`、`journalctl -u blog-server -n 50` |
| 刷新详情页 404 | Nginx 缺 `try_files ... /index.html` | 用第 8 步的配置 |
| 上传报 413 | `client_max_body_size` 小于上传大小 | 已配 8m，仍不够就同步调大 `UPLOAD_MAX_MB` |
| 评论 IP 全是 127.0.0.1 | `TRUSTED_PROXIES` 没配 | `.env` 设 `127.0.0.1` 后重启 |
| 二进制报 `GLIBC not found` | 动态链接 + 服务器 glibc 旧 | 用改过的 `build.sh`（`CGO_ENABLED=0`）重新编译 |
| 发版后页面还是旧的 | `index.html` 被缓存 | 配置里已对 `index.html` 设 no-cache |
| 中文乱码 | 库/表/连接字符集不一致 | 库用 `utf8mb4`，客户端 `default-character-set=utf8mb4` |
