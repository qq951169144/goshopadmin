package controllers

import (
	"strconv"

	svcErrors "search-service/errors"
	"search-service/services"
	"search-service/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================
// ProductController 商品搜索控制器
// 处理商品搜索相关的 HTTP 请求
// 支持关键词搜索、分类筛选、商户筛选、价格区间筛选、排序和分页
// ============================================================

// ProductController 商品搜索控制器结构体
// 嵌入 BaseController 获得统一的响应处理能力
type ProductController struct {
	BaseController
}

// SearchProductsRequest 商品搜索请求参数
// 所有参数通过 Query String 传递
type SearchProductsRequest struct {
	// Keyword 搜索关键词，使用 IK 分词器匹配商品名称和描述
	Keyword string `form:"keyword"`

	// CategoryID 分类ID，精确筛选指定分类下的商品
	CategoryID int `form:"category_id"`

	// MerchantID 商户ID，精确筛选指定商户的商品
	MerchantID int `form:"merchant_id"`

	// Status 商品状态筛选：active 上架 / inactive 下架
	Status string `form:"status"`

	// MinPrice 最低价格筛选，与 MaxPrice 配合实现价格区间
	MinPrice float64 `form:"min_price"`

	// MaxPrice 最高价格筛选，与 MinPrice 配合实现价格区间
	MaxPrice float64 `form:"max_price"`

	// Sort 排序字段：默认 relevance（相关度），可选 price（价格）、sales（销量）、created_at（创建时间）
	Sort string `form:"sort"`

	// Order 排序方向：desc 降序（默认）、asc 升序
	Order string `form:"order"`

	// Page 页码，从 1 开始，默认 1
	Page int `form:"page"`

	// PageSize 每页记录数，默认 20，最大 100
	PageSize int `form:"page_size"`
}

// SearchProducts 商品搜索接口
// GET /api/search/products
// 接收搜索参数，调用商品搜索服务，返回搜索结果
func (pc *ProductController) SearchProducts(ctx *gin.Context) {
	var req SearchProductsRequest

	// 绑定 Query String 参数到请求结构体
	if err := ctx.ShouldBindQuery(&req); err != nil {
		pc.ResponseError(ctx, svcErrors.CodeParamError, err)
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
	params := services.ProductSearchParams{
		Keyword:    req.Keyword,
		CategoryID: req.CategoryID,
		MerchantID: req.MerchantID,
		Status:     req.Status,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		Sort:       req.Sort,
		Order:      req.Order,
		Page:       req.Page,
		PageSize:   req.PageSize,
	}

	// 调用商品搜索服务
	result, err := services.SearchProducts(params)
	if err != nil {
		utils.Error("商品搜索失败: %v", err)
		pc.ResponseError(ctx, svcErrors.CodeSearchError, err)
		return
	}

	utils.Info("商品搜索成功, 关键词: %s, 页码: %s, 结果数: %d",
		req.Keyword, strconv.Itoa(req.Page), result.Total)

	pc.ResponseSuccess(ctx, result)
}
