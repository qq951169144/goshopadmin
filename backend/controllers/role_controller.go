package controllers

import (
	"goshopadmin/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RoleController 角色控制器
type RoleController struct {
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取角色列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取角色列表成功",
		"data":    roles,
	})
}

// GetRole 获取单个角色信息
func (c *RoleController) GetRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	role, err := c.authService.GetRoleByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "角色不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取角色信息成功",
		"data":    role,
	})
}

// CreateRole 创建角色
func (c *RoleController) CreateRole(ctx *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	role, err := c.authService.CreateRole(req.Name, req.Description, req.Status)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建角色成功",
		"data":    role,
	})
}

// UpdateRole 更新角色
func (c *RoleController) UpdateRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	role, err := c.authService.UpdateRole(id, req.Name, req.Description, req.Status)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新角色成功",
		"data":    role,
	})
}

// DeleteRole 删除角色
func (c *RoleController) DeleteRole(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := c.authService.DeleteRole(id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除角色成功",
	})
}

// AssignPermissions 为角色分配权限
func (c *RoleController) AssignPermissions(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req struct {
		PermissionIDs []int `json:"permission_ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := c.authService.AssignPermissions(id, req.PermissionIDs); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "分配权限成功",
	})
}
