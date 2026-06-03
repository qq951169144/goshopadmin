package controllers

import (
	svcErrors "search-service/errors"
	"search-service/services"
	"search-service/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================
// UserController 用户搜索控制器
// 处理后台管理系统用户搜索相关的 HTTP 请求
// 支持关键词搜索（用户名精确匹配 + 邮箱模糊匹配）、角色和状态筛选
// ============================================================

// UserController 用户搜索控制器结构体
type UserController struct {
	BaseController
}

// SearchUsersRequest 用户搜索请求参数
type SearchUsersRequest struct {
	// Keyword 搜索关键词，匹配用户名（精确）或邮箱（模糊）
	Keyword string `form:"keyword"`

	// RoleID 角色ID，筛选指定角色的用户
	RoleID int `form:"role_id"`

	// Status 用户状态筛选：active 启用 / inactive 禁用
	Status string `form:"status"`

	// Page 页码，从 1 开始
	Page int `form:"page"`

	// PageSize 每页记录数，默认 20，最大 100
	PageSize int `form:"page_size"`
}

// SearchUsers 用户搜索接口
// GET /api/search/users
func (uc *UserController) SearchUsers(ctx *gin.Context) {
	var req SearchUsersRequest

	if err := ctx.ShouldBindQuery(&req); err != nil {
		uc.ResponseError(ctx, svcErrors.CodeParamError, err)
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
	params := services.UserSearchParams{
		Keyword:  req.Keyword,
		RoleID:   req.RoleID,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// 调用用户搜索服务
	result, err := services.SearchUsers(params)
	if err != nil {
		utils.Error("用户搜索失败: %v", err)
		uc.ResponseError(ctx, svcErrors.CodeSearchError, err)
		return
	}

	utils.Info("用户搜索成功, 关键词: %s, 结果数: %d", req.Keyword, result.Total)

	uc.ResponseSuccess(ctx, result)
}
