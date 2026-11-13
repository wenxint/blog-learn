// Package response 规定所有接口都返回同一种 JSON 外壳：
//
//	{"code":0,"message":"ok","data":{...}}
//
// HTTP 状态码（200/400/401/...）和业务码 code 同时设置。
// 网关、浏览器看 HTTP 状态码；前端业务逻辑只认 code。
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
	CodeTooManyReqs  = 10007 // 请求过于频繁     -> 429
)

// Body 是所有接口统一的响应结构。
// 反引号里的 json:"code" 叫结构体标签：c.JSON 把它转成 JSON 时用这个名字，
// 而不是 Go 字段名 Code。前端拿到的键是小写 code。
type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"` // any 表示「任意类型」：列表、对象、null 都能放进来
}

// OK 成功响应，HTTP 200。data 会放进 Body.Data。
func OK(c *gin.Context, data any) {
	// c.JSON 设置 Content-Type 并写出 JSON。StatusOK 就是数字 200，用常量避免满篇魔法数字。
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

// TooManyRequests 限流：告诉调用方「慢点」，配合 Retry-After 头体验更好
func TooManyRequests(c *gin.Context, message string) {
	Fail(c, http.StatusTooManyRequests, CodeTooManyReqs, message)
}
