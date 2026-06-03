package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"search-service/config"
	"search-service/utils"

	"github.com/olivere/elastic/v7"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ============================================================
// Elasticsearch 客户端和 MySQL 连接管理
// 提供全局的 ES 客户端和 MySQL 连接实例
// 包含熔断器机制，防止 ES 故障时拖垮整个服务
// ============================================================

// 全局客户端实例
var (
	// esClient Elasticsearch 客户端实例
	esClient *elastic.Client

	// db MySQL 数据库连接实例，用于数据补齐同步
	db *gorm.DB

	// clientOnce 确保 ES 客户端只初始化一次
	clientOnce sync.Once

	// dbOnce 确保 MySQL 连接只初始化一次
	dbOnce sync.Once
)

// ESHealthResult ES 健康检查结果
type ESHealthResult struct {
	// Connected 是否已连接到 ES 集群
	Connected bool

	// ClusterStatus 集群状态：green / yellow / red
	ClusterStatus string

	// IKPlugin 是否检测到 IK 分词插件
	IKPlugin bool
}

// ============================================================
// 熔断器（Circuit Breaker）
// 当 ES 连续失败超过阈值时，自动切断请求，避免雪崩效应
// 状态机：Closed（正常）→ Open（熔断）→ HalfOpen（半开）→ Closed
// ============================================================

// CircuitState 熔断器状态类型
type CircuitState int

const (
	// StateClosed 正常状态，请求正常通过
	StateClosed CircuitState = iota

	// StateOpen 熔断状态，所有请求被拒绝
	StateOpen

	// StateHalfOpen 半开状态，允许少量请求通过以测试是否恢复
	StateHalfOpen
)

// CircuitBreaker 熔断器结构体
type CircuitBreaker struct {
	// state 当前状态
	state CircuitState

	// failures 连续失败次数
	failures int

	// maxFailures 最大允许连续失败次数，超过此值触发熔断
	maxFailures int

	// resetTimeout 熔断后等待多久进入半开状态
	resetTimeout time.Duration

	// lastFailure 上次失败时间
	lastFailure time.Time

	// mu 互斥锁，保证并发安全
	mu sync.Mutex
}

// circuitBreaker 全局熔断器实例
var circuitBreaker = &CircuitBreaker{
	state:        StateClosed,
	maxFailures:  5,
	resetTimeout: 30 * time.Second,
}

// Allow 判断是否允许请求通过
// Closed 状态：允许
// Open 状态：如果超过 resetTimeout 则转为 HalfOpen 并允许，否则拒绝
// HalfOpen 状态：允许
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// 检查是否超过冷却时间
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			utils.Info("熔断器进入半开状态，尝试恢复")
			return true
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// RecordSuccess 记录一次成功请求
// 如果当前是 HalfOpen 状态，则恢复为 Closed
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		utils.Info("熔断器恢复正常状态")
	}
}

// RecordFailure 记录一次失败请求
// 如果连续失败次数达到阈值，则进入 Open 状态
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	if cb.failures >= cb.maxFailures {
		if cb.state != StateOpen {
			cb.state = StateOpen
			utils.Warn("熔断器进入熔断状态，连续失败 %d 次", cb.failures)
		}
	}
}

// GetState 获取当前熔断器状态
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// InitESClient 初始化 Elasticsearch 客户端
// esHosts: ES 节点地址列表，多个地址用逗号分隔
// 设置 Sniff=false（不自动发现节点，适合 Docker 环境）
// 启动后进行健康检查，确保连接正常
func InitESClient(esHosts string) error {
	var initErr error

	clientOnce.Do(func() {
		// 解析 ES 地址列表
		hosts := strings.Split(esHosts, ",")
		for i := range hosts {
			hosts[i] = strings.TrimSpace(hosts[i])
		}

		// 创建 ES 客户端
		// SetURL: 设置 ES 节点地址
		// SetSniff: 禁用节点自动发现（Docker 环境下节点 IP 可能不可达）
		// SetHealthcheck: 启用健康检查
		// SetHealthcheckInterval: 每 30 秒检查一次
		client, err := elastic.NewClient(
			elastic.SetURL(hosts...),
			elastic.SetSniff(false),
			elastic.SetHealthcheck(true),
			elastic.SetHealthcheckInterval(30*time.Second),
		)
		if err != nil {
			initErr = fmt.Errorf("创建 ES 客户端失败: %w", err)
			utils.Error("创建 ES 客户端失败: %v", err)
			return
		}

		esClient = client

		// 验证连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		info, code, err := client.Ping(hosts[0]).Do(ctx)
		if err != nil {
			initErr = fmt.Errorf("ES 连接验证失败: %w", err)
			utils.Error("ES 连接验证失败: %v", err)
			return
		}

		utils.Info("ES 连接成功, 版本: %s, 状态码: %d", info.Version.Number, code)
	})

	return initErr
}

