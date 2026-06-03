package main

import (
	"fmt"
	"log"

	"customer-service/config"
	httpHandler "customer-service/interfaces/http"
	"customer-service/infrastructure/persistence"

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
	customerRepo := persistence.NewGormCustomerRepository(db)

	// 启动 HTTP 服务
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	httpHandler.SetupRouter(engine, customerRepo)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("客户服务启动, 监听地址: %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("客户服务启动失败: %v", err)
	}
}
