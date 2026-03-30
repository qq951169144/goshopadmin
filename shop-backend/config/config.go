package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
	JWTSecret     string
	JWTExpireHour int
	// RabbitMQ配置
	MQHost     string
	MQPort     int
	MQUser     string
	MQPassword string
	MQVHost    string

	// 布隆过滤器配置
	EnableBloomFilter bool
}

// LoadConfig 返回配置实例（非全局变量）
func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	jwtExpireHour, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOUR", "24"))
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	// RabbitMQ配置
	mqPort, err := strconv.Atoi(getEnv("MQ_PORT", "5672"))
	if err != nil {
		return nil, fmt.Errorf("invalid MQ_PORT: %v", err)
	}

	// 布隆过滤器配置
	enableBloomFilter, err := strconv.ParseBool(getEnv("ENABLE_BLOOM_FILTER", "true"))
	if err != nil {
		return nil, fmt.Errorf("invalid ENABLE_BLOOM_FILTER: %v", err)
	}

	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "8081"),
		DBHost:        getEnv("DB_HOST", "mysql"),
		DBPort:        getEnv("DB_PORT", "3306"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPassword:    getEnv("DB_PASSWORD", "password"),
		DBName:        getEnv("DB_NAME", "goshopadmin"),
		RedisHost:     getEnv("REDIS_HOST", "redis"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		JWTSecret:     getEnv("JWT_SECRET", "1a4tx4pQczv8y1HMX9KytdUeP2rrVt9q"),
		JWTExpireHour: jwtExpireHour,

		// RabbitMQ配置
		MQHost:     getEnv("MQ_HOST", "localhost"),
		MQPort:     mqPort,
		MQUser:     getEnv("MQ_USER", "guest"),
		MQPassword: getEnv("MQ_PASSWORD", "guest"),
		MQVHost:    getEnv("MQ_VHOST", "/"),

		// 布隆过滤器配置
		EnableBloomFilter: enableBloomFilter,
	}, nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
