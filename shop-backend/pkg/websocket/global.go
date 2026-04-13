package websocket

import (
	"log"
	"sync"
)

var (
	globalPublisher *MessagePublisher
	publisherMu    sync.RWMutex
)

func SetGlobalPublisher(p *MessagePublisher) {
	publisherMu.Lock()
	globalPublisher = p
	publisherMu.Unlock()
	log.Println("Global WebSocket publisher set")
}

func GetGlobalPublisher() *MessagePublisher {
	publisherMu.RLock()
	defer publisherMu.RUnlock()
	return globalPublisher
}

func SendToCustomer(customerID int, messageType string, data interface{}) bool {
	publisher := GetGlobalPublisher()
	if publisher == nil {
		log.Printf("WebSocket publisher not initialized, cannot send message to customer %d", customerID)
		return false
	}
	return publisher.SendToCustomer(customerID, messageType, data) == nil
}

func SendToCustomerAsync(customerID int, messageType string, data interface{}) {
	go func() {
		if !SendToCustomer(customerID, messageType, data) {
			log.Printf("Failed to send async message to customer %d: %s", customerID, messageType)
		}
	}()
}
