package mq

import (
	"encoding/json"
	"shop-backend/services"
	"shop-backend/utils"
	ws "shop-backend/pkg/websocket"
)

// OrderConsumer 订单消费者
type OrderConsumer struct {
	orderService *services.OrderService
}

// NewOrderConsumer 创建订单消费者
func NewOrderConsumer(orderService *services.OrderService) *OrderConsumer {
	return &OrderConsumer{
		orderService: orderService,
	}
}

// HandleTimeoutOrder 处理超时订单
func (oc *OrderConsumer) HandleTimeoutOrder(msg []byte) error {
	// 解析消息
	var message struct {
		OrderID   int    `json:"order_id"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(msg, &message); err != nil {
		return err
	}

	// 获取订单信息
	order, err := oc.orderService.GetOrderByID(message.OrderID)
	if err != nil {
		return err
	}

	// 直接调用CancelOrder方法取消订单
	// CancelOrder方法内部已包含状态检查等逻辑
	err = oc.orderService.CancelOrder(order.OrderNo, order.CustomerID)
	if err != nil {
		return err
	}

	// 发送WebSocket站内信通知
	cancelData := map[string]interface{}{
		"order_id": order.ID,
		"order_no": order.OrderNo,
		"status":   "cancelled",
		"reason":   "超时未支付",
	}
	ws.SendToCustomerAsync(order.CustomerID, ws.MessageTypeOrderCanceled, cancelData)

	utils.Info("订单 %d 超时未支付，已自动取消", message.OrderID)
	return nil
}
