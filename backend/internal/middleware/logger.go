package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 记录每个请求的 request id、方法、路径、状态码、耗时、来源 IP。
// 耗时必须包住 c.Next()：Next 之前记开始时间，Next 返回说明后面的处理都结束了。
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next() // 先放行，等后续所有 handler 执行完才继续往下走

		latency := time.Since(start)
		log.Printf("[HTTP] %s | %3d | %-6s %-30s | %8s | %s",
			c.GetString(ContextKeyRequestID),
			c.Writer.Status(),
			c.Request.Method,
			c.Request.URL.Path,
			latency,
			c.ClientIP(),
		)
	}
}
