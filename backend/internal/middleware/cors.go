// Package middleware 放「每个请求都要经过、但跟具体业务无关」的处理。
// Gin 的中间件本质是一个函数：它收到 c，干自己的事，然后调用 c.Next()
// 把请求交给下一个中间件或最终的 handler；Next 返回后，还能继续写「请求结束之后」的逻辑。
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件（开发期简单版：请求从哪个网站来，就允许哪个网站）。
// 浏览器里，网页在 :5173、接口在 :8090，端口不同就算跨域，不带下面这些响应头会被浏览器拦下。
// curl 不存在这个限制。生产环境应改成白名单，不能谁来都放行，见第 20 课。
//
// 返回值是 gin.HandlerFunc。函数套函数：外层 CORS() 在启动时调用一次，
// 内层 func(c) 才是每个请求真正执行的那段。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Origin 是浏览器自动带上的「我从哪个网站发出的」。没有它，多半不是浏览器跨域请求。
		origin := c.GetHeader("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// 浏览器在真正发 POST / 带自定义头之前，会先发一次 OPTIONS，问服务器允不允许。
		// 这种预检没有业务要处理，直接 204（成功但没有响应体），并且 Abort 掉，不再往下走。
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Next 把请求交给后面的中间件和 handler。不写 Next，请求就停在这里。
		c.Next()
	}
}
