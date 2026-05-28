package mq

import (
	"fmt"

	"shop-backend/constants"
	"shop-backend/utils"

	"github.com/rabbitmq/amqp091-go"
)

const (
	MaxRetryCount = 3
)

// Consumer 消息消费者
type Consumer struct {
	conn *Connection
}

// NewConsumer 创建新的消费者
func NewConsumer(conn *Connection) *Consumer {
	return &Consumer{
		conn: conn,
	}
}

func getRetryCount(msg amqp091.Delivery) int {
	if msg.Headers == nil {
		return 0
	}

	xDeath, ok := msg.Headers["x-death"]
	if !ok {
		return 0
	}

	deaths, ok := xDeath.([]interface{})
	if !ok || len(deaths) == 0 {
		return 0
	}

	for _, death := range deaths {
		deathMap, ok := death.(amqp091.Table)
		if !ok {
			continue
		}

		if count, exists := deathMap["count"]; exists {
			if c, ok := count.(int64); ok {
				return int(c)
			}
		}
	}

	return 0
}

func sendToAlertQueue(conn *Connection, queue string, msg amqp091.Delivery) error {
	producer := NewProducer(conn)
	body := map[string]interface{}{
		"original_body":  string(msg.Body),
		"retry_count":    getRetryCount(msg),
		"original_queue": queue,
		"arrival_time":   msg.Timestamp,
	}

	err := producer.Publish("", constants.MQQueueActivityOrderAlert, body)
	if err != nil {
		utils.Error("[MQ] 发送告警消息失败 | 错误: %v", err)
		return err
	}

	utils.Info("[MQ] 消息已发送到告警队列 | 原始队列: %s | 重试次数: %d", queue, getRetryCount(msg))
	return nil
}

// Consume 消费消息
func (c *Consumer) Consume(queue string, handler func([]byte) error) error {
	msgs, err := c.conn.Channel().Consume(
		queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("注册消费者失败: %w", err)
	}

	go func() {
		for msg := range msgs {
			utils.Info("收到消息: %s", string(msg.Body))

			err := handler(msg.Body)
			if err != nil {
				utils.Error("处理消息失败: %v", err)

				retryCount := getRetryCount(msg)
				utils.Info("消息重试次数: %d | 阈值: %d", retryCount, MaxRetryCount)

				if retryCount >= MaxRetryCount {
					utils.Info("消息重试次数超限，发送到告警队列")
					sendToAlertQueue(c.conn, queue, msg)
					msg.Ack(false)
					continue
				}

				msg.Nack(false, true)
				continue
			}

			msg.Ack(false)
		}
	}()

	return nil
}

// BindQueue 绑定队列到交换机
func (c *Consumer) BindQueue(queue, exchange, routingKey string) error {
	return c.conn.Channel().QueueBind(
		queue,
		routingKey,
		exchange,
		false,
		nil,
	)
}

// DeclareQueue 声明队列
func (c *Consumer) DeclareQueue(name string, durable bool) (amqp091.Queue, error) {
	return c.conn.Channel().QueueDeclare(
		name,
		durable, // 持久化
		false,   // 自动删除
		false,   // 独占
		false,   // 无等待
		nil,     // 参数
	)
}

// DeclareExchange 声明交换机
func (c *Consumer) DeclareExchange(name, kind string, durable bool) error {
	return c.conn.Channel().ExchangeDeclare(
		name,
		kind,
		durable,
		false,
		false,
		false,
		nil,
	)
}
