package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"shop-backend/constants"
	"shop-backend/models"
	"shop-backend/utils"
	"time"

	"gorm.io/gorm"
)

// ActivityOrderService 活动订单服务
type ActivityOrderService struct {
	DB              *gorm.DB
	activityService *ActivityService
}

// NewActivityOrderService 创建活动订单服务实例
func NewActivityOrderService(db *gorm.DB) *ActivityOrderService {
	return &ActivityOrderService{
		DB:              db,
		activityService: NewActivityService(db),
	}
}

// ActivityOrderItem 活动订单商品项
type ActivityOrderItem struct {
	AddressID int
	ProductID int
	SkuID     int
	Quantity  int
}

// ActivityOrderResponse 活动订单响应结构体
type ActivityOrderResponse struct {
	ID           int                         `json:"id"`
	OrderNo      string                      `json:"order_no"`
	CustomerID   int                         `json:"customer_id"`
	ActivityID   int                         `json:"activity_id"`
	ActivityName string                      `json:"activity_name"`
	TotalAmount  float64                     `json:"total_amount"`
	Status       string                      `json:"status"`
	CreatedAt    time.Time                   `json:"created_at"`
	Items        []ActivityOrderItemResponse `json:"items"`
}

// ActivityOrderItemResponse 活动订单项响应结构体
type ActivityOrderItemResponse struct {
	ID            int     `json:"id"`
	OrderID       int     `json:"order_id"`
	ProductID     int     `json:"product_id"`
	SkuID         int     `json:"sku_id"`
	ProductName   string  `json:"product_name"`
	SkuAttributes string  `json:"sku_attributes"`
	ProductImage  string  `json:"product_image"`
	Price         float64 `json:"price"`
	Quantity      int     `json:"quantity"`
	TotalAmount   float64 `json:"total_amount"`
}

