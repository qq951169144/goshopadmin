package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"search-service/models"
	"search-service/utils"

	"github.com/olivere/elastic/v7"
)

// ============================================================
// 订单搜索服务
// 提供基于 Elasticsearch 的订单搜索功能
// 支持关键词搜索（订单号精确匹配 + 商品名称模糊匹配）、多状态筛选、时间范围和金额范围筛选
// 订单明细使用 Nested 类型，需要使用 NestedQuery 搜索嵌套字段
// ============================================================

// OrderSearchParams 订单搜索参数
type OrderSearchParams struct {
	// Keyword 搜索关键词，匹配订单号（精确）或订单明细中的商品名称（模糊）
	Keyword string

	// CustomerID 客户ID
	CustomerID int

	// MerchantID 商户ID
	MerchantID int

	// Status 订单状态
	Status string

	// PaymentStatus 支付状态
	PaymentStatus string

	// ShippingStatus 物流状态
	ShippingStatus string

	// StartDate 开始日期，格式：2006-01-02
	StartDate string

	// EndDate 结束日期，格式：2006-01-02
	EndDate string

	// MinAmount 最小金额
	MinAmount float64

	// MaxAmount 最大金额
	MaxAmount float64

	// Sort 排序字段
	Sort string

	// Order 排序方向
	Order string

	// Page 页码
	Page int

	// PageSize 每页记录数
	PageSize int
}

// SearchOrders 订单搜索
// 查询逻辑：
// 1. 如果有关键词，使用 Should 组合：订单号精确匹配 + 订单明细商品名称模糊匹配
// 2. 订单明细中的商品名称搜索需要使用 NestedQuery（因为 items 是嵌套类型）
// 3. 支持客户、商户、状态、支付状态、物流状态精确筛选
// 4. 支持时间范围和金额范围筛选
// 设置 3 秒超时
func SearchOrders(params OrderSearchParams) (*models.SearchResult, error) {
	client := GetESClient()
	if client == nil {
		return nil, errors.New("Elasticsearch 客户端未初始化")
	}

	var result *models.SearchResult

	err := executeWithCircuitBreaker(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		boolQuery := elastic.NewBoolQuery()

		// 关键词搜索
		if params.Keyword != "" {
			// 使用 Should 组合多个匹配条件，至少需要匹配一个
			shouldQuery := elastic.NewBoolQuery()

			// 订单号精确匹配
			shouldQuery.Should(elastic.NewTermQuery("order_no", params.Keyword))

			// 订单明细中的商品名称模糊匹配（使用 NestedQuery）
			// 因为 items 是嵌套类型（nested），需要使用 NestedQuery 才能搜索内部字段
			nestedQuery := elastic.NewNestedQuery("items",
				elastic.NewMatchQuery("items.product_name", params.Keyword).
					Analyzer("ik_smart"), // 使用 IK 分词
			)
			shouldQuery.Should(nestedQuery)

			// 设置至少匹配一个条件
			shouldQuery.MinimumNumberShouldMatch(1)
			boolQuery.Must(shouldQuery)
		}

		// 客户筛选
		if params.CustomerID > 0 {
			boolQuery.Filter(elastic.NewTermQuery("customer_id", params.CustomerID))
		}

		// 商户筛选
		if params.MerchantID > 0 {
			boolQuery.Filter(elastic.NewTermQuery("merchant_id", params.MerchantID))
		}

		// 订单状态筛选
		if params.Status != "" {
			boolQuery.Filter(elastic.NewTermQuery("status", params.Status))
		}

		// 支付状态筛选
		if params.PaymentStatus != "" {
			boolQuery.Filter(elastic.NewTermQuery("payment_status", params.PaymentStatus))
		}

		// 物流状态筛选
		if params.ShippingStatus != "" {
			boolQuery.Filter(elastic.NewTermQuery("shipping_status", params.ShippingStatus))
		}

		// 时间范围筛选
		if params.StartDate != "" || params.EndDate != "" {
			rangeQuery := elastic.NewRangeQuery("created_at")
			if params.StartDate != "" {
				startTime, err := time.Parse("2006-01-02", params.StartDate)
				if err == nil {
					rangeQuery.Gte(startTime)
				}
			}
			if params.EndDate != "" {
				endTime, err := time.Parse("2006-01-02", params.EndDate)
				if err == nil {
					// 结束日期包含当天，所以加一天再减 1 纳秒
					rangeQuery.Lte(endTime.AddDate(0, 0, 1).Add(-time.Nanosecond))
				}
			}
			boolQuery.Filter(rangeQuery)
		}

		// 金额范围筛选
		if params.MinAmount > 0 || params.MaxAmount > 0 {
			rangeQuery := elastic.NewRangeQuery("total_amount")
			if params.MinAmount > 0 {
				rangeQuery.Gte(params.MinAmount)
			}
			if params.MaxAmount > 0 {
				rangeQuery.Lte(params.MaxAmount)
			}
			boolQuery.Filter(rangeQuery)
		}

		// 构建搜索请求
		search := client.Search().
			Index("orders").
			Query(boolQuery)

		// 设置排序
		search = applyOrderSort(search, params.Sort, params.Order, params.Keyword)

		// 设置分页
		from := (params.Page - 1) * params.PageSize
		search = search.From(from).Size(params.PageSize)

		// 执行搜索
		searchResult, err := search.Do(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("订单搜索超时: %w", err)
			}
			return fmt.Errorf("订单搜索执行失败: %w", err)
		}

		// 解析搜索结果
		var orders []models.OrderDoc
		for _, hit := range searchResult.Hits.Hits {
			var order models.OrderDoc
			if err := json.Unmarshal(hit.Source, &order); err != nil {
				utils.Warn("订单搜索结果解析失败: %v", err)
				continue
			}
			orders = append(orders, order)
		}

		if orders == nil {
			orders = []models.OrderDoc{}
		}

		result = &models.SearchResult{
			Total:    searchResult.TotalHits(),
			Page:     params.Page,
			PageSize: params.PageSize,
			Items:    orders,
		}

		return nil
	})

	return result, err
}

// applyOrderSort 应用订单排序规则
func applyOrderSort(search *elastic.SearchService, sort, order, keyword string) *elastic.SearchService {
	ascending := order == "asc"

	switch sort {
	case "total_amount":
		search = search.Sort("total_amount", ascending)
	case "created_at":
		search = search.Sort("created_at", ascending)
	default:
		if keyword != "" {
			search = search.Sort("_score", false)
		} else {
			search = search.Sort("created_at", false)
		}
	}

	return search
}
