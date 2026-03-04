package routes

import (
	"goshopadmin/controllers"
	"goshopadmin/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupAuthRoutes 设置认证相关路由
func SetupAuthRoutes(router *gin.RouterGroup, db *gorm.DB) {
	authController := controllers.NewAuthController(db)

	// 认证相关路由
	authGroup := router.Group("/auth")
	{
		// 登录和验证码不需要认证
		authGroup.POST("/login", authController.Login)
		authGroup.GET("/captcha", authController.GetCaptcha)
		authGroup.POST("/captcha/verify", authController.VerifyCaptcha)

		// 需要认证的路由
		authRequired := authGroup.Group("/")
		authRequired.Use(middleware.AuthMiddleware())
		{
			authRequired.POST("/logout", authController.Logout)
			authRequired.POST("/refresh", authController.RefreshToken)
			authRequired.GET("/me", authController.GetCurrentUser)
		}
	}
}
