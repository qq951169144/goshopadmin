package controllers

import (
	"goshopadmin/errors"
	"goshopadmin/services"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserController 用户控制器
type UserController struct {
	BaseController
	authService *services.AuthService
}

// NewUserController 创建用户控制器实例
func NewUserController(db *gorm.DB, jwtSecret string, jwtExpireHour int) *UserController {
	return &UserController{
		authService: services.NewAuthService(db, jwtSecret, jwtExpireHour),
	}
}

// GetUsers 获取用户列表
func (c *UserController) GetUsers(ctx *gin.Context) {
	// 这里可以添加分页和筛选逻辑
	users, err := c.authService.GetUsers()
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, users)
}

// GetUser 获取单个用户信息
func (c *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	user, err := c.authService.GetUserByID(id)
	if err != nil {
		c.ResponseError(ctx, errors.CodeUserNotFound, err)
		return
	}

	c.ResponseSuccess(ctx, user)
}

// CreateUser 创建用户
func (c *UserController) CreateUser(ctx *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		RoleID   int    `json:"role_id" binding:"required"`
		Status   string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	user, err := c.authService.CreateUser(req.Username, req.Password, req.RoleID, req.Status)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDuplicate, err)
		return
	}

	c.ResponseSuccess(ctx, user)
}

// UpdateUser 更新用户
func (c *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	var req struct {
		Password string `json:"password"`
		RoleID   int    `json:"role_id"`
		Status   string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	user, err := c.authService.UpdateUser(id, req.Password, req.RoleID, req.Status)
	if err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, user)
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.ResponseError(ctx, errors.CodeParamInvalid, err)
		return
	}

	if err := c.authService.DeleteUser(id); err != nil {
		c.ResponseError(ctx, errors.CodeDBError, err)
		return
	}

	c.ResponseSuccess(ctx, nil)
}
