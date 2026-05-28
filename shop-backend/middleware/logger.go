package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"strings"
	"time"

	"shop-backend/controllers"
	"shop-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// maxBodySize 定义请求体和响应体的最大日志记录大小（100KB）
// 超过此大小的内容将被截断，避免日志文件过大
const (
	maxBodySize = 1024 * 100
)

// skipContentTypes 定义需要跳过日志记录的Content-Type类型列表
// 这些类型通常是二进制文件或大文件，记录它们会消耗大量资源
var skipContentTypes = []string{
	"image/",                       // 图片类型（JPEG、PNG、GIF等）
	"application/octet-stream",     // 二进制流
	"application/pdf",              // PDF文件
	"application/zip",              // ZIP压缩文件
	"application/x-rar-compressed", // RAR压缩文件
	"application/x-7z-compressed",  // 7Z压缩文件
	"audio/",                       // 音频类型（MP3、WAV等）
	"video/",                       // 视频类型（MP4、AVI等）
	"image/svg+xml",                // SVG图片
}

// skipPaths 定义需要跳过日志记录的路径列表
// 这些路径通常返回大文件或敏感数据
var skipPaths = []string{
	"/api/captcha",  // 验证码接口
	"/api/upload",   // 文件上传接口
	"/api/download", // 文件下载接口
	"/api/image",    // 图片相关接口
	"/api/export",   // 数据导出接口
}

// shouldSkipLogging 判断是否应该跳过日志记录
// 参数:
//
//	contentType - HTTP请求或响应的Content-Type头部值
//
// 返回值:
//
//	true - 应该跳过日志记录（二进制文件等）
//	false - 需要记录日志
func shouldSkipLogging(contentType string) bool {
	if contentType == "" {
		return false
	}
	// 使用mime.ParseMediaType解析Content-Type，提取主要类型
	mediaType, _, _ := mime.ParseMediaType(contentType)
	for _, skipType := range skipContentTypes {
		if strings.HasPrefix(mediaType, skipType) {
			return true
		}
	}
	return false
}

// shouldSkipPath 判断请求路径是否应该跳过日志记录
// 参数:
//
//	path - 请求路径
//
// 返回值:
//
//	true - 应该跳过日志记录
//	false - 需要记录日志
func shouldSkipPath(path string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// isBase64Content 判断内容是否包含大量base64编码数据
// base64编码通常用于传输图片等二进制数据，会大幅增加日志大小
// 参数:
//
//	content - 待检测的内容
//
// 返回值:
//
//	true - 包含大量base64数据
//	false - 不包含或仅包含少量base64数据
func isBase64Content(content []byte) bool {
	if len(content) < 100 {
		return false
	}

	contentStr := string(content)

	// 检查是否有base64数据URL格式（如 data:image/png;base64,xxx）
	if strings.Contains(contentStr, "data:image/") && strings.Contains(contentStr, ";base64,") {
		return true
	}

	// 检查纯base64编码内容（不含data URL前缀）
	// base64编码特点：只包含A-Z, a-z, 0-9, +, /, =
	// 计算疑似base64的比例
	base64Chars := 0
	totalChars := 0
	for _, c := range contentStr {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=' {
			base64Chars++
		}
		totalChars++
	}

	// 如果超过80%的字符是base64字符，且内容长度超过一定阈值，视为base64数据
	if totalChars > 0 && float64(base64Chars)/float64(totalChars) > 0.8 && len(content) > 1024 {
		return true
	}

	return false
}

// truncateBody 截断过长的body内容
// 参数:
//
//	body - 原始body内容
//	maxSize - 最大允许大小
//
// 返回值:
//
//	截断后的body内容，末尾添加"...(truncated)"标记
func truncateBody(body []byte, maxSize int) []byte {
	if len(body) <= maxSize {
		return body
	}
	return append(body[:maxSize], []byte("...(truncated)")...)
}