// InitDB 初始化 MySQL 数据库连接
// cfg: 配置对象，包含数据库连接信息
// 用于数据补齐同步，从 MySQL 读取最近更新的数据写入 ES
func InitDB(cfg *config.Config) error {
	var initErr error

	dbOnce.Do(func() {
		// 构建 DSN（数据源名称）
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.DBUser,
			cfg.DBPassword,
			cfg.DBHost,
			cfg.DBPort,
			cfg.DBName,
		)

		// 打开数据库连接
		conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			initErr = fmt.Errorf("MySQL 连接失败: %w", err)
			utils.Error("MySQL 连接失败: %v", err)
			return
		}

		// 配置连接池
		sqlDB, err := conn.DB()
		if err != nil {
			initErr = fmt.Errorf("获取 MySQL 连接池失败: %w", err)
			return
		}

		// SetMaxIdleConns 设置空闲连接池中的最大连接数
		sqlDB.SetMaxIdleConns(5)
		// SetMaxOpenConns 设置数据库的最大打开连接数
		sqlDB.SetMaxOpenConns(20)
		// SetConnMaxLifetime 设置连接可复用的最长时间
		sqlDB.SetConnMaxLifetime(time.Hour)

		db = conn
		utils.Info("MySQL 连接成功: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	})

	return initErr
}

// GetESClient 获取全局 ES 客户端实例
// 必须在 InitESClient 之后调用
func GetESClient() *elastic.Client {
	return esClient
}

// GetDB 获取全局 MySQL 连接实例
// 必须在 InitDB 之后调用
func GetDB() *gorm.DB {
	return db
}

// CheckESHealth 检查 Elasticsearch 健康状态
// 检查内容包括：连接是否正常、集群状态、IK 分词插件是否安装
// 返回 ESHealthResult 结构体
func CheckESHealth() ESHealthResult {
	result := ESHealthResult{
		Connected:     false,
		ClusterStatus: "unknown",
		IKPlugin:      false,
	}

	if esClient == nil {
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 检查集群健康状态
	health, err := esClient.ClusterHealth().Do(ctx)
	if err != nil {
		utils.Warn("ES 集群健康检查失败: %v", err)
		return result
	}

	result.Connected = true
	result.ClusterStatus = health.Status

	// 检查 IK 分词插件是否安装
	// 通过 ES 的 NodesInfo API 检测已安装的插件
	nodesInfo, err := esClient.NodesInfo().Do(ctx)
	if err == nil {
		for _, node := range nodesInfo.Nodes {
			if node.Plugins != nil {
				for _, plugin := range node.Plugins {
					if strings.Contains(strings.ToLower(plugin.Name), "ik") {
						result.IKPlugin = true
						break
					}
				}
			}
			if result.IKPlugin {
				break
			}
		}
	}

	return result
}

// executeWithCircuitBreaker 使用熔断器保护执行 ES 操作
// 如果熔断器处于 Open 状态，直接返回错误
// 执行成功则记录成功，失败则记录失败
func executeWithCircuitBreaker(fn func() error) error {
	if !circuitBreaker.Allow() {
		return errors.New("搜索服务暂时不可用，请稍后重试")
	}

	err := fn()
	if err != nil {
		circuitBreaker.RecordFailure()
		return err
	}

	circuitBreaker.RecordSuccess()
	return nil
}

// parseESHosts 解析 ES 地址字符串为切片
// 输入格式: "http://localhost:9200" 或 "http://node1:9200,http://node2:9200"
func parseESHosts(esHosts string) []string {
	hosts := strings.Split(esHosts, ",")
	for i := range hosts {
		hosts[i] = strings.TrimSpace(hosts[i])
	}
	return hosts
}

// parseIntOrDefault 将字符串解析为整数，失败则返回默认值
func parseIntOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
