package routes

import (
	"search-service/controllers"
	"search-service/middleware"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 路由配置
// 定义搜索服务的所有 HTTP 路由
// 使用 Dependencies 结构体注入控制器，保持与项目风格一致
// ============================================================

// Dependencies 路由依赖注入结构体
// 包含所有控制器实例，由 main.go 创建并传入
type Dependencies struct {
	// ProductController 商品搜索控制器
	ProductController *controllers.ProductController

	// OrderController 订单搜索控制器
	OrderController *controllers.OrderController

	// UserController 用户搜索控制器
	UserController *controllers.UserController

	// CustomerController 客户搜索控制器
	CustomerController *controllers.CustomerController

	// HealthController 健康检查控制器
	HealthController *controllers.HealthController

	// SuggestController 搜索建议控制器
	SuggestController *controllers.SuggestController

	// SyncController 同步管理控制器
	SyncController *controllers.SyncController
}

// SetupRoutes 设置路由
// r: Gin 引擎实例
// deps: 依赖注入结构体，包含所有控制器
func SetupRoutes(r *gin.Engine, deps *Dependencies) {
	// 健康检查路由（无需限流，供监控系统调用）
	r.GET("/health", deps.HealthController.HealthCheck)

	// 搜索 API 路由组
	// 应用 CORS 跨域中间件、请求日志中间件和限流中间件
	api := r.Group("/api/search")
	api.Use(middleware.CORS())
	api.Use(middleware.RequestLogger())
	api.Use(middleware.RateLimit())
	{
		// 公开接口（无需认证）
		// 商品搜索
		// GET /api/search/products?keyword=xxx&category_id=1&page=1&page_size=20
		api.GET("/products", deps.ProductController.SearchProducts)

		// 搜索建议
		// GET /api/search/suggest?prefix=xxx&type=product
		api.GET("/suggest", deps.SuggestController.Suggest)

		// 管理端接口（需要管理端认证）
		admin := api.Group("/admin")
		admin.Use(middleware.AdminAuth())
		{
			// 管理端订单搜索
			// GET /api/search/admin/orders?keyword=xxx&status=paid&page=1&page_size=20
			admin.GET("/orders", deps.OrderController.SearchAdminOrders)

			// 用户搜索
			// GET /api/search/admin/users?keyword=xxx&role_id=1&page=1&page_size=20
			admin.GET("/users", deps.UserController.SearchUsers)

			// 客户搜索
			// GET /api/search/admin/customers?keyword=xxx&status=active&page=1&page_size=20
			admin.GET("/customers", deps.CustomerController.SearchCustomers)

			// 同步管理
			// POST /api/search/admin/sync/full  手动触发全量同步
			// GET /api/search/admin/sync/status 查询同步状态
			admin.POST("/sync/full", deps.SyncController.TriggerFullSync)
			admin.GET("/sync/status", deps.SyncController.GetSyncStatus)
		}

		// C端接口（需要C端认证）
		customer := api.Group("/customer")
		customer.Use(middleware.CustomerAuth())
		{
			// C端订单搜索（只查自己的订单）
			// GET /api/search/customer/orders?keyword=xxx&status=paid&page=1&page_size=20
			customer.GET("/orders", deps.OrderController.SearchCustomerOrders)
		}
	}
}
