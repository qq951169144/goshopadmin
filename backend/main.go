package main

import (
	"fmt"
	"log"

	"goshopadmin/config"
	"goshopadmin/middleware"
	"goshopadmin/models"
	"goshopadmin/routes"
	"goshopadmin/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志记录器（确保日志目录创建）
	utils.Info("日志系统初始化成功")

	// 确保在程序退出时关闭日志记录器
	defer utils.CloseLogger()

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

	// 4. 自动迁移模型
	conn.DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Merchant{},
		&models.MerchantUser{},
		&models.MerchantAudit{},
		&models.Product{},
		&models.ProductCategory{},
		&models.ProductImage{},
		&models.ProductSKU{},
		&models.Activity{},
	)

	// 5. 创建Gin引擎
	r := gin.New()

	// 6. 注册中间件（注意顺序）
	// 1. Logger 中间件（最先执行，生成 RequestID）
	r.Use(middleware.RequestLogger())

	// 2. CORS 中间件
	r.Use(middleware.CORSMiddleware())

	// 3. Recovery 中间件
	r.Use(gin.Recovery())

	// 7. 配置静态文件服务
	r.Static("/uploads", "./uploads")

	// 8. 设置路由
	routes.SetupRoutes(r, conn.DB, cfg)

	// 9. 启动服务器
	port := cfg.ServerPort
	fmt.Printf("Server starting on port %d...\n", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
