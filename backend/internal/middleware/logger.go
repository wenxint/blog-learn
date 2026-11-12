package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 记录每个请求的方法、路径、状态码、耗时、来源 IP
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // 先放行，等后续所有 handler 执行完才继续往下走

		latency := time.Since(start)
		log.Printf("[HTTP] %3d | %-6s %-30s | %8s | %s",
			c.Writer.Status(),
			c.Request.Method,
			c.Request.URL.Path,
			latency,
			c.ClientIP(),
		)
	}
}
