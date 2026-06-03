package http

import (
	"customer-service/domain/customer"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter(engine *gin.Engine, repo customer.CustomerRepository) {
	handler := NewCustomerHandler(repo)

	// 内部 API（供其他微服务调用，不走认证）
	internal := engine.Group("/api/internal")
	{
		internal.GET("/customers/:customerID/addresses/:addressID/verify", handler.VerifyAddress)
	}
}
