package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"goshopadmin/utils"

	"github.com/gin-gonic/gin"
)

// RequestLogger 记录请求和响应的日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录开始时间
		start := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重置请求体，否则后续处理会读不到
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 保存原始的 ResponseWriter
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// 执行后续处理
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 记录请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		if query != "" {
			path = path + "?" + query
		}

		// 解析请求体（如果是 JSON）
		var requestData interface{}
		if len(requestBody) > 0 {
			json.Unmarshal(requestBody, &requestData)
		}

		// 解析响应体（如果是 JSON）
		var responseData interface{}
		if len(blw.body.Bytes()) > 0 {
			json.Unmarshal(blw.body.Bytes(), &responseData)
		}

		// 格式化请求体为JSON字符串
		requestJSON, _ := json.MarshalIndent(requestData, "", "  ")
		// 格式化响应体为JSON字符串
		responseJSON, _ := json.MarshalIndent(responseData, "", "  ")

		// 记录日志
		utils.Info("API请求:\n"+
			"  方法: %s\n"+
			"  路径: %s\n"+
			"  状态码: %d\n"+
			"  耗时: %v\n"+
			"  请求体:\n%s\n"+
			"  响应体:\n%s",
			method, path, c.Writer.Status(), latency, string(requestJSON), string(responseJSON))
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
