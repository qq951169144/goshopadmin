package routes

import (
	"github.com/gin-gonic/gin"
	"goshopadmin/controllers"
)

// RegisterCommonRoutes 注册通用路由
func RegisterCommonRoutes(r *gin.Engine) {
	// 创建控制器实例
	commonController := controllers.NewCommonController()

	// 健康检查路由
	r.GET("/health", commonController.HealthCheck)

	// 根路径HelloWorld路由
	r.GET("/", commonController.HelloWorld)
}
