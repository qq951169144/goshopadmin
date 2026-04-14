package websocket

import (
	"shop-backend/utils"
)

type MessagePublisher struct {
	hub *Hub
}

func NewMessagePublisher(hub *Hub) *MessagePublisher {
	return &MessagePublisher{
		hub: hub,
	}
}

func (p *MessagePublisher) SendToCustomer(customerID int, messageType string, data interface{}) error {
	msg := NewMessage(messageType, data)
	jsonMsg, err := msg.ToJSON()
	if err != nil {
		utils.Error("[WS] 消息序列化失败 | customerID: %d | 类型: %s | 错误: %v", customerID, messageType, err)
		return err
	}

	if !p.hub.SendToCustomer(customerID, jsonMsg) {
		utils.Warn("[WS] 发送消息失败，客户未连接或发送缓冲区满 | customerID: %d | 类型: %s", customerID, messageType)
		return nil
	}

	utils.Info("[WS] 消息发送成功 | customerID: %d | 类型: %s", customerID, messageType)
	return nil
}

func (p *MessagePublisher) Broadcast(messageType string, data interface{}) error {
	msg := NewMessage(messageType, data)
	jsonMsg, err := msg.ToJSON()
	if err != nil {
		utils.Error("[WS] 广播消息序列化失败 | 类型: %s | 错误: %v", messageType, err)
		return err
	}

	p.hub.Broadcast <- jsonMsg
	utils.Info("[WS] 广播消息已发送 | 类型: %s", messageType)
	return nil
}

const (
	MessageTypeOrderCreated  = "order_created"
	MessageTypeOrderPaid    = "order_paid"
	MessageTypeOrderShipped = "order_shipped"
	MessageTypeOrderReceived = "order_received"
	MessageTypeOrderCanceled = "order_canceled"
	MessageTypeSystemNotice = "system_notice"
)
