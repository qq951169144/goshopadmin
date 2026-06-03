package middleware

import (
	"time"

	"search-service/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 请求日志中间件
// 记录每个 HTTP 请求的方法、路径、状态码和响应时间
// 便于排查问题和性能分析
// ============================================================

// RequestLogger 请求日志中间件
// 在请求处理完成后记录日志，包含以下信息：
// - 请求方法（GET/POST 等）
// - 请求路径
// - 响应状态码
// - 请求处理耗时
func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 记录请求开始时间
		startTime := time.Now()

		// 获取请求路径（在中间件执行前获取，避免路径被后续处理修改）
		path := ctx.Request.URL.Path

		// 获取请求方法
		method := ctx.Request.Method

		// 如果有查询参数，也记录下来
		if ctx.Request.URL.RawQuery != "" {
			path = path + "?" + ctx.Request.URL.RawQuery
		}

		// 继续处理后续中间件和路由
		ctx.Next()

		// 请求处理完成后，计算耗时
		latency := time.Since(startTime)

		// 获取响应状态码
		statusCode := ctx.Writer.Status()

		// 记录请求日志
		utils.Info("%s %s %d %v", method, path, statusCode, latency)
	}
}
