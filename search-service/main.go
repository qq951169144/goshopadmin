package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"search-service/config"
	"search-service/controllers"
	"search-service/routes"
	"search-service/services"
	"search-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// ============================================================
// 搜索服务入口文件
// 启动流程：
// 1. 初始化日志系统
// 2. 加载配置（环境变量 + .env 文件）
// 3. 初始化 Elasticsearch 客户端
// 4. 初始化 MySQL 连接（用于数据补齐同步）
// 5. 启动数据同步定时任务
// 6. 创建 Gin HTTP 引擎
// 7. 设置路由
// 8. 启动 HTTP 服务器
// 9. 优雅关闭
// ============================================================

func main() {
	// ========================================
	// 步骤 1: 初始化日志系统
	// 服务名称为 "search-service"，会写入每条日志的 service 字段
	// ========================================
	utils.InitLogger("search-service")
	defer utils.CloseLogger()

	utils.Info("搜索服务启动中...")

	// ========================================
	// 步骤 2: 加载配置
	// 尝试加载 .env 文件（开发环境），失败则忽略（生产环境使用系统环境变量）
	// ========================================
	if err := godotenv.Load(); err != nil {
		utils.Warn(".env 文件未找到，使用系统环境变量")
	}

	cfg := config.LoadConfig()
	jwtAdminPreview := "****"
	if len(cfg.JWTSecretAdmin) >= 4 {
		jwtAdminPreview = cfg.JWTSecretAdmin[:4] + "..."
	}
	jwtCustomerPreview := "****"
	if len(cfg.JWTSecretCustomer) >= 4 {
		jwtCustomerPreview = cfg.JWTSecretCustomer[:4] + "..."
	}
	utils.Info("配置加载完成: ES_HOSTS=%s, DB_HOST=%s, JWT_SECRET_ADMIN=%s, JWT_SECRET_CUSTOMER=%s..., SERVER_PORT=%s",
		cfg.ESHosts, cfg.DBHost, jwtAdminPreview, jwtCustomerPreview, cfg.ServerPort)

	// ========================================
	// 步骤 3: 初始化 Elasticsearch 客户端
	// 连接 ES 集群，设置 Sniff=false（适合 Docker 环境）
	// ========================================
	if err := services.InitESClient(cfg.ESHosts); err != nil {
		utils.Error("Elasticsearch 初始化失败: %v", err)
		utils.Warn("搜索服务将在降级模式下运行，搜索功能不可用")
	} else {
		utils.Info("Elasticsearch 连接成功")
	}

	// ========================================
	// 步骤 4: 初始化 MySQL 连接
	// 用于数据补齐同步，从 MySQL 读取最近更新的数据写入 ES
	// ========================================
	if err := services.InitDB(cfg); err != nil {
		utils.Error("MySQL 初始化失败: %v", err)
		utils.Warn("数据同步功能不可用")
	} else {
		utils.Info("MySQL 连接成功")

		// ========================================
		// 步骤 5: 启动数据同步定时任务
		// 每 60 秒同步一次最近 5 分钟更新的数据
		// ========================================
		services.StartSyncService()
		defer services.StopSyncService()
	}

	// ========================================
	// 步骤 6: 创建 Gin HTTP 引擎
	// 设置为 Release 模式，减少日志输出
	// ========================================
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// 使用 Gin 内置的 Recovery 中间件，防止 panic 导致服务崩溃
	engine.Use(gin.Recovery())

	// ========================================
	// 步骤 7: 设置路由
	// 创建控制器实例并注入到路由依赖中
	// ========================================
	deps := &routes.Dependencies{
		ProductController:  &controllers.ProductController{},
		OrderController:    &controllers.OrderController{},
		UserController:     &controllers.UserController{},
		CustomerController: &controllers.CustomerController{},
		HealthController:   &controllers.HealthController{},
		SuggestController:  &controllers.SuggestController{},
		SyncController:     &controllers.SyncController{},
	}
	routes.SetupRoutes(engine, deps)

	// ========================================
	// 步骤 8: 启动 HTTP 服务器
	// ========================================
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	utils.Info("搜索服务启动, 监听地址: %s", addr)

	// 在协程中启动服务器，以便主协程监听退出信号
	go func() {
		if err := engine.Run(addr); err != nil {
			utils.Error("HTTP 服务器启动失败: %v", err)
			os.Exit(1)
		}
	}()

	// ========================================
	// 步骤 9: 优雅关闭
	// 监听系统信号（SIGINT/SIGTERM），收到信号后执行清理工作
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	utils.Info("收到退出信号: %v, 正在关闭服务...", sig)
	utils.Info("搜索服务已停止")
}
