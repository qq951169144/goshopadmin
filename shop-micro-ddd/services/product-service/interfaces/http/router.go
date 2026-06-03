package http

import (
	"product-service/domain/product"

	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter(engine *gin.Engine, repo product.ProductRepository) {
	handler := NewProductHandler(repo)

	// 内部 API（供其他微服务调用，不走认证）
	internal := engine.Group("/api/internal")
	{
		internal.GET("/products/:productID/skus/:skuID", handler.GetProductAndSKU)
		internal.POST("/skus/:skuID/deduct", handler.DeductStock)
		internal.POST("/skus/:skuID/restore", handler.RestoreStock)
	}
}
