package middleware

import (
	"log"
	"runtime/debug"

	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// Recovery 捕获 panic，打印堆栈，返回统一的 500 JSON（而不是让连接直接断开）。
// panic 是 Go 里「这里已经没法继续」的中断。没有这层兜底，一次空指针会把整个进程打死。
// 业务上的可预期失败不要用 panic，用 error 返回值；panic 只留给真的程序错误。
func Recovery() gin.HandlerFunc {
	// CustomRecovery 的回调在 panic 发生时被调用。err 的类型是 any，因为 panic 可以扔任意值。
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Printf("[PANIC] %s | %v\n%s",
			c.GetString(ContextKeyRequestID), err, debug.Stack())
		response.ServerError(c, "服务器内部错误")
		c.Abort()
	})
}