// CreateActivityOrder 创建活动订单
func (s *ActivityOrderService) CreateActivityOrder(customerID int, activityID int, addressID int, items []ActivityOrderItem) (*OrderInfo, error) {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var activity models.Activity
	result := tx.First(&activity, activityID)
	if result.Error != nil {
		tx.Rollback()
		return nil, errors.New("活动不存在")
	}

	now := time.Now()
	isActivityActive := activity.Status == constants.ActivityStatusActive && activity.StartTime.Before(now) && activity.EndTime.After(now)
	if !isActivityActive {
		tx.Rollback()
		return nil, errors.New("活动已结束或未开始")
	}

	var address models.Address
	if err := tx.Where("id = ? AND customer_id = ?", addressID, customerID).First(&address).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("收货地址不存在")
	}

	var totalAmount float64
	orderItems := make([]models.OrderItem, 0, len(items))

	for _, item := range items {
		err := s.activityService.CheckActivityStock(activityID, item.ProductID, item.SkuID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		var sku models.ProductSku
		result := tx.First(&sku, item.SkuID)
		if result.Error != nil {
			tx.Rollback()
			return nil, errors.New("商品SKU不存在")
		}

		itemAmount := sku.Price * float64(item.Quantity)
		totalAmount += itemAmount

		var product models.Product
		s.DB.Where("id = ?", item.ProductID).First(&product)

		skuAttrs := "{}"
		var skuSpecs []models.ProductSkuSpec
		s.DB.Where("sku_id = ?", item.SkuID).Find(&skuSpecs)
		if len(skuSpecs) > 0 {
			attrs := make(map[string]string)
			for _, spec := range skuSpecs {
				var specification models.ProductSpecification
				s.DB.Where("id = ?", spec.SpecID).First(&specification)
				var specValue models.ProductSpecificationValue
				s.DB.Where("id = ?", spec.SpecValueID).First(&specValue)
				if specification.ID > 0 && specValue.ID > 0 {
					attrs[specification.Name] = specValue.Value
				}
			}
			if len(attrs) > 0 {
				attrsJSON, _ := json.Marshal(attrs)
				skuAttrs = string(attrsJSON)
			}
		}

		orderItem := models.OrderItem{
			ProductID:     item.ProductID,
			SkuID:         item.SkuID,
			Quantity:      item.Quantity,
			Price:         sku.Price,
			TotalAmount:   itemAmount,
			ProductName:   product.Name,
			SkuAttributes: skuAttrs,
		}
		orderItems = append(orderItems, orderItem)

		err = s.activityService.ReduceActivityStock(activityID, item.ProductID, item.SkuID, item.Quantity)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	orderNo := s.generateOrderNo()

	order := &models.Order{
		CustomerID:     customerID,
		OrderNo:        orderNo,
		ActivityID:     activityID,
		TotalAmount:    totalAmount,
		AddressID:      addressID,
		Status:         constants.OrderStatusPending,
		PaymentStatus:  constants.PaymentStatusPending,
		ShippingStatus: constants.ShippingStatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	utils.Info("AddressID = %v, order.ActivityID = %v", addressID, order.AddressID)
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	paymentURL := fmt.Sprintf("/api/payment/fake-pay?orderNo=%s", orderNo)

	return &OrderInfo{
		ID:         order.ID,
		OrderID:    orderNo,
		Amount:     totalAmount,
		PaymentURL: paymentURL,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
	}, nil
}

// GetActivityOrders 获取用户活动订单列表
func (s *ActivityOrderService) GetActivityOrders(customerID int, page, pageSize int) ([]ActivityOrderResponse, int64, error) {
	var orders []models.Order
	var total int64

	s.DB.Model(&models.Order{}).Where("customer_id = ? AND activity_id > 0", customerID).Count(&total)

	offset := (page - 1) * pageSize
	result := s.DB.Where("customer_id = ? AND activity_id > 0", customerID).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders)
	if result.Error != nil {
		utils.Error("获取活动订单失败: %v", result.Error)
		return nil, 0, result.Error
	}

	responseOrders := make([]ActivityOrderResponse, len(orders))
	for i, order := range orders {
		// 查询活动名称
		var activity models.Activity
		s.DB.Select("name").Where("id = ?", order.ActivityID).First(&activity)

		// 查询订单商品
		var items []models.OrderItem
		s.DB.Where("order_id = ?", order.ID).Find(&items)

		// 转换商品项
		responseItems := make([]ActivityOrderItemResponse, len(items))
		for j, item := range items {
			// 查询商品主图
			var productImage models.ProductImage
			s.DB.Where("product_id = ? AND is_main = ?", item.ProductID, true).First(&productImage)
			productImageURL := ""
			if productImage.ID > 0 {
				productImageURL = productImage.ImageURL
			} else {
				// 如果没有主图，查询第一张图片
				s.DB.Where("product_id = ?", item.ProductID).First(&productImage)
				if productImage.ID > 0 {
					productImageURL = productImage.ImageURL
				}
			}

			responseItems[j] = ActivityOrderItemResponse{
				ID:            item.ID,
				OrderID:       item.OrderID,
				ProductID:     item.ProductID,
				SkuID:         item.SkuID,
				ProductName:   item.ProductName,
				SkuAttributes: item.SkuAttributes,
				ProductImage:  productImageURL,
				Price:         item.Price,
				Quantity:      item.Quantity,
				TotalAmount:   item.TotalAmount,
			}
		}

		responseOrders[i] = ActivityOrderResponse{
			ID:           order.ID,
			OrderNo:      order.OrderNo,
			CustomerID:   order.CustomerID,
			ActivityID:   order.ActivityID,
			ActivityName: activity.Name,
			TotalAmount:  order.TotalAmount,
			Status:       order.Status,
			CreatedAt:    order.CreatedAt,
			Items:        responseItems,
		}
	}

	return responseOrders, total, nil
}

// GetActivityOrderByID 根据ID获取活动订单详情
func (s *ActivityOrderService) GetActivityOrderByID(orderID int, customerID int) (*ActivityOrderResponse, error) {
	var order models.Order

	result := s.DB.Where("id = ? AND customer_id = ? AND activity_id > 0", orderID, customerID).First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		utils.Error("获取订单详情失败: %v", result.Error)
		return nil, result.Error
	}

	// 查询活动名称
	var activity models.Activity
	s.DB.Select("name").Where("id = ?", order.ActivityID).First(&activity)

	// 查询订单商品
	var items []models.OrderItem
	s.DB.Where("order_id = ?", order.ID).Find(&items)

	// 转换商品项
	responseItems := make([]ActivityOrderItemResponse, len(items))
	for j, item := range items {
		// 查询商品主图
		var productImage models.ProductImage
		s.DB.Where("product_id = ? AND is_main = ?", item.ProductID, true).First(&productImage)
		productImageURL := ""
		if productImage.ID > 0 {
			productImageURL = productImage.ImageURL
		} else {
			// 如果没有主图，查询第一张图片
			s.DB.Where("product_id = ?", item.ProductID).First(&productImage)
			if productImage.ID > 0 {
				productImageURL = productImage.ImageURL
			}
		}

		responseItems[j] = ActivityOrderItemResponse{
			ID:            item.ID,
			OrderID:       item.OrderID,
			ProductID:     item.ProductID,
			SkuID:         item.SkuID,
			ProductName:   item.ProductName,
			SkuAttributes: item.SkuAttributes,
			ProductImage:  productImageURL,
			Price:         item.Price,
			Quantity:      item.Quantity,
			TotalAmount:   item.TotalAmount,
		}
	}

	responseOrder := &ActivityOrderResponse{
		ID:           order.ID,
		OrderNo:      order.OrderNo,
		CustomerID:   order.CustomerID,
		ActivityID:   order.ActivityID,
		ActivityName: activity.Name,
		TotalAmount:  order.TotalAmount,
		Status:       order.Status,
		CreatedAt:    order.CreatedAt,
		Items:        responseItems,
	}

	return responseOrder, nil
}

// generateOrderNo 生成订单号
func (s *ActivityOrderService) generateOrderNo() string {
	timestamp := time.Now().Format("20060102150405")
	random := rand.Intn(10000)
	return fmt.Sprintf("ACT%s%04d", timestamp, random)
}

// CancelActivityOrder 取消活动订单
func (s *ActivityOrderService) CancelActivityOrder(orderID int, customerID int) error {
	tx := s.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order models.Order
	result := tx.Where("id = ? AND customer_id = ? AND activity_id > 0 AND status IN ?",
		orderID, customerID, []string{constants.OrderStatusPending, constants.OrderStatusPaid}).First(&order)
	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在或状态不正确")
		}
		return result.Error
	}

	var items []models.OrderItem
	result = tx.Where("order_id = ?", orderID).Find(&items)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	for _, item := range items {
		var sku models.ProductSku
		result := tx.First(&sku, item.SkuID)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		if err := tx.Model(&sku).Update("stock", sku.Stock+item.Quantity).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status":         constants.OrderStatusCancelled,
		"payment_status": constants.PaymentStatusFailed,
		"updated_at":     time.Now(),
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
