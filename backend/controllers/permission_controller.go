package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionController 权限控制器
type PermissionController struct {
	BaseController
	authService *services.AuthService
}

// NewPermissionController 创建权限控制器实例
func NewPermissionController(db *gorm.DB, jwtSecret string, jwtExpireHour int) *PermissionController {
	return &PermissionController{
		authService: services.NewAuthService(db, jwtSecret, jwtExpireHour),
	}
}

// GetPermissions 获取权限列表
func (c *PermissionController) GetPermissions(ctx *gin.Context) {
	permissions, err := c.authService.GetPermissions()
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, permissions)
}

// GetPermission 获取单个权限信息
func (c *PermissionController) GetPermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	permission, err := c.authService.GetPermissionByID(id)
	if err != nil {
		c.ResponseError(ctx, errors.CodeNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, permission)
}

// CreatePermission 创建权限
func (c *PermissionController) CreatePermission(ctx *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	permission, err := c.authService.CreatePermission(req.Name, req.Code, req.Description, req.Status)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDuplicate, err)
		return
	}

	c.ResponseSuccess(ctx, permission)
}

// UpdatePermission 更新权限
func (c *PermissionController) UpdatePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Code        string `json:"code"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	permission, err := c.authService.UpdatePermission(id, req.Name, req.Code, req.Description, req.Status)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, permission)
}

// DeletePermission 删除权限
func (c *PermissionController) DeletePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	if err := c.authService.DeletePermission(id); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}
