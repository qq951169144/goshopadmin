package mq

import (
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
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

// Consume 消费消息
func (c *Consumer) Consume(queue string, handler func([]byte) error) error {
	// 消费消息
	msgs, err := c.conn.Channel().Consume(
		queue,
		"",    // 消费者标签
		false, // 自动确认
		false, // 独占
		false, // 本地
		false, // 无等待
		nil,   // 参数
	)

	if err != nil {
		return fmt.Errorf("注册消费者失败: %w", err)
	}

	// 处理消息
	go func() {
		for msg := range msgs {
			log.Printf("收到消息: %s", string(msg.Body))
			
			// 处理消息
			err := handler(msg.Body)
			if err != nil {
				log.Printf("处理消息失败: %v", err)
				// 拒绝消息并重新入队
				msg.Nack(false, true)
				continue
			}
			
			// 确认消息
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
		durable,  // 持久化
		false,    // 自动删除
		false,    // 独占
		false,    // 无等待
		nil,      // 参数
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