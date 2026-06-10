package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"search-service/utils"

	"github.com/olivere/elastic/v7"
)

// esTimestamp 返回 ES 兼容的时间戳格式（毫秒精度）
func esTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

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
	ID            int             `gorm:"column:id"`
	ProductID     int             `gorm:"column:product_id"`
	SkuCode       string          `gorm:"column:sku_code"`
	Price         float64         `gorm:"column:price"`
	OriginalPrice float64         `gorm:"column:original_price"`
	Stock         int             `gorm:"column:stock"`
	Attributes    json.RawMessage `gorm:"column:attributes"`
	UpdatedAt     time.Time       `gorm:"column:updated_at"`
}

// 订单明细同步用的 MySQL 查询结果结构
type orderItemRow struct {
	ID            int             `gorm:"column:id"`
	OrderID       int             `gorm:"column:order_id"`
	ProductID     int             `gorm:"column:product_id"`
	ProductName   string          `gorm:"column:product_name"`
	SkuID         int             `gorm:"column:sku_id"`
	SkuAttributes json.RawMessage `gorm:"column:sku_attributes"`
	Price         float64         `gorm:"column:price"`
	Quantity      int             `gorm:"column:quantity"`
	TotalAmount   float64         `gorm:"column:total_amount"`
	UpdatedAt     time.Time       `gorm:"column:updated_at"`
	// ProductImage 商品主图URL
	ProductImage string `gorm:"column:product_image"`
}

