package mq

import (
	"fmt"

	"shop-backend/constants"
	"github.com/rabbitmq/amqp091-go"
)

// SetupDelayQueue 设置延迟队列
// 使用死信队列 + TTL 实现延迟功能
func (c *Consumer) SetupDelayQueue(delayQueue, deadLetterQueue string) error {
	// 声明死信交换机
	err := c.DeclareExchange(constants.MQExchangeDeadLetter, "direct", true)
	if err != nil {
		return fmt.Errorf("声明死信交换机失败: %w", err)
	}

	// 声明死信队列
	_, err = c.DeclareQueue(deadLetterQueue, true)
	if err != nil {
		return fmt.Errorf("声明死信队列失败: %w", err)
	}

	// 绑定死信队列到死信交换机
	err = c.BindQueue(deadLetterQueue, constants.MQExchangeDeadLetter, constants.MQRoutingKeyDeadLetter)
	if err != nil {
		return fmt.Errorf("绑定死信队列失败: %w", err)
	}

	// 声明延迟队列（设置死信参数）
	_, err = c.conn.Channel().QueueDeclare(
		delayQueue,
		true, // 持久化
		false,
		false,
		false,
		amqp091.Table{
			"x-dead-letter-exchange":    constants.MQExchangeDeadLetter,
			"x-dead-letter-routing-key": constants.MQRoutingKeyDeadLetter,
		},
	)

	if err != nil {
		return fmt.Errorf("声明延迟队列失败: %w", err)
	}

	return nil
}