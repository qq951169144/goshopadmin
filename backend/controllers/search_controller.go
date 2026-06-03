package controllers

import (
	"goshopadmin/services"

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

// SearchUsers 搜索用户
func (sc *SearchController) SearchUsers(ctx *gin.Context) {
	sc.searchProxyService.ProxySearch(ctx, "/users")
}

// SearchCustomers 搜索客户
func (sc *SearchController) SearchCustomers(ctx *gin.Context) {
	sc.searchProxyService.ProxySearch(ctx, "/customers")
}

// Suggest 搜索建议
func (sc *SearchController) Suggest(ctx *gin.Context) {
	sc.searchProxyService.ProxySearch(ctx, "/suggest")
}
