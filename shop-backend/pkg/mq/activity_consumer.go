package mq

import (
	"encoding/json"

	"shop-backend/constants"
	"shop-backend/services"
	"shop-backend/utils"
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
	// 解析消息
	var req struct {
		CustomerID int `json:"customer_id"`
		ActivityID int `json:"activity_id"`
		Items      []struct {
			ProductID int `json:"product_id"`
			SkuID     int `json:"sku_id"`
			Quantity  int `json:"quantity"`
		}
	}

	if err := json.Unmarshal(msg, &req); err != nil {
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
	order, err := ac.activityOrderService.CreateActivityOrder(req.CustomerID, req.ActivityID, activityOrderItems)
	if err != nil {
		utils.Error("创建活动订单失败: %v", err)
		return err
	}

	// 发送延迟消息，30分钟后检查订单状态
	go func() {
		conn, err := NewConnection()
		if err != nil {
			utils.Error("创建MQ连接失败: %v", err)
			return
		}
		defer conn.Close()

		producer := NewProducer(conn)

		// 构建延迟消息
		delayMsg := map[string]interface{}{
			"order_id":   order.ID,
			"created_at": order.CreatedAt,
		}

		// 30分钟超时
		err = producer.PublishWithTTL("", constants.MQQueueActivityOrderDelay, delayMsg, constants.MQOrderTimeoutTTL)
		if err != nil {
			utils.Error("发送活动订单延迟消息失败: %v", err)
		}
	}()

	utils.Info("活动订单创建成功")
	return nil
}

// HandleTimeoutActivityOrder 处理超时活动订单
func (ac *ActivityConsumer) HandleTimeoutActivityOrder(msg []byte) error {
	// 解析消息
	var message struct {
		OrderID   int    `json:"order_id"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(msg, &message); err != nil {
		return err
	}

	// 获取活动订单
	order, err := ac.activityOrderService.GetActivityOrderByID(message.OrderID, 0) // 0表示不验证customerID
	if err != nil {
		return err
	}

	// 直接调用CancelActivityOrder方法取消订单
	// CancelActivityOrder方法内部已包含状态检查等逻辑
	err = ac.activityOrderService.CancelActivityOrder(message.OrderID, order.CustomerID)
	if err != nil {
		return err
	}

	utils.Info("活动订单 %d 超时未支付，已自动取消", message.OrderID)
	return nil
}
