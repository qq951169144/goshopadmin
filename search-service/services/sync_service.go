package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"search-service/utils"

	"github.com/olivere/elastic/v7"
)

// ============================================================
// 数据同步服务
// 定时从 MySQL 读取最近更新的数据，同步到 Elasticsearch
// 确保搜索数据与业务数据库保持一致
// 同步策略：每 60 秒查询一次最近 5 分钟更新的数据，批量写入 ES
// ============================================================

// 全局变量
var (
	// lastSyncTime 上次同步时间，用于健康检查报告数据新鲜度
	lastSyncTime string

	// syncTicker 定时器，控制同步间隔
	syncTicker *time.Ticker

	// syncDone 通道，用于通知同步协程退出
	syncDone chan struct{}
)

// SKU同步用的 MySQL 查询结果结构
type skuRow struct {
	ID          int             `gorm:"column:id"`
	ProductID   int             `gorm:"column:product_id"`
	SkuName     string          `gorm:"column:sku_name"`
	Price       float64         `gorm:"column:price"`
	Stock       int             `gorm:"column:stock"`
	Image       string          `gorm:"column:image"`
	SpecValues  json.RawMessage `gorm:"column:spec_values"`
	UpdatedAt   time.Time       `gorm:"column:updated_at"`
	ProductName string          `gorm:"column:product_name"`
}

