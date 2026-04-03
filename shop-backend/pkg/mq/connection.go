package mq

import (
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"shop-backend/config"
)

// Connection RabbitMQ连接管理
type Connection struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

// NewConnection 创建新的MQ连接
func NewConnection() (*Connection, error) {
	cfg := config.GetMQConfig()
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)

	// 建立连接
	conn, err := amqp091.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("连接RabbitMQ失败: %w", err)
	}

	// 创建通道
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("创建通道失败: %w", err)
	}

	return &Connection{
		conn:    conn,
		channel: ch,
	}, nil
}

// Close 关闭连接
func (c *Connection) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			log.Printf("关闭通道失败: %v", err)
		}
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Channel 获取通道
func (c *Connection) Channel() *amqp091.Channel {
	return c.channel
}

// Reconnect 重新连接
func (c *Connection) Reconnect() error {
	if err := c.Close(); err != nil {
		log.Printf("关闭旧连接失败: %v", err)
	}

	cfg := config.GetMQConfig()
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/%s",
		cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.VHost)

	// 尝试重连
	var err error
	for i := 0; i < 5; i++ {
		c.conn, err = amqp091.Dial(dsn)
		if err == nil {
			break
		}
		log.Printf("重连失败，%d秒后重试: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil {
		return fmt.Errorf("重连失败: %w", err)
	}

	// 创建新通道
	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("创建通道失败: %w", err)
	}

	return nil
}