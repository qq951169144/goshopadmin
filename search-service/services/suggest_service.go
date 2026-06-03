package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
)

// ============================================================
// 搜索建议服务
// 提供搜索自动补全/建议功能
// 根据用户输入的前缀，在指定索引和字段中使用 PrefixQuery 搜索
// 返回去重的建议词列表
// ============================================================

// SuggestParams 搜索建议参数
type SuggestParams struct {
	// Prefix 用户输入的搜索前缀
	Prefix string

	// Type 建议类型：product / order / user / customer
	Type string
}

// suggestConfig 建议配置，定义每种类型对应的索引和字段
type suggestConfig struct {
	// Index Elasticsearch 索引名
	Index string

	// Field 搜索字段名
	Field string
}

// suggestConfigs 建议类型配置映射表
// 每种建议类型对应一个 ES 索引和搜索字段
var suggestConfigs = map[string]suggestConfig{
	"product": {
		Index: "products",
		Field: "name", // 商品名称
	},
	"order": {
		Index: "orders",
		Field: "order_no", // 订单号
	},
	"user": {
		Index: "users",
		Field: "username", // 用户名
	},
	"customer": {
		Index: "customers",
		Field: "nickname", // 客户昵称
	},
}

// Suggest 搜索建议
// 根据类型确定索引和字段，使用 PrefixQuery 搜索前缀匹配的文档
// 从搜索结果中提取指定字段的值，去重后返回
// 设置 2 秒超时
func Suggest(params SuggestParams) ([]string, error) {
	client := GetESClient()
	if client == nil {
		return nil, errors.New("Elasticsearch 客户端未初始化")
	}

	// 获取建议配置
	cfg, ok := suggestConfigs[params.Type]
	if !ok {
		return nil, fmt.Errorf("不支持的建议类型: %s", params.Type)
	}

	var suggestions []string

	err := executeWithCircuitBreaker(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 使用 PrefixQuery 进行前缀匹配搜索
		prefixQuery := elastic.NewPrefixQuery(cfg.Field, params.Prefix)

		// 构建搜索请求
		searchResult, err := client.Search().
			Index(cfg.Index).
			Query(prefixQuery).
			Size(20). // 最多返回 20 条建议
			Do(ctx)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("搜索建议超时: %w", err)
			}
			return fmt.Errorf("搜索建议执行失败: %w", err)
		}

		// 从搜索结果中提取建议词并去重
		seen := make(map[string]bool)
		for _, hit := range searchResult.Hits.Hits {
			// 从 _source 中解析指定字段的值
			// hit.Source 是 json.RawMessage 类型，需要先反序列化为 map
			var source map[string]interface{}
			if err := json.Unmarshal(hit.Source, &source); err != nil {
				continue
			}
			if fieldValue, ok := source[cfg.Field]; ok {
				if strVal, ok := fieldValue.(string); ok && strVal != "" {
					if !seen[strVal] {
						seen[strVal] = true
						suggestions = append(suggestions, strVal)
					}
				}
			}
		}

		if suggestions == nil {
			suggestions = []string{}
		}

		return nil
	})

	return suggestions, err
}
