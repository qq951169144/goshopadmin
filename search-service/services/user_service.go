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
// 用户搜索服务
// 提供基于 Elasticsearch 的后台管理系统用户搜索功能
// 支持关键词搜索（用户名精确匹配 + 邮箱模糊匹配）、角色和状态筛选
// ============================================================

// UserSearchParams 用户搜索参数
type UserSearchParams struct {
	// Keyword 搜索关键词，匹配用户名（精确）或邮箱（模糊）
	Keyword string

	// RoleID 角色ID
	RoleID int

	// Status 用户状态
	Status string

	// Page 页码
	Page int

	// PageSize 每页记录数
	PageSize int
}

// SearchUsers 用户搜索
// 查询逻辑：
// 1. 如果有关键词，使用 Should 组合：用户名精确匹配 + 邮箱模糊匹配
// 2. 支持角色和状态精确筛选
func SearchUsers(params UserSearchParams) (*models.SearchResult, error) {
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

			// 邮箱模糊匹配
			shouldQuery.Should(elastic.NewMatchQuery("email", params.Keyword))

			// 至少匹配一个条件
			shouldQuery.MinimumNumberShouldMatch(1)
			boolQuery.Must(shouldQuery)
		}

		// 角色筛选
		if params.RoleID > 0 {
			boolQuery.Filter(elastic.NewTermQuery("role_id", params.RoleID))
		}

		// 状态筛选
		if params.Status != "" {
			boolQuery.Filter(elastic.NewTermQuery("status", params.Status))
		}

		// 构建搜索请求
		search := client.Search().
			Index("users").
			Query(boolQuery)

		// 设置排序
		search = applyUserSort(search, params.Keyword)

		// 设置分页
		from := (params.Page - 1) * params.PageSize
		search = search.From(from).Size(params.PageSize)

		// 执行搜索
		searchResult, err := search.Do(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("用户搜索超时: %w", err)
			}
			return fmt.Errorf("用户搜索执行失败: %w", err)
		}

		// 解析搜索结果
		var users []models.UserDoc
		for _, hit := range searchResult.Hits.Hits {
			var user models.UserDoc
			if err := json.Unmarshal(hit.Source, &user); err != nil {
				utils.Warn("用户搜索结果解析失败: %v", err)
				continue
			}
			users = append(users, user)
		}

		if users == nil {
			users = []models.UserDoc{}
		}

		result = &models.SearchResult{
			Total:    searchResult.TotalHits(),
			Page:     params.Page,
			PageSize: params.PageSize,
			Items:    users,
		}

		return nil
	})

	return result, err
}

// applyUserSort 应用用户排序规则
func applyUserSort(search *elastic.SearchService, keyword string) *elastic.SearchService {
	if keyword != "" {
		search = search.Sort("_score", false)
	} else {
		search = search.Sort("created_at", false)
	}
	return search
}
