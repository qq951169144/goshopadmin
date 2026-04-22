package services

import (
	"errors"
	"time"

	"shop-backend/models"

	"gorm.io/gorm"
)

// CustomerService 客户服务
type CustomerService struct {
	db *gorm.DB
}

// NewCustomerService 创建客户服务实例
func NewCustomerService(db *gorm.DB) *CustomerService {
	return &CustomerService{db: db}
}

// CustomerResponse 客户响应结构
type CustomerResponse struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// GetProfile 获取个人信息
func (s *CustomerService) GetProfile(customerID int) (*CustomerResponse, error) {
	var customer models.Customer
	result := s.db.First(&customer, customerID)
	if result.RowsAffected == 0 {
		return nil, errors.New("用户不存在")
	}

	return &CustomerResponse{
		ID:        customer.ID,
		Username:  customer.Username,
		Email:     customer.Email,
		CreatedAt: customer.CreatedAt,
	}, nil
}

// UpdateProfileRequest 更新个人信息请求结构
type UpdateProfileRequest struct {
	Username string
	Email    string
}

// UpdateProfile 更新个人信息
func (s *CustomerService) UpdateProfile(customerID int, req UpdateProfileRequest) (*CustomerResponse, error) {
	var customer models.Customer
	result := s.db.First(&customer, customerID)
	if result.RowsAffected == 0 {
		return nil, errors.New("用户不存在")
	}

	// 更新客户信息
	if req.Username != "" {
		customer.Username = req.Username
	}
	if req.Email != "" {
		customer.Email = req.Email
	}

	if err := s.db.Save(&customer).Error; err != nil {
		return nil, errors.New("更新个人信息失败")
	}

	return &CustomerResponse{
		ID:        customer.ID,
		Username:  customer.Username,
		Email:     customer.Email,
		CreatedAt: customer.CreatedAt,
	}, nil
}

// OrderItemResponse 订单项响应结构体
type OrderItemResponse struct {
	ProductID     int     `json:"product_id"`
	ProductName   string  `json:"product_name"`
	ProductImage  string  `json:"product_image"`
	SkuCode       string  `json:"sku_code"`
	SkuAttributes string  `json:"sku_attributes"`
	Price         float64 `json:"price"`
	Quantity      int     `json:"quantity"`
}

// OrderResponse 订单响应结构体
type OrderResponse struct {
	OrderID   int               `json:"order_id"`
	OrderNo   string            `json:"order_no"`
	Amount    float64           `json:"amount"`
	Status    string            `json:"status"`
	CreatedAt time.Time         `json:"created_at"`
	Items     []OrderItemResponse `json:"items"`
}

// GetOrders 获取订单列表
func (s *CustomerService) GetOrders(customerID int, page, limit int, status string) ([]OrderResponse, int64, error) {
	// 构建查询，过滤掉活动订单
	query := s.db.Model(&models.Order{}).Where("customer_id = ? AND activity_id = 0", customerID)

	// 如果指定了状态，添加状态筛选
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.New("统计订单数量失败")
	}

	// 分页
	offset := (page - 1) * limit
	var orders []models.Order
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Preload("Items").Find(&orders).Error; err != nil {
		return nil, 0, errors.New("获取订单列表失败")
	}

	// 转换为前端需要的格式
	var orderList []OrderResponse
	for _, order := range orders {
		// 构建订单项列表
		var items []OrderItemResponse
		for _, item := range order.Items {
			// 查询商品主图（优先is_main=1，否则按sort排序取第一张）
			var productImage models.ProductImage
			imageURL := ""
			s.db.Where("product_id = ?", item.ProductID).Order("is_main DESC, sort ASC").First(&productImage)
			if productImage.ID > 0 {
				imageURL = productImage.ImageURL
			}

			// 查询SKU编码
			skuCode := ""
			if item.SkuID > 0 {
				var sku models.ProductSku
				if err := s.db.Where("id = ?", item.SkuID).First(&sku).Error; err == nil {
					skuCode = sku.SkuCode
				}
			}

			items = append(items, OrderItemResponse{
				ProductID:     item.ProductID,
				ProductName:   item.ProductName,
				ProductImage:  imageURL,
				SkuCode:       skuCode,
				SkuAttributes: item.SkuAttributes,
				Price:         item.Price,
				Quantity:      item.Quantity,
			})
		}

		orderList = append(orderList, OrderResponse{
			OrderID:   order.ID,
			OrderNo:   order.OrderNo,
			Amount:    order.TotalAmount,
			Status:    order.Status,
			CreatedAt: order.CreatedAt,
			Items:     items,
		})
	}

	return orderList, total, nil
}
