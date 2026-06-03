package controllers

import (
	"errors"

	svcErrors "search-service/errors"
	"search-service/services"
	"search-service/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================
// SuggestController 搜索建议控制器
// 提供搜索自动补全/建议功能
// 根据用户输入的前缀，返回匹配的建议词列表
// ============================================================

// SuggestController 搜索建议控制器结构体
type SuggestController struct {
	BaseController
}

// SuggestRequest 搜索建议请求参数
type SuggestRequest struct {
	// Prefix 用户输入的搜索前缀，至少需要 1 个字符
	Prefix string `form:"prefix" binding:"required"`

	// Type 建议类型，决定搜索哪个索引和字段
	// 可选值：product（商品名称）、order（订单号）、user（用户名）、customer（客户昵称）
	Type string `form:"type" binding:"required"`
}

// SuggestResponse 搜索建议响应
type SuggestResponse struct {
	// Suggestions 建议词列表，已去重
	Suggestions []string `json:"suggestions"`
}

// Suggest 搜索建议接口
// GET /api/search/suggest
// 根据前缀和类型返回匹配的建议词
func (sc *SuggestController) Suggest(ctx *gin.Context) {
	var req SuggestRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		sc.ResponseError(ctx, svcErrors.CodeParamError, err)
		return
	}

	// 验证建议类型是否合法
	validTypes := map[string]bool{
		"product":  true,
		"order":    true,
		"user":     true,
		"customer": true,
	}
	if !validTypes[req.Type] {
		sc.ResponseError(ctx, svcErrors.CodeParamInvalid, errors.New("type 参数必须是 product, order, user, customer 之一"))
		return
	}

	// 构建搜索建议参数
	params := services.SuggestParams{
		Prefix: req.Prefix,
		Type:   req.Type,
	}

	// 调用搜索建议服务
	suggestions, err := services.Suggest(params)
	if err != nil {
		utils.Error("搜索建议失败: %v", err)
		sc.ResponseError(ctx, svcErrors.CodeSearchError, err)
		return
	}

	utils.Info("搜索建议成功, 前缀: %s, 类型: %s, 建议数: %d", req.Prefix, req.Type, len(suggestions))

	sc.ResponseSuccess(ctx, SuggestResponse{
		Suggestions: suggestions,
	})
}
