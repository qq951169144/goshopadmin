package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"shop-backend/constants"
	"shop-backend/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderService 订单服务
type OrderService struct {
	db *gorm.DB
}

// NewOrderService 创建订单服务实例
func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

// OrderItemRequest 订单项请求（前端传入）
type OrderItemRequest struct {
	ProductID int `json:"product_id"`
	SkuID     int `json:"sku_id"`
	Quantity  int `json:"quantity"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	CustomerID int
	AddressID  int
	Items      []OrderItemRequest
	Remark     string
}

// OrderInfo 订单信息
type OrderInfo struct {
	OrderID    string    `json:"order_id"`
	Amount     float64   `json:"amount"`
	PaymentURL string    `json:"payment_url"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(req CreateOrderRequest) (*OrderInfo, error) {
	// 开始事务
	tx := s.db.Begin()

	// 验证地址是否存在且属于当前客户
	var address models.Address
	if err := tx.Where("id = ? AND customer_id = ?", req.AddressID, req.CustomerID).First(&address).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("收货地址不存在")
		}
		return nil, errors.New("查询地址失败")
	}

	// 计算订单金额并检查库存
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		// 查询商品信息（加锁防止并发问题）
		var product models.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", item.ProductID).First(&product).Error; err != nil {
			tx.Rollback()
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("商品ID %d 不存在", item.ProductID)
			}
			return nil, errors.New("查询商品失败")
		}

		// 检查库存
		if product.Stock < item.Quantity {
			tx.Rollback()
			return nil, fmt.Errorf("商品 '%s' 库存不足，当前库存: %d，需要: %d", product.Name, product.Stock, item.Quantity)
		}

		// 扣减库存
		product.Stock -= item.Quantity
		if err := tx.Save(&product).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("扣减库存失败")
		}

		// 查询商品主图
		var productImage models.ProductImage
		tx.Where("product_id = ? AND is_main = ?", item.ProductID, true).First(&productImage)

		// 处理SKU信息
		skuID := item.SkuID
		skuAttrs := "{}"
		itemPrice := product.Price

		if skuID > 0 {
			var sku models.ProductSKU
			if err := tx.Where("id = ? AND product_id = ?", skuID, item.ProductID).First(&sku).Error; err == nil {
				itemPrice = sku.Price
				// 查询SKU规格属性
				var skuSpecs []models.ProductSKUSpec
				tx.Where("sku_id = ?", skuID).Find(&skuSpecs)
				if len(skuSpecs) > 0 {
					attrs := make(map[string]string)
					for _, spec := range skuSpecs {
						// 查询规格名称
						var specification models.ProductSpecification
						tx.Where("id = ?", spec.SpecID).First(&specification)
						// 查询规格值
						var specValue models.ProductSpecificationValue
						tx.Where("id = ?", spec.SpecValueID).First(&specValue)
						if specification.ID > 0 && specValue.ID > 0 {
							attrs[specification.Name] = specValue.Value
						}
					}
					if len(attrs) > 0 {
						attrsJSON, _ := json.Marshal(attrs)
						skuAttrs = string(attrsJSON)
					}
				}
			}
		}

		// 计算金额
		itemAmount := itemPrice * float64(item.Quantity)
		totalAmount += itemAmount

		// 构建订单项
		orderItem := models.OrderItem{
			ProductID:     item.ProductID,
			SkuID:         skuID,
			ProductName:   product.Name,
			SkuAttributes: skuAttrs,
			Price:         itemPrice,
			Quantity:      item.Quantity,
			TotalAmount:   itemAmount,
		}
		orderItems = append(orderItems, orderItem)
	}

	// 生成订单号
	var count int64
	today := time.Now().Format("20060102")
	tx.Model(&models.Order{}).Where("order_no LIKE ?", "ORD"+today+"%").Count(&count)
	orderNo := fmt.Sprintf("ORD%s%04d", today, count+1)

	// 创建订单
	order := models.Order{
		OrderNo:       orderNo,
		CustomerID:    req.CustomerID,
		MerchantID:    1, // 默认商户ID
		TotalAmount:   totalAmount,
		Status:        constants.OrderStatusPending,
		AddressID:     req.AddressID,
		PaymentMethod: "fake",
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("创建订单失败")
	}

	// 保存订单项
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("创建订单项失败")
		}
	}

	// 创建订单项后，清空客户购物车
	var cart models.Cart
	if err := tx.Where("customer_id = ?", req.CustomerID).First(&cart).Error; err == nil {
		// 先删除购物车项
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("清空购物车项失败")
		}
		// 再删除购物车
		if err := tx.Delete(&cart).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("清空购物车失败")
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交事务失败")
	}

	// 生成支付URL
	paymentURL := fmt.Sprintf("/api/payment/fake-pay?order_id=%s", orderNo)

	return &OrderInfo{
		OrderID:    orderNo,
		Amount:     totalAmount,
		PaymentURL: paymentURL,
		Status:     constants.OrderStatusPending,
		CreatedAt:  order.CreatedAt,
	}, nil
}

