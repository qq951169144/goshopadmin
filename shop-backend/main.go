package main

import (
	"log"
	"shop-backend/config"
	"shop-backend/routes"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	if err := config.InitDB(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 初始化Redis
	if err := config.InitRedis(); err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}

	// 初始化路由
	router := routes.SetupRouter()

	// 启动服务器
	port := config.AppConfig.ServerPort
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}