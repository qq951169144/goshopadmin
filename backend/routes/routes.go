package routes

import (
	"goshopadmin/controllers"
	"goshopadmin/middleware"

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

	// 设置需要认证的路由
	authRequired := apiGroup.Group("/")
	authRequired.Use(middleware.AuthMiddleware())
	{
		// 用户管理路由
		userController := controllers.NewUserController(db)
		userGroup := authRequired.Group("/users")
		{
			userGroup.GET("", userController.GetUsers)
			userGroup.GET("/:id", userController.GetUser)
			userGroup.POST("", userController.CreateUser)
			userGroup.PUT("/:id", userController.UpdateUser)
			userGroup.DELETE("/:id", userController.DeleteUser)
		}

		// 角色管理路由
		roleController := controllers.NewRoleController(db)
		roleGroup := authRequired.Group("/roles")
		{
			roleGroup.GET("", roleController.GetRoles)
			roleGroup.GET("/:id", roleController.GetRole)
			roleGroup.POST("", roleController.CreateRole)
			roleGroup.PUT("/:id", roleController.UpdateRole)
			roleGroup.DELETE("/:id", roleController.DeleteRole)
			roleGroup.POST("/:id/permissions", roleController.AssignPermissions)
		}

		// 权限管理路由
		permissionController := controllers.NewPermissionController(db)
		permissionGroup := authRequired.Group("/permissions")
		{
			permissionGroup.GET("", permissionController.GetPermissions)
			permissionGroup.GET("/:id", permissionController.GetPermission)
			permissionGroup.POST("", permissionController.CreatePermission)
			permissionGroup.PUT("/:id", permissionController.UpdatePermission)
			permissionGroup.DELETE("/:id", permissionController.DeletePermission)
		}
	}
}
