package websocket

import (
	"net/http"
	"strconv"

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

func webSocketResponseError(c *gin.Context, bizCode int, err error) {
	message := errors.GetErrorMessage(bizCode)
	c.JSON(http.StatusOK, gin.H{
		"code":    bizCode,
		"message": message,
		"data":    nil,
	})
}

func WebSocketHandler(c *gin.Context) {
	customerIDStr := c.Query("customer_id")
	customerID, err := strconv.Atoi(customerIDStr)
	if err != nil || customerID <= 0 {
		webSocketResponseError(c, errors.CodeParamError, err)
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
		CustomerID: customerID,
	}

	hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
