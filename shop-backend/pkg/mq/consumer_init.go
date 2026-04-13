package mq

import (
	"shop-backend/constants"
	"shop-backend/pkg/websocket"
	"shop-backend/services"
	"shop-backend/utils"
)

// InitConsumers 初始化消费者
func InitConsumers(orderService *services.OrderService, activityOrderService *services.ActivityOrderService, productService *services.ProductService) error {
	// 初始化WebSocket Hub和Publisher
	hub := websocket.InitHub()
	publisher := websocket.NewMessagePublisher(hub)
	websocket.SetGlobalPublisher(publisher)
	utils.Info("WebSocket初始化完成")

	// 创建MQ连接
	conn, err := NewConnection()
	if err != nil {
		return err
	}

	// 创建消费者
	consumer := NewConsumer(conn)

	// 设置普通订单延迟队列
	err = consumer.SetupDelayQueue(constants.MQQueueOrderDelay, constants.MQQueueOrderDeadLetter)
	if err != nil {
		return err
	}

	// 设置活动订单延迟队列
	err = consumer.SetupDelayQueue(constants.MQQueueActivityOrderDelay, constants.MQQueueActivityOrderDeadLetter)
	if err != nil {
		return err
	}

	// 声明活动订单交换机和队列
	err = consumer.DeclareExchange(constants.MQExchangeActivity, "direct", true)
	if err != nil {
		return err
	}
	_, err = consumer.DeclareQueue(constants.MQQueueActivityOrder, true)
	if err != nil {
		return err
	}
	err = consumer.BindQueue(constants.MQQueueActivityOrder, constants.MQExchangeActivity, constants.MQRoutingKeyActivityOrder)
	if err != nil {
		return err
	}

	// 声明订单状态交换机和队列
	err = consumer.DeclareExchange(constants.MQExchangeOrderStatus, "fanout", true)
	if err != nil {
		return err
	}
	_, err = consumer.DeclareQueue(constants.MQQueueOrderStatus, true)
	if err != nil {
		return err
	}
	err = consumer.BindQueue(constants.MQQueueOrderStatus, constants.MQExchangeOrderStatus, "")
	if err != nil {
		return err
	}

	// 启动订单超时消费者
	orderConsumer := NewOrderConsumer(orderService)
	err = consumer.Consume(constants.MQQueueOrderDeadLetter, orderConsumer.HandleTimeoutOrder)
	if err != nil {
		return err
	}

	// 启动活动订单消费者
	activityConsumer := NewActivityConsumer(activityOrderService, productService)
	err = consumer.Consume(constants.MQQueueActivityOrder, activityConsumer.HandleActivityOrder)
	if err != nil {
		return err
	}

	// 启动活动订单超时消费者
	err = consumer.Consume(constants.MQQueueActivityOrderDeadLetter, activityConsumer.HandleTimeoutActivityOrder)
	if err != nil {
		return err
	}

	// 启动状态变更消费者
	statusConsumer := NewStatusConsumer()
	err = consumer.Consume(constants.MQQueueOrderStatus, statusConsumer.HandleOrderStatus)
	if err != nil {
		return err
	}

	utils.Info("消费者初始化完成")
	return nil
}
