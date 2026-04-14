package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret string

// SetJWTSecret 设置JWT密钥
func SetJWTSecret(secret string) {
	jwtSecret = secret
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		wsProtocolHeader := c.GetHeader("Sec-WebSocket-Protocol")
		queryToken := c.Query("token")

		if authHeader == "" && wsProtocolHeader == "" && queryToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			c.Abort()
			return
		}

		tokenString := authHeader
		if tokenString == "" && wsProtocolHeader != "" {
			tokenString = wsProtocolHeader
		}
		if tokenString == "" && queryToken != "" {
			tokenString = queryToken
		}

		if strings.HasPrefix(tokenString, "Bearer ") {
			parts := strings.SplitN(tokenString, " ", 2)
			if len(parts) == 2 {
				tokenString = parts[1]
			}
		}

		// 使用已设置的密钥（强制从配置读取）
		secret := jwtSecret
		if secret == "" {
			// 密钥未配置，返回服务器错误
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		customerID, ok := claims["customer_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid customer ID in token"})
			c.Abort()
			return
		}

		c.Set("customer_id", int(customerID))
		c.Next()
	}
}
