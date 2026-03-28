package controllers

import (
	stderrors "errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"goshopadmin/errors"
	"goshopadmin/services"
	"gorm.io/gorm"
)

// BaseController 基础控制器
type BaseController struct{}

// GetUserID 从上下文获取用户ID，如果不存在则返回错误
func (c *BaseController) GetUserID(ctx *gin.Context) (int, bool) {
	userID, exists := ctx.Get("userID")
	if !exists {
		c.ResponseError(ctx, errors.CodeUnauthorized, nil)
		return 0, false
	}
	return userID.(int), true
}

// GetMerchantIDFromContext 从上下文获取商户ID
func (c *BaseController) GetMerchantIDFromContext(ctx *gin.Context, db *gorm.DB) (int, error) {
	userID, ok := c.GetUserID(ctx)
	if !ok {
		return 0, stderrors.New("未授权")
	}
	
	productService := services.NewProductService(db, nil)
	return productService.GetMerchantIDByUserID(userID)
}

// ResponseSuccess 返回成功响应
func (c *BaseController) ResponseSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    errors.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// ResponseError 返回错误响应（HTTP 始终 200，错误信息放在 body 中）
// 参数:
//   - ctx: gin context
//   - bizCode: 业务错误码
//   - err: 原始错误（用于日志记录，不会返回给前端）
func (c *BaseController) ResponseError(ctx *gin.Context, bizCode int, err error) {
	// 从 context 获取 RequestID（由 Logger 中间件生成）
	requestID, _ := ctx.Get(RequestIDKey)
	if requestID == nil {
		requestID = "unknown"
	}

	// 获取友好的错误信息
	message := errors.GetErrorMessage(bizCode)

	// 构建错误详情（用于日志）
	errorDetail := &ErrorDetail{
		Code:      bizCode,
		Message:   message,
		Detail:    "",
		RequestID: requestID.(string),
	}

	// 如果有原始错误，记录详细信息
	if err != nil {
		errorDetail.Detail = err.Error()
	}

	// 将错误详情存入 context，供中间件使用
	SetErrorDetail(ctx, errorDetail)

	// HTTP 始终返回 200，错误信息放在 body 中
	// 前端根据 body.code 判断成功/失败
	ctx.JSON(http.StatusOK, Response{
		Code:    bizCode,
		Message: message,
		Data:    nil,
	})
}

// ResponseErrorWithMessage 返回自定义错误消息（特殊场景使用）
func (c *BaseController) ResponseErrorWithMessage(ctx *gin.Context, bizCode int, message string, err error) {
	requestID, _ := ctx.Get(RequestIDKey)
	if requestID == nil {
		requestID = "unknown"
	}

	errorDetail := &ErrorDetail{
		Code:      bizCode,
		Message:   message,
		Detail:    "",
		RequestID: requestID.(string),
	}

	if err != nil {
		errorDetail.Detail = err.Error()
	}

	SetErrorDetail(ctx, errorDetail)

	// HTTP 始终返回 200
	ctx.JSON(http.StatusOK, Response{
		Code:    bizCode,
		Message: message,
		Data:    nil,
	})
}
