package controllers

import (
	"net/http"
	"strconv"

	"shop-backend/config"
	"shop-backend/models"

	"github.com/gin-gonic/gin"
)

// 获取购物车
func GetCart(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		// 未登录用户，返回空购物车
		ResponseSuccess(c, gin.H{"items": []CartItem{}})
		return
	}

	// 从数据库获取购物车
	var cart models.Cart
	result := config.DB.Where("user_id = ?", userID).Preload("Items").First(&cart)
	if result.RowsAffected == 0 {
		// 购物车不存在，返回空购物车
		ResponseSuccess(c, gin.H{"items": []CartItem{}})
		return
	}

	// 转换为前端需要的格式
	var items []CartItem
	for _, item := range cart.Items {
		items = append(items, CartItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			SKU:       item.SKU,
		})
	}

	ResponseSuccess(c, gin.H{"items": items})
}

// 添加商品到购物车
func AddToCart(c *gin.Context) {
	var item CartItem
	if err := c.ShouldBindJSON(&item); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 查找或创建购物车
	var cart models.Cart
	result := config.DB.Where("user_id = ?", userID).First(&cart)
	if result.RowsAffected == 0 {
		// 创建新购物车
		cart = models.Cart{
			UserID: userID.(uint),
		}
		if err := config.DB.Create(&cart).Error; err != nil {
			ResponseError(c, http.StatusInternalServerError, "Failed to create cart")
			return
		}
	}

	// 检查商品是否已在购物车中
	var existingItem models.CartItem
	result = config.DB.Where("cart_id = ? AND product_id = ? AND sku = ?", cart.ID, item.ProductID, item.SKU).First(&existingItem)
	if result.RowsAffected > 0 {
		// 更新数量
		existingItem.Quantity += item.Quantity
		if err := config.DB.Save(&existingItem).Error; err != nil {
			ResponseError(c, http.StatusInternalServerError, "Failed to update cart item")
			return
		}
	} else {
		// 添加新商品
		newItem := models.CartItem{
			CartID:    cart.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			SKU:       item.SKU,
		}
		if err := config.DB.Create(&newItem).Error; err != nil {
			ResponseError(c, http.StatusInternalServerError, "Failed to add item to cart")
			return
		}
	}

	ResponseSuccess(c, gin.H{
		"message": "Item added to cart",
		"item":    item,
	})
}

// 更新购物车项
func UpdateCartItem(c *gin.Context) {
	itemID := c.Param("id")
	itemIDUint, err := strconv.ParseUint(itemID, 10, 32)
	if err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid item ID")
		return
	}

	type UpdateRequest struct {
		Quantity int `json:"quantity" binding:"required"`
	}

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 查找购物车项并验证所有权
	var cartItem models.CartItem
	result := config.DB.Joins("JOIN carts ON cart_items.cart_id = carts.id").Where("cart_items.id = ? AND carts.user_id = ?", itemIDUint, userID).First(&cartItem)
	if result.RowsAffected == 0 {
		ResponseError(c, http.StatusNotFound, "Cart item not found")
		return
	}

	// 更新数量
	cartItem.Quantity = req.Quantity
	if err := config.DB.Save(&cartItem).Error; err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to update cart item")
		return
	}

	ResponseSuccess(c, gin.H{
		"message":  "Cart item updated",
		"item_id":  itemID,
		"quantity": req.Quantity,
	})
}

// 移除购物车项
func RemoveCartItem(c *gin.Context) {
	itemID := c.Param("id")
	itemIDUint, err := strconv.ParseUint(itemID, 10, 32)
	if err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid item ID")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 查找购物车项并验证所有权
	var cartItem models.CartItem
	result := config.DB.Joins("JOIN carts ON cart_items.cart_id = carts.id").Where("cart_items.id = ? AND carts.user_id = ?", itemIDUint, userID).First(&cartItem)
	if result.RowsAffected == 0 {
		ResponseError(c, http.StatusNotFound, "Cart item not found")
		return
	}

	// 删除购物车项
	if err := config.DB.Delete(&cartItem).Error; err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to remove cart item")
		return
	}

	ResponseSuccess(c, gin.H{
		"message": "Cart item removed",
		"item_id": itemID,
	})
}

// 同步购物车
func SyncCart(c *gin.Context) {
	var cart Cart
	if err := c.ShouldBindJSON(&cart); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 查找或创建购物车
	var dbCart models.Cart
	result := config.DB.Where("user_id = ?", userID).First(&dbCart)
	if result.RowsAffected == 0 {
		// 创建新购物车
		dbCart = models.Cart{
			UserID: userID.(uint),
		}
		if err := config.DB.Create(&dbCart).Error; err != nil {
			ResponseError(c, http.StatusInternalServerError, "Failed to create cart")
			return
		}
	} else {
		// 清空现有购物车项
		if err := config.DB.Where("cart_id = ?", dbCart.ID).Delete(&models.CartItem{}).Error; err != nil {
			ResponseError(c, http.StatusInternalServerError, "Failed to clear cart")
			return
		}
	}

	// 添加新购物车项
	for _, item := range cart.Items {
		newItem := models.CartItem{
			CartID:    dbCart.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			SKU:       item.SKU,
		}
		if err := config.DB.Create(&newItem).Error; err != nil {
			ResponseError(c, http.StatusInternalServerError, "Failed to add item to cart")
			return
		}
	}

	ResponseSuccess(c, gin.H{
		"message": "Cart synced",
		"items":   cart.Items,
	})
}
