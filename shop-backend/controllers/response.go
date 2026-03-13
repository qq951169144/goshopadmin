package controllers

import "github.com/gin-gonic/gin"

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`           // 业务错误码
	Message string      `json:"message"`        // 友好提示信息
	Data    interface{} `json:"data,omitempty"` // 数据
}

// ErrorDetail 错误详情（用于日志记录）
type ErrorDetail struct {
	Code      int    `json:"code"`       // 错误码
	Message   string `json:"message"`    // 友好提示
	Detail    string `json:"detail"`     // 详细错误信息（用于日志）
	RequestID string `json:"request_id"` // 请求ID
}

// Context key 定义
const (
	ErrorDetailKey = "error_detail"
	RequestIDKey   = "request_id"
)

// SetErrorDetail 将错误详情设置到 context
func SetErrorDetail(ctx *gin.Context, detail *ErrorDetail) {
	ctx.Set(ErrorDetailKey, detail)
}

// GetErrorDetail 从 context 获取错误详情
func GetErrorDetail(ctx *gin.Context) (*ErrorDetail, bool) {
	val, exists := ctx.Get(ErrorDetailKey)
	if !exists {
		return nil, false
	}
	return val.(*ErrorDetail), true
}
