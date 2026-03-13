package controllers

import (
	"goshopadmin/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserController 用户控制器
type UserController struct {
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取用户列表失败"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取用户列表成功",
		"data":    users,
	})
}

// GetUser 获取单个用户信息
func (c *UserController) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	user, err := c.authService.GetUserByID(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取用户信息成功",
		"data":    user,
	})
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
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	user, err := c.authService.CreateUser(req.Username, req.Password, req.RoleID, req.Status)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建用户成功",
		"data":    user,
	})
}

// UpdateUser 更新用户
func (c *UserController) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	var req struct {
		Password string `json:"password"`
		RoleID   int    `json:"role_id"`
		Status   string `json:"status"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	user, err := c.authService.UpdateUser(id, req.Password, req.RoleID, req.Status)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新用户成功",
		"data":    user,
	})
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	if err := c.authService.DeleteUser(id); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除用户成功",
	})
}
