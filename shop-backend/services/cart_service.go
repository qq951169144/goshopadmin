package services

import (
	"errors"

	"gorm.io/gorm"
	"shop-backend/models"
)

// CartService 购物车服务
type CartService struct {
	db *gorm.DB
}

// NewCartService 创建购物车服务实例
func NewCartService(db *gorm.DB) *CartService {
	return &CartService{db: db}
}

// CartItemInfo 购物车项信息
type CartItemInfo struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	SKU       string  `json:"sku"`
}

// CartInfo 购物车信息
type CartInfo struct {
	Items []CartItemInfo `json:"items"`
}

// GetCart 获取购物车
func (s *CartService) GetCart(userID int) (*CartInfo, error) {
	// 从数据库获取购物车
	var cart models.Cart
	result := s.db.Where("user_id = ?", int(userID)).Preload("Items").First(&cart)
	if result.RowsAffected == 0 {
		// 购物车不存在，返回空购物车
		return &CartInfo{Items: []CartItemInfo{}}, nil
	}

	// 转换为前端需要的格式
	var items []CartItemInfo
	for _, item := range cart.Items {
		items = append(items, CartItemInfo{
			ProductID: int(item.ProductID),
			Quantity:  item.Quantity,
			Price:     item.Price,
			SKU:       item.SKU,
		})
	}

	return &CartInfo{Items: items}, nil
}

// AddToCartRequest 添加到购物车请求
type AddToCartRequest struct {
	UserID    int
	ProductID int
	Quantity  int
	Price     float64
	SKU       string
}

// AddToCart 添加商品到购物车
func (s *CartService) AddToCart(req AddToCartRequest) error {
	// 查找或创建购物车
	var cart models.Cart
	result := s.db.Where("user_id = ?", int(req.UserID)).First(&cart)
	if result.RowsAffected == 0 {
		// 创建新购物车
		cart = models.Cart{
			UserID: int(req.UserID),
		}
		if err := s.db.Create(&cart).Error; err != nil {
			return errors.New("创建购物车失败")
		}
	}

	// 检查商品是否已在购物车中
	var existingItem models.CartItem
	result = s.db.Where("cart_id = ? AND product_id = ? AND sku = ?", int(cart.ID), req.ProductID, req.SKU).First(&existingItem)
	if result.RowsAffected > 0 {
		// 更新数量
		existingItem.Quantity += req.Quantity
		if err := s.db.Save(&existingItem).Error; err != nil {
			return errors.New("更新购物车项失败")
		}
	} else {
		// 添加新商品
		newItem := models.CartItem{
			CartID:    int(cart.ID),
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
			Price:     req.Price,
			SKU:       req.SKU,
		}
		if err := s.db.Create(&newItem).Error; err != nil {
			return errors.New("添加商品到购物车失败")
		}
	}

	return nil
}

// UpdateCartItemRequest 更新购物车项请求
type UpdateCartItemRequest struct {
	UserID   int
	ItemID   int
	Quantity int
}

// UpdateCartItem 更新购物车项
func (s *CartService) UpdateCartItem(req UpdateCartItemRequest) error {
	// 查找购物车项并验证所有权
	var cartItem models.CartItem
	result := s.db.Joins("JOIN carts ON cart_items.cart_id = carts.id").Where("cart_items.id = ? AND carts.user_id = ?", req.ItemID, int(req.UserID)).First(&cartItem)
	if result.RowsAffected == 0 {
		return errors.New("购物车项不存在")
	}

	// 更新数量
	cartItem.Quantity = req.Quantity
	if err := s.db.Save(&cartItem).Error; err != nil {
		return errors.New("更新购物车项失败")
	}

	return nil
}

// RemoveCartItemRequest 移除购物车项请求
type RemoveCartItemRequest struct {
	UserID int
	ItemID int
}

// RemoveCartItem 移除购物车项
func (s *CartService) RemoveCartItem(req RemoveCartItemRequest) error {
	// 查找购物车项并验证所有权
	var cartItem models.CartItem
	result := s.db.Joins("JOIN carts ON cart_items.cart_id = carts.id").Where("cart_items.id = ? AND carts.user_id = ?", req.ItemID, int(req.UserID)).First(&cartItem)
	if result.RowsAffected == 0 {
		return errors.New("购物车项不存在")
	}

	// 删除购物车项
	if err := s.db.Delete(&cartItem).Error; err != nil {
		return errors.New("删除购物车项失败")
	}

	return nil
}

// SyncCartRequest 同步购物车请求
type SyncCartRequest struct {
	UserID int
	Items  []CartItemInfo
}

// SyncCart 同步购物车
func (s *CartService) SyncCart(req SyncCartRequest) error {
	// 查找或创建购物车
	var cart models.Cart
	result := s.db.Where("user_id = ?", int(req.UserID)).First(&cart)
	if result.RowsAffected == 0 {
		// 创建新购物车
		cart = models.Cart{
			UserID: int(req.UserID),
		}
		if err := s.db.Create(&cart).Error; err != nil {
			return errors.New("创建购物车失败")
		}
	} else {
		// 清空现有购物车项
		if err := s.db.Where("cart_id = ?", int(cart.ID)).Delete(&models.CartItem{}).Error; err != nil {
			return errors.New("清空购物车失败")
		}
	}

	// 添加新购物车项
	for _, item := range req.Items {
		newItem := models.CartItem{
			CartID:    int(cart.ID),
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			SKU:       item.SKU,
		}
		if err := s.db.Create(&newItem).Error; err != nil {
			return errors.New("添加购物车项失败")
		}
	}

	return nil
}
