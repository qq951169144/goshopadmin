package config

import (
	"context"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConnection 封装数据库连接（替代全局变量）
type DBConnection struct {
	DB    *gorm.DB
	Redis *redis.Client
}

// InitDB 初始化数据库连接，返回实例
func InitDB(cfg *Config) (*DBConnection, error) {
	dsn := cfg.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	log.Println("Database connected successfully")

	// 初始化Redis连接
	var redisClient *redis.Client
	redisAddr := cfg.GetRedisAddr()
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// 测试Redis连接
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Failed to connect Redis: %v", err)
		// Redis连接失败，不中断启动
		redisClient = nil
	} else {
		log.Println("Redis connected successfully")
	}

	return &DBConnection{
		DB:    db,
		Redis: redisClient,
	}, nil
}

// Close 关闭数据库和Redis连接
func (conn *DBConnection) Close() error {
	sqlDB, err := conn.DB.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Close(); err != nil {
		return err
	}

	if conn.Redis != nil {
		if err := conn.Redis.Close(); err != nil {
			return err
		}
	}

	return nil
}
