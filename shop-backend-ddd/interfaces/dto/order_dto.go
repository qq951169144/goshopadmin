package dto

// CreateOrderRequest 创建订单请求 DTO
type CreateOrderRequest struct {
	AddressID int              `json:"address_id" binding:"required"`
	Items     []OrderItemInput `json:"items" binding:"required"`
	Remark    string           `json:"remark"`
}

// OrderItemInput 订单项输入
type OrderItemInput struct {
	ProductID int `json:"product_id" binding:"required"`
	SkuID     int `json:"sku_id" binding:"required"`
	Quantity  int `json:"quantity" binding:"required,min=1"`
}

// CreateOrderResponse 创建订单响应 DTO
type CreateOrderResponse struct {
	OrderNo     string `json:"order_no"`
	TotalAmount string `json:"total_amount"`
	Status      string `json:"status"`
	PaymentURL  string `json:"payment_url"`
}

// OrderDetailResponse 订单详情响应 DTO
type OrderDetailResponse struct {
	OrderID     string              `json:"order_id"`
	OrderNo     string              `json:"order_no"`
	TotalAmount string              `json:"total_amount"`
	Status      string              `json:"status"`
	CreatedAt   string              `json:"created_at"`
	Items       []OrderItemResponse `json:"items"`
}

// OrderItemResponse 订单项响应
type OrderItemResponse struct {
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	SkuID       int    `json:"sku_id"`
	Price       string `json:"price"`
	Quantity    int    `json:"quantity"`
	TotalAmount string `json:"total_amount"`
}

// APIResponse 统一 API 响应格式
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
