package routes

import (
	"github.com/gin-gonic/gin"
	"shop-backend/controllers"
	"shop-backend/middleware"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// 跨域中间件
	router.Use(middleware.CORS())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API路由组
	api := router.Group("/api")
	api.Use(middleware.RequestLogger())
	{
		// 验证码路由
		api.GET("/captcha", controllers.GenerateCaptcha)
		api.POST("/captcha/verify", controllers.VerifyCaptcha)

		// 认证路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
			auth.POST("/logout", middleware.Auth(), controllers.Logout)
		}

		// 用户路由
		user := api.Group("/user", middleware.Auth())
		{
			user.GET("/profile", controllers.GetProfile)
			user.PUT("/profile", controllers.UpdateProfile)
			user.GET("/orders", controllers.GetOrders)
		}

		// 商品路由
		products := api.Group("/products")
		{
			products.GET("", controllers.GetProducts)
			products.GET("/:id", controllers.GetProductDetail)
		}

		// 购物车路由
		cart := api.Group("/cart")
		{
			cart.GET("", controllers.GetCart)
			cart.POST("/items", controllers.AddToCart)
			cart.PUT("/items/:id", controllers.UpdateCartItem)
			cart.DELETE("/items/:id", controllers.RemoveCartItem)
			cart.POST("/sync", middleware.Auth(), controllers.SyncCart)
		}

		// 订单路由
		orders := api.Group("/orders", middleware.Auth())
		{
			orders.POST("", controllers.CreateOrder)
			orders.GET("/:id", controllers.GetOrderDetail)
		}

		// 支付路由
		payment := api.Group("/payment")
		{
			payment.GET("/fake-pay", controllers.FakePay)
			payment.POST("/callback", controllers.PaymentCallback)
		}
	}

	return router
}