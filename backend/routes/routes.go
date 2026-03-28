package routes

import (
	"context"
	"goshopadmin/cache"
	"goshopadmin/config"
	"goshopadmin/controllers"
	"goshopadmin/middleware"
	"goshopadmin/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Dependencies 包含所有依赖
type Dependencies struct {
	CommonController        *controllers.CommonController
	AuthController          *controllers.AuthController
	UserController          *controllers.UserController
	RoleController          *controllers.RoleController
	PermissionController    *controllers.PermissionController
	MerchantController      *controllers.MerchantController
	ProductController       *controllers.ProductController
	SpecificationController *controllers.SpecificationController
	SKUController           *controllers.SKUController
	ActivityController      *controllers.ActivityController
	RedeemCodeController    *controllers.RedeemCodeController
}

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, cfg *config.Config) {
	// 初始化缓存工具并预热布隆过滤器
	ctx := context.Background()
	cacheUtil := cache.NewCacheUtil(db, redisClient)
	if err := cacheUtil.InitBloomFilters(ctx); err != nil {
		// 记录错误但不中断启动
		utils.Error("布隆过滤器初始化失败: %v", err)
	} else {
		utils.Info("布隆过滤器初始化成功并预热完成")
	}

	// 创建控制器实例
	deps := &Dependencies{
		CommonController:        controllers.NewCommonController(),
		AuthController:          controllers.NewAuthController(db, cfg.JWTSecret, cfg.JWTExpireHour),
		UserController:          controllers.NewUserController(db, cfg.JWTSecret, cfg.JWTExpireHour),
		RoleController:          controllers.NewRoleController(db, cfg.JWTSecret, cfg.JWTExpireHour),
		PermissionController:    controllers.NewPermissionController(db, cfg.JWTSecret, cfg.JWTExpireHour),
		MerchantController:      controllers.NewMerchantController(db),
		ProductController:       controllers.NewProductController(db, redisClient),
		SpecificationController: controllers.NewSpecificationController(db),
		SKUController:           controllers.NewSKUController(db),
		ActivityController:      controllers.NewActivityController(db),
		RedeemCodeController:    controllers.NewRedeemCodeController(db),
	}

	// 1. 通用路由（无需认证）
	// 路径: /health, /
	r.GET("/health", deps.CommonController.HealthCheck)
	r.GET("/", deps.CommonController.HelloWorld)

	// 2. API路由组
	// 路径前缀: /api
	// 注意：RequestLogger 已在 main.go 中全局注册，这里不需要重复注册
	api := r.Group("/api")
	{
		// 2.1 认证相关路由（部分需要认证）
		// 路径前缀: /api/auth
		auth := api.Group("/auth")
		{
			// 无需认证的认证路由
			// 路径: /api/auth/login, /api/auth/captcha, /api/auth/captcha/verify
			auth.POST("/login", deps.AuthController.Login)
			auth.GET("/captcha", deps.AuthController.GetCaptcha)
			auth.POST("/captcha/verify", deps.AuthController.VerifyCaptcha)

			// 需要认证的认证路由
			// 路径: /api/auth/logout, /api/auth/refresh, /api/auth/me
			authProtected := auth.Group("/")
			authProtected.Use(middleware.AuthMiddleware())
			{
				authProtected.POST("/logout", deps.AuthController.Logout)
				authProtected.POST("/refresh", deps.AuthController.RefreshToken)
				authProtected.GET("/me", deps.AuthController.GetCurrentUser)
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
				users.GET("", deps.UserController.GetUsers)
				users.GET("/:id", deps.UserController.GetUser)
				users.POST("", deps.UserController.CreateUser)
				users.PUT("/:id", deps.UserController.UpdateUser)
				users.DELETE("/:id", deps.UserController.DeleteUser)
			}

			// 角色管理路由
			// 路径: /api/roles, /api/roles/:id
			roles := protected.Group("/roles")
			{
				roles.GET("", deps.RoleController.GetRoles)
				roles.GET("/:id", deps.RoleController.GetRole)
				roles.POST("", deps.RoleController.CreateRole)
				roles.PUT("/:id", deps.RoleController.UpdateRole)
				roles.DELETE("/:id", deps.RoleController.DeleteRole)
				roles.POST("/:id/permissions", deps.RoleController.AssignPermissions)
			}

			// 权限管理路由
			// 路径: /api/permissions, /api/permissions/:id
			permissions := protected.Group("/permissions")
			{
				permissions.GET("", deps.PermissionController.GetPermissions)
				permissions.GET("/:id", deps.PermissionController.GetPermission)
				permissions.POST("", deps.PermissionController.CreatePermission)
				permissions.PUT("/:id", deps.PermissionController.UpdatePermission)
				permissions.DELETE("/:id", deps.PermissionController.DeletePermission)
			}

			// 商户管理路由
			// 路径: /api/merchants, /api/merchants/:id
			merchants := protected.Group("/merchants")
			{
				merchants.GET("", deps.MerchantController.GetMerchants)
				merchants.GET("/:id", deps.MerchantController.GetMerchant)
				merchants.POST("", deps.MerchantController.CreateMerchant)
				merchants.PUT("/:id", deps.MerchantController.UpdateMerchant)
				merchants.DELETE("/:id", deps.MerchantController.DeleteMerchant)
				merchants.PUT("/:id/audit", deps.MerchantController.AuditMerchant)
				merchants.GET("/:id/users", deps.MerchantController.GetMerchantUsers)
				merchants.POST("/:id/users", deps.MerchantController.AddMerchantUser)
				merchants.DELETE("/:id/users/:user_id", deps.MerchantController.RemoveMerchantUser)
			}

			// 商品管理路由
			// 路径: /api/products, /api/products/:id
			products := protected.Group("/products")
			{
				products.GET("", deps.ProductController.GetProducts)
				products.GET("/:id", deps.ProductController.GetProduct)
				products.POST("", deps.ProductController.CreateProduct)
				products.PUT("/:id", deps.ProductController.UpdateProduct)
				products.DELETE("/:id", deps.ProductController.DeleteProduct)
			}

			// 商品分类管理路由
			// 路径: /api/product-categories, /api/product-categories/:id
			categories := protected.Group("/product-categories")
			{
				categories.GET("", deps.ProductController.GetCategories)
				categories.GET("/:id", deps.ProductController.GetCategory)
				categories.POST("", deps.ProductController.CreateCategory)
				categories.PUT("/:id", deps.ProductController.UpdateCategory)
				categories.DELETE("/:id", deps.ProductController.DeleteCategory)
			}

			// 商品图片管理路由
			// 路径: /api/product-images, /api/product-images/:id
			images := protected.Group("/product-images")
			{
				images.POST("", deps.ProductController.AddProductImage)
				images.DELETE("/:id", deps.ProductController.DeleteProductImage)
				images.PUT("/:id", deps.ProductController.UpdateProductImage)
			}

			// 规格管理路由
			// 路径: /api/products/:id/specifications
			products.GET("/:id/specifications", deps.SpecificationController.GetSpecificationsByProductID)
			products.POST("/:id/specifications", deps.SpecificationController.CreateSpecification)

			// 规格值管理路由
			// 路径: /api/specifications/:id/values
			specifications := protected.Group("/specifications")
			{
				specifications.PUT("/:id", deps.SpecificationController.UpdateSpecification)
				specifications.DELETE("/:id", deps.SpecificationController.DeleteSpecification)
				specifications.POST("/:id/values", deps.SpecificationController.CreateSpecificationValue)
			}

			// 规格值管理路由（独立路径）
			specValues := protected.Group("/specification-values")
			{
				specValues.PUT("/:id", deps.SpecificationController.UpdateSpecificationValue)
				specValues.DELETE("/:id", deps.SpecificationController.DeleteSpecificationValue)
			}

			// SKU管理路由（新）
			// 路径: /api/products/:id/skus
			products.POST("/:id/skus", deps.SKUController.CreateSKU)
			products.POST("/:id/skus/batch", deps.SKUController.BatchCreateSKU)
			products.GET("/:id/skus", deps.SKUController.GetSKUsByProductID)
			products.POST("/:id/skus/generate", deps.SKUController.GenerateSKUsFromSpecs)

			// SKU管理路由（独立路径）
			skuRoutes := protected.Group("/skus")
			{
				skuRoutes.PUT("/:id", deps.SKUController.UpdateSKU)
				skuRoutes.DELETE("/:id", deps.SKUController.DeleteSKU)
			}

			// 活动管理路由
			activities := protected.Group("/activities")
			{
				activities.GET("", deps.ActivityController.GetActivities)
				activities.GET("/:id", deps.ActivityController.GetActivity)
				activities.POST("", deps.ActivityController.CreateActivity)
				activities.PUT("/:id", deps.ActivityController.UpdateActivity)
				activities.DELETE("/:id", deps.ActivityController.DeleteActivity)
				activities.PUT("/:id/status", deps.ActivityController.UpdateActivityStatus)

				// 活动兑换码管理路由
				activities.POST("/:id/redeem-codes/generate", deps.RedeemCodeController.GenerateRedeemCodes)
				activities.GET("/:id/redeem-codes", deps.RedeemCodeController.GetRedeemCodes)
				activities.GET("/:id/redeem-codes/export", deps.RedeemCodeController.ExportRedeemCodes)
				activities.POST("/:id/redeem-codes/import", deps.RedeemCodeController.ImportRedeemCodes)
			}

			// 兑换码管理路由
			redeemCodes := protected.Group("/redeem-codes")
			{
				redeemCodes.POST("/verify", deps.RedeemCodeController.VerifyRedeemCode)
				redeemCodes.GET("/logs", deps.RedeemCodeController.GetRedeemCodeLogs)
				redeemCodes.PUT("/:id/status", deps.RedeemCodeController.UpdateRedeemCodeStatus)
			}
		}
	}
}
