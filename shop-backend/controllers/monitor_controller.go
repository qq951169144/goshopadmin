package controllers

import (
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
)

type MonitorController struct {
	monitor *utils.Monitor
}

func NewMonitorController(monitor *utils.Monitor) *MonitorController {
	return &MonitorController{monitor: monitor}
}

// GetCurrentStats 获取最新运行时统计
func (c *MonitorController) GetCurrentStats(ctx *gin.Context) {
	stats := c.monitor.GetCurrentStats()
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
}

// GetHistoryStats 获取历史统计列表
func (c *MonitorController) GetHistoryStats(ctx *gin.Context) {
	history := c.monitor.GetHistoryStats()
	ctx.JSON(200, gin.H{
		"code":    0,
		"message": "success",
		"data":    history,
	})
}
