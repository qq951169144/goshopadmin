package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	Hub         *Hub
	Conn        *websocket.Conn
	Send        chan []byte
	CustomerID  int
	mu          sync.Mutex
}

type Hub struct {
	Clients      map[int]*Client
	Register     chan *Client
	Unregister   chan *Client
	Broadcast    chan []byte
	mu           sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[int]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Broadcast:  make(chan []byte),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.CustomerID] = client
			h.mu.Unlock()
			log.Printf("WebSocket client registered: customerID=%d", client.CustomerID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.CustomerID]; ok {
				delete(h.Clients, client.CustomerID)
				close(client.Send)
				log.Printf("WebSocket client unregistered: customerID=%d", client.CustomerID)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			h.mu.RLock()
			for _, client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client.CustomerID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToCustomer(customerID int, message []byte) bool {
	h.mu.RLock()
	client, ok := h.Clients[customerID]
	h.mu.RUnlock()

	if !ok {
		return false
	}

	select {
	case client.Send <- message:
		return true
	default:
		return false
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.mu.Lock()
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			c.mu.Unlock()
			if err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type Message struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Time    int64       `json:"time"`
}

func NewMessage(messageType string, data interface{}) *Message {
	return &Message{
		Type: messageType,
		Data: data,
		Time: time.Now().Unix(),
	}
}

func (m *Message) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
