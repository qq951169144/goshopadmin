package controllers

import (
	"shop-backend/services"

	"github.com/gin-gonic/gin"
)

// SearchController 搜索控制器
// 代理搜索请求到 search-service
type SearchController struct {
	BaseController
	searchProxyService *services.SearchProxyService
}

// NewSearchController 创建搜索控制器
func NewSearchController() *SearchController {
	return &SearchController{
		searchProxyService: services.NewSearchProxyService(),
	}
}

// SearchProducts 搜索商品
func (sc *SearchController) SearchProducts(ctx *gin.Context) {
	sc.searchProxyService.ProxySearch(ctx, "/products")
}

// SearchOrders 搜索订单
func (sc *SearchController) SearchOrders(ctx *gin.Context) {
	sc.searchProxyService.ProxySearch(ctx, "/orders")
}

// Suggest 搜索建议
func (sc *SearchController) Suggest(ctx *gin.Context) {
	sc.searchProxyService.ProxySearch(ctx, "/suggest")
}
