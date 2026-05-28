package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"shop-backend/cache"
	"shop-backend/config"
	"shop-backend/middleware"
	"shop-backend/pkg/mq"
	"shop-backend/pkg/pool"
	"shop-backend/routes"
	"shop-backend/services"
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
)

// 全局池实例
var (
	mqPool     *mq.ConnectionPool
	workerPool *utils.WorkerPool
)

func main() {
	// 初始化日志记录器
	utils.Info("日志系统初始化成功")
	defer utils.CloseLogger()

	// 1. 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化数据库和Redis连接
	conn, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer conn.Close()

	// 3. 设置JWT密钥到中间件
	middleware.SetJWTSecret(cfg.JWTSecret)

	// 4. 创建Gin引擎
	r := gin.New()

	// 5. 注册中间件
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS())
	r.Use(gin.Recovery())

	// 6. 配置静态文件服务
	r.Static("/uploads", "./uploads")

	// ========== 新增：初始化MQ连接池 ==========
	mqPool, err = mq.NewConnectionPool(5, 50)
	if err != nil {
		log.Fatalf("创建MQ连接池失败: %v", err)
	}
	defer mqPool.Close()
	utils.Info("MQ连接池初始化成功")

	// ========== 新增：初始化工作池 ==========
	workerPool = utils.NewWorkerPool(2, runtime.NumCPU()*4, 5000)
	defer workerPool.Close()
	utils.Info("工作池初始化成功")

	// ========== 新增：注册池获取函数 ==========
	pool.SetMQConnGetters(
		func() (interface{}, error) { return mqPool.Get() },
		func(conn interface{}) { mqPool.Put(conn.(*mq.Connection)) },
	)
	pool.SetSubmitTask(func(fn func()) { workerPool.Submit(fn) })

	// ========== 新增：初始化协程监控 ==========
	monitor := utils.NewMonitor(1000, 10*time.Second, 100)
	monitor.Start()
	defer monitor.Stop()
	utils.Info("协程监控初始化成功")

	// 7. 设置路由
	routes.SetupRoutes(r, conn.DB, conn.Redis, cfg)

	// 8. 初始化MQ消费者
	go func() {
		// 初始化缓存工具
		cacheUtil := cache.NewCacheUtil(conn.DB, conn.Redis)

		// 初始化服务
		orderService := services.NewOrderService(conn.DB, cacheUtil)
		activityOrderService := services.NewActivityOrderService(conn.DB)
		productService := services.NewProductService(conn.DB, cacheUtil)

		// 初始化MQ消费者
		err := mq.InitConsumers(orderService, activityOrderService, productService)
		if err != nil {
			log.Printf("初始化MQ消费者失败: %v", err)
		}
	}()

	// 9. 启动服务器
	port := cfg.ServerPort
	fmt.Printf("Server starting on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// GetMQPool 获取MQ连接池
func GetMQPool() *mq.ConnectionPool {
	return mqPool
}

// GetWorkerPool 获取工作池
func GetWorkerPool() *utils.WorkerPool {
	return workerPool
}
