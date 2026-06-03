package main

import (
	"fmt"
	"log"

	"order-service/config"
	"order-service/domain/order"
	"order-service/infrastructure/client"
	"order-service/infrastructure/persistence"
	httpHandler "order-service/interfaces/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	// 初始化数据库
	db, err := gorm.Open(mysql.Open(cfg.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 组装依赖
	orderRepo := persistence.NewGormOrderRepository(db)
	txManager := persistence.NewGormTransactionManager(db)

	// 微服务关键区别：通过 HTTP 客户端调用其他服务
	productClient := client.NewProductHTTPClient(cfg.ProductServiceURL)
	customerClient := client.NewCustomerHTTPClient(cfg.CustomerServiceURL)

	// 领域服务：注入 HTTP 客户端而不是 Repository
	orderService := order.NewOrderService(orderRepo, productClient, customerClient, txManager)

	// 启动 HTTP 服务
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	httpHandler.SetupRouter(engine, orderService)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("订单服务启动, 监听地址: %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("订单服务启动失败: %v", err)
	}
}
