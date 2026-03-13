package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"shop-backend/services"
)

// CartController 购物车控制器
type CartController struct {
	BaseController
	cartService *services.CartService
}

// NewCartController 创建购物车控制器实例
func NewCartController(cartService *services.CartService) *CartController {
	return &CartController{
		cartService: cartService,
	}
}

// CartItemRequest 购物车项请求结构
type CartItemRequest struct {
	ProductID int     `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	SKU       string  `json:"sku"`
}

// GetCart 获取购物车
func (c *CartController) GetCart(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		// 未登录用户，返回空购物车
		c.ResponseSuccess(ctx, gin.H{"items": []services.CartItemInfo{}})
		return
	}

	// 从服务层获取购物车
	cart, err := c.cartService.GetCart(userID.(uint))
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{"items": cart.Items})
}

// AddToCart 添加商品到购物车
func (c *CartController) AddToCart(ctx *gin.Context) {
	var req CartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 添加到购物车
	err := c.cartService.AddToCart(services.AddToCartRequest{
		UserID:    userID.(uint),
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
		Price:     req.Price,
		SKU:       req.SKU,
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "Item added to cart",
		"item":    req,
	})
}

// UpdateCartItemRequest 更新购物车项请求结构
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required"`
}

// UpdateCartItem 更新购物车项
func (c *CartController) UpdateCartItem(ctx *gin.Context) {
	itemID := ctx.Param("id")
	itemIDUint, err := strconv.ParseUint(itemID, 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid item ID")
		return
	}

	var req UpdateCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 更新购物车项
	err = c.cartService.UpdateCartItem(services.UpdateCartItemRequest{
		UserID:   userID.(uint),
		ItemID:   uint(itemIDUint),
		Quantity: req.Quantity,
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message":  "Cart item updated",
		"item_id":  itemID,
		"quantity": req.Quantity,
	})
}

// RemoveCartItem 移除购物车项
func (c *CartController) RemoveCartItem(ctx *gin.Context) {
	itemID := ctx.Param("id")
	itemIDUint, err := strconv.ParseUint(itemID, 10, 32)
	if err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid item ID")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 移除购物车项
	err = c.cartService.RemoveCartItem(services.RemoveCartItemRequest{
		UserID: userID.(uint),
		ItemID: uint(itemIDUint),
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "Cart item removed",
		"item_id": itemID,
	})
}

// SyncCartRequest 同步购物车请求结构
type SyncCartRequest struct {
	Items []services.CartItemInfo `json:"items"`
}

// SyncCart 同步购物车
func (c *CartController) SyncCart(ctx *gin.Context) {
	var req SyncCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := ctx.Get("user_id")
	if !exists {
		c.ResponseError(ctx, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 同步购物车
	err := c.cartService.SyncCart(services.SyncCartRequest{
		UserID: userID.(uint),
		Items:  req.Items,
	})
	if err != nil {
		c.ResponseError(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "Cart synced",
		"items":   req.Items,
	})
}
