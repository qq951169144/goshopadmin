package routes

import (
	"context"

	"shop-backend/cache"
	"shop-backend/config"
	"shop-backend/controllers"
	"shop-backend/middleware"
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// Dependencies 包含所有依赖
type Dependencies struct {
	AuthController          *controllers.AuthController
	CustomerController      *controllers.CustomerController
	CaptchaController       *controllers.CaptchaController
	ProductController       *controllers.ProductController
	CartController          *controllers.CartController
	OrderController         *controllers.OrderController
	PaymentController       *controllers.PaymentController
	AddressController       *controllers.AddressController
	SpecificationController *controllers.SpecificationController
	ActivityController      *controllers.ActivityController
	RedeemCodeController    *controllers.RedeemCodeController
	ActivityOrderController *controllers.ActivityOrderController
	HealthController        *controllers.HealthController
}

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, cfg *config.Config) {
	// 初始化缓存工具并预热布隆过滤器
	ctx := context.Background()
	cacheUtil := cache.NewCacheUtil(db, redisClient)

	// 根据配置决定是否初始化布隆过滤器
	if cfg.EnableBloomFilter {
		if err := cacheUtil.InitBloomFilters(ctx); err != nil {
			// 记录错误但不中断启动
			utils.Error("布隆过滤器初始化失败: %v", err)
		} else {
			utils.Info("布隆过滤器初始化成功并预热完成")
		}
	} else {
		utils.Info("布隆过滤器已禁用")
	}

	// 创建控制器实例
	deps := &Dependencies{
		AuthController:          controllers.NewAuthController(db, redisClient, cfg.JWTSecret, cfg.JWTExpireHour),
		CustomerController:      controllers.NewCustomerController(db),
		CaptchaController:       controllers.NewCaptchaController(redisClient),
		ProductController:       controllers.NewProductController(db, cacheUtil),
		CartController:          controllers.NewCartController(db),
		OrderController:         controllers.NewOrderController(db, cacheUtil),
		PaymentController:       controllers.NewPaymentController(db, cacheUtil),
		AddressController:       controllers.NewAddressController(db),
		SpecificationController: controllers.NewSpecificationController(db, cacheUtil),
		ActivityController:      controllers.NewActivityController(db),
		RedeemCodeController:    controllers.NewRedeemCodeController(db),
		ActivityOrderController: controllers.NewActivityOrderController(db),
		HealthController:        controllers.NewHealthController(),
	}

	// 1. 健康检查
	// 路径: /health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 2. API路由组
	// 路径前缀: /api
	api := r.Group("/api")
	{
		// 2.0 健康检查路由
		// 路径: /api/health/mq
		health := api.Group("/health")
		{
			health.GET("/mq", deps.HealthController.CheckMQ)
		}

		// 2.1 验证码路由
		// 路径: /api/captcha, /api/captcha/verify
		api.GET("/captcha", deps.CaptchaController.GenerateCaptcha)
		api.POST("/captcha/verify", deps.CaptchaController.VerifyCaptcha)

		// 2.2 认证路由
		// 路径前缀: /api/auth
		auth := api.Group("/auth")
		{
			// 无需认证的认证路由
			// 路径: /api/auth/register, /api/auth/login
			auth.POST("/register", deps.AuthController.Register)
			auth.POST("/login", deps.AuthController.Login)

			// 需要认证的认证路由
			// 路径: /api/auth/logout
			authProtected := auth.Group("/")
			authProtected.Use(middleware.Auth())
			{
				authProtected.POST("/logout", deps.AuthController.Logout)
			}
		}

		// 2.3 用户路由（需要认证）
		// 路径前缀: /api/user
		user := api.Group("/user")
		user.Use(middleware.Auth())
		{
			// 路径: /api/user/profile, /api/user/orders
			user.GET("/profile", deps.CustomerController.GetProfile)
			user.PUT("/profile", deps.CustomerController.UpdateProfile)
			user.GET("/orders", deps.CustomerController.GetOrders)
		}

		// 2.4 客户相关路由（使用 customer 前缀，需要认证）
		// 路径前缀: /api/customer
		customer := api.Group("/customer")
		customer.Use(middleware.Auth())
		{
			// 地址管理
			// 路径: /api/customer/addresses, /api/customer/addresses/:id
			customer.GET("/addresses", deps.AddressController.GetAddresses)
			customer.POST("/addresses", deps.AddressController.CreateAddress)
			customer.GET("/addresses/:id", deps.AddressController.GetAddress)
			customer.PUT("/addresses/:id", deps.AddressController.UpdateAddress)
			customer.DELETE("/addresses/:id", deps.AddressController.DeleteAddress)
			customer.PUT("/addresses/:id/default", deps.AddressController.SetDefaultAddress)
			customer.GET("/addresses/default", deps.AddressController.GetDefaultAddress)
		}

		// 2.5 商品路由
		// 路径前缀: /api/products
		products := api.Group("/products")
		{
			// 路径: /api/products, /api/products/:id
			products.GET("", deps.ProductController.GetProducts)
			products.GET("/:id", deps.SpecificationController.GetProductDetail)
			products.GET("/:id/skus", deps.SpecificationController.GetProductSkus)
			products.GET("/:id/sku", deps.SpecificationController.GetSkuBySpecCombination)
		}

		// 2.6 购物车路由（需要认证）
		// 路径前缀: /api/cart
		cart := api.Group("/cart")
		cart.Use(middleware.Auth())
		{
			// 路径: /api/cart, /api/cart/items, /api/cart/items/:id
			cart.GET("", deps.CartController.GetCart)
			cart.POST("/items", deps.CartController.AddToCart)
			cart.PUT("/items/:id", deps.CartController.UpdateCartItem)
			cart.DELETE("/items/:id", deps.CartController.RemoveCartItem)
			cart.POST("/sync", deps.CartController.SyncCart)
		}

		// 2.7 订单路由（需要认证）
		// 路径前缀: /api/orders
		orders := api.Group("/orders")
		orders.Use(middleware.Auth())
		{
			// 路径: /api/orders, /api/orders/:orderNo
			orders.POST("", deps.OrderController.CreateOrder)
			orders.GET("/:orderNo", deps.OrderController.GetOrderDetail)
			orders.PUT("/:orderNo/cancel", deps.OrderController.CancelOrder)
			orders.PUT("/:orderNo/confirm", deps.OrderController.ConfirmReceipt)
		}

		// 2.8 支付路由
		// 路径前缀: /api/payment
		payment := api.Group("/payment")
		{
			// 路径: /api/payment/fake-pay, /api/payment/callback
			payment.GET("/fake-pay", deps.PaymentController.FakePay)
			payment.POST("/callback", deps.PaymentController.PaymentCallback)
		}

		// 2.9 活动路由
		// 路径前缀: /api/activities
		activities := api.Group("/activities")
		{
			// 路径: /api/activities, /api/activities/:id
			activities.GET("", deps.ActivityController.GetActiveActivities)
			activities.GET("/:id", deps.ActivityController.GetActivity)
			activities.GET("/:id/products", deps.ActivityController.GetActivityProducts)
			activities.GET("/:id/skus", deps.ActivityController.GetActivityProductSkus)
		}

		// 2.10 兑换码路由
		// 路径前缀: /api/redeem-codes
		redeemCodes := api.Group("/redeem-codes")
		{
			// 路径: /api/redeem-codes/verify
			redeemCodes.POST("/verify", deps.RedeemCodeController.VerifyRedeemCode)

			// 需要认证的兑换码路由
			redeemProtected := redeemCodes.Group("/")
			redeemProtected.Use(middleware.Auth())
			{
				redeemProtected.POST("/redeem", deps.RedeemCodeController.RedeemCode)
				redeemProtected.GET("/logs", deps.RedeemCodeController.GetRedeemCodeLogs)
			}
		}

		// 2.11 活动订单路由（需要认证）
		// 路径前缀: /api/activity-orders
		activityOrders := api.Group("/activity-orders")
		activityOrders.Use(middleware.Auth())
		{
			// 路径: /api/activity-orders, /api/activity-orders/:id
			activityOrders.POST("", deps.ActivityOrderController.CreateActivityOrder)
			activityOrders.GET("", deps.ActivityOrderController.GetActivityOrders)
			activityOrders.GET("/:id", deps.ActivityOrderController.GetActivityOrder)
			activityOrders.PUT("/:id/cancel", deps.ActivityOrderController.CancelActivityOrder)
		}
	}
}
