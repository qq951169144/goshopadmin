package routes

import (
	"github.com/gin-gonic/gin"
	"shop-backend/controllers"
	"shop-backend/middleware"
)

// Dependencies 包含所有依赖
type Dependencies struct {
	AuthController     *controllers.AuthController
	CustomerController *controllers.CustomerController
	CaptchaController  *controllers.CaptchaController
	ProductController  *controllers.ProductController
	CartController     *controllers.CartController
	OrderController    *controllers.OrderController
	PaymentController  *controllers.PaymentController
}

// SetupRouter 设置路由
func SetupRouter(deps *Dependencies) *gin.Engine {
	// 创建Gin引擎
	router := gin.New()

	// 注册中间件（注意顺序）
	// 1. Logger 中间件（最先执行，生成 RequestID）
	router.Use(middleware.RequestLogger())

	// 2. CORS 中间件
	router.Use(middleware.CORS())

	// 3. Recovery 中间件
	router.Use(gin.Recovery())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API路由组
	api := router.Group("/api")
	{
		// 验证码路由
		api.GET("/captcha", deps.CaptchaController.GenerateCaptcha)
		api.POST("/captcha/verify", deps.CaptchaController.VerifyCaptcha)

		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", deps.AuthController.Register)
			auth.POST("/login", deps.AuthController.Login)
			auth.POST("/logout", middleware.Auth(), deps.AuthController.Logout)
		}

		// 用户路由
		user := api.Group("/user", middleware.Auth())
		{
			user.GET("/profile", deps.CustomerController.GetProfile)
			user.PUT("/profile", deps.CustomerController.UpdateProfile)
			user.GET("/orders", deps.CustomerController.GetOrders)
		}

		// 商品路由
		products := api.Group("/products")
		{
			products.GET("", deps.ProductController.GetProducts)
			products.GET("/:id", deps.ProductController.GetProductDetail)
		}

		// 购物车路由
		cart := api.Group("/cart")
		{
			cart.GET("", deps.CartController.GetCart)
			cart.POST("/items", deps.CartController.AddToCart)
			cart.PUT("/items/:id", deps.CartController.UpdateCartItem)
			cart.DELETE("/items/:id", deps.CartController.RemoveCartItem)
			cart.POST("/sync", middleware.Auth(), deps.CartController.SyncCart)
		}

		// 订单路由
		orders := api.Group("/orders", middleware.Auth())
		{
			orders.POST("", deps.OrderController.CreateOrder)
			orders.GET("/:id", deps.OrderController.GetOrderDetail)
		}

		// 支付路由
		payment := api.Group("/payment")
		{
			payment.GET("/fake-pay", deps.PaymentController.FakePay)
			payment.POST("/callback", deps.PaymentController.PaymentCallback)
		}
	}

	return router
}
