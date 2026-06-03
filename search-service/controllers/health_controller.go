package controllers

import (
	"search-service/services"
	"search-service/utils"

	"github.com/gin-gonic/gin"
)

// ============================================================
// HealthController 健康检查控制器
// 提供服务健康状态检查接口，用于监控和运维
// 返回服务状态、Elasticsearch 连接状态和数据新鲜度
// ============================================================

// HealthController 健康检查控制器结构体
type HealthController struct {
	BaseController
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	// Status 服务整体健康状态：healthy（健康）/ degraded（降级）/ unhealthy（不可用）
	Status string `json:"status"`

	// Elasticsearch ES 连接状态信息
	Elasticsearch ESHealthInfo `json:"elasticsearch"`

	// DataFreshness 数据新鲜度信息
	DataFreshness DataFreshnessInfo `json:"data_freshness"`
}

// ESHealthInfo Elasticsearch 健康状态信息
type ESHealthInfo struct {
	// Connected 是否已连接到 ES 集群
	Connected bool `json:"connected"`

	// ClusterStatus 集群状态：green（健康）/ yellow（警告）/ red（故障）
	ClusterStatus string `json:"cluster_status"`

	// IKPlugin 是否检测到 IK 分词插件
	IKPlugin bool `json:"ik_plugin"`
}

// DataFreshnessInfo 数据新鲜度信息
type DataFreshnessInfo struct {
	// LastSyncTime 上次数据同步时间
	LastSyncTime string `json:"last_sync_time"`

	// SyncInterval 同步间隔（秒）
	SyncInterval int `json:"sync_interval"`
}

// HealthCheck 健康检查接口
// GET /health
// 不需要认证和限流，供监控系统调用
func (hc *HealthController) HealthCheck(ctx *gin.Context) {
	result := HealthCheckResult{
		Status: "healthy",
		Elasticsearch: ESHealthInfo{
			Connected:     false,
			ClusterStatus: "unknown",
			IKPlugin:      false,
		},
		DataFreshness: DataFreshnessInfo{
			LastSyncTime: services.GetLastSyncTime(),
			SyncInterval: 60,
		},
	}

	// 检查 Elasticsearch 健康状态
	esHealth := services.CheckESHealth()
	result.Elasticsearch.Connected = esHealth.Connected
	result.Elasticsearch.ClusterStatus = esHealth.ClusterStatus
	result.Elasticsearch.IKPlugin = esHealth.IKPlugin

	// 根据检查结果确定整体健康状态
	if !esHealth.Connected {
		// ES 无法连接，服务不可用
		result.Status = "unhealthy"
		utils.Warn("健康检查: Elasticsearch 不可用")
	} else if esHealth.ClusterStatus == "red" {
		// ES 集群状态为红色，服务降级
		result.Status = "degraded"
		utils.Warn("健康检查: Elasticsearch 集群状态为 red")
	} else if !esHealth.IKPlugin {
		// IK 插件未安装，搜索质量降级
		result.Status = "degraded"
		utils.Warn("健康检查: IK 分词插件未安装")
	}

	utils.Info("健康检查完成, 状态: %s", result.Status)

	hc.ResponseSuccess(ctx, result)
}
