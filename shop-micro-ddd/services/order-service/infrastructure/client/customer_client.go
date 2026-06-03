package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CustomerHTTPClient 客户服务 HTTP 客户端
// 实现 order.CustomerServiceProvider 接口
type CustomerHTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewCustomerHTTPClient 创建客户服务 HTTP 客户端
func NewCustomerHTTPClient(baseURL string) *CustomerHTTPClient {
	return &CustomerHTTPClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// verifyAddressResponse 验证地址响应
type verifyAddressResponse struct {
	MerchantID int `json:"merchant_id"`
}

// VerifyAddress 验证地址（HTTP 调用客户服务）
func (c *CustomerHTTPClient) VerifyAddress(ctx context.Context, customerID, addressID int) (int, error) {
	url := fmt.Sprintf("%s/api/internal/customers/%d/addresses/%d/verify", c.baseURL, customerID, addressID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("调用客户服务失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return 0, err
	}

	if apiResp.Code != 0 {
		return 0, fmt.Errorf("客户服务返回错误: %s", apiResp.Message)
	}

	var data verifyAddressResponse
	if err := json.Unmarshal(apiResp.Data, &data); err != nil {
		return 0, err
	}

	return data.MerchantID, nil
}
