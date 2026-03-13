package services

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"shop-backend/models"
)

// OrderService 订单服务
type OrderService struct {
	db *gorm.DB
}

// NewOrderService 创建订单服务实例
func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{db: db}
}

// OrderItemInfo 订单项信息
type OrderItemInfo struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	SKU       string  `json:"sku"`
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID uint
	Items  []OrderItemInfo
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
	// 计算订单金额
	var amount float64
	for _, item := range req.Items {
		amount += item.Price * float64(item.Quantity)
	}

	// 生成订单ID
	// 获取当天订单数量
	var count int64
	today := time.Now().Format("20060102")
	s.db.Model(&models.Order{}).Where("order_no LIKE ?", "ORD"+today+"%").Count(&count)
	orderID := fmt.Sprintf("ORD%s%04d", today, count+1)

	// 生成支付URL
	paymentURL := fmt.Sprintf("/api/payment/fake-pay?order_id=%s", orderID)

	// 开始事务
	tx := s.db.Begin()

	// 创建订单
	order := models.Order{
		OrderNo:       orderID,
		CustomerID:    int(req.UserID),
		MerchantID:    1, // 默认商户ID
		TotalAmount:   amount,
		Status:        "pending",
		AddressID:     1, // 默认地址ID
		PaymentMethod: "fake",
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("创建订单失败")
	}

	// 创建订单项
	for _, item := range req.Items {
		orderItem := models.OrderItem{
			OrderID:       order.ID,
			ProductID:     item.ProductID,
			SkuID:         0,    // 默认SKU ID
			ProductName:   "",   // 产品名称
			SkuAttributes: "{}", // SKU属性
			Price:         item.Price,
			Quantity:      item.Quantity,
			TotalAmount:   item.Price * float64(item.Quantity),
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("创建订单项失败")
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交事务失败")
	}

	return &OrderInfo{
		OrderID:    orderID,
		Amount:     amount,
		PaymentURL: paymentURL,
		Status:     "pending",
		CreatedAt:  order.CreatedAt,
	}, nil
}

// GetOrderDetail 获取订单详情
func (s *OrderService) GetOrderDetail(orderID string, userID uint) (map[string]interface{}, error) {
	// 从数据库获取订单详情
	var order models.Order
	result := s.db.Where("order_no = ? AND customer_id = ?", orderID, userID).Preload("Items").First(&order)
	if result.RowsAffected == 0 {
		return nil, errors.New("订单不存在")
	}

	// 转换为前端需要的格式
	var items []OrderItemInfo
	for _, item := range order.Items {
		items = append(items, OrderItemInfo{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			SKU:       "",
		})
	}

	return map[string]interface{}{
		"order_id":   order.OrderNo,
		"amount":     order.TotalAmount,
		"status":     order.Status,
		"items":      items,
		"created_at": order.CreatedAt,
		"paid_at":    order.UpdatedAt,
	}, nil
}

// UpdateOrderStatus 更新订单状态
func (s *OrderService) UpdateOrderStatus(orderID, status, transactionID string) error {
	var order models.Order
	result := s.db.Where("order_no = ?", orderID).First(&order)
	if result.RowsAffected == 0 {
		return errors.New("订单不存在")
	}

	order.Status = status
	order.TransactionID = transactionID
	if err := s.db.Save(&order).Error; err != nil {
		return errors.New("更新订单状态失败")
	}

	return nil
}

// GetOrderByID 根据ID获取订单
func (s *OrderService) GetOrderByID(orderID string) (*models.Order, error) {
	var order models.Order
	result := s.db.Where("order_no = ?", orderID).First(&order)
	if result.RowsAffected == 0 {
		return nil, errors.New("订单不存在")
	}
	return &order, nil
}
