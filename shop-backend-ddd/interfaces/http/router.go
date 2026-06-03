package http

import (
	"shop-backend-ddd/domain/order"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
// 路由只做 URL → Handler 的映射，不包含业务逻辑
func SetupRouter(engine *gin.Engine, orderService *order.OrderService) {
	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 路由组
	api := engine.Group("/api")
	{
		// 订单路由（需要认证）
		orders := api.Group("/orders")
		orders.Use(AuthMiddleware())
		{
			handler := NewOrderHandler(orderService)
			orders.POST("", handler.CreateOrder)
			orders.GET("/:orderNo", handler.GetOrderDetail)
			orders.PUT("/:orderNo/cancel", handler.CancelOrder)
		}
	}
}
