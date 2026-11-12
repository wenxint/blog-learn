package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// HeaderRequestID 响应头名字
const HeaderRequestID = "X-Request-Id"

// ContextKeyRequestID 在 gin.Context 里存取 request id 的 key
const ContextKeyRequestID = "request_id"

// RequestID 给每个请求生成唯一 ID：
//  1. 客户端已带 X-Request-Id 就沿用（便于跨服务串联）
//  2. 没有就生成一个 16 位十六进制随机串
//  3. 写进响应头 + 存进 Context，供 Logger / Recovery / 业务代码读取
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderRequestID)
		if id == "" {
			id = newRequestID()
		}

		c.Set(ContextKeyRequestID, id)
		c.Header(HeaderRequestID, id)

		c.Next()
	}
}

func newRequestID() string {
	var b [8]byte // 8 字节 = 16 位十六进制
	if _, err := rand.Read(b[:]); err != nil {
		// 极端情况下 crypto/rand 不可用，退化成时间戳，保证不 panic
		return "fallback"
	}
	return hex.EncodeToString(b[:])
}
