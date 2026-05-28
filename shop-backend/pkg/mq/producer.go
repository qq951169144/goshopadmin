package mq

import (
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
	"shop-backend/utils"
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
	body, err := json.Marshal(msg)
	if err != nil {
		utils.Error("[MQ Producer] 序列化消息失败 | 交换机: %s | 路由键: %s | 错误: %v", exchange, routingKey, err)
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	utils.Info("[MQ Producer] 发布消息 | 交换机: %s | 路由键: %s | 消息: %s", exchange, routingKey, string(body))

	err = p.conn.Channel().Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		utils.Error("[MQ Producer] 发布消息失败 | 交换机: %s | 路由键: %s | 错误: %v", exchange, routingKey, err)
		return fmt.Errorf("发布消息失败: %w", err)
	}

	utils.Info("[MQ Producer] 发布消息成功 | 交换机: %s | 路由键: %s", exchange, routingKey)
	return nil
}

// PublishWithTTL 发布带TTL的消息（用于延迟队列）
func (p *Producer) PublishWithTTL(exchange, routingKey string, msg interface{}, ttl int64) error {
	body, err := json.Marshal(msg)
	if err != nil {
		utils.Error("[MQ Producer] 序列化延迟消息失败 | 交换机: %s | 路由键: %s | TTL: %dms | 错误: %v", exchange, routingKey, ttl, err)
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	utils.Info("[MQ Producer] 发布延迟消息 | 交换机: %s | 路由键: %s | TTL: %dms | 消息: %s", exchange, routingKey, ttl, string(body))

	err = p.conn.Channel().Publish(
		exchange,
		routingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Expiration:  fmt.Sprintf("%d", ttl),
		},
	)

	if err != nil {
		utils.Error("[MQ Producer] 发布延迟消息失败 | 交换机: %s | 路由键: %s | TTL: %dms | 错误: %v", exchange, routingKey, ttl, err)
		return fmt.Errorf("发布消息失败: %w", err)
	}

	utils.Info("[MQ Producer] 发布延迟消息成功 | 交换机: %s | 路由键: %s | TTL: %dms", exchange, routingKey, ttl)
	return nil
}