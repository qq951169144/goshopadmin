package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"shop-backend/utils"

	"github.com/gin-gonic/gin"
)

// SearchProxyService 搜索代理服务
// 负责将搜索请求转发到 search-service，并处理降级逻辑
type SearchProxyService struct {
	searchServiceURL string       // search-service 的地址
	httpClient       *http.Client // HTTP 客户端，带超时
}

// NewSearchProxyService 创建搜索代理服务
func NewSearchProxyService() *SearchProxyService {
	// 默认使用 Docker 内部网络地址
	searchURL := "http://search-service:8082"

	return &SearchProxyService{
		searchServiceURL: searchURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // 5秒超时，防止 search-service 响应慢导致请求堆积
		},
	}
}

// ProxySearch 代理搜索请求到 search-service
// searchPath 为搜索路径，如 "/products"、"/orders" 等
func (s *SearchProxyService) ProxySearch(ctx *gin.Context, searchPath string) {
	// 构造完整的 search-service URL
	url := fmt.Sprintf("%s/api/search%s?%s", s.searchServiceURL, searchPath, ctx.Request.URL.RawQuery)

	// 发送 GET 请求到 search-service
	resp, err := s.httpClient.Get(url)
	if err != nil {
		// 调用失败，返回搜索不可用错误
		utils.Warn("search-service 调用失败: %v", err)
		ctx.JSON(200, gin.H{
			"code":    4081,
			"message": "搜索服务暂时不可用",
			"data":    nil,
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Error("读取 search-service 响应失败: %v", err)
		ctx.JSON(200, gin.H{
			"code":    4081,
			"message": "搜索服务暂时不可用",
			"data":    nil,
		})
		return
	}

	// 解析并转发 search-service 的响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		utils.Error("解析 search-service 响应失败: %v", err)
		ctx.Data(200, "application/json", body)
		return
	}

	ctx.JSON(200, result)
}
