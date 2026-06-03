package controllers

import (
	svcErrors "search-service/errors"
	"search-service/services"
	"search-service/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================
// CustomerController 客户搜索控制器
// 处理 C 端商城客户搜索相关的 HTTP 请求
// 支持关键词搜索（用户名/手机号精确匹配 + 邮箱模糊匹配 + 昵称 IK 分词搜索）、状态筛选
// ============================================================

// CustomerController 客户搜索控制器结构体
type CustomerController struct {
	BaseController
}

// SearchCustomersRequest 客户搜索请求参数
type SearchCustomersRequest struct {
	// Keyword 搜索关键词，匹配用户名/手机号（精确）或邮箱（模糊）或昵称（IK 分词）
	Keyword string `form:"keyword"`

	// Status 客户状态筛选：active 启用 / inactive 禁用
	Status string `form:"status"`

	// Page 页码，从 1 开始
	Page int `form:"page"`

	// PageSize 每页记录数，默认 20，最大 100
	PageSize int `form:"page_size"`
}

// SearchCustomers 客户搜索接口
// GET /api/search/customers
func (cc *CustomerController) SearchCustomers(ctx *gin.Context) {
	var req SearchCustomersRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		cc.ResponseError(ctx, svcErrors.CodeParamError, err)
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
	params := services.CustomerSearchParams{
		Keyword:  req.Keyword,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 调用客户搜索服务
	result, err := services.SearchCustomers(params)
	if err != nil {
		utils.Error("客户搜索失败: %v", err)
		cc.ResponseError(ctx, svcErrors.CodeSearchError, err)
		return
	}

	utils.Info("客户搜索成功, 关键词: %s, 结果数: %d", req.Keyword, result.Total)

	cc.ResponseSuccess(ctx, result)
}
