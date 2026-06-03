package http

import "github.com/gin-gonic/gin"

// AuthMiddleware 认证中间件（简化版）
// 实际项目中会验证 JWT Token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(200, gin.H{
				"code":    4010,
				"message": "未认证",
				"data":    nil,
			})
			c.Abort()
			return
		}

		// 简化：从 header 中解析客户ID
		// 实际项目中会解析 JWT Token
		c.Set("customer_id", 1)
		c.Next()
	}
}
