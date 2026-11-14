# 笔记索引

> 文档总入口见 [../README.md](../README.md)，成稿文章见 [../articles/](../articles/)。

这个目录放「随时要翻」的东西，不放进架构文档。

| 文件 | 什么时候看 |
|---|---|
| [踩坑记录.md](踩坑记录.md) | 遇到诡异报错先翻它（字符集、端口占用、Makefile 缩进、Homebrew 等） |
| [后端速查.md](后端速查.md) | 要新增一个接口时，照着「五步走」抄 |
| [前端速查.md](前端速查.md) | 要新增一个页面时，照着抄 |
| [自测题.md](自测题.md) | 复习 / 面试前过一遍，答不上来就是没懂 |

---

## 这个项目一句话总结

```
浏览器 → Vue3 页面 → Axios(拦截器带 token) → Vite/Nginx 代理
      → Gin(中间件链) → handler(参数) → service(业务+事务) → repository(数据) → GORM → MySQL
```

**记住这条链路，任何「数据为什么不对」都能按方向排查：**
1. 数据不对 → 先看数据库里本来是什么（`mysql -e "SELECT ..."`）
2. 数据库对但接口不对 → 看 service 逻辑 / 是否被软删除过滤
3. 接口对但页面不对 → 看 Network 面板 + Console
4. 全对但偶尔错 → 看并发 / 事务 / N+1

---

## 常用命令

```bash
# 后端
cd backend
make run        # 启动
make test       # 单测
make db         # 进 MySQL
curl -s localhost:8090/api/v1/debug/db | python3 -m json.tool   # 数据库健康

# 前端
cd frontend
pnpm dev        # 开发
pnpm build      # 构建

# 一键看后端到底注册了哪些路由（启动日志里）
make run 2>&1 | grep "GIN-debug.*-->"
```

---

## 排查顺序（血泪版）

| 现象 | 先查什么 |
|---|---|
| 改了代码没生效 | **端口是否被旧进程占用**（`lsof -nP -iTCP:8090 -sTCP:LISTEN`） |
| 接口 404 | `grep "GIN-debug.*-->"` 看路由到底注册了什么 |
| 接口 401 | token 是否过期 / `Authorization` 格式是否 `Bearer xxx` |
| 接口 403 | 当前用户 role 是不是 admin |
| 中文乱码 | 三层字符集（服务端 / 表 / **客户端连接**） |
| 前端请求失败 | Network 面板看真实状态码，别只看 Console |
| 数据莫名少了一行 | 是不是被软删除了（`deleted_at` 有值） |
