package main

import (
	"fmt"
	"log"

	"shop-backend-ddd/config"
	"shop-backend-ddd/domain/order"
	httpHandler "shop-backend-ddd/interfaces/http"
	"shop-backend-ddd/infrastructure/persistence"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// ========================================
	// 1. 加载配置
	// ========================================
	cfg := config.LoadConfig()

	// ========================================
	// 2. 初始化数据库
	// ========================================
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// ========================================
	// 3. 组装依赖（Composition Root）
	// 这是整个应用唯一知道所有具体实现的地方
	// ========================================

	// 基础设施层：创建具体实现
	orderRepo := persistence.NewGormOrderRepository(db)
	productRepo := persistence.NewGormProductRepository(db)
	customerRepo := persistence.NewGormCustomerRepository(db)
	txManager := persistence.NewGormTransactionManager(db)

	// 领域层：注入接口依赖
	// 注意：OrderService 只依赖接口，不知道具体实现
	orderService := order.NewOrderService(
		orderRepo,    // OrderRepository 接口
		productRepo,  // ProductQuerier 接口（GormProductRepository 同时实现了两个接口）
		customerRepo, // CustomerQuerier 接口（GormCustomerRepository 同时实现了两个接口）
		txManager,    // TransactionManager 接口
	)

	// ========================================
	// 4. 创建 HTTP 引擎和路由
	// ========================================
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())

	httpHandler.SetupRouter(engine, orderService)

	// ========================================
	// 5. 启动服务器
	// ========================================
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("DDD 示例服务启动, 监听地址: %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