// RequestLogger 记录请求和响应的日志中间件
// 功能：
// 1. 为每个请求生成唯一的RequestID，用于链路追踪
// 2. 记录请求的基本信息（方法、路径、查询参数、状态码等）
// 3. 根据Content-Type判断是否记录请求体和响应体
// 4. 对大文件请求/响应进行跳过或截断处理
// 5. 根据业务错误码决定日志级别
// 过滤策略（按优先级）：
// 1. 路径过滤：跳过指定路径（如验证码、上传、下载接口）
// 2. Content-Type过滤：跳过二进制文件类型（图片、音频、视频、PDF等）
// 3. Base64检测：检测并跳过包含大量base64编码的数据
// 4. 大小限制：超过100KB的内容自动截断
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// ========== 阶段1：初始化请求上下文 ==========
		// 生成唯一RequestID，用于请求追踪和日志关联
		requestID := uuid.New().String()
		// 将RequestID存入context，供后续中间件和控制器使用
		c.Set(controllers.RequestIDKey, requestID)
		// 将RequestID添加到响应头，便于客户端排查问题
		c.Header("X-Request-ID", requestID)

		// 记录请求开始时间，用于计算请求耗时
		start := time.Now()

		// ========== 阶段2：检查是否需要跳过日志记录 ==========
		// 策略1：路径过滤 - 跳过验证码、上传下载等接口
		isPathSkipped := shouldSkipPath(c.Request.URL.Path)

		// 获取请求的Content-Type，判断是否为需要跳过的类型
		contentType := c.Request.Header.Get("Content-Type")
		// 策略2：Content-Type过滤 - 跳过二进制文件类型
		isBinaryRequest := shouldSkipLogging(contentType)

		var requestBody []byte
		var isRequestBase64 bool
		// 如果不是跳过路径且不是二进制请求，才读取请求体
		if !isPathSkipped && !isBinaryRequest && c.Request.Body != nil {
			// 读取完整请求体
			body, _ := io.ReadAll(c.Request.Body)
			// 策略3：Base64检测 - 检测是否包含大量base64数据
			isRequestBase64 = isBase64Content(body)
			// 如果不是base64内容，才进行截断处理
			if !isRequestBase64 {
				requestBody = truncateBody(body, maxBodySize)
			}
			// 将原始请求体写回，供后续处理使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		// ========== 阶段3：准备响应体捕获 ==========
		isBinaryResponse := false
		var blw *bodyLogWriter
		// 如果不是跳过路径且不是二进制请求，才创建响应体捕获器
		if !isPathSkipped && !isBinaryRequest {
			blw = &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw
		}

		// ========== 阶段4：执行后续中间件和控制器 ==========
		c.Next()

		// ========== 阶段5：处理响应体 ==========
		var isResponseBase64 bool
		if !isPathSkipped && !isBinaryRequest {
			responseContentType := c.Writer.Header().Get("Content-Type")
			// 策略2：Content-Type过滤 - 检查响应是否为二进制类型
			isBinaryResponse = shouldSkipLogging(responseContentType)

			// 如果不是二进制响应，检查是否包含base64数据
			if !isBinaryResponse && blw != nil && len(blw.body.Bytes()) > 0 {
				isResponseBase64 = isBase64Content(blw.body.Bytes())
			}
		}

		// ========== 阶段6：构建日志字段 ==========
		// 计算请求耗时（毫秒）
		latency := time.Since(start)

		// 构建基础日志字段（始终记录，不受过滤影响）
		logFields := map[string]interface{}{
			"request_id": requestID,              // 请求唯一标识
			"method":     c.Request.Method,       // HTTP方法（GET/POST/PUT/DELETE等）
			"path":       c.Request.URL.Path,     // 请求路径
			"query":      c.Request.URL.RawQuery, // 查询参数
			"status":     c.Writer.Status(),      // HTTP状态码
			"latency_ms": latency.Milliseconds(), // 请求耗时（毫秒）
			"client_ip":  c.ClientIP(),           // 客户端IP
			"user_agent": c.Request.UserAgent(),  // 用户代理（浏览器/客户端信息）
		}

		// 记录跳过原因（便于排查问题）
		if isPathSkipped {
			logFields["skip_reason"] = "path_filter"
		} else if isBinaryRequest {
			logFields["skip_reason"] = "binary_request"
		} else if isBinaryResponse {
			logFields["skip_reason"] = "binary_response"
		} else if isRequestBase64 {
			logFields["skip_reason"] = "base64_request"
		} else if isResponseBase64 {
			logFields["skip_reason"] = "base64_response"
		}

		// 添加请求体到日志（仅当没有被过滤时）
		if !isPathSkipped && !isBinaryRequest && !isRequestBase64 && len(requestBody) > 0 {
			var requestData interface{}
			// 尝试解析为JSON格式
			if err := json.Unmarshal(requestBody, &requestData); err == nil {
				logFields["request_body"] = requestData
			} else {
				// 非JSON格式，直接存储原始字符串
				logFields["request_body_raw"] = string(requestBody)
			}
		}

		// 添加响应体到日志（仅当没有被过滤时）
		if !isPathSkipped && !isBinaryResponse && !isResponseBase64 && blw != nil && len(blw.body.Bytes()) > 0 {
			// 策略4：大小限制 - 截断超过最大大小的内容
			truncatedBody := truncateBody(blw.body.Bytes(), maxBodySize)
			var responseData interface{}
			// 尝试解析为JSON格式
			if err := json.Unmarshal(truncatedBody, &responseData); err == nil {
				logFields["response_body"] = responseData
			} else {
				// 非JSON格式，直接存储原始字符串
				logFields["response_body_raw"] = string(truncatedBody)
			}
		}

		// ========== 阶段7：处理错误信息 ==========
		// 如果控制器设置了错误详情，添加到日志（错误信息始终记录）
		if errorDetail, exists := controllers.GetErrorDetail(c); exists {
			logFields["error"] = map[string]interface{}{
				"code":    errorDetail.Code,    // 业务错误码
				"message": errorDetail.Message, // 错误消息
				"detail":  errorDetail.Detail,  // 错误详情
			}
		}

		// ========== 阶段8：记录日志 ==========
		// 将日志字段序列化为JSON字符串
		logJSON, _ := json.Marshal(logFields)

		// 从响应体中获取业务码，用于决定日志级别
		var bizCode int
		if responseData, ok := logFields["response_body"].(map[string]interface{}); ok {
			if code, exists := responseData["code"]; exists {
				bizCode = int(code.(float64))
			}
		}

		// 根据业务码决定日志级别
		switch {
		case bizCode >= 5000:
			// 服务器错误 - 使用Error级别
			utils.Error("API请求错误: %s", string(logJSON))
		case bizCode >= 4000:
			// 客户端错误 - 使用Warn级别
			utils.Warn("API请求警告: %s", string(logJSON))
		default:
			// 正常请求 - 使用Info级别
			utils.Info("API请求: %s", string(logJSON))
		}
	}
}

// bodyLogWriter 自定义ResponseWriter，用于捕获响应体
// 嵌入gin.ResponseWriter保持原有功能，同时使用bytes.Buffer记录响应内容
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法，同时将内容写入缓冲区和原始ResponseWriter
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 重写WriteString方法，同时将内容写入缓冲区和原始ResponseWriter
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
