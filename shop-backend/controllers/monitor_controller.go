package controllers

import (
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
)

// MonitorController 监控控制器
type MonitorController struct {
	BaseController
	monitor *utils.Monitor
}

// NewMonitorController 创建监控控制器
func NewMonitorController(monitor *utils.Monitor) *MonitorController {
	return &MonitorController{monitor: monitor}
}

// GetCurrentStats 获取最新运行时统计
func (c *MonitorController) GetCurrentStats(ctx *gin.Context) {
	stats := c.monitor.GetCurrentStats()
	c.ResponseSuccess(ctx, stats)
}

// GetHistoryStats 获取历史统计列表
func (c *MonitorController) GetHistoryStats(ctx *gin.Context) {
	history := c.monitor.GetHistoryStats()
	c.ResponseSuccess(ctx, history)
}
