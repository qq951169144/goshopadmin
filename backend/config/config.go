package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config 应用配置结构
type Config struct {
	// 服务器配置
	ServerPort int
	Domain     string

	// 数据库配置
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// Redis配置
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// JWT配置
	JWTSecret     string
	JWTExpireHour int

	// RabbitMQ配置
	MQHost     string
	MQPort     int
	MQUser     string
	MQPassword string
	MQVHost    string
}

// LoadConfig 加载配置，返回配置实例
func LoadConfig() (*Config, error) {
	// 服务器配置
	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %v", err)
	}

	// 数据库配置
	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "3306"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}

	// Redis配置
	redisPort, err := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_PORT: %v", err)
	}

	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %v", err)
	}

	// JWT配置
	jwtExpireHour, err := strconv.Atoi(getEnv("JWT_EXPIRE_HOUR", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRE_HOUR: %v", err)
	}

	// RabbitMQ配置
	mqPort, err := strconv.Atoi(getEnv("MQ_PORT", "5672"))
	if err != nil {
		return nil, fmt.Errorf("invalid MQ_PORT: %v", err)
	}

	// 返回配置实例
	return &Config{
		// 服务器配置
		ServerPort: serverPort,
		Domain:     getEnv("DOMAIN", "http://localhost:8080"),

		// 数据库配置
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     dbPort,
		DBUser:     getEnv("DB_USER", "root"),
		DBPassword: getEnv("DB_PASSWORD", "password"),
		DBName:     getEnv("DB_NAME", "goshopadmin"),

		// Redis配置
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     redisPort,
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		// JWT配置
		JWTSecret:     getEnv("JWT_SECRET", "pqe9SIYpdZf4SPZAcjo4c8T6UcKChcTi"),
		JWTExpireHour: jwtExpireHour,

		// RabbitMQ配置
		MQHost:     getEnv("MQ_HOST", "localhost"),
		MQPort:     mqPort,
		MQUser:     getEnv("MQ_USER", "guest"),
		MQPassword: getEnv("MQ_PASSWORD", "guest"),
		MQVHost:    getEnv("MQ_VHOST", "/"),
	}, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
