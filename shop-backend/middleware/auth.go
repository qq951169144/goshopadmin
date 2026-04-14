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

		if authHeader == "" && wsProtocolHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token required"})
			c.Abort()
			return
		}

		tokenString := authHeader
		if tokenString == "" && wsProtocolHeader != "" {
			protocols := strings.Split(wsProtocolHeader, ",")
			for _, protocol := range protocols {
				protocol = strings.TrimSpace(protocol)
				if strings.HasPrefix(protocol, "Bearer ") {
					tokenString = protocol
					break
				}
			}
		}

		parts := strings.SplitN(tokenString, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			c.Abort()
			return
		}
		tokenString = parts[1]

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
