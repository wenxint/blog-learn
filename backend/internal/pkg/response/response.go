package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 业务错误码：与 HTTP 状态码分离，前端只认 code
const (
	CodeSuccess      = 0     // 成功
	CodeInvalidParam = 10001 // 参数错误         -> 400
	CodeUnauthorized = 10002 // 未登录/token无效 -> 401
	CodeForbidden    = 10003 // 无权限           -> 403
	CodeNotFound     = 10004 // 资源不存在       -> 404
	CodeConflict     = 10005 // 冲突             -> 409
	CodeServerError  = 10006 // 服务器内部错误   -> 500
)

// Body 是所有接口统一的响应结构
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK 成功响应，HTTP 200
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: CodeSuccess, Message: "ok", Data: data})
}

// Fail 失败响应，调用方指定 HTTP 状态码和业务码
func Fail(c *gin.Context, httpStatus int, code int, message string) {
	c.JSON(httpStatus, Body{Code: code, Message: message, Data: nil})
}

// 以下是语义化的快捷方法，避免业务代码里满是魔法数字

func InvalidParam(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, CodeInvalidParam, message)
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, CodeForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, CodeNotFound, message)
}

func Conflict(c *gin.Context, message string) {
	Fail(c, http.StatusConflict, CodeConflict, message)
}

func ServerError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, CodeServerError, message)
}
