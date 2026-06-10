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
// 商品搜索服务
// 提供基于 Elasticsearch 的商品搜索功能
// 支持关键词搜索（IK 分词）、分类/商户/状态/价格区间筛选、高亮、排序和分页
// ============================================================

// ProductSearchParams 商品搜索参数
type ProductSearchParams struct {
	// Keyword 搜索关键词，使用 IK 分词器匹配商品名称和描述
	Keyword string

	// CategoryID 分类ID，精确筛选
	CategoryID int

	// MerchantID 商户ID，精确筛选
	MerchantID int

	// Status 商品状态筛选：active / inactive
	Status string

	// MinPrice 最低价格
	MinPrice float64

	// MaxPrice 最高价格
	MaxPrice float64

	// Sort 排序字段：relevance / price / sales / created_at
	Sort string

	// Order 排序方向：desc / asc
	Order string

	// Page 页码
	Page int

	// PageSize 每页记录数
	PageSize int
}

// SearchProducts 商品搜索
// 使用 BoolQuery 组合多个查询条件：
// - MultiMatchQuery: 关键词搜索（name 和 description 字段，使用 ik_smart 分词器）
// - TermQuery: 分类、商户、状态精确筛选
// - RangeQuery: 价格区间筛选
// 支持高亮显示匹配的关键词
// 设置 3 秒超时，防止慢查询影响服务
func SearchProducts(params ProductSearchParams) (*models.SearchResult, error) {
	client := GetESClient()
	if client == nil {
		return nil, errors.New("Elasticsearch 客户端未初始化")
	}

	var result *models.SearchResult

	err := executeWithCircuitBreaker(func() error {
		// 创建带超时的上下文，3 秒超时
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		// 构建 Bool 查询
		boolQuery := elastic.NewBoolQuery()

		// 关键词搜索：使用 MultiMatchQuery 同时搜索 name 和 description 字段
		// 使用 ik_smart 分词器进行中文分词
		if params.Keyword != "" {
			shouldQuery := elastic.NewBoolQuery()

			// IK 分词匹配（中文搜索主要靠这个）
			multiMatch := elastic.NewMultiMatchQuery(params.Keyword, "name", "description").
				Type("best_fields"). // best_fields 策略：取最高分字段的得分
				Analyzer("ik_smart") // 使用 IK 智能分词
			shouldQuery.Should(multiMatch)

			// 通配符匹配（补充：支持英文+数字混合词的子串搜索，如 "iphone" 匹配 "iphone17"）
			wildcardName := elastic.NewWildcardQuery("name.keyword", "*"+params.Keyword+"*")
			shouldQuery.Should(wildcardName)

			shouldQuery.MinimumNumberShouldMatch(1)
			boolQuery.Must(shouldQuery)
		}

		// 分类筛选：精确匹配 category_id
		if params.CategoryID > 0 {
			boolQuery.Filter(elastic.NewTermQuery("category_id", params.CategoryID))
		}

		// 商户筛选：精确匹配 merchant_id
		if params.MerchantID > 0 {
			boolQuery.Filter(elastic.NewTermQuery("merchant_id", params.MerchantID))
		}

		// 状态筛选：精确匹配 status
		if params.Status != "" {
			boolQuery.Filter(elastic.NewTermQuery("status", params.Status))
		}

		// 价格区间筛选：使用 RangeQuery 筛选 min_price 和 max_price
		// 商品的 min_price <= 用户设定的 max_price 且 max_price >= 用户设定的 min_price
		if params.MinPrice > 0 || params.MaxPrice > 0 {
			if params.MinPrice > 0 {
				boolQuery.Filter(elastic.NewRangeQuery("max_price").Gte(params.MinPrice))
			}
			if params.MaxPrice > 0 {
				boolQuery.Filter(elastic.NewRangeQuery("min_price").Lte(params.MaxPrice))
			}
		}

		// 构建搜索请求
		search := client.Search().
			Index("products"). // 搜索 products 索引
			Query(boolQuery)

		// 设置高亮：在 name 和 description 字段中高亮匹配的关键词
		if params.Keyword != "" {
			highlight := elastic.NewHighlight().
				Field("name").         // 高亮商品名称
				Field("description").  // 高亮商品描述
				PreTags("<em>").       // 高亮前缀标签
				PostTags("</em>").     // 高亮后缀标签
				FragmentSize(200)      // 高亮片段最大长度
			search = search.Highlight(highlight)
		}

		// 设置排序
		search = applyProductSort(search, params.Sort, params.Order, params.Keyword)

		// 设置分页
		from := (params.Page - 1) * params.PageSize
		search = search.From(from).Size(params.PageSize)

		// 执行搜索
		searchResult, err := search.Do(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("商品搜索超时: %w", err)
			}
			return fmt.Errorf("商品搜索执行失败: %w", err)
		}

		// 解析搜索结果
		var products []models.ProductDoc
		for _, hit := range searchResult.Hits.Hits {
			var product models.ProductDoc
			if err := json.Unmarshal(hit.Source, &product); err != nil {
				utils.Warn("商品搜索结果解析失败: %v", err)
				continue
			}

			// 处理高亮：用高亮文本替换原始文本
			if hit.Highlight != nil {
				if names, ok := hit.Highlight["name"]; ok && len(names) > 0 {
					product.Name = names[0]
				}
				if descs, ok := hit.Highlight["description"]; ok && len(descs) > 0 {
					product.Description = descs[0]
				}
			}

			products = append(products, product)
		}

		if products == nil {
			products = []models.ProductDoc{}
		}

		result = &models.SearchResult{
			Total:    searchResult.TotalHits(),
			Page:     params.Page,
			PageSize: params.PageSize,
			Items:    products,
		}

		return nil
	})

	return result, err
}

// applyProductSort 应用商品排序规则
// sort: 排序字段名
// order: 排序方向（asc/desc）
// keyword: 搜索关键词（有关键词时默认按相关度排序）
func applyProductSort(search *elastic.SearchService, sort, order, keyword string) *elastic.SearchService {
	// 确定排序方向，默认降序
	ascending := order == "asc"

	switch sort {
	case "price":
		// 按最低价格排序
		search = search.Sort("min_price", ascending)
	case "sales":
		// 按销量排序
		search = search.Sort("sales", ascending)
	case "created_at":
		// 按创建时间排序
		search = search.Sort("created_at", ascending)
	default:
		// 默认排序：有关键词时按相关度排序，无关键词时按创建时间降序
		if keyword != "" {
			search = search.Sort("_score", false)
		} else {
			search = search.Sort("created_at", false)
		}
	}

	return search
}
