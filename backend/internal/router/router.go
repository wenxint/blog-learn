// Package router 把「哪个网址交给哪个函数」登记在一起。
// main 只负责启动进程，不在 main 里堆路由。
package router

import (
	"os"
	"time"

	"blog/internal/handler"
	"blog/internal/middleware"
	"blog/internal/pkg/jwt"
	"blog/internal/pkg/response"
	"blog/internal/repository"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// New 组装路由，返回 Gin 引擎。main 把它塞进 http.Server 的 Handler 字段。
// db 通过参数传进来，handler 不自己去连数据库，也不用全局变量。
// 这样换一个数据库对象（比如测试里的）时，不用改 handler 内部。
// jwtMgr 同理：中间件需要它验签，但 middleware 包不该自己去读配置。
// uploadSvc 负责把上传的文件落到磁盘，router 只负责把 /uploads 这个路径映射过去。
func New(db *gorm.DB, jwtMgr *jwt.Manager, uploadSvc *service.UploadService) *gin.Engine {
	// 生产模式关掉 Gin 的调试路由打印。开发环境保持默认的 DebugMode。
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.New() 是空引擎，不自带日志和崩溃恢复，下面的中间件是我们自己挂的。
	// gin.Default() 会自带一份，和我们的 Logger / Recovery 重复，所以不用它。
	r := gin.New()

	// 上传文件在内存里缓冲的上限：超过 8MB 的部分会落到临时文件，
	// 防止有人传大文件把内存吃光
	r.MaxMultipartMemory = 8 << 20

	// 不信任任何代理，ClientIP 直接取 RemoteAddr（避免伪造 X-Forwarded-For）
	_ = r.SetTrustedProxies(nil)

	// 把上传目录挂成静态资源：/uploads/2026/11/xxxx.png 直接可访问。
	// 注意这里只允许 GET，而且只能读 uploads 目录内部，不会读到源码。
	r.Static("/uploads", uploadSvc.Dir())

	// 全局中间件：每个请求都按这个顺序穿过，再进入具体接口。
	// 可以想成安检：先发编号，再记日志，再兜住崩溃，再处理跨域。
	// RequestID 必须在最外层，Logger / Recovery 才能读到它写进 Context 的 ID。
	r.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.CORS(),
	)

	// 未匹配的路由也要返回统一 JSON，否则前端拦截器解析会崩
	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "接口不存在: "+c.Request.Method+" "+c.Request.URL.Path)
	})

	// ---------- 依赖装配 ----------
	// 这是全项目唯一「谁依赖谁」的地方，从上往下读就是调用方向：
	//   handler -> service -> repository -> db
	// 每一层只拿到自己需要的东西：
	//   repository 只拿到 db（它只管存取数据）
	//   service 拿到 db + repository（它要开事务，也要调数据访问）
	//   handler 只拿到 service（它不该知道数据库长什么样）
	articleRepo := repository.NewArticleRepository(db)
	articleSvc := service.NewArticleService(db, articleRepo)
	articleHandler := handler.NewArticleHandler(articleSvc)

	// 还没分层的 handler：仍然直接用 db（category、debug 属于查看类，暂不重构）
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(db, userRepo, jwtMgr)
	authHandler := handler.NewAuthHandler(authSvc)

	// 分类与标签也走同样的三层：repo -> service -> handler
	categoryRepo := repository.NewCategoryRepository(db)
	categorySvc := service.NewCategoryService(categoryRepo)
	categoryHandler := handler.NewCategoryHandler(categorySvc)

	tagRepo := repository.NewTagRepository(db)
	tagSvc := service.NewTagService(tagRepo)
	tagHandler := handler.NewTagHandler(tagSvc)

	commentRepo := repository.NewCommentRepository(db)
	commentSvc := service.NewCommentService(commentRepo, articleRepo)
	commentHandler := handler.NewCommentHandler(commentSvc)

	uploadHandler := handler.NewUploadHandler(uploadSvc)

	// debug 是纯诊断接口，不涉及业务规则，直接用 db
	debugHandler := handler.NewDebugHandler(db)

	// 路由分组：花括号里登记的路径都会自动加上 /api/v1。
	// 以后要出 v2，另开一个 Group，旧接口可以留着。
	api := r.Group("/api/v1")
	{
		// GET 对应浏览器地址栏、curl 不带 -X 时的那种请求。
		// 第二个参数是处理函数：c 里装着这次请求的全部信息（路径、参数、响应）。
		api.GET("/ping", func(c *gin.Context) {
			// gin.H 是 map[string]any 的简写，用来临时组一个 JSON 对象。
			response.OK(c, gin.H{"pong": true})
		})

		// 前台只读接口：浏览器、未登录用户都能访问
		api.GET("/articles", articleHandler.List)
		// /:id 是路径参数，curl /articles/1 时 handler 里用 c.Param("id") 取到 "1"
		api.GET("/articles/:id", articleHandler.Detail)
		api.GET("/categories", categoryHandler.List)
		api.GET("/tags", tagHandler.List)
		// 归档是跨文章的聚合统计，路径放在 /archives 而不是 /articles/archives，
		// 避免和 /articles/:id 的 :id 段冲突
		api.GET("/archives", articleHandler.Archives)

		// 评论：前台可读可写。写入后进入待审核，不直接展示。
		api.GET("/articles/:id/comments", commentHandler.List)
		api.POST("/articles/:id/comments", commentHandler.Create)

		// 认证接口。分组前缀 /api/v1/auth。
		// register / login 必须公开：用户还没登录，哪来的 token。
		// profile / password 需要登录：在单个路由上挂 AuthRequired 中间件，
		// 这样比再开一个 group 更直观——看这一行就知道它要不要登录。
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile", middleware.AuthRequired(jwtMgr), authHandler.Profile)
			auth.PUT("/password", middleware.AuthRequired(jwtMgr), authHandler.ChangePassword)
		}

		// 后台写接口。
		// Group("/admin") 相当于给里面所有路由再加一段前缀 /api/v1/admin。
		// 单独分组的意义：第 10 课会给这个组整体挂上 JWT 鉴权中间件，
		// 一行代码就能保护组内所有接口，不用逐个加。
		admin := api.Group("/admin")
		// 整个后台组统一挂鉴权：组内每个接口都必须带 Bearer token。
		// 写在一处而不是逐个加，就不会出现「新加了接口忘记鉴权」这种低级漏洞。
		admin.Use(middleware.AuthRequired(jwtMgr))
		{
			// POST / PUT / DELETE 是「有副作用」的请求：
			// 浏览器地址栏访问不到，必须用表单、fetch 或 curl -X 指定。
			// 图片上传：需要登录（author 也能传，写文章时要配图）
			admin.POST("/upload", uploadHandler.Upload)

			admin.POST("/articles", articleHandler.Create)
			admin.PUT("/articles/:id", articleHandler.Update)
			// 删除是危险操作，额外要求 admin 角色（author 只能写，不能删）
			admin.DELETE("/articles/:id", middleware.AdminOnly(), articleHandler.Delete)

			// 分类和标签的增删改都只给 admin：
			// 它们是全站共享的分类体系，author 随手改会打乱所有人的文章归类
			admin.POST("/categories", middleware.AdminOnly(), categoryHandler.Create)
			admin.PUT("/categories/:id", middleware.AdminOnly(), categoryHandler.Update)
			admin.DELETE("/categories/:id", middleware.AdminOnly(), categoryHandler.Delete)

			// 评论审核
			admin.GET("/comments", commentHandler.AdminList)
			admin.PUT("/comments/:id/status", middleware.AdminOnly(), commentHandler.SetStatus)
			admin.DELETE("/comments/:id", middleware.AdminOnly(), commentHandler.Delete)

			admin.POST("/tags", tagHandler.Create)
			admin.PUT("/tags/:id", tagHandler.Update)
			admin.DELETE("/tags/:id", middleware.AdminOnly(), tagHandler.Delete)
		}

		// 仅开发环境暴露的诊断接口
		if os.Getenv("APP_ENV") != "prod" {
			api.GET("/debug/db", debugHandler.DB)
			api.GET("/debug/models", debugHandler.Models)
			// 验证事务回滚：故意在事务里失败，看数据有没有留下来
			api.GET("/debug/tx-rollback", debugHandler.TxRollback)
			api.GET("/debug/panic", func(c *gin.Context) {
				panic("这是一次故意的 panic，用来验证 Recovery 中间件")
			})
			// 慢接口：用来验证优雅关闭时进行中的请求会被排空而不是被砍断
			api.GET("/debug/slow", func(c *gin.Context) {
				const sec = 3
				select {
				case <-time.After(sec * time.Second):
					response.OK(c, gin.H{"slept_seconds": sec})
				case <-c.Request.Context().Done():
					// 客户端主动断开
					return
				}
			})
		}
	}

	return r
}
