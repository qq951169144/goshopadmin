package controllers

import (
	"strconv"

	"shop-backend/errors"
	"shop-backend/services"
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
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
	SkuID     int     `json:"sku_id"`
	Quantity  int     `json:"quantity" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
}

// GetCart 获取购物车
func (c *CartController) GetCart(ctx *gin.Context) {
	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		// 未登录用户，返回空购物车
		c.ResponseSuccess(ctx, gin.H{"items": []services.CartItemInfo{}})
		return
	}

	// 从服务层获取购物车
	cart, err := c.cartService.GetCart(customerID.(int))
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{"items": cart.Items})
}

// AddToCart 添加商品到购物车
func (c *CartController) AddToCart(ctx *gin.Context) {
	var req CartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	utils.Info("获取customerID = %v", customerID)
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 添加到购物车
	err := c.cartService.AddToCart(services.AddToCartRequest{
		CustomerID: customerID.(int),
		ProductID:  req.ProductID,
		SkuID:      req.SkuID,
		Quantity:   req.Quantity,
		Price:      req.Price,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
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
	itemIDInt, err := strconv.Atoi(itemID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req UpdateCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 更新购物车项
	err = c.cartService.UpdateCartItem(services.UpdateCartItemRequest{
		CustomerID: customerID.(int),
		ItemID:     itemIDInt,
		Quantity:   req.Quantity,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
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
	itemIDInt, err := strconv.Atoi(itemID)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 移除购物车项
	err = c.cartService.RemoveCartItem(services.RemoveCartItemRequest{
		CustomerID: customerID.(int),
		ItemID:     itemIDInt,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
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
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	// 从上下文中获取客户ID
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return
	}

	// 同步购物车
	err := c.cartService.SyncCart(services.SyncCartRequest{
		CustomerID: customerID.(int),
		Items:      req.Items,
	})
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, gin.H{
		"message": "Cart synced",
		"items":   req.Items,
	})
}
