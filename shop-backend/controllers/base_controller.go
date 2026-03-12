package controllers

import (
	"github.com/gin-gonic/gin"
)

// 购物车项结构
type CartItem struct {
	ProductID uint    `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	SKU       string  `json:"sku"`
}

// 购物车结构
type Cart struct {
	Items []CartItem `json:"items"`
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(200, data)
}

func ResponseError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}