// StartSyncService 启动数据同步定时任务
// 每 60 秒执行一次同步，查询最近 5 分钟更新的数据
// 使用 5 分钟窗口而非 60 秒，是为了避免边界情况导致数据遗漏
// 首次启动时执行全量同步，确保 ES 数据完整
func StartSyncService() {
	syncTicker = time.NewTicker(60 * time.Second)
	syncDone = make(chan struct{})

	// 记录初始同步时间
	lastSyncTime = time.Now().Format(time.RFC3339)

	// 启动同步协程
	go func() {
		// 首次启动时执行全量同步
		utils.Info("数据同步服务启动，开始首次全量同步")
		runFullSync()

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

// runSync 执行一次增量数据同步
// 依次同步 SKU 数据和订单明细数据（仅最近5分钟更新的）
func runSync() {
	startTime := time.Now()

	// 同步 SKU 数据（关联商品信息）
	syncProductSkus()

	// 同步订单明细数据
	syncOrderItems()

	// 更新同步时间
	lastSyncTime = time.Now().Format(time.RFC3339)

	utils.Info("增量同步完成, 耗时: %v", time.Since(startTime))
}

// runFullSync 执行全量数据同步
// 首次启动时调用，同步所有订单明细数据到 ES
// 确保订单文档中的 items nested 数组数据完整
func runFullSync() {
	startTime := time.Now()

	// 全量同步 SKU 数据
	syncProductSkusFull()

	// 全量同步订单明细数据
	syncOrderItemsFull()

	// 更新同步时间
	lastSyncTime = time.Now().Format(time.RFC3339)

	utils.Info("全量同步完成, 耗时: %v", time.Since(startTime))
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
		SELECT s.id, s.product_id, s.sku_code, s.price, s.original_price, s.stock,
		       s.attributes, s.updated_at
		FROM product_skus s
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
			var attributes map[string]string
			if len(sku.Attributes) > 0 {
				json.Unmarshal(sku.Attributes, &attributes)
			}

			skuDocs = append(skuDocs, map[string]interface{}{
				"id":             sku.ID,
				"product_id":     sku.ProductID,
				"sku_code":       sku.SkuCode,
				"price":          sku.Price,
				"original_price": sku.OriginalPrice,
				"stock":          sku.Stock,
				"attributes":     attributes,
			})
		}

		// 创建批量更新请求
		// 使用 script 方式更新 skus 字段
		updateDoc := map[string]interface{}{
			"skus":       skuDocs,
			"updated_at": esTimestamp(),
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
					utils.Warn("SKU 同步失败, 文档ID: %s, 错误: %s", updateItem.Id, updateItem.Error.Reason)
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
		       oi.sku_id, oi.sku_attributes, oi.price, oi.quantity,
		       oi.total_amount, oi.updated_at,
		       COALESCE(pi.image_url, '') AS product_image
		FROM order_items oi
		LEFT JOIN (
		    SELECT product_id, MIN(image_url) AS image_url
		    FROM product_images
		    WHERE is_main = 1
		    GROUP BY product_id
		) pi ON oi.product_id = pi.product_id
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
			// sku_attributes 在 ES mapping 中是 keyword 类型，需要存储为字符串
			skuAttrStr := ""
			if len(item.SkuAttributes) > 0 {
				skuAttrStr = string(item.SkuAttributes)
			}

			itemDocs = append(itemDocs, map[string]interface{}{
				"id":             item.ID,
				"product_id":     item.ProductID,
				"product_name":   item.ProductName,
				"product_image":  item.ProductImage,
				"sku_id":         item.SkuID,
				"sku_attributes": skuAttrStr,
				"price":          item.Price,
				"quantity":       item.Quantity,
				"total_amount":   item.TotalAmount,
			})
		}

		// 创建批量更新请求
		updateDoc := map[string]interface{}{
			"items":      itemDocs,
			"updated_at": esTimestamp(),
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
					utils.Warn("订单明细同步失败, 文档ID: %s, 错误: %s", updateItem.Id, updateItem.Error.Reason)
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

// syncProductSkusFull 全量同步 SKU 数据到 Elasticsearch
// 查询所有 SKU 数据，按 product_id 分组后批量更新到 ES
func syncProductSkusFull() {
	database := GetDB()
	client := GetESClient()

	if database == nil || client == nil {
		utils.Warn("全量同步 SKU 跳过: MySQL 或 ES 客户端未初始化")
		return
	}

	var skus []skuRow
	err := database.Raw(`
		SELECT s.id, s.product_id, s.sku_code, s.price, s.original_price, s.stock,
		       s.attributes, s.updated_at
		FROM product_skus s
		ORDER BY s.product_id, s.id
	`).Scan(&skus).Error

	if err != nil {
		utils.Error("全量同步 SKU 数据查询失败: %v", err)
		return
	}

	if len(skus) == 0 {
		utils.Info("全量同步 SKU: 无数据需要同步")
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
		var skuDocs []map[string]interface{}
		for _, sku := range skuList {
			var attributes map[string]string
			if len(sku.Attributes) > 0 {
				json.Unmarshal(sku.Attributes, &attributes)
			}

			skuDocs = append(skuDocs, map[string]interface{}{
				"id":             sku.ID,
				"product_id":     sku.ProductID,
				"sku_code":       sku.SkuCode,
				"price":          sku.Price,
				"original_price": sku.OriginalPrice,
				"stock":          sku.Stock,
				"attributes":     attributes,
			})
		}

		updateDoc := map[string]interface{}{
			"skus":       skuDocs,
			"updated_at": esTimestamp(),
		}

		bulkRequest = bulkRequest.Add(
			elastic.NewBulkUpdateRequest().
				Index("products").
				Id(fmt.Sprintf("%d", productID)).
				Doc(updateDoc),
		)
	}

	if bulkRequest.NumberOfActions() > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		bulkResponse, err := bulkRequest.Do(ctx)
		if err != nil {
			utils.Error("全量同步 SKU 到 ES 失败: %v", err)
			return
		}

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

		utils.Info("全量同步 SKU 完成, 成功: %d, 失败: %d", successCount, failCount)
	}
}

// syncOrderItemsFull 全量同步订单明细数据到 Elasticsearch
// 查询所有订单明细，按 order_id 分组后批量更新到 ES
// 关联 product_images 表获取商品主图URL
func syncOrderItemsFull() {
	database := GetDB()
	client := GetESClient()

	if database == nil || client == nil {
		utils.Warn("全量同步订单明细跳过: MySQL 或 ES 客户端未初始化")
		return
	}

	var items []orderItemRow
	err := database.Raw(`
		SELECT oi.id, oi.order_id, oi.product_id, oi.product_name,
		       oi.sku_id, oi.sku_attributes, oi.price, oi.quantity,
		       oi.total_amount, oi.updated_at,
		       COALESCE(pi.image_url, '') AS product_image
		FROM order_items oi
		LEFT JOIN (
		    SELECT product_id, MIN(image_url) AS image_url
		    FROM product_images
		    WHERE is_main = 1
		    GROUP BY product_id
		) pi ON oi.product_id = pi.product_id
		ORDER BY oi.order_id, oi.id
	`).Scan(&items).Error

	if err != nil {
		utils.Error("全量同步订单明细查询失败: %v", err)
		return
	}

	if len(items) == 0 {
		utils.Info("全量同步订单明细: 无数据需要同步")
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
		var itemDocs []map[string]interface{}
		for _, item := range itemList {
			// sku_attributes 在 ES mapping 中是 keyword 类型，需要存储为字符串
			skuAttrStr := ""
			if len(item.SkuAttributes) > 0 {
				skuAttrStr = string(item.SkuAttributes)
			}

			itemDocs = append(itemDocs, map[string]interface{}{
				"id":             item.ID,
				"product_id":     item.ProductID,
				"product_name":   item.ProductName,
				"product_image":  item.ProductImage,
				"sku_id":         item.SkuID,
				"sku_attributes": skuAttrStr,
				"price":          item.Price,
				"quantity":       item.Quantity,
				"total_amount":   item.TotalAmount,
			})
		}

		updateDoc := map[string]interface{}{
			"items":      itemDocs,
			"updated_at": esTimestamp(),
		}

		bulkRequest = bulkRequest.Add(
			elastic.NewBulkUpdateRequest().
				Index("orders").
				Id(fmt.Sprintf("%d", orderID)).
				Doc(updateDoc),
		)
	}

	if bulkRequest.NumberOfActions() > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		bulkResponse, err := bulkRequest.Do(ctx)
		if err != nil {
			utils.Error("全量同步订单明细到 ES 失败: %v", err)
			return
		}

		successCount := 0
		failCount := 0
		for _, item := range bulkResponse.Items {
			if updateItem, ok := item["update"]; ok {
				if updateItem.Error != nil {
					failCount++
					utils.Warn("全量同步订单明细失败, 文档ID: %s, 错误: %s", updateItem.Id, updateItem.Error.Reason)
				} else {
					successCount++
				}
			}
		}

		utils.Info("全量同步订单明细完成, 成功: %d, 失败: %d", successCount, failCount)
	}
}
