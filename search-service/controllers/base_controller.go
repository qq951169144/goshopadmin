package controllers

import (
	"net/http"

	svcErrors "search-service/errors"

	"github.com/gin-gonic/gin"
)

// ============================================================
// BaseController 基础控制器
// 提供统一的成功/错误响应方法，所有控制器应嵌入此结构体
// 响应格式：HTTP 状态码始终为 200，业务状态通过 body 中的 code 字段区分
// ============================================================

// BaseController 基础控制器结构体
// 其他控制器通过嵌入此结构体获得统一的响应处理能力
type BaseController struct{}

// ResponseSuccess 返回成功响应
// HTTP 状态码始终为 200，body 中 code=0 表示成功
// ctx: Gin 上下文
// data: 返回给前端的数据，可以是任意类型
func (bc *BaseController) ResponseSuccess(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, gin.H{
		"code":    svcErrors.CodeSuccess,
		"message": "success",
		"data":    data,
	})
}

// ResponseError 返回错误响应
// HTTP 状态码始终为 200，错误信息放在 body 中
// 前端根据 body.code 判断成功/失败，非 0 表示失败
// ctx: Gin 上下文
// bizCode: 业务错误码，定义在 errors/code.go 中
// err: 原始错误对象，用于日志记录，前端显示 ErrorMessage 中的友好提示
func (bc *BaseController) ResponseError(ctx *gin.Context, bizCode int, err error) {
	// 获取错误码对应的友好提示信息
	message := svcErrors.GetErrorMessage(bizCode)

	// 如果有原始错误，将错误详情附加到消息中，方便调试
	if err != nil {
		message = message + ": " + err.Error()
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    bizCode,
		"message": message,
		"data":    nil,
	})
}
