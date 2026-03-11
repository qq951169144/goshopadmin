package main

import (
	"fmt"
	"log"

	"goshopadmin/config"
	"goshopadmin/middleware"
	"goshopadmin/routes"
	"goshopadmin/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 确保在程序退出时关闭日志记录器
	defer utils.CloseLogger()

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.DBUser,
		config.AppConfig.DBPassword,
		config.AppConfig.DBHost,
		config.AppConfig.DBPort,
		config.AppConfig.DBName,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 应用CORS中间件
	r.Use(middleware.CORSMiddleware())

	// 配置静态文件服务
	r.Static("/uploads", "./uploads")

	// 设置路由
	routes.SetupRoutes(r, db)

	// 启动服务器
	port := config.AppConfig.ServerPort
	fmt.Printf("Server starting on port %d...\n", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
