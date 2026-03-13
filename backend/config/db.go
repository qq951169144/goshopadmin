package config

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBConnection 封装数据库连接（替代全局变量）
type DBConnection struct {
	DB *gorm.DB
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

	return &DBConnection{
		DB: db,
	}, nil
}

// Close 关闭数据库连接
func (conn *DBConnection) Close() error {
	sqlDB, err := conn.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
