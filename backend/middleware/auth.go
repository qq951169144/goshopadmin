package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"goshopadmin/services"
	"goshopadmin/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

var jwtSecret string
var db *gorm.DB
var redisClient *redis.Client

// SetJWTSecret 设置JWT密钥
func SetJWTSecret(secret string) {
	jwtSecret = secret
}

// SetDB 设置数据库连接
func SetDB(database *gorm.DB) {
	db = database
}

// SetRedis 设置Redis客户端
func SetRedis(client *redis.Client) {
	redisClient = client
}

// ClearPermissionCache 清除用户权限缓存
func ClearPermissionCache(userID int) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:permissions", userID)

	if redisClient != nil {
		redisClient.Del(ctx, cacheKey)
	}
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
func PermissionMiddleware(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
			c.Abort()
			return
		}

		// 检查用户ID是否为0
		userIDInt, ok := userID.(int)
		if !ok || userIDInt <= 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
			c.Abort()
			return
		}

		// 从缓存获取用户权限
		permissions, err := getCachedPermissions(userIDInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取权限失败"})
			c.Abort()
			return
		}
		utils.Info("登录用户的权限 = %s, 需要用到的权限 = %s", permissions, requiredPermissions)
		// 验证权限
		if !hasPermission(permissions, requiredPermissions) {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// 缓存用户权限
func getCachedPermissions(userID int) ([]string, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:permissions", userID)

	// 1. 先从Redis缓存获取
	if redisClient != nil {
		cachedPermissions, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil {
			var permissions []string
			if err := json.Unmarshal([]byte(cachedPermissions), &permissions); err == nil {
				return permissions, nil
			}
		}
	}

	// 2. 缓存不存在时从数据库获取
	permissionService := services.NewPermissionService(db, redisClient)
	permissions, err := permissionService.GetUserPermissionCodes(userID)
	if err != nil {
		return nil, err
	}

	// 3. 将权限缓存到Redis
	if redisClient != nil {
		permissionsJSON, err := json.Marshal(permissions)
		if err == nil {
			// 缓存1小时
			redisClient.Set(ctx, cacheKey, permissionsJSON, time.Hour)
		}
	}

	return permissions, nil
}

// 验证权限
func hasPermission(userPermissions []string, requiredPermissions []string) bool {
	// 实现权限验证逻辑
	// 支持AND/OR逻辑
	for _, requiredPerm := range requiredPermissions {
		// 检查是否包含OR逻辑
		if strings.Contains(requiredPerm, "|") {
			// OR逻辑：只要有一个权限满足即可
			orPerms := strings.Split(requiredPerm, "|")
			hasOrPerm := false
			for _, orPerm := range orPerms {
				for _, userPerm := range userPermissions {
					if userPerm == orPerm {
						hasOrPerm = true
						break
					}
				}
				if hasOrPerm {
					break
				}
			}
			if !hasOrPerm {
				return false
			}
		} else {
			// AND逻辑：所有权限都必须满足
			hasPerm := false
			for _, userPerm := range userPermissions {
				if userPerm == requiredPerm {
					hasPerm = true
					break
				}
			}
			if !hasPerm {
				return false
			}
		}
	}
	return true
}