// OrderDetailItem 订单详情项
type OrderDetailItem struct {
	ProductID     int             `json:"product_id"`
	ProductName   string          `json:"product_name"`
	ProductImage  string          `json:"product_image"`
	SkuCode       string          `json:"sku_code"`
	SkuAttributes json.RawMessage `json:"sku_attributes"`
	Price         float64         `json:"price"`
	Quantity      int             `json:"quantity"`
}

// OrderDetailInfo 订单详情信息
type OrderDetailInfo struct {
	OrderID   string            `json:"order_id"`
	Amount    float64           `json:"amount"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	Address   map[string]string `json:"address"`
	Items     []OrderDetailItem `json:"items"`
}

// GetOrderDetail 获取订单详情
func (s *OrderService) GetOrderDetail(orderNo string, customerID int) (map[string]interface{}, error) {
	// 从数据库获取订单详情，支持按order_no查询
	var order models.Order
	result := s.db.Where("order_no = ? AND customer_id = ?", orderNo, customerID).Preload("Items").First(&order)
	if result.RowsAffected == 0 {
		return nil, errors.New("订单不存在")
	}

	// 查询地址信息
	var address models.Address
	addressInfo := map[string]string{
		"name":           "",
		"phone":          "",
		"province":       "",
		"city":           "",
		"district":       "",
		"detail_address": "",
	}
	if err := s.db.Where("id = ?", order.AddressID).First(&address).Error; err == nil {
		addressInfo["name"] = address.Name
		addressInfo["phone"] = address.Phone
		addressInfo["province"] = address.Province
		addressInfo["city"] = address.City
		addressInfo["district"] = address.District
		addressInfo["detail_address"] = address.DetailAddress
	}

	// 转换为前端需要的格式
	var items []OrderDetailItem
	for _, item := range order.Items {
		// 查询商品主图
		var productImage models.ProductImage
		imageURL := ""
		s.db.Where("product_id = ? AND is_main = ?", item.ProductID, true).First(&productImage)
		if productImage.ID > 0 {
			imageURL = productImage.ImageURL
		}

		// 查询SKU编码
		skuCode := ""
		if item.SkuID > 0 {
			var sku models.ProductSKU
			if err := s.db.Where("id = ?", item.SkuID).First(&sku).Error; err == nil {
				skuCode = sku.SKUCode
			}
		}

		items = append(items, OrderDetailItem{
			ProductID:     item.ProductID,
			ProductName:   item.ProductName,
			ProductImage:  imageURL,
			SkuCode:       skuCode,
			SkuAttributes: json.RawMessage(item.SkuAttributes),
			Price:         item.Price,
			Quantity:      item.Quantity,
		})
	}

	return map[string]interface{}{
		"order_id":   order.ID,
		"order_no":   order.OrderNo,
		"amount":     order.TotalAmount,
		"status":     order.Status,
		"created_at": order.CreatedAt,
		"address":    addressInfo,
		"items":      items,
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(orderNo, status, transactionID string) error {
	var order models.Order
	result := s.db.Where("order_no = ?", orderNo).First(&order)
	if result.RowsAffected == 0 {
		return errors.New("订单不存在")
	}

	order.Status = status
	order.TransactionID = transactionID
	if status == constants.OrderStatusPaid {
		now := time.Now()
		order.PaidAt = &now
	}
	if err := s.db.Save(&order).Error; err != nil {
		return errors.New("更新订单状态失败")
	}

	return nil
}

// CancelOrder 取消订单
func (s *OrderService) CancelOrder(orderNo string, customerID int) error {
	var order models.Order
	result := s.db.Where("order_no = ? AND customer_id = ?", orderNo, customerID).First(&order)
	if result.RowsAffected == 0 {
		return errors.New("订单不存在")
	}

	// 只有待付款或已支付状态的订单可以取消
	if order.Status != constants.OrderStatusPending && order.Status != constants.OrderStatusPaid {
		return errors.New("当前订单状态不允许取消")
	}

	order.Status = constants.OrderStatusCancelled
	now := time.Now()
	order.CancelledAt = &now

	if err := s.db.Save(&order).Error; err != nil {
		return errors.New("取消订单失败")
	}

	return nil
}

// ConfirmReceipt 确认收货
func (s *OrderService) ConfirmReceipt(orderNo string, customerID int) error {
	var order models.Order
	result := s.db.Where("order_no = ? AND customer_id = ?", orderNo, customerID).First(&order)
	if result.RowsAffected == 0 {
		return errors.New("订单不存在")
	}

	// 只有已发货状态的订单可以确认收货
	if order.Status != constants.OrderStatusShipped {
		return errors.New("当前订单状态不允许确认收货")
	}

	order.Status = constants.OrderStatusCompleted
	now := time.Now()
	order.DeliveredAt = &now

	if err := s.db.Save(&order).Error; err != nil {
		return errors.New("确认收货失败")
	}

	return nil
}

// GetOrderByOrderNo 根据订单号获取订单
func (s *OrderService) GetOrderByOrderNo(orderNo string) (*models.Order, error) {
	var order models.Order
	result := s.db.Where("order_no = ?", orderNo).First(&order)
	if result.RowsAffected == 0 {
		return nil, errors.New("订单不存在")
	}
	return &order, nil
}
