package main

import (
	"fmt"
	"log"
	"time"

	"goshopadmin/config"
	"goshopadmin/middleware"
	"goshopadmin/routes"
	"goshopadmin/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化日志记录器
	utils.InitLogger()
	utils.Info("日志系统初始化成功")

	// 确保在程序退出时关闭日志记录器
	defer utils.CloseLogger()

	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	jwtSecretPreview := "****"
	if len(cfg.JWTSecret) >= 4 {
		jwtSecretPreview = cfg.JWTSecret[:4] + "..."
	}
	utils.Info("配置加载完成: DB_HOST=%s, DB_PORT=%s, JWT_SECRET_PREVIEW=%s, REDIS_HOST=%s, SERVER_PORT=%d",
		cfg.DBHost, cfg.DBPort, jwtSecretPreview, cfg.RedisHost, cfg.ServerPort)

	// 2. 初始化数据库连接
	conn, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// 3. 设置JWT密钥到中间件
	middleware.SetJWTSecret(cfg.JWTSecret)

	// 4. 设置数据库连接到中间件
	middleware.SetDB(conn.DB)

	// 5. 设置Redis客户端到中间件
	middleware.SetRedis(conn.Redis)

	// 6. 创建Gin引擎
	r := gin.New()

	// 7. 注册中间件（注意顺序）
	// 1. Logger 中间件（最先执行，生成 RequestID）
	r.Use(middleware.RequestLogger())

	// 2. CORS 中间件
	r.Use(middleware.CORSMiddleware())

	// 3. Recovery 中间件
	r.Use(gin.Recovery())

	// 7. 配置静态文件服务
	r.Static("/uploads", "./uploads")

	// 9. 初始化协程监控器
	monitor := utils.NewMonitor()
	monitor.Start(5 * time.Second)
	utils.Info("协程监控器初始化成功")

	// 9. 设置路由
	routes.SetupRoutes(r, conn.DB, conn.Redis, cfg, monitor)

	// 11. 启动服务器
	port := cfg.ServerPort
	fmt.Printf("Server starting on port %d...\n", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
