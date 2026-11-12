// Package router 把「哪个网址交给哪个函数」登记在一起。
// main 只负责启动进程，不在 main 里堆路由。
package router

import (
	"os"
	"time"

	"blog/internal/handler"
	"blog/internal/middleware"
	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// New 组装路由，返回 Gin 引擎。main 把它塞进 http.Server 的 Handler 字段。
// db 通过参数传进来，handler 不自己去连数据库，也不用全局变量。
// 这样换一个数据库对象（比如测试里的）时，不用改 handler 内部。
func New(db *gorm.DB) *gin.Engine {
	// 生产模式关掉 Gin 的调试路由打印。开发环境保持默认的 DebugMode。
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// gin.New() 是空引擎，不自带日志和崩溃恢复，下面的中间件是我们自己挂的。
	// gin.Default() 会自带一份，和我们的 Logger / Recovery 重复，所以不用它。
	r := gin.New()

	// 不信任任何代理，ClientIP 直接取 RemoteAddr（避免伪造 X-Forwarded-For）
	_ = r.SetTrustedProxies(nil)

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

	// 依赖注入：需要 db 的 handler 在这里创建
	articleHandler := handler.NewArticleHandler(db)
	categoryHandler := handler.NewCategoryHandler(db)
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

		api.GET("/articles", articleHandler.List)
		api.GET("/categories", categoryHandler.List)

		// 仅开发环境暴露的诊断接口
		if os.Getenv("APP_ENV") != "prod" {
			api.GET("/debug/db", debugHandler.DB)
			api.GET("/debug/models", debugHandler.Models)
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
