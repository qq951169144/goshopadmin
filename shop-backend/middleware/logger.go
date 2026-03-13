package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"shop-backend/controllers"
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLogger 记录请求和响应的日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 【单体应用】每次请求生成新的 RequestID
		// 不需要从 header 获取，因为不需要透传
		requestID := uuid.New().String()
		c.Set(controllers.RequestIDKey, requestID)

		// 返回给客户端（方便问题排查时提供）
		c.Header("X-Request-ID", requestID)

		// 记录开始时间
		start := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 保存原始的 ResponseWriter
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 执行后续处理
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 构建日志字段
		logFields := map[string]interface{}{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"status":     c.Writer.Status(),
			"latency_ms": latency.Milliseconds(),
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}

		// 解析请求体
		if len(requestBody) > 0 {
			var requestData interface{}
			if err := json.Unmarshal(requestBody, &requestData); err == nil {
				logFields["request_body"] = requestData
			} else {
				logFields["request_body_raw"] = string(requestBody)
			}
		}

		// 解析响应体
		if len(blw.body.Bytes()) > 0 {
			var responseData interface{}
			if err := json.Unmarshal(blw.body.Bytes(), &responseData); err == nil {
				logFields["response_body"] = responseData
			} else {
				logFields["response_body_raw"] = blw.body.String()
			}
		}

		// 如果有错误详情，添加到日志
		if errorDetail, exists := controllers.GetErrorDetail(c); exists {
			logFields["error"] = map[string]interface{}{
				"code":    errorDetail.Code,
				"message": errorDetail.Message,
				"detail":  errorDetail.Detail,
			}
		}

		// 记录日志（根据响应 body 中的 code 决定日志级别）
		logJSON, _ := json.Marshal(logFields)

		// 从响应体中获取业务码
		var bizCode int
		if responseData, ok := logFields["response_body"].(map[string]interface{}); ok {
			if code, exists := responseData["code"]; exists {
				bizCode = int(code.(float64))
			}
		}

		switch {
		case bizCode >= 5000:
			// 服务器错误 - Error 级别
			utils.Error("API请求错误: %s", string(logJSON))
		case bizCode >= 4000:
			// 客户端错误 - Warn 级别
			utils.Warn("API请求警告: %s", string(logJSON))
		default:
			// 正常请求 - Info 级别
			utils.Info("API请求: %s", string(logJSON))
		}
	}
}

// bodyLogWriter 用于捕获响应体
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写 Write 方法
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 重写 WriteString 方法
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
