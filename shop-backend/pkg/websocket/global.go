package websocket

import (
	"sync"

	"shop-backend/utils"
)

var (
	globalPublisher *MessagePublisher
	publisherMu     sync.RWMutex
)

func SetGlobalPublisher(p *MessagePublisher) {
	publisherMu.Lock()
	globalPublisher = p
	publisherMu.Unlock()
	utils.Info("[WS] Global WebSocket publisher 已设置")
}

func GetGlobalPublisher() *MessagePublisher {
	publisherMu.RLock()
	defer publisherMu.RUnlock()
	return globalPublisher
}

func SendToCustomer(customerID int, messageType string, data interface{}) bool {
	publisher := GetGlobalPublisher()
	if publisher == nil {
		utils.Error("[WS] WebSocket publisher 未初始化，无法发送消息 | customerID: %d | 类型: %s", customerID, messageType)
		return false
	}
	return publisher.SendToCustomer(customerID, messageType, data) == nil
}

func SendToCustomerAsync(customerID int, messageType string, data interface{}) {
	go func() {
		if !SendToCustomer(customerID, messageType, data) {
			utils.Error("[WS] 异步发送消息失败 | customerID: %d | 类型: %s", customerID, messageType)
		}
	}()
}
