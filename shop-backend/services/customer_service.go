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

// GetOrders 获取订单列表
func (s *CustomerService) GetOrders(customerID int, page, limit int, status string) ([]map[string]interface{}, int64, error) {
	// 构建查询
	query := s.db.Model(&models.Order{}).Where("customer_id = ?", customerID)

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
	var orderList []map[string]interface{}
	for _, order := range orders {
		// 构建订单项列表
		var items []map[string]interface{}
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
				var sku models.ProductSKU
				if err := s.db.Where("id = ?", item.SkuID).First(&sku).Error; err == nil {
					skuCode = sku.SKUCode
				}
			}

			items = append(items, map[string]interface{}{
				"product_id":     item.ProductID,
				"product_name":   item.ProductName,
				"product_image":  imageURL,
				"sku_code":       skuCode,
				"sku_attributes": item.SkuAttributes,
				"price":          item.Price,
				"quantity":       item.Quantity,
			})
		}

		orderList = append(orderList, map[string]interface{}{
			"order_id":   order.ID,
			"order_no":   order.OrderNo,
			"amount":     order.TotalAmount,
			"status":     order.Status,
			"created_at": order.CreatedAt,
			"items":      items,
		})
	}

	return orderList, total, nil
}
