package routes

import (
	"context"
	"net"
	"net/http"

	"shop-backend/cache"
	"shop-backend/config"
	"shop-backend/controllers"
	"shop-backend/middleware"
	"shop-backend/utils"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	MonitorController       *controllers.MonitorController
}

// metricsIPWhitelist 创建 IP 白名单中间件，仅允许内部网络访问 /metrics 端点
// 允许的网段：172.16.0.0/12（Docker 默认网段）、192.168.0.0/16（局域网）、127.0.0.1（本机）
// 防止 /metrics 端点暴露运行时指标到公网
func metricsIPWhitelist() gin.HandlerFunc {
	allowedCIDRs := []string{
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.1/32",
		"::1/128",
	}

	parsedNetworks := make([]*net.IPNet, 0, len(allowedCIDRs))
	for _, cidr := range allowedCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil {
			parsedNetworks = append(parsedNetworks, network)
		}
	}

	return func(c *gin.Context) {
		ip := net.ParseIP(c.ClientIP())
		if ip == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		for _, network := range parsedNetworks {
			if network.Contains(ip) {
				c.Next()
				return
			}
		}

		c.AbortWithStatus(http.StatusForbidden)
	}
}

// SetupRoutes 设置所有路由
func SetupRoutes(r *gin.Engine, db *gorm.DB, redisClient *redis.Client, cfg *config.Config, monitor *utils.Monitor) {
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
		MonitorController:       controllers.NewMonitorController(monitor),
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
			activities.GET("/:id/skus/:sku_id", deps.ActivityController.GetActivitySkuDetail)
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

		// 2.12 监控路由（需要认证）
		// 路径前缀: /api/monitor
		monitor := api.Group("/monitor")
		monitor.Use(middleware.Auth())
		{
			// 路径: /api/monitor/stats, /api/monitor/stats/history
			monitor.GET("/stats", deps.MonitorController.GetCurrentStats)
			monitor.GET("/stats/history", deps.MonitorController.GetHistoryStats)
		}
	}

	// Prometheus metrics 端点
	// 使用 IP 白名单中间件保护，仅允许 Docker 内部网络和本机访问
	// Prometheus 在同一 Docker 网络内通过容器名访问此端点
	r.GET("/metrics", metricsIPWhitelist(), gin.WrapH(promhttp.Handler()))

	// pprof 性能分析端点（需要认证保护）
	// pprof 暴露 CPU、内存、协程等敏感运行时数据，必须认证后才能访问
	pprofGroup := r.Group("/debug/pprof")
	pprofGroup.Use(middleware.Auth())
	pprof.RouteRegister(pprofGroup)
}
