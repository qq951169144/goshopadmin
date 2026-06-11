package controllers

import (
	svcErrors "search-service/errors"
	"search-service/services"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 同步管理控制器
// 提供全量同步手动触发和同步状态查询接口
// 仅限管理端使用，需要 AdminAuth 中间件保护
// ============================================================

// SyncController 同步管理控制器结构体
type SyncController struct {
	BaseController
}

// TriggerFullSyncRequest 手动触发全量同步请求参数
type TriggerFullSyncRequest struct {
	// Confirm 确认标识，必须为 "yes" 才会执行
	Confirm string `json:"confirm" binding:"required"`
}

// TriggerFullSync 手动触发全量同步
// POST /api/search/admin/sync/full
// 需要管理端认证
func (sc *SyncController) TriggerFullSync(ctx *gin.Context) {
	var req TriggerFullSyncRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		sc.ResponseError(ctx, svcErrors.CodeParamError, err)
		return
	}

	if req.Confirm != "yes" {
		sc.ResponseError(ctx, svcErrors.CodeParamError, nil)
		return
	}

	if err := services.TriggerFullSync(); err != nil {
		sc.ResponseError(ctx, svcErrors.CodeSyncInProgress, err)
		return
	}

	sc.ResponseSuccess(ctx, gin.H{
		"message": "全量同步已触发",
	})
}

// GetSyncStatus 获取同步状态
// GET /api/search/admin/sync/status
// 需要管理端认证
func (sc *SyncController) GetSyncStatus(ctx *gin.Context) {
	status := services.GetSyncStatus()
	sc.ResponseSuccess(ctx, status)
}
