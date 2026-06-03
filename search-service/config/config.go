package config

import "os"

// Config 搜索服务的全局配置结构体
// 包含服务端口、Elasticsearch 连接信息和 MySQL 数据库连接信息
type Config struct {
	// ServerPort HTTP 服务监听端口
	ServerPort string

	// ESHosts Elasticsearch 集群节点地址列表，多个地址用逗号分隔
	// 例如: "http://localhost:9200" 或 "http://node1:9200,http://node2:9200"
	ESHosts string

	// DBHost MySQL 数据库主机地址
	DBHost string

	// DBPort MySQL 数据库端口号
	DBPort string

	// DBUser MySQL 数据库用户名
	DBUser string

	// DBPassword MySQL 数据库密码
	DBPassword string

	// DBName MySQL 数据库名称
	DBName string
}

// globalConfig 全局配置变量，包内私有，通过 GetConfig 获取
var globalConfig *Config

// LoadConfig 从环境变量加载配置
// 优先读取系统环境变量，如果存在 .env 文件则自动加载
// 环境变量名与 Config 结构体字段对应，例如 SERVER_PORT 对应 ServerPort
func LoadConfig() *Config {
	// 尝试加载 .env 文件，如果文件不存在则忽略（使用系统环境变量）
	// godotenv.Load() 不会覆盖已存在的环境变量

	globalConfig = &Config{
		ServerPort: getEnv("SERVER_PORT", "8082"),
		ESHosts:    getEnv("ES_HOSTS", "http://localhost:9200"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "goshopadmin"),
	}

	return globalConfig
}

// GetConfig 获取全局配置实例
// 必须在 LoadConfig 之后调用，否则返回默认配置
func GetConfig() *Config {
	if globalConfig == nil {
		return LoadConfig()
	}
	return globalConfig
}

// getEnv 获取环境变量值，如果不存在则返回默认值
// key: 环境变量名
// defaultValue: 环境变量不存在时使用的默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
