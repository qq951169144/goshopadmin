package controllers

import (
	"net/http"

	"shop-backend/pkg/mq"
	"github.com/gin-gonic/gin"
)

// HealthController 健康检查控制器
type HealthController struct {
	BaseController
}

// NewHealthController 创建健康检查控制器
func NewHealthController() *HealthController {
	return &HealthController{}
}

// CheckMQ 检查MQ连接
func (c *HealthController) CheckMQ(ctx *gin.Context) {
	conn, err := mq.NewConnection()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    5000,
			"message": "MQ连接失败",
			"data":    nil,
		})
		return
	}
	defer conn.Close()

	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "MQ连接正常",
		"data":    nil,
	})
}