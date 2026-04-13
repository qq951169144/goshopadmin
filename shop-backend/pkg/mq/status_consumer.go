package mq

import (
	"encoding/json"
	"shop-backend/utils"
)

// StatusConsumer 状态变更消费者
type StatusConsumer struct {}

// NewStatusConsumer 创建状态变更消费者
func NewStatusConsumer() *StatusConsumer {
	return &StatusConsumer{}
}

// HandleOrderStatus 处理订单状态变更
func (sc *StatusConsumer) HandleOrderStatus(msg []byte) error {
	// 解析消息
	var message struct {
		OrderID    int    `json:"order_id"`
		Status     string `json:"status"`
		UpdatedAt  string `json:"updated_at"`
	}

	if err := json.Unmarshal(msg, &message); err != nil {
		return err
	}

	// 处理状态变更（这里可以根据需要扩展，比如发送通知、更新缓存等）
	utils.Info("订单 %d 状态变更为: %s", message.OrderID, message.Status)

	// 示例：发送通知
	// 这里可以调用通知服务发送短信、邮件等

	// TODO: 后续追加发送站内信服务

	return nil
}