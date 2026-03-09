package routes

import (
	"goshopadmin/controllers"
	"goshopadmin/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// 创建控制器实例
	commonController := controllers.NewCommonController()
	authController := controllers.NewAuthController(db)
	userController := controllers.NewUserController(db)
	roleController := controllers.NewRoleController(db)
	permissionController := controllers.NewPermissionController(db)
	merchantController := controllers.NewMerchantController(db)
	productController := controllers.NewProductController(db)

	// 1. 通用路由（无需认证）
	// 路径: /health, /
	r.GET("/health", commonController.HealthCheck)
	r.GET("/", commonController.HelloWorld)

	// 2. API路由组
	// 路径前缀: /api
	api := r.Group("/api")
	api.Use(middleware.RequestLogger())
	{
		// 2.1 认证相关路由（部分需要认证）
		// 路径前缀: /api/auth
		auth := api.Group("/auth")
		{
			// 无需认证的认证路由
			// 路径: /api/auth/login, /api/auth/captcha, /api/auth/captcha/verify
			auth.POST("/login", authController.Login)
			auth.GET("/captcha", authController.GetCaptcha)
			auth.POST("/captcha/verify", authController.VerifyCaptcha)

			// 需要认证的认证路由
			// 路径: /api/auth/logout, /api/auth/refresh, /api/auth/me
			authProtected := auth.Group("/")
			authProtected.Use(middleware.AuthMiddleware())
			{
				authProtected.POST("/logout", authController.Logout)
				authProtected.POST("/refresh", authController.RefreshToken)
				authProtected.GET("/me", authController.GetCurrentUser)
			}
		}

		// 2.2 业务管理路由（均需要认证）
		// 路径前缀: /api
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// 用户管理路由
			// 路径: /api/users, /api/users/:id
			users := protected.Group("/users")
			{
				users.GET("", userController.GetUsers)
				users.GET("/:id", userController.GetUser)
				users.POST("", userController.CreateUser)
				users.PUT("/:id", userController.UpdateUser)
				users.DELETE("/:id", userController.DeleteUser)
			}

			// 角色管理路由
			// 路径: /api/roles, /api/roles/:id
			roles := protected.Group("/roles")
			{
				roles.GET("", roleController.GetRoles)
				roles.GET("/:id", roleController.GetRole)
				roles.POST("", roleController.CreateRole)
				roles.PUT("/:id", roleController.UpdateRole)
				roles.DELETE("/:id", roleController.DeleteRole)
				roles.POST("/:id/permissions", roleController.AssignPermissions)
			}

			// 权限管理路由
			// 路径: /api/permissions, /api/permissions/:id
			permissions := protected.Group("/permissions")
			{
				permissions.GET("", permissionController.GetPermissions)
				permissions.GET("/:id", permissionController.GetPermission)
				permissions.POST("", permissionController.CreatePermission)
				permissions.PUT("/:id", permissionController.UpdatePermission)
				permissions.DELETE("/:id", permissionController.DeletePermission)
			}

			// 商户管理路由
			// 路径: /api/merchants, /api/merchants/:id
			merchants := protected.Group("/merchants")
			{
				merchants.GET("", merchantController.GetMerchants)
				merchants.GET("/:id", merchantController.GetMerchant)
				merchants.POST("", merchantController.CreateMerchant)
				merchants.PUT("/:id", merchantController.UpdateMerchant)
				merchants.DELETE("/:id", merchantController.DeleteMerchant)
				merchants.PUT("/:id/audit", merchantController.AuditMerchant)
				merchants.GET("/:id/users", merchantController.GetMerchantUsers)
				merchants.POST("/:id/users", merchantController.AddMerchantUser)
				merchants.DELETE("/:id/users/:user_id", merchantController.RemoveMerchantUser)
			}

			// 商品管理路由
			// 路径: /api/products, /api/products/:id
			products := protected.Group("/products")
			{
				products.GET("", productController.GetProducts)
				products.GET("/:id", productController.GetProduct)
				products.POST("", productController.CreateProduct)
				products.PUT("/:id", productController.UpdateProduct)
				products.DELETE("/:id", productController.DeleteProduct)
			}

			// 商品分类管理路由
			// 路径: /api/product-categories, /api/product-categories/:id
			categories := protected.Group("/product-categories")
			{
				categories.GET("", productController.GetCategories)
				categories.GET("/:id", productController.GetCategory)
				categories.POST("", productController.CreateCategory)
				categories.PUT("/:id", productController.UpdateCategory)
				categories.DELETE("/:id", productController.DeleteCategory)
			}

			// 商品图片管理路由
			// 路径: /api/product-images, /api/product-images/:id
			images := protected.Group("/product-images")
			{
				images.POST("", productController.AddProductImage)
				images.DELETE("/:id", productController.DeleteProductImage)
			}

			// 商品SKU管理路由
			// 路径: /api/product-skus, /api/product-skus/:id
			skus := protected.Group("/product-skus")
			{
				skus.POST("", productController.AddProductSKU)
				skus.PUT("/:id", productController.UpdateProductSKU)
				skus.DELETE("/:id", productController.DeleteProductSKU)
			}
		}
	}
}
