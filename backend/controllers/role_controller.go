package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RoleController 角色控制器
type RoleController struct {
	BaseController
	authService *services.AuthService
}

// NewRoleController 创建角色控制器实例
func NewRoleController(db *gorm.DB, jwtSecret string, jwtExpireHour int) *RoleController {
	return &RoleController{
		authService: services.NewAuthService(db, jwtSecret, jwtExpireHour),
	}
}

// GetRoles 获取角色列表
func (c *RoleController) GetRoles(ctx *gin.Context) {
	roles, err := c.authService.GetRoles()
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, roles)
}

// GetRole 获取单个角色信息
func (c *RoleController) GetRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	role, err := c.authService.GetRoleByID(id)
	if err != nil {
		c.ResponseError(ctx, errors.CodeNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, role)
}

// CreateRole 创建角色
func (c *RoleController) CreateRole(ctx *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	role, err := c.authService.CreateRole(req.Name, req.Description, req.Status)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDuplicate, err)
		return
	}

	c.ResponseSuccess(ctx, role)
}

// UpdateRole 更新角色
func (c *RoleController) UpdateRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	role, err := c.authService.UpdateRole(id, req.Name, req.Description, req.Status)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, role)
}

// DeleteRole 删除角色
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	if err := c.authService.DeleteRole(id); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}

// AssignPermissions 为角色分配权限
func (c *RoleController) AssignPermissions(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req struct {
		PermissionIDs []int `json:"permission_ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	if err := c.authService.AssignPermissions(id, req.PermissionIDs); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}
