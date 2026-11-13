package middleware

import (
	"strings"

	"blog/internal/pkg/authctx"
	"blog/internal/pkg/jwt"
	"blog/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthRequired 校验 Authorization: Bearer <token>。
// 校验通过后把用户信息放进 Context，后面的 handler 用 authctx.UserID(c) 取。
// 校验失败必须 c.Abort()，否则会继续执行后面的 handler（这是新手最常犯的错）。
func AuthRequired(mgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			response.Unauthorized(c, "缺少 Authorization 请求头")
			c.Abort()
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(raw, prefix) {
			response.Unauthorized(c, "Authorization 格式应为: Bearer <token>")
			c.Abort()
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(raw, prefix))
		if token == "" {
			response.Unauthorized(c, "token 为空")
			c.Abort()
			return
		}

		claims, err := mgr.Parse(token)
		if err != nil {
			response.Unauthorized(c, err.Error())
			c.Abort()
			return
		}

		authctx.Set(c, claims.UserID, claims.Username, claims.Role)
		c.Next()
	}
}

// AdminOnly 要求当前用户角色是 admin。
//
// 401 和 403 的区别（面试常问）：
//
//	401 Unauthorized = 「你是谁我不知道」→ 请先登录
//	403 Forbidden    = 「我知道你是谁，但你没权限」→ 别登了，登了也没用
//
// 前端处理也不同：401 跳登录页，403 提示「无权限」。
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		if authctx.Role(c) != "admin" {
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
