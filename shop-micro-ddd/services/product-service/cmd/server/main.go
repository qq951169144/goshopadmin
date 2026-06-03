package main

import (
	"fmt"
	"log"

	"product-service/config"
	httpHandler "product-service/interfaces/http"
	"product-service/infrastructure/persistence"

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
	productRepo := persistence.NewGormProductRepository(db)

	// 启动 HTTP 服务
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	httpHandler.SetupRouter(engine, productRepo)

	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("商品服务启动, 监听地址: %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("商品服务启动失败: %v", err)
	}
}
