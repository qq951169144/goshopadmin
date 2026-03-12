package controllers

import (
	"net/http"
	"strconv"
	"time"

	"shop-backend/config"
	"shop-backend/models"

	"github.com/gin-gonic/gin"
)

// 用户结构
type User struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// 更新个人信息请求结构
type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// 获取个人信息
func GetProfile(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 从数据库获取用户信息
	var user models.User
	result := config.DB.First(&user, userID)
	if result.RowsAffected == 0 {
		ResponseError(c, http.StatusNotFound, "User not found")
		return
	}

	// 转换为前端需要的格式
	ResponseSuccess(c, User{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

// 更新个人信息
func UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 从数据库获取用户信息
	var user models.User
	result := config.DB.First(&user, userID)
	if result.RowsAffected == 0 {
		ResponseError(c, http.StatusNotFound, "User not found")
		return
	}

	// 更新用户信息
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := config.DB.Save(&user).Error; err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	ResponseSuccess(c, gin.H{
		"message":  "Profile updated",
		"username": user.Username,
		"email":    user.Email,
	})
}

// 获取订单列表
func GetOrders(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseError(c, http.StatusUnauthorized, "User not logged in")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// 构建查询
	query := config.DB.Model(&models.Order{}).Where("customer_id = ?", userID)

	// 计算总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to count orders")
		return
	}

	// 分页
	offset := (page - 1) * limit
	var orders []models.Order
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error; err != nil {
		ResponseError(c, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	// 转换为前端需要的格式
	var orderList []gin.H
	for _, order := range orders {
		orderList = append(orderList, gin.H{
			"order_id":   order.OrderNo,
			"amount":     order.TotalAmount,
			"status":     order.Status,
			"created_at": order.CreatedAt,
		})
	}

	ResponseSuccess(c, gin.H{
		"orders": orderList,
		"total":  total,
	})
}
