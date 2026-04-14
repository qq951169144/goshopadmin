package mq

import (
	"encoding/json"
	"errors"
	"time"

	"shop-backend/constants"
	"shop-backend/models"
	ws "shop-backend/pkg/websocket"
	"shop-backend/services"
	"shop-backend/utils"

	"gorm.io/gorm"
)

// ActivityConsumer 活动订单消费者
type ActivityConsumer struct {
	activityOrderService *services.ActivityOrderService
	productService       *services.ProductService
}

// NewActivityConsumer 创建活动订单消费者
func NewActivityConsumer(activityOrderService *services.ActivityOrderService, productService *services.ProductService) *ActivityConsumer {
	return &ActivityConsumer{
		activityOrderService: activityOrderService,
		productService:       productService,
	}
}

// HandleActivityOrder 处理活动订单
func (ac *ActivityConsumer) HandleActivityOrder(msg []byte) error {
	utils.Info("[MQ] 开始处理活动订单消息 | 队列: %s | 消息: %s", constants.MQQueueActivityOrder, string(msg))

	// 解析消息
	var req struct {
		CustomerID int `json:"customer_id"`
		ActivityID int `json:"activity_id"`
		AddressID  int `json:"address_id"`
		Items      []struct {
			ProductID int `json:"product_id"`
			SkuID     int `json:"sku_id"`
			Quantity  int `json:"quantity"`
		}
	}

	if err := json.Unmarshal(msg, &req); err != nil {
		utils.Error("[MQ] 解析活动订单消息失败 | 队列: %s | 错误: %v", constants.MQQueueActivityOrder, err)
		return err
	}

	// 转换为ActivityOrderItem
	activityOrderItems := make([]services.ActivityOrderItem, len(req.Items))
	for i, item := range req.Items {
		activityOrderItems[i] = services.ActivityOrderItem{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
		}
	}

	// 创建活动订单
	order, err := ac.activityOrderService.CreateActivityOrder(req.CustomerID, req.ActivityID, req.AddressID, activityOrderItems)
	if err != nil {
		utils.Error("[MQ] 创建活动订单失败 | 队列: %s | customerID: %d | activityID: %d | 错误: %v", constants.MQQueueActivityOrder, req.CustomerID, req.ActivityID, err)
		return err
	}

	utils.Info("[MQ] 活动订单创建成功 | 队列: %s | orderID: %d | orderNo: %s | amount: %s | customerID: %d", constants.MQQueueActivityOrder, order.ID, order.OrderID, order.Amount, req.CustomerID)

	// 发送WebSocket站内信通知
	messageData := map[string]interface{}{
		"order_id":     order.ID,
		"order_no":     order.OrderID,
		"total_amount": order.Amount,
		"status":       order.Status,
		"created_at":   order.CreatedAt,
	}
	ws.SendToCustomerAsync(req.CustomerID, ws.MessageTypeOrderCreated, messageData)
	utils.Info("[WS] 发送订单创建消息 | customerID: %d | 类型: %s | 数据: %v", req.CustomerID, ws.MessageTypeOrderCreated, messageData)

	// 发送延迟消息，30分钟后检查订单状态
	go func() {
		conn, err := NewConnection()
		if err != nil {
			utils.Error("[MQ] 创建MQ连接失败 | 错误: %v", err)
			return
		}
		defer conn.Close()

		producer := NewProducer(conn)

		// 构建延迟消息
		delayMsg := map[string]interface{}{
			"order_id":   order.ID,
			"created_at": order.CreatedAt,
		}

		utils.Info("[MQ] 发送活动订单延迟消息 | 交换机: %s | 队列: %s | TTL: %dms | 数据: %v", constants.MQExchangeActivity, constants.MQQueueActivityOrderDelay, constants.MQOrderTimeoutTTL, delayMsg)

		// 30分钟超时
		err = producer.PublishWithTTL("", constants.MQQueueActivityOrderDelay, delayMsg, constants.MQOrderTimeoutTTL)
		if err != nil {
			utils.Error("[MQ] 发送活动订单延迟消息失败 | 交换机: %s | 队列: %s | TTL: %dms | 错误: %v", constants.MQExchangeActivity, constants.MQQueueActivityOrderDelay, constants.MQOrderTimeoutTTL, err)
		} else {
			utils.Info("[MQ] 活动订单延迟消息发送成功 | 交换机: %s | 队列: %s | TTL: %dms", constants.MQExchangeActivity, constants.MQQueueActivityOrderDelay, constants.MQOrderTimeoutTTL)
		}
	}()

	utils.Info("[MQ] 活动订单处理完成 | 队列: %s | orderID: %d", constants.MQQueueActivityOrder, order.ID)
	return nil
}

