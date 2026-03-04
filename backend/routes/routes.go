package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// 注册通用路由
	RegisterCommonRoutes(r)

	// API路由组
	apiGroup := r.Group("/api")
	
	// 注册认证路由
	SetupAuthRoutes(apiGroup, db)
}