// 订单明细同步用的 MySQL 查询结果结构
type orderItemRow struct {
	ID          int       `gorm:"column:id"`
	OrderID     int       `gorm:"column:order_id"`
	ProductID   int       `gorm:"column:product_id"`
	ProductName string    `gorm:"column:product_name"`
	SkuID       int       `gorm:"column:sku_id"`
	SkuName     string    `gorm:"column:sku_name"`
	Price       float64   `gorm:"column:price"`
	Quantity    int       `gorm:"column:quantity"`
	Subtotal    float64   `gorm:"column:subtotal"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

// StartSyncService 启动数据同步定时任务
// 每 60 秒执行一次同步，查询最近 5 分钟更新的数据
// 使用 5 分钟窗口而非 60 秒，是为了避免边界情况导致数据遗漏
func StartSyncService() {
	syncTicker = time.NewTicker(60 * time.Second)
	syncDone = make(chan struct{})

	// 记录初始同步时间
	lastSyncTime = time.Now().Format(time.RFC3339)

	// 启动同步协程
	go func() {
		// 首次启动时立即执行一次同步
		utils.Info("数据同步服务启动，开始首次同步")
		runSync()

		for {
			select {
			case <-syncTicker.C:
				runSync()
			case <-syncDone:
				utils.Info("数据同步服务停止")
				return
			}
		}
	}()

	utils.Info("数据同步服务已启动，同步间隔: 60 秒")
}

// StopSyncService 停止数据同步定时任务
func StopSyncService() {
	if syncTicker != nil {
		syncTicker.Stop()
	}
	if syncDone != nil {
		close(syncDone)
	}
}

// runSync 执行一次数据同步
// 依次同步 SKU 数据和订单明细数据
func runSync() {
	startTime := time.Now()

	// 同步 SKU 数据（关联商品信息）
	syncProductSkus()

	// 同步订单明细数据
	syncOrderItems()

	// 更新同步时间
	lastSyncTime = time.Now().Format(time.RFC3339)

	utils.Info("数据同步完成, 耗时: %v", time.Since(startTime))
}

// syncProductSkus 同步 SKU 数据到 Elasticsearch
// 查询最近 5 分钟更新的 SKU，按 product_id 分组
// 对每个商品，批量更新其 SKU 列表到 ES 的 products 索引
func syncProductSkus() {
	database := GetDB()
	client := GetESClient()

	if database == nil || client == nil {
		utils.Warn("数据同步跳过: MySQL 或 ES 客户端未初始化")
		return
	}

	// 查询最近 5 分钟更新的 SKU 数据
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	var skus []skuRow
	err := database.Raw(`
		SELECT s.id, s.product_id, s.sku_name, s.price, s.stock, s.image,
		       s.spec_values, s.updated_at, p.name as product_name
		FROM skus s
		JOIN products p ON s.product_id = p.id
		WHERE s.updated_at >= ?
		ORDER BY s.product_id, s.id
	`, fiveMinutesAgo).Scan(&skus).Error

	if err != nil {
		utils.Error("同步 SKU 数据查询失败: %v", err)
		return
	}

	if len(skus) == 0 {
		return
	}

	// 按 product_id 分组
	productSkus := make(map[int][]skuRow)
	for _, sku := range skus {
		productSkus[sku.ProductID] = append(productSkus[sku.ProductID], sku)
	}

	// 批量更新 ES
	bulkRequest := client.Bulk()
	for productID, skuList := range productSkus {
		// 构建 SKU 文档列表
		var skuDocs []map[string]interface{}
		for _, sku := range skuList {
			var specValues map[string]string
			if len(sku.SpecValues) > 0 {
				json.Unmarshal(sku.SpecValues, &specValues)
			}

			skuDocs = append(skuDocs, map[string]interface{}{
				"id":          sku.ID,
				"product_id":  sku.ProductID,
				"sku_name":    sku.SkuName,
				"price":       sku.Price,
				"stock":       sku.Stock,
				"image":       sku.Image,
				"spec_values": specValues,
			})
		}

		// 创建批量更新请求
		// 使用 script 方式更新 skus 字段
		updateDoc := map[string]interface{}{
			"skus":       skuDocs,
			"updated_at": time.Now(),
		}

		bulkRequest = bulkRequest.Add(
			elastic.NewBulkUpdateRequest().
				Index("products").
				Id(fmt.Sprintf("%d", productID)).
				Doc(updateDoc),
		)
	}

	// 执行批量更新
	if bulkRequest.NumberOfActions() > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		bulkResponse, err := bulkRequest.Do(ctx)
		if err != nil {
			utils.Error("SKU 批量同步到 ES 失败: %v", err)
			return
		}

		// 统计成功和失败数
		successCount := 0
		failCount := 0
		for _, item := range bulkResponse.Items {
			if updateItem, ok := item["update"]; ok {
				if updateItem.Error != nil {
					failCount++
				} else {
					successCount++
				}
			}
		}

		utils.Info("SKU 同步完成, 成功: %d, 失败: %d", successCount, failCount)
	}
}

// syncOrderItems 同步订单明细数据到 Elasticsearch
// 查询最近 5 分钟更新的订单明细，按 order_id 分组
// 对每个订单，批量更新其订单明细列表到 ES 的 orders 索引
func syncOrderItems() {
	database := GetDB()
	client := GetESClient()

	if database == nil || client == nil {
		return
	}

	// 查询最近 5 分钟更新的订单明细
	fiveMinutesAgo := time.Now().Add(-5 * time.Minute)

	var items []orderItemRow
	err := database.Raw(`
		SELECT oi.id, oi.order_id, oi.product_id, oi.product_name,
		       oi.sku_id, oi.sku_name, oi.price, oi.quantity,
		       oi.subtotal, oi.updated_at
		FROM order_items oi
		WHERE oi.updated_at >= ?
		ORDER BY oi.order_id, oi.id
	`, fiveMinutesAgo).Scan(&items).Error

	if err != nil {
		utils.Error("同步订单明细查询失败: %v", err)
		return
	}

	if len(items) == 0 {
		return
	}

	// 按 order_id 分组
	orderItems := make(map[int][]orderItemRow)
	for _, item := range items {
		orderItems[item.OrderID] = append(orderItems[item.OrderID], item)
	}

	// 批量更新 ES
	bulkRequest := client.Bulk()
	for orderID, itemList := range orderItems {
		// 构建订单明细文档列表
		var itemDocs []map[string]interface{}
		for _, item := range itemList {
			itemDocs = append(itemDocs, map[string]interface{}{
				"id":           item.ID,
				"product_id":   item.ProductID,
				"product_name": item.ProductName,
				"sku_id":       item.SkuID,
				"sku_name":     item.SkuName,
				"price":        item.Price,
				"quantity":     item.Quantity,
				"subtotal":     item.Subtotal,
			})
		}

		// 创建批量更新请求
		updateDoc := map[string]interface{}{
			"items":      itemDocs,
			"updated_at": time.Now(),
		}

		bulkRequest = bulkRequest.Add(
			elastic.NewBulkUpdateRequest().
				Index("orders").
				Id(fmt.Sprintf("%d", orderID)).
				Doc(updateDoc),
		)
	}

	// 执行批量更新
	if bulkRequest.NumberOfActions() > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		bulkResponse, err := bulkRequest.Do(ctx)
		if err != nil {
			utils.Error("订单明细批量同步到 ES 失败: %v", err)
			return
		}

		// 统计成功和失败数
		successCount := 0
		failCount := 0
		for _, item := range bulkResponse.Items {
			if updateItem, ok := item["update"]; ok {
				if updateItem.Error != nil {
					failCount++
				} else {
					successCount++
				}
			}
		}

		utils.Info("订单明细同步完成, 成功: %d, 失败: %d", successCount, failCount)
	}
}

// GetLastSyncTime 获取上次同步时间
// 供健康检查接口调用，报告数据新鲜度
func GetLastSyncTime() string {
	return lastSyncTime
}
