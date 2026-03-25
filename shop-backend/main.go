package main

import (
	"context"
	"log"

	"shop-backend/cache"
	"shop-backend/config"
	"shop-backend/controllers"
	"shop-backend/middleware"
	"shop-backend/routes"
	"shop-backend/services"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化数据库连接
	conn, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// 3. 设置JWT密钥到中间件
	middleware.SetJWTSecret(cfg.JWTSecret)

	// 4. 初始化缓存工具
	cacheUtil := cache.NewCacheUtil(conn.DB, conn.Redis)

	// 5. 初始化布隆过滤器并预热数据
	ctx := context.Background()
	if err := cacheUtil.InitBloomFilters(ctx); err != nil {
		log.Printf("Warning: Failed to initialize bloom filters: %v", err)
	} else {
		log.Println("Bloom filters initialized and data warmed up successfully")
	}

	// 6. 创建服务层实例（依赖注入）
	// 注意：CaptchaService 需要先创建，因为 AuthService 依赖它
	captchaService := services.NewCaptchaService(conn.Redis)

	authService := services.NewAuthService(conn.DB, captchaService, cfg.JWTSecret, cfg.JWTExpireHour)
	customerService := services.NewCustomerService(conn.DB)
	productService := services.NewProductService(conn.DB, cacheUtil)
	cartService := services.NewCartService(conn.DB)
	orderService := services.NewOrderService(conn.DB, cacheUtil)
	addressService := services.NewAddressService(conn.DB)
	specificationService := services.NewSpecificationService(conn.DB)

	// 6. 创建控制器实例（依赖注入）
	deps := &routes.Dependencies{
		AuthController:          controllers.NewAuthController(authService),
		CustomerController:      controllers.NewCustomerController(customerService),
		CaptchaController:       controllers.NewCaptchaController(captchaService),
		ProductController:       controllers.NewProductController(productService),
		CartController:          controllers.NewCartController(cartService),
		OrderController:         controllers.NewOrderController(orderService),
		PaymentController:       controllers.NewPaymentController(orderService),
		AddressController:       controllers.NewAddressController(addressService),
		SpecificationController: controllers.NewSpecificationController(specificationService),
	}

	// 7. 设置路由
	router := routes.SetupRouter(deps)

	// 8. 启动服务器
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
