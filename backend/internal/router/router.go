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

// New 组装路由。main.go 只负责启动，路由配置都在这里。
// db 通过参数注入，handler 不直接用全局变量——这是后面分层重构的基础。
func New(db *gorm.DB) *gin.Engine {
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 不信任任何代理，ClientIP 直接取 RemoteAddr（避免伪造 X-Forwarded-For）
	_ = r.SetTrustedProxies(nil)

	// 全局中间件：顺序即执行顺序
	// RequestID 必须在最外层，Logger / Recovery 才能读到它写进 Context 的 ID
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

	// 路由分组：所有接口统一 /api/v1 前缀
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
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
