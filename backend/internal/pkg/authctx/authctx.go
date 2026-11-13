// Package authctx 定义「当前登录用户」在 gin.Context 里的存取方式。
//
// 为什么单独抽一个包，而不是在 middleware 里定义常量让 handler import：
// handler 一旦 import middleware，架构上就变成 handler 依赖中间件（反向依赖），
// 两者都依赖 authctx 更干净，也不会出现循环 import。
package authctx

import "github.com/gin-gonic/gin"

const (
	keyUserID   = "auth.user_id"
	keyUsername = "auth.username"
	keyRole     = "auth.role"
)

// Set 由鉴权中间件在验签成功后调用
func Set(c *gin.Context, userID uint64, username, role string) {
	c.Set(keyUserID, userID)
	c.Set(keyUsername, username)
	c.Set(keyRole, role)
}

// UserID 取当前登录用户 id；未登录返回 0
func UserID(c *gin.Context) uint64 {
	if v, ok := c.Get(keyUserID); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

// Username 取当前登录用户名
func Username(c *gin.Context) string {
	return c.GetString(keyUsername)
}

// Role 取当前登录用户角色（admin / author）
func Role(c *gin.Context) string {
	return c.GetString(keyRole)
}
