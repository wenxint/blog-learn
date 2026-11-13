package handler

import (
	"errors"
	"log"

	"blog/internal/pkg/response"
	"blog/internal/service"

	"github.com/gin-gonic/gin"
)

// fail 把 service 的错误统一翻译成 HTTP 响应，所有 handler 共用。
//
// 分层原则：service 只表达「业务上出了什么事」，HTTP 状态码是接口层的事。
// 这样同一份 service 换成 gRPC / 命令行调用时，错误语义不变。
//
// 安全原则：只有业务错误（用户能理解、能自己改）才把原文返回；
// 其他错误（数据库挂了、SQL 报错）只回笼统信息，细节写日志。
// 否则错误信息里的表名/字段名/SQL 就是免费的情报。
func fail(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.NotFound(c, err.Error()) // 404
	case errors.Is(err, service.ErrUnauthorized):
		response.Unauthorized(c, err.Error()) // 401
	case errors.Is(err, service.ErrConflict):
		response.Conflict(c, err.Error()) // 409
	case errors.Is(err, service.ErrInvalidInput):
		response.InvalidParam(c, err.Error()) // 400
	default:
		log.Printf("[ERROR] %s: %v", fallback, err)
		response.ServerError(c, fallback) // 500
	}
}
