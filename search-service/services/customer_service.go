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
// 客户搜索服务
// 提供基于 Elasticsearch 的 C 端商城客户搜索功能
// 支持关键词搜索（用户名/手机号精确匹配 + 邮箱模糊匹配 + 昵称 IK 分词搜索）、状态筛选
// ============================================================

// CustomerSearchParams 客户搜索参数
type CustomerSearchParams struct {
	// Keyword 搜索关键词，匹配用户名/手机号（精确）或邮箱（模糊）或昵称（IK 分词）
	Keyword string

	// Status 客户状态
	Status string

	// Page 页码
	Page int

	// PageSize 每页记录数
	PageSize int
}

// SearchCustomers 客户搜索
// 查询逻辑：
// 1. 如果有关键词，使用 Should 组合：
//    - 用户名精确匹配
//    - 手机号精确匹配
//    - 邮箱模糊匹配
//    - 昵称 IK 分词搜索（支持中文分词）
// 2. 支持状态精确筛选
func SearchCustomers(params CustomerSearchParams) (*models.SearchResult, error) {
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
			shouldQuery := elastic.NewBoolQuery()

			// 用户名精确匹配
			shouldQuery.Should(elastic.NewTermQuery("username", params.Keyword))

			// 手机号精确匹配
			shouldQuery.Should(elastic.NewTermQuery("phone", params.Keyword))

			// 邮箱模糊匹配
			shouldQuery.Should(elastic.NewMatchQuery("email", params.Keyword))

			// 昵称 IK 分词搜索，支持中文分词
			shouldQuery.Should(elastic.NewMatchQuery("nickname", params.Keyword).
				Analyzer("ik_smart")) // 使用 IK 智能分词

			// 至少匹配一个条件
			shouldQuery.MinimumNumberShouldMatch(1)
			boolQuery.Must(shouldQuery)
		}

		// 状态筛选
		if params.Status != "" {
			boolQuery.Filter(elastic.NewTermQuery("status", params.Status))
		}

		// 构建搜索请求
		search := client.Search().
			Index("customers").
			Query(boolQuery)

		// 设置排序
		search = applyCustomerSort(search, params.Keyword)

		// 设置分页
		from := (params.Page - 1) * params.PageSize
		search = search.From(from).Size(params.PageSize)

		// 执行搜索
		searchResult, err := search.Do(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("客户搜索超时: %w", err)
			}
			return fmt.Errorf("客户搜索执行失败: %w", err)
		}

		// 解析搜索结果
		var customers []models.CustomerDoc
		for _, hit := range searchResult.Hits.Hits {
			var customer models.CustomerDoc
			if err := json.Unmarshal(hit.Source, &customer); err != nil {
				utils.Warn("客户搜索结果解析失败: %v", err)
				continue
			}
			customers = append(customers, customer)
		}

		if customers == nil {
			customers = []models.CustomerDoc{}
		}

		result = &models.SearchResult{
			Total:    searchResult.TotalHits(),
			Page:     params.Page,
			PageSize: params.PageSize,
			Items:    customers,
		}

		return nil
	})

	return result, err
}

// applyCustomerSort 应用客户排序规则
func applyCustomerSort(search *elastic.SearchService, keyword string) *elastic.SearchService {
	if keyword != "" {
		search = search.Sort("_score", false)
	} else {
		search = search.Sort("created_at", false)
	}
	return search
}
