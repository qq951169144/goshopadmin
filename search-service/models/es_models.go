package models

import "time"

// ============================================================
// Elasticsearch 文档模型
// 定义了搜索服务中各个索引的文档结构
// 每个结构体对应 Elasticsearch 中的一个索引类型
// JSON tag 使用 snake_case 命名，与 ES 索引字段名保持一致
// ============================================================

// ProductDoc 商品文档，对应 Elasticsearch 中的 products 索引
// 包含商品基本信息及其关联的 SKU 和规格信息
type ProductDoc struct {
	// ID 商品主键ID
	ID int `json:"id"`

	// Name 商品名称，使用 IK 分词器进行中文分词
	Name string `json:"name"`

	// Description 商品描述，使用 IK 分词器进行中文分词
	Description string `json:"description"`

	// CategoryID 分类ID，用于按分类筛选
	CategoryID int `json:"category_id"`

	// CategoryName 分类名称，用于搜索结果展示
	CategoryName string `json:"category_name"`

	// MerchantID 商户ID，用于按商户筛选
	MerchantID int `json:"merchant_id"`

	// MerchantName 商户名称，用于搜索结果展示
	MerchantName string `json:"merchant_name"`

	// Status 商品状态：active 上架 / inactive 下架
	Status string `json:"status"`

	// MinPrice 最低价格（取自所有 SKU 中的最低价），用于价格区间筛选
	MinPrice float64 `json:"min_price"`

	// MaxPrice 最高价格（取自所有 SKU 中的最高价），用于价格区间筛选
	MaxPrice float64 `json:"max_price"`

	// MainImage 主图URL，用于搜索结果展示
	MainImage string `json:"main_image"`

	// Skus SKU 列表，嵌套文档，包含商品的规格和价格信息
	Skus []SkuDoc `json:"skus"`

	// Specs 规格列表，嵌套文档，包含商品的规格名和规格值
	Specs []SpecDoc `json:"specs"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间，用于数据同步判断
	UpdatedAt time.Time `json:"updated_at"`
}

// SkuDoc SKU 文档，嵌套在 ProductDoc 中
// 对应商品的某个规格组合的具体信息
type SkuDoc struct {
	// ID SKU 主键ID
	ID int `json:"id"`

	// ProductID 所属商品ID
	ProductID int `json:"product_id"`

	// SkuName SKU 名称，如 "红色-XL"
	SkuName string `json:"sku_name"`

	// Price SKU 价格
	Price float64 `json:"price"`

	// Stock 库存数量
	Stock int `json:"stock"`

	// Image SKU 图片URL
	Image string `json:"image"`

	// SpecValues 规格值组合，如 {"颜色":"红色","尺码":"XL"}
	SpecValues map[string]string `json:"spec_values"`
}

// SpecDoc 规格文档，嵌套在 ProductDoc 中
// 描述商品的某个规格维度及其可选值
type SpecDoc struct {
	// SpecName 规格名称，如 "颜色"、"尺码"
	SpecName string `json:"spec_name"`

	// SpecValues 规格可选值列表，如 ["红色", "蓝色", "黑色"]
	SpecValues []string `json:"spec_values"`
}

// OrderDoc 订单文档，对应 Elasticsearch 中的 orders 索引
// 包含订单基本信息及其关联的订单明细
type OrderDoc struct {
	// ID 订单主键ID
	ID int `json:"id"`

	// OrderNo 订单编号，唯一标识一个订单
	OrderNo string `json:"order_no"`

	// CustomerID 客户ID
	CustomerID int `json:"customer_id"`

	// CustomerName 客户名称，用于搜索结果展示
	CustomerName string `json:"customer_name"`

	// MerchantID 商户ID
	MerchantID int `json:"merchant_id"`

	// Status 订单状态：pending 待支付 / paid 已支付 / shipped 已发货 / completed 已完成 / cancelled 已取消
	Status string `json:"status"`

	// PaymentStatus 支付状态：pending 待支付 / success 支付成功 / failed 支付失败
	PaymentStatus string `json:"payment_status"`

	// ShippingStatus 物流状态：pending 待发货 / shipped 已发货 / delivered 已签收 / returned 已退回
	ShippingStatus string `json:"shipping_status"`

	// TotalAmount 订单总金额
	TotalAmount float64 `json:"total_amount"`

	// Items 订单明细列表，嵌套文档，用于按商品名称搜索订单
	Items []OrderItem `json:"items"`

	// CreatedAt 下单时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// OrderItem 订单明细，嵌套在 OrderDoc 中
// 表示订单中的某个商品购买记录
type OrderItem struct {
	// ID 订单明细ID
	ID int `json:"id"`

	// ProductID 商品ID
	ProductID int `json:"product_id"`

	// ProductName 商品名称，用于订单搜索
	ProductName string `json:"product_name"`

	// SkuID SKU ID
	SkuID int `json:"sku_id"`

	// SkuName SKU 名称
	SkuName string `json:"sku_name"`

	// Price 购买单价
	Price float64 `json:"price"`

	// Quantity 购买数量
	Quantity int `json:"quantity"`

	// Subtotal 小计金额
	Subtotal float64 `json:"subtotal"`
}

// UserDoc 用户文档，对应 Elasticsearch 中的 users 索引
// 用于后台管理系统的用户搜索
type UserDoc struct {
	// ID 用户主键ID
	ID int `json:"id"`

	// Username 用户名，用于精确匹配搜索
	Username string `json:"username"`

	// Email 邮箱地址，用于模糊匹配搜索
	Email string `json:"email"`

	// Phone 手机号码
	Phone string `json:"phone"`

	// RoleID 角色ID，用于按角色筛选
	RoleID int `json:"role_id"`

	// RoleName 角色名称，用于搜索结果展示
	RoleName string `json:"role_name"`

	// Status 用户状态：active 启用 / inactive 禁用
	Status string `json:"status"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// CustomerDoc 客户文档，对应 Elasticsearch 中的 customers 索引
// 用于 C 端商城的客户搜索
type CustomerDoc struct {
	// ID 客户主键ID
	ID int `json:"id"`

	// Username 客户用户名，用于精确匹配搜索
	Username string `json:"username"`

	// Email 邮箱地址，用于模糊匹配搜索
	Email string `json:"email"`

	// Phone 手机号码，用于精确匹配搜索
	Phone string `json:"phone"`

	// Nickname 昵称，使用 IK 分词器进行中文分词搜索
	Nickname string `json:"nickname"`

	// Avatar 头像URL
	Avatar string `json:"avatar"`

	// Status 客户状态：active 启用 / inactive 禁用
	Status string `json:"status"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchResult 通用搜索结果结构体
// 所有搜索接口统一返回此结构，包含分页信息和结果列表
type SearchResult struct {
	// Total 符合条件的总记录数
	Total int64 `json:"total"`

	// Page 当前页码，从 1 开始
	Page int `json:"page"`

	// PageSize 每页记录数
	PageSize int `json:"page_size"`

	// Items 当前页的搜索结果列表
	Items interface{} `json:"items"`
}
