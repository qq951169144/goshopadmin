package middleware

import (
	"github.com/gin-gonic/gin"
)

// ============================================================
// CORS 跨域中间件
// 允许前端跨域访问搜索服务 API
// 在开发环境中前端和后端通常运行在不同端口，需要 CORS 支持
// ============================================================

// CORS 跨域资源共享中间件
// 设置允许的 Origin、Methods、Headers 等
// 允许所有来源访问（生产环境建议限制具体域名）
func CORS() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Access-Control-Allow-Origin: 允许所有来源访问
		// 生产环境应替换为具体的前端域名
		ctx.Header("Access-Control-Allow-Origin", "*")

		// Access-Control-Allow-Methods: 允许的 HTTP 方法
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Access-Control-Allow-Headers: 允许的请求头
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		// Access-Control-Expose-Headers: 允许前端读取的响应头
		ctx.Header("Access-Control-Expose-Headers", "Content-Length")

		// Access-Control-Max-Age: 预检请求缓存时间，减少 OPTIONS 请求次数
		ctx.Header("Access-Control-Max-Age", "86400")

		// 处理 OPTIONS 预检请求
		// 浏览器在发送跨域请求前会先发送 OPTIONS 请求
		// 直接返回 204 No Content 即可
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}

		// 继续处理后续中间件和路由
		ctx.Next()
	}
}
