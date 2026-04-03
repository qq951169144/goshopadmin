package mq

import (
	"fmt"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

// Producer 消息生产者
type Producer struct {
	conn *Connection
}

// NewProducer 创建新的生产者
func NewProducer(conn *Connection) *Producer {
	return &Producer{
		conn: conn,
	}
}

// Publish 发布消息
func (p *Producer) Publish(exchange, routingKey string, msg interface{}) error {
	// 序列化消息
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发布消息
	err = p.conn.Channel().Publish(
		exchange,
		routingKey,
		false, // 强制投递
		false, // 立即投递
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		return fmt.Errorf("发布消息失败: %w", err)
	}

	return nil
}

// PublishWithTTL 发布带TTL的消息（用于延迟队列）
func (p *Producer) PublishWithTTL(exchange, routingKey string, msg interface{}, ttl int64) error {
	// 序列化消息
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 发布消息
	err = p.conn.Channel().Publish(
		exchange,
		routingKey,
		false, // 强制投递
		false, // 立即投递
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Expiration:   fmt.Sprintf("%d", ttl), // TTL（毫秒）
		},
	)

	if err != nil {
		return fmt.Errorf("发布消息失败: %w", err)
	}

	return nil
}