// HandleTimeoutActivityOrder 处理超时活动订单
func (ac *ActivityConsumer) HandleTimeoutActivityOrder(msg []byte) error {
	utils.Info("[MQ] 开始处理超时活动订单消息 | 队列: %s | 消息: %s", constants.MQQueueActivityOrderDeadLetter, string(msg))

	var message struct {
		OrderID   int    `json:"order_id"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(msg, &message); err != nil {
		utils.Error("[MQ] 解析超时活动订单消息失败 | 队列: %s | 错误: %v", constants.MQQueueActivityOrderDeadLetter, err)
		return err
	}

	order, err := ac.getOrderForTimeout(message.OrderID)
	if err != nil {
		utils.Error("[MQ] 获取超时活动订单失败 | 队列: %s | orderID: %d | 错误: %v", constants.MQQueueActivityOrderDeadLetter, message.OrderID, err)
		return err
	}

	utils.Info("[MQ] 查询到超时活动订单 | 队列: %s | orderID: %d | customerID: %d | status: %s", constants.MQQueueActivityOrderDeadLetter, order.ID, order.CustomerID, order.Status)

	if ac.isTerminalStatus(order.Status) {
		utils.Info("[MQ] 超时活动订单已是终态，无需处理 | 队列: %s | orderID: %d | status: %s", constants.MQQueueActivityOrderDeadLetter, order.ID, order.Status)
		return nil
	}

	err = ac.activityOrderService.CancelActivityOrder(message.OrderID, order.CustomerID)
	if err != nil {
		utils.Error("[MQ] 取消超时活动订单失败 | 队列: %s | orderID: %d | customerID: %d | 错误: %v", constants.MQQueueActivityOrderDeadLetter, message.OrderID, order.CustomerID, err)

		currentOrder, checkErr := ac.getOrderForTimeout(message.OrderID)
		if checkErr == nil && currentOrder.Status == constants.OrderStatusCancelled {
			utils.Info("[MQ] 订单已处于取消状态，视为处理成功 | 队列: %s | orderID: %d", constants.MQQueueActivityOrderDeadLetter, message.OrderID)
			return nil
		}

		return err
	}

	utils.Info("[MQ] 超时活动订单已取消 | 队列: %s | orderID: %d | customerID: %d", constants.MQQueueActivityOrderDeadLetter, message.OrderID, order.CustomerID)

	cancelData := map[string]interface{}{
		"order_id": order.ID,
		"order_no": order.OrderNo,
		"status":   "cancelled",
		"reason":   "超时未支付",
	}
	ws.SendToCustomerAsync(order.CustomerID, ws.MessageTypeOrderCanceled, cancelData)
	utils.Info("[WS] 发送订单取消消息 | customerID: %d | 类型: %s | 数据: %v", order.CustomerID, ws.MessageTypeOrderCanceled, cancelData)

	utils.Info("[MQ] 超时活动订单处理完成 | 队列: %s | orderID: %d", constants.MQQueueActivityOrderDeadLetter, message.OrderID)
	return nil
}

func (ac *ActivityConsumer) getOrderForTimeout(orderID int) (*models.Order, error) {
	var order models.Order
	result := ac.activityOrderService.DB.Where("id = ? AND activity_id > 0", orderID).First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		utils.Error("获取超时订单详情失败: %v", result.Error)
		return nil, result.Error
	}

	var items []models.OrderItem
	ac.activityOrderService.DB.Where("order_id = ?", order.ID).Find(&items)
	order.Items = items

	return &order, nil
}

func (ac *ActivityConsumer) isTerminalStatus(status string) bool {
	return status == constants.OrderStatusCancelled ||
		status == constants.OrderStatusCompleted ||
		status == constants.OrderStatusShipped
}

// HandleAlertMessage 处理告警消息（重试超限）
func (ac *ActivityConsumer) HandleAlertMessage(msg []byte) error {
	utils.Info("[MQ] 收到告警消息 | 队列: %s | 消息: %s | 时间: %s", constants.MQQueueActivityOrderAlert, string(msg), time.Now().Format("2006-01-02 15:04:05"))

	utils.Info("[MQ] TODO: 实现邮件通知运维功能")

	return nil
}
