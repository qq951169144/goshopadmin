package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"search-service/config"
	"search-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ============================================================
// JWT 认证中间件
// 支持管理端和C端双 JWT Secret 验证
// 通过 URL 路径前缀判断使用哪个 Secret：
//   /api/search/admin/*    → JWT_SECRET_ADMIN（管理端）
//   /api/search/customer/* → JWT_SECRET_CUSTOMER（C端）
//   /api/search/products, /api/search/suggest → 公开，无需认证
// ============================================================

// AdminAuth 管理端认证中间件
// 验证管理端 JWT token，解析 user_id、role_id
// 与 backend 的 AuthMiddleware 行为保持一致：严格要求 Bearer 格式、校验签名方法
func AdminAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4010,
				"message": "未提供认证信息",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		tokenString, err := extractToken(authHeader)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4012,
				"message": "认证格式错误",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		secret := config.GetConfig().JWTSecretAdmin
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 校验签名方法为 HMAC，防止算法混淆攻击
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4012,
				"message": "无效的token",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4012,
				"message": "无效的token",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		// 管理端 token 包含 user_id 和 role_id
		userID, ok := claims["user_id"].(float64)
		if !ok {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4012,
				"message": "无效的管理端token",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		roleID, _ := claims["role_id"].(float64)

		// 同时设置两种键名，兼容不同消费方
		ctx.Set("user_type", "admin")
		ctx.Set("user_id", int(userID))
		ctx.Set("role_id", int(roleID))
		ctx.Set("userID", int(userID))
		ctx.Set("roleID", int(roleID))

		utils.Info("管理端认证成功, user_id: %d, role_id: %d", int(userID), int(roleID))
		ctx.Next()
	}
}

// CustomerAuth C端认证中间件
// 验证C端 JWT token，解析 customer_id
// 与 shop-backend 的 Auth 中间件行为保持一致：支持 Authorization/WebSocket/Query 三种 token 来源
func CustomerAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		wsProtocolHeader := ctx.GetHeader("Sec-WebSocket-Protocol")
		queryToken := ctx.Query("token")

		if authHeader == "" && wsProtocolHeader == "" && queryToken == "" {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4010,
				"message": "未提供认证信息",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		// 按优先级提取 token：Authorization > WebSocket > Query
		tokenString := authHeader
		if tokenString == "" && wsProtocolHeader != "" {
			tokenString = wsProtocolHeader
		}
		if tokenString == "" && queryToken != "" {
			tokenString = queryToken
		}

		// 处理 Bearer 前缀
		if strings.HasPrefix(tokenString, "Bearer ") {
			parts := strings.SplitN(tokenString, " ", 2)
			if len(parts) == 2 {
				tokenString = parts[1]
			}
		}

		secret := config.GetConfig().JWTSecretCustomer
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 校验签名方法为 HMAC，防止算法混淆攻击
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4012,
				"message": "无效的token",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4012,
				"message": "无效的token",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		// C端 token 包含 customer_id
		customerID, ok := claims["customer_id"].(float64)
		if !ok {
			ctx.JSON(http.StatusOK, gin.H{
				"code":    4012,
				"message": "无效的C端token",
				"data":    nil,
			})
			ctx.Abort()
			return
		}

		ctx.Set("user_type", "customer")
		ctx.Set("customer_id", int(customerID))

		utils.Info("C端认证成功, customer_id: %d", int(customerID))
		ctx.Next()
	}
}

// extractToken 从 Authorization header 提取 Bearer token
// 严格要求 "Bearer <token>" 格式，与 backend 保持一致
func extractToken(authHeader string) (string, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		return "", errors.New("认证格式错误，需要 Bearer token")
	}
	return parts[1], nil
}
