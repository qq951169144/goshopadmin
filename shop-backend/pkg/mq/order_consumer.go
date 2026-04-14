package mq

import (
	"encoding/json"
	"shop-backend/constants"
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
	utils.Info("[MQ] 开始处理超时订单消息 | 队列: %s | 消息: %s", constants.MQQueueOrderDeadLetter, string(msg))

	// 解析消息
	var message struct {
		OrderID   int    `json:"order_id"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(msg, &message); err != nil {
		utils.Error("[MQ] 解析超时订单消息失败 | 队列: %s | 错误: %v", constants.MQQueueOrderDeadLetter, err)
		return err
	}

	// 获取订单信息
	order, err := oc.orderService.GetOrderByID(message.OrderID)
	if err != nil {
		utils.Error("[MQ] 获取超时订单失败 | 队列: %s | orderID: %d | 错误: %v", constants.MQQueueOrderDeadLetter, message.OrderID, err)
		return err
	}

	utils.Info("[MQ] 查询到超时订单 | 队列: %s | orderID: %d | customerID: %d | orderNo: %s | status: %s", constants.MQQueueOrderDeadLetter, order.ID, order.CustomerID, order.OrderNo, order.Status)

	// 直接调用CancelOrder方法取消订单
	// CancelOrder方法内部已包含状态检查等逻辑
	err = oc.orderService.CancelOrder(order.OrderNo, order.CustomerID)
	if err != nil {
		utils.Error("[MQ] 取消超时订单失败 | 队列: %s | orderID: %d | customerID: %d | 错误: %v", constants.MQQueueOrderDeadLetter, message.OrderID, order.CustomerID, err)
		return err
	}

	utils.Info("[MQ] 超时订单已取消 | 队列: %s | orderID: %d | customerID: %d", constants.MQQueueOrderDeadLetter, message.OrderID, order.CustomerID)

	// 发送WebSocket站内信通知
	cancelData := map[string]interface{}{
		"order_id": order.ID,
		"order_no": order.OrderNo,
		"status":   "cancelled",
		"reason":   "超时未支付",
	}
	ws.SendToCustomerAsync(order.CustomerID, ws.MessageTypeOrderCanceled, cancelData)
	utils.Info("[WS] 发送订单取消消息 | customerID: %d | 类型: %s | 数据: %v", order.CustomerID, ws.MessageTypeOrderCanceled, cancelData)

	utils.Info("[MQ] 超时订单处理完成 | 队列: %s | orderID: %d", constants.MQQueueOrderDeadLetter, message.OrderID)
	return nil
}
