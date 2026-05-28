package mq

import (
	"encoding/json"
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

func getRetryCountFromXDeath(msg amqp091.Delivery) int {
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

func getRetryCountFromBody(body []byte) int {
	var msg map[string]interface{}
	if err := json.Unmarshal(body, &msg); err != nil {
		return 0
	}

	if retryCount, ok := msg["retry_count"]; ok {
		switch v := retryCount.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case int64:
			return int(v)
		}
	}

	return 0
}

func incrementRetryCountAndResend(conn *Connection, body []byte, delayQueue string, ttl int64) error {
	var msg map[string]interface{}
	if err := json.Unmarshal(body, &msg); err != nil {
		return err
	}

	currentRetry := getRetryCountFromBody(body)
	msg["retry_count"] = currentRetry + 1

	producer := NewProducer(conn)
	err := producer.PublishWithTTL("", delayQueue, msg, ttl)
	if err != nil {
		utils.Error("[MQ] 重新发送消息到延迟队列失败 | 队列: %s | retry_count: %d | 错误: %v", delayQueue, currentRetry+1, err)
		return err
	}

	utils.Info("[MQ] 消息已重新发送到延迟队列 | 队列: %s | retry_count: %d", delayQueue, currentRetry+1)
	return nil
}

func sendToAlertQueue(conn *Connection, queue string, msg amqp091.Delivery, retryCount int) error {
	producer := NewProducer(conn)
	body := map[string]interface{}{
		"original_body":  string(msg.Body),
		"retry_count":    retryCount,
		"original_queue": queue,
		"arrival_time":   msg.Timestamp,
	}

	err := producer.Publish("", constants.MQQueueActivityOrderAlert, body)
	if err != nil {
		utils.Error("[MQ] 发送告警消息失败 | 错误: %v", err)
		return err
	}

	utils.Info("[MQ] 消息已发送到告警队列 | 原始队列: %s | 重试次数: %d", queue, retryCount)
	return nil
}

// RetryConfig 重试配置
type RetryConfig struct {
	DelayQueue string
	TTL        int64
}

// Consume 消费消息（支持自定义重试次数）
func (c *Consumer) Consume(queue string, handler func([]byte) error, retryConfig *RetryConfig) error {
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

				retryCount := getRetryCountFromBody(msg.Body)
				xDeathCount := getRetryCountFromXDeath(msg)
				
				utils.Info("消息重试次数(body): %d | x-death次数: %d | 阈值: %d", retryCount, xDeathCount, MaxRetryCount)

				if retryCount >= MaxRetryCount {
					utils.Info("消息重试次数超限，发送到告警队列")
					sendToAlertQueue(c.conn, queue, msg, retryCount)
					msg.Ack(false)
					continue
				}

				if retryConfig != nil {
					err := incrementRetryCountAndResend(c.conn, msg.Body, retryConfig.DelayQueue, retryConfig.TTL)
					if err != nil {
						utils.Error("[MQ] 重新发送失败，执行Nack | 错误: %v", err)
						msg.Nack(false, true)
					} else {
						msg.Ack(false)
					}
				} else {
					msg.Nack(false, true)
				}
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
		durable,
		false,
		false,
		false,
		nil,
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