package mq

import (
	"encoding/json"
	"shop-backend/constants"
	"shop-backend/utils"
)

// StatusConsumer 状态变更消费者
type StatusConsumer struct {
}

// NewStatusConsumer 创建状态变更消费者
func NewStatusConsumer() *StatusConsumer {
	return &StatusConsumer{}
}

// HandleOrderStatus 处理订单状态变更
func (sc *StatusConsumer) HandleOrderStatus(msg []byte) error {
	utils.Info("[MQ] 开始处理订单状态变更消息 | 队列: %s | 消息: %s", constants.MQQueueOrderStatus, string(msg))

	// 解析消息
	var message struct {
		CustomerID int    `json:"customer_id"`
		OrderID   int    `json:"order_id"`
		OrderNo   string `json:"order_no"`
		Status    string `json:"status"`
		UpdatedAt string `json:"updated_at"`
	}

	if err := json.Unmarshal(msg, &message); err != nil {
		utils.Error("[MQ] 解析订单状态变更消息失败 | 队列: %s | 错误: %v", constants.MQQueueOrderStatus, err)
		return err
	}

	utils.Info("[MQ] 订单状态变更 | 队列: %s | orderID: %d | orderNo: %s | 新状态: %s | customerID: %d", constants.MQQueueOrderStatus, message.OrderID, message.OrderNo, message.Status, message.CustomerID)

	utils.Info("[MQ] 订单状态变更消息处理完成 | 队列: %s | orderID: %d", constants.MQQueueOrderStatus, message.OrderID)
	return nil
}
