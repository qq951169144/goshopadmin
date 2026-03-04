package controllers

import (
	"goshopadmin/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PermissionController 权限控制器
type PermissionController struct {
	authService *services.AuthService
}

// NewPermissionController 创建权限控制器实例
func NewPermissionController(db *gorm.DB) *PermissionController {
	return &PermissionController{
		authService: services.NewAuthService(db),
	}
}

// GetPermissions 获取权限列表
func (c *PermissionController) GetPermissions(ctx *gin.Context) {
	permissions, err := c.authService.GetPermissions()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取权限列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取权限列表成功",
		"data":    permissions,
	})
}

// GetPermission 获取单个权限信息
func (c *PermissionController) GetPermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	permission, err := c.authService.GetPermissionByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取权限信息失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取权限信息成功",
		"data":    permission,
	})
}

// CreatePermissionRequest 创建权限请求结构
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
}

// CreatePermission 创建权限
func (c *PermissionController) CreatePermission(ctx *gin.Context) {
	var req CreatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	permission, err := c.authService.CreatePermission(req.Name, req.Code, req.Description)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建权限失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建权限成功",
		"data":    permission,
	})
}

// UpdatePermissionRequest 更新权限请求结构
type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

// UpdatePermission 更新权限
func (c *PermissionController) UpdatePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req UpdatePermissionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	permission, err := c.authService.UpdatePermission(id, req.Name, req.Code, req.Description)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新权限失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新权限成功",
		"data":    permission,
	})
}

// DeletePermission 删除权限
func (c *PermissionController) DeletePermission(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	err = c.authService.DeletePermission(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除权限失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除权限成功",
	})
}
