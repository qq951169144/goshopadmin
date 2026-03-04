package controllers

import (
	"github.com/gin-gonic/gin"
)

// CommonController 通用控制器
type CommonController struct{}

// NewCommonController 创建通用控制器实例
func NewCommonController() *CommonController {
	return &CommonController{}
}

// HealthCheck 健康检查
func (c *CommonController) HealthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status":  "ok",
		"message": "Goshopadmin backend service is running",
	})
}

// HelloWorld 首页HelloWorld
func (c *CommonController) HelloWorld(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"message": "Hello World!",
		"status":  "success",
		"service": "Goshopadmin Backend",
	})
}
