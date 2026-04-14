package websocket

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"shop-backend/errors"
	"shop-backend/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub *Hub

func InitHub() *Hub {
	hub = NewHub()
	go hub.Run()
	return hub
}

func GetHub() *Hub {
	return hub
}

func webSocketResponseError(c *gin.Context, bizCode int) {
	message := errors.GetErrorMessage(bizCode)
	c.JSON(http.StatusOK, gin.H{
		"code":    bizCode,
		"message": message,
		"data":    nil,
	})
}

func WebSocketHandler(c *gin.Context) {
	customerID, exists := c.Get("customer_id")
	if !exists {
		utils.Error("CodeUnauthorized: %v, bool = %v", customerID, exists)
		webSocketResponseError(c, errors.CodeUnauthorized)
		return
	}

	customerIDInt, ok := customerID.(int)
	if !ok || customerIDInt <= 0 {
		utils.Error("CodeParamError: %v", customerIDInt)
		webSocketResponseError(c, errors.CodeParamError)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Error("Failed to upgrade connection: %v", err)
		return
	}

	client := &Client{
		Hub:        hub,
		Conn:       conn,
		Send:       make(chan []byte, 256),
		CustomerID: customerIDInt,
	}

	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
