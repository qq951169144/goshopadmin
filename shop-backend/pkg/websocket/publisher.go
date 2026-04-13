package websocket

import (
	"log"
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
		log.Printf("Failed to marshal message: %v", err)
		return err
	}

	if !p.hub.SendToCustomer(customerID, jsonMsg) {
		log.Printf("Failed to send message to customer %d: client not connected or send buffer full", customerID)
		return nil
	}

	log.Printf("Message sent to customer %d: type=%s", customerID, messageType)
	return nil
}

func (p *MessagePublisher) Broadcast(messageType string, data interface{}) error {
	msg := NewMessage(messageType, data)
	jsonMsg, err := msg.ToJSON()
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return err
	}

	p.hub.Broadcast <- jsonMsg
	log.Printf("Message broadcasted: type=%s", messageType)
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
