package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseController 基础控制器
type BaseController struct{}

// ResponseSuccess 返回成功响应
func (c *BaseController) ResponseSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, data)
}

// ResponseError 返回错误响应
func (c *BaseController) ResponseError(ctx *gin.Context, code int, message string) {
	ctx.JSON(code, gin.H{"error": message})
}

// GetUserID 从上下文获取用户ID
func (c *BaseController) GetUserID(ctx *gin.Context) (uint, bool) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}
