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
	ID          int     `json:"id"`           // 购物车项ID
	ProductID   int     `json:"product_id"`   // 商品ID
	ProductName string  `json:"product_name"` // 商品名称
	MainImage   string  `json:"main_image"`   // 商品主图
	SkuID       int     `json:"sku_id"`       // SKU ID
	SkuCode     string  `json:"sku_code"`     // SKU编码
	Quantity    int     `json:"quantity"`     // 数量
	Price       float64 `json:"price"`        // 价格
}

// CartInfo 购物车信息
type CartInfo struct {
	Items []CartItemInfo `json:"items"`
}

// GetCart 获取购物车
func (s *CartService) GetCart(customerID int) (*CartInfo, error) {
	// 从数据库获取购物车
	var cart models.Cart
	result := s.db.Where("customer_id = ?", int(customerID)).First(&cart)
	if result.RowsAffected == 0 {
		// 购物车不存在，返回空购物车
		return &CartInfo{Items: []CartItemInfo{}}, nil
	}

	// 查询购物车项，预加载商品和SKU信息
	var cartItems []models.CartItem
	err := s.db.Where("cart_id = ?", cart.ID).
		Preload("Product").
		Preload("SKU").
		Find(&cartItems).Error
	if err != nil {
		return nil, err
	}

	// 获取所有商品ID，批量查询主图
	var productIDs []int
	for _, item := range cartItems {
		productIDs = append(productIDs, item.ProductID)
	}

	// 批量查询商品主图
	var mainImages []models.ProductImage
	s.db.Where("product_id IN ? AND is_main = ?", productIDs, true).Find(&mainImages)

	// 构建productID到主图的映射
	imageMap := make(map[int]string)
	for _, img := range mainImages {
		imageMap[img.ProductID] = img.ImageURL
	}

	// 转换为前端需要的格式
	var items []CartItemInfo
	for _, item := range cartItems {
		cartItemInfo := CartItemInfo{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			SkuID:     item.SkuID,
		}

		// 设置商品名称
		if item.Product.ID != 0 {
			cartItemInfo.ProductName = item.Product.Name
		}

		// 设置主图
		if imageURL, ok := imageMap[item.ProductID]; ok {
			cartItemInfo.MainImage = imageURL
		}

		// 设置SKU编码
		if item.SKU.ID != 0 {
			cartItemInfo.SkuCode = item.SKU.SKUCode
		}

		items = append(items, cartItemInfo)
	}

	return &CartInfo{Items: items}, nil
}

// AddToCartRequest 添加到购物车请求
type AddToCartRequest struct {
	CustomerID int
	ProductID  int
	SkuID      int
	Quantity   int
	Price      float64
}

// AddToCart 添加商品到购物车
func (s *CartService) AddToCart(req AddToCartRequest) error {
	// 查找或创建购物车
	var cart models.Cart
	result := s.db.Where("customer_id = ?", int(req.CustomerID)).First(&cart)
	if result.RowsAffected == 0 {
		// 创建新购物车
		cart = models.Cart{
			CustomerID: int(req.CustomerID),
		}
		if err := s.db.Create(&cart).Error; err != nil {
			return errors.New("创建购物车失败")
		}
	}

	// 检查商品是否已在购物车中（使用sku_id）
	var existingItem models.CartItem
	result = s.db.Where("cart_id = ? AND product_id = ? AND sku_id = ?", int(cart.ID), req.ProductID, req.SkuID).First(&existingItem)
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
			SkuID:     req.SkuID,
			Quantity:  req.Quantity,
			Price:     req.Price,
		}
		if err := s.db.Create(&newItem).Error; err != nil {
			return errors.New("添加商品到购物车失败")
		}
	}

	return nil
}

// UpdateCartItemRequest 更新购物车项请求
type UpdateCartItemRequest struct {
	CustomerID int
	ItemID     int
	Quantity   int
}

// UpdateCartItem 更新购物车项
func (s *CartService) UpdateCartItem(req UpdateCartItemRequest) error {
	// 查找购物车项并验证所有权
	var cartItem models.CartItem
	result := s.db.Joins("JOIN carts ON cart_items.cart_id = carts.id").Where("cart_items.id = ? AND carts.customer_id = ?", req.ItemID, int(req.CustomerID)).First(&cartItem)
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
	CustomerID int
	ItemID     int
}

// RemoveCartItem 移除购物车项
func (s *CartService) RemoveCartItem(req RemoveCartItemRequest) error {
	// 查找购物车项并验证所有权
	var cartItem models.CartItem
	result := s.db.Joins("JOIN carts ON cart_items.cart_id = carts.id").Where("cart_items.id = ? AND carts.customer_id = ?", req.ItemID, int(req.CustomerID)).First(&cartItem)
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
	CustomerID int
	Items      []CartItemInfo
}

// SyncCart 同步购物车
func (s *CartService) SyncCart(req SyncCartRequest) error {
	// 查找或创建购物车
	var cart models.Cart
	result := s.db.Where("customer_id = ?", int(req.CustomerID)).First(&cart)
	if result.RowsAffected == 0 {
		// 创建新购物车
		cart = models.Cart{
			CustomerID: int(req.CustomerID),
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
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
			Price:     item.Price,
		}
		if err := s.db.Create(&newItem).Error; err != nil {
			return errors.New("添加购物车项失败")
		}
	}

	return nil
}
