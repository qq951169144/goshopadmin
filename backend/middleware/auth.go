package middleware

import (
	"goshopadmin/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var jwtSecret string

// SetJWTSecret 设置JWT密钥
func SetJWTSecret(secret string) {
	jwtSecret = secret
}

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未提供认证信息"})
			c.Abort()
			return
		}

		// 检查Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证格式错误"})
			c.Abort()
			return
		}

		// 解析token
		tokenString := parts[1]

		// 使用已设置的密钥（强制从配置读取）
		secret := jwtSecret
		if secret == "" {
			// 密钥未配置，返回服务器错误
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "JWT密钥未配置"})
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(tokenString, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "无效的token"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("roleID", claims.RoleID)

		c.Next()
	}
}

// PermissionMiddleware 权限验证中间件
func PermissionMiddleware(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 这里需要从数据库中获取用户的权限信息
		// 暂时简化处理，实际项目中应该从数据库或缓存中获取
		c.Next()
	}
}
