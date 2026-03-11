package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseController 基础控制器
type BaseController struct{}

// GetUserID 从上下文获取用户ID，如果不存在则返回错误
func (c *BaseController) GetUserID(ctx *gin.Context) (int, bool) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未认证"})
		return 0, false
	}
	return userID.(int), true
}
