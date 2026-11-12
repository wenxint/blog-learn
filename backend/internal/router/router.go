package router

import (
	"os"

	"blog/internal/handler"
	"blog/internal/middleware"
	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// New 组装路由。main.go 只负责启动，路由配置都在这里。
func New() *gin.Engine {
	if os.Getenv("APP_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 全局中间件：顺序即执行顺序
	r.Use(middleware.Logger(), middleware.Recovery(), middleware.CORS())

	// 路由分组：所有接口统一 /api/v1 前缀
	api := r.Group("/api/v1")
	{
		api.GET("/ping", func(c *gin.Context) {
			response.OK(c, gin.H{"pong": true})
		})

		api.GET("/articles", handler.List)

		// 仅开发环境暴露，用于验证 panic 恢复是否生效
		if os.Getenv("APP_ENV") != "prod" {
			api.GET("/debug/panic", func(c *gin.Context) {
				panic("这是一次故意的 panic，用来验证 Recovery 中间件")
			})
		}
	}

	return r
}
