package http

import (
	"order-service/domain/order"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter(engine *gin.Engine, orderService *order.OrderService) {
	handler := NewOrderHandler(orderService)

	api := engine.Group("/api")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", handler.CreateOrder)
			orders.GET("/:orderNo", handler.GetOrderDetail)
			orders.POST("/:orderNo/cancel", handler.CancelOrder)
		}
	}
}
