package middleware

import (
	"log"
	"runtime/debug"

	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery 捕获 panic，打印堆栈，返回统一的 500 JSON（而不是让连接直接断开）
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("[PANIC] %v\n%s", err, debug.Stack())
		response.ServerError(c, "服务器内部错误")
		c.Abort()
	})
}
