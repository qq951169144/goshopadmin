package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// ProductHTTPClient 商品服务 HTTP 客户端
// 实现 order.ProductServiceProvider 接口
type ProductHTTPClient struct {
	baseURL string // 商品服务地址，如 http://localhost:8091
	client  *http.Client
}

// NewProductHTTPClient 创建商品服务 HTTP 客户端
func NewProductHTTPClient(baseURL string) *ProductHTTPClient {
	return &ProductHTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// apiResponse 通用 API 响应
type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// productSKUResponse 商品SKU查询响应
type productSKUResponse struct {
	Name     string          `json:"name"`
	SkuAttrs string          `json:"sku_attrs"`
	Price    decimal.Decimal `json:"price"`
	Stock    int             `json:"stock"`
}

// GetProductAndSKU 获取商品和SKU信息（HTTP 调用商品服务）
func (c *ProductHTTPClient) GetProductAndSKU(ctx context.Context, productID, skuID int) (string, string, decimal.Decimal, int, error) {
	url := fmt.Sprintf("%s/api/internal/products/%d/skus/%d", c.baseURL, productID, skuID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", decimal.Zero, 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", decimal.Zero, 0, fmt.Errorf("调用商品服务失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", decimal.Zero, 0, err
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", "", decimal.Zero, 0, err
	}

	if apiResp.Code != 0 {
		return "", "", decimal.Zero, 0, fmt.Errorf("商品服务返回错误: %s", apiResp.Message)
	}

	var data productSKUResponse
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return "", "", decimal.Zero, 0, err
	}

	return data.Name, data.SkuAttrs, data.Price, data.Stock, nil
}

// DeductStock 扣减库存（HTTP 调用商品服务）
func (c *ProductHTTPClient) DeductStock(ctx context.Context, skuID int, quantity int) error {
	url := fmt.Sprintf("%s/api/internal/skus/%d/deduct", c.baseURL, skuID)
	body := fmt.Sprintf(`{"quantity": %d}`, quantity)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("调用商品服务扣减库存失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("商品服务扣减库存返回异常状态码: %d", resp.StatusCode)
	}

	return nil
}

// RestoreStock 恢复库存（HTTP 调用商品服务）
func (c *ProductHTTPClient) RestoreStock(ctx context.Context, skuID int, quantity int) error {
	url := fmt.Sprintf("%s/api/internal/skus/%d/restore", c.baseURL, skuID)
	body := fmt.Sprintf(`{"quantity": %d}`, quantity)
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("调用商品服务恢复库存失败: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
