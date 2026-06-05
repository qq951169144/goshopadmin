package controllers

import (
	"errors"

	svcErrors "search-service/errors"
	"search-service/services"
	"search-service/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================
// OrderController 订单搜索控制器
// 处理订单搜索相关的 HTTP 请求
// 支持关键词搜索（订单号精确匹配 + 商品名称模糊匹配）、多状态筛选、时间范围和金额范围筛选
// ============================================================

// OrderController 订单搜索控制器结构体
type OrderController struct {
	BaseController
}

// SearchOrdersRequest 订单搜索请求参数
type SearchOrdersRequest struct {
	// Keyword 搜索关键词，匹配订单号（精确）或订单明细中的商品名称（模糊）
	Keyword string `form:"keyword"`

	// CustomerID 客户ID，筛选指定客户的订单
	CustomerID int `form:"customer_id"`

	// MerchantID 商户ID，筛选指定商户的订单
	MerchantID int `form:"merchant_id"`

	// Status 订单状态筛选：pending / paid / shipped / completed / cancelled
	Status string `form:"status"`

	// PaymentStatus 支付状态筛选：pending / success / failed
	PaymentStatus string `form:"payment_status"`

	// ShippingStatus 物流状态筛选：pending / shipped / delivered / returned
	ShippingStatus string `form:"shipping_status"`

	// StartDate 开始日期，筛选下单时间 >= 此日期的订单，格式：2006-01-02
	StartDate string `form:"start_date"`

	// EndDate 结束日期，筛选下单时间 <= 此日期的订单，格式：2006-01-02
	EndDate string `form:"end_date"`

	// MinAmount 最小金额筛选
	MinAmount float64 `form:"min_amount"`

	// MaxAmount 最大金额筛选
	MaxAmount float64 `form:"max_amount"`

	// Sort 排序字段：默认 relevance，可选 total_amount、created_at
	Sort string `form:"sort"`

	// Order 排序方向：desc 降序（默认）、asc 升序
	Order string `form:"order"`

	// Page 页码，从 1 开始
	Page int `form:"page"`

	// PageSize 每页记录数，默认 20，最大 100
	PageSize int `form:"page_size"`
}

// SearchAdminOrders 管理端订单搜索接口
// GET /api/search/admin/orders
// 管理端可查看该商户的订单，支持所有筛选参数
func (oc *OrderController) SearchAdminOrders(ctx *gin.Context) {
	var req SearchOrdersRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		oc.ResponseError(ctx, svcErrors.CodeParamError, err)
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建搜索参数
	params := services.OrderSearchParams{
		Keyword:        req.Keyword,
		CustomerID:     req.CustomerID,
		MerchantID:     req.MerchantID,
		Status:         req.Status,
		PaymentStatus:  req.PaymentStatus,
		ShippingStatus: req.ShippingStatus,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		MinAmount:      req.MinAmount,
		MaxAmount:      req.MaxAmount,
		Sort:           req.Sort,
		Order:          req.Order,
		Page:           req.Page,
		PageSize:       req.PageSize,
	}

	// 调用订单搜索服务
	result, err := services.SearchOrders(params)
	if err != nil {
		utils.Error("订单搜索失败: %v", err)
		oc.ResponseError(ctx, svcErrors.CodeSearchError, err)
		return
	}

	utils.Info("订单搜索成功, 关键词: %s, 结果数: %d", req.Keyword, result.Total)

	oc.ResponseSuccess(ctx, result)
}

// SearchCustomerOrdersRequest C端订单搜索请求参数
// 与管理端不同，C端不允许指定 customer_id 和 merchant_id，由 JWT token 自动注入
type SearchCustomerOrdersRequest struct {
	// Keyword 搜索关键词，匹配订单号（精确）或订单明细中的商品名称（模糊）
	Keyword string `form:"keyword"`

	// Status 订单状态筛选：pending / paid / shipped / completed / cancelled
	Status string `form:"status"`

	// PaymentStatus 支付状态筛选：pending / success / failed
	PaymentStatus string `form:"payment_status"`

	// ShippingStatus 物流状态筛选：pending / shipped / delivered / returned
	ShippingStatus string `form:"shipping_status"`

	// StartDate 开始日期，筛选下单时间 >= 此日期的订单，格式：2006-01-02
	StartDate string `form:"start_date"`

	// EndDate 结束日期，筛选下单时间 <= 此日期的订单，格式：2006-01-02
	EndDate string `form:"end_date"`

	// MinAmount 最小金额筛选
	MinAmount float64 `form:"min_amount"`

	// MaxAmount 最大金额筛选
	MaxAmount float64 `form:"max_amount"`

	// Sort 排序字段：默认 relevance，可选 total_amount、created_at
	Sort string `form:"sort"`

	// Order 排序方向：desc 降序（默认）、asc 升序
	Order string `form:"order"`

	// Page 页码，从 1 开始
	Page int `form:"page"`

	// PageSize 每页记录数，默认 20，最大 100
	PageSize int `form:"page_size"`
}

// SearchCustomerOrders C端订单搜索接口
// GET /api/search/customer/orders
// C端只能查看自己的订单，customer_id 从 JWT token 中获取
func (oc *OrderController) SearchCustomerOrders(ctx *gin.Context) {
	var req SearchCustomerOrdersRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		oc.ResponseError(ctx, svcErrors.CodeParamError, err)
		return
	}

	// 从 JWT token 中获取 customer_id，强制过滤
	customerID, exists := ctx.Get("customer_id")
	if !exists {
		oc.ResponseError(ctx, svcErrors.CodeUnauthorized, errors.New("未获取到客户ID"))
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// 构建搜索参数，强制注入 customer_id
	params := services.OrderSearchParams{
		Keyword:        req.Keyword,
		CustomerID:     customerID.(int), // 强制过滤：只查自己的订单
		Status:         req.Status,
		PaymentStatus:  req.PaymentStatus,
		ShippingStatus: req.ShippingStatus,
		StartDate:      req.StartDate,
		EndDate:        req.EndDate,
		MinAmount:      req.MinAmount,
		MaxAmount:      req.MaxAmount,
		Sort:           req.Sort,
		Order:          req.Order,
		Page:           req.Page,
		PageSize:       req.PageSize,
	}

	// 调用订单搜索服务
	result, err := services.SearchOrders(params)
	if err != nil {
		utils.Error("C端订单搜索失败: %v", err)
		oc.ResponseError(ctx, svcErrors.CodeSearchError, err)
		return
	}

	utils.Info("C端订单搜索成功, customer_id: %d, 关键词: %s, 结果数: %d", customerID.(int), req.Keyword, result.Total)

	oc.ResponseSuccess(ctx, result)
}
