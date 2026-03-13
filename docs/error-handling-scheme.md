# 后端错误处理与日志记录方案（单体应用版）

## 目标

1. **前端友好**：返回标准化的错误码和友好的错误信息
2. **日志完整**：记录详细的错误信息用于排查问题
3. **不侵入业务**：Controller 层保持简洁，错误处理逻辑下沉
4. **简单易用**：单体应用无需复杂的 request_id 传递机制

---

## 整体架构

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              客户端（前端）                                    │
│                         发起请求（无需携带 RequestID）                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              API 服务（单体应用）                               │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Logger Middleware（最先执行）                                       │   │
│  │  1. 生成 RequestID（uuid）                                          │   │
│  │  2. 存入 context                                                    │   │
│  │  3. 设置响应头 X-Request-ID（返回给前端，用于问题排查）                │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                      │
│                                      ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Controller 层                                                      │   │
│  │  c.ResponseError(ctx, code.ParamInvalid, err)                       │   │
│  │  // 只传递错误码和原始错误                                           │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                      │
│                                      ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  BaseController                                                     │   │
│  │  1. 从 ctx 获取 RequestID（由 Logger 中间件生成）                    │   │
│  │  2. 将错误信息记录到 ctx（供中间件使用）                              │   │
│  │  3. 根据错误码获取友好提示信息                                        │   │
│  │  4. 返回标准化响应给前端                                              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                      │                                      │
│                                      ▼                                      │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  Logger Middleware（最后执行）                                       │   │
│  │  1. 从 ctx 读取 RequestID 和错误详情                                  │   │
│  │  2. 记录完整请求/响应日志                                             │   │
│  │  3. 包含原始错误堆栈等调试信息                                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              客户端（前端）                                    │
│  1. 获取响应头 X-Request-ID（错误时可用于排查）                              │
│  2. 下一次请求**不需要**携带 RequestID（每次请求独立生成）                    │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## RequestID 使用说明

### 单体应用 vs 微服务

| 场景 | RequestID 行为 | 前端是否需要传递 |
|------|----------------|------------------|
| **单体应用** | 每次请求生成新的 RequestID | ❌ 不需要 |
| **微服务** | 服务间透传同一个 RequestID | ✅ 需要 |

### 为什么单体应用不需要传递？

```
【正确做法 - 单体应用】
请求1: 前端 ──→ API  生成 RequestID: abc-111  ──→ 返回给前端
请求2: 前端 ──→ API  生成 RequestID: abc-222  ──→ 返回给前端  
请求3: 前端 ──→ API  生成 RequestID: abc-333  ──→ 返回给前端

每次请求独立，日志中通过不同的 RequestID 区分

【错误做法 - 不要这样做】
请求1: 前端 ──→ API  生成 RequestID: abc-111  ──→ 返回给前端
请求2: 前端 ──→ API  复用 RequestID: abc-111  ──→ 返回给前端

这样会导致多个请求的日志混在一起，无法区分
```

### RequestID 的作用

1. **问题排查**：用户反馈问题时，提供 RequestID 给开发快速定位
2. **日志关联**：一次请求的所有日志都有相同的 RequestID
3. **性能分析**：通过 RequestID 查看请求耗时

---

## 1. 错误码定义 (errors/code.go)

```go
package errors

// 错误码定义
// 格式：HTTP状态码(1位) + 模块(1位) + 序号(2位)
const (
    // 成功
    CodeSuccess = 0

    // 4000 - 参数错误 (4xx 客户端错误)
    CodeParamError      = 4001 // 通用参数错误
    CodeParamMissing    = 4002 // 参数缺失
    CodeParamInvalid    = 4003 // 参数格式无效
    CodeParamOutOfRange = 4004 // 参数超出范围

    // 4010 - 认证错误
    CodeUnauthorized = 4010 // 未认证
    CodeTokenExpired = 4011 // Token 过期
    CodeTokenInvalid = 4012 // Token 无效
    CodeLoginFailed  = 4013 // 登录失败

    // 4030 - 权限错误
    CodeForbidden      = 4030 // 权限不足
    CodeResourceDenied = 4031 // 资源访问被拒绝

    // 4040 - 资源错误
    CodeNotFound        = 4040 // 资源不存在
    CodeUserNotFound    = 4041 // 用户不存在
    CodeProductNotFound = 4042 // 商品不存在

    // 4090 - 业务冲突
    CodeConflict  = 4090 // 资源冲突
    CodeDuplicate = 4091 // 数据重复

    // 5000 - 服务器错误 (5xx 服务端错误)
    CodeInternalError = 5000 // 内部错误
    CodeDBError       = 5001 // 数据库错误
    CodeCacheError    = 5002 // 缓存错误
    CodeExternalError = 5003 // 外部服务错误
)

// ErrorMessage 错误码对应的前端友好提示
var ErrorMessage = map[int]string{
    CodeSuccess:         "操作成功",
    CodeParamError:      "请求参数错误",
    CodeParamMissing:    "缺少必要参数",
    CodeParamInvalid:    "参数格式不正确",
    CodeParamOutOfRange: "参数超出有效范围",
    CodeUnauthorized:    "请先登录",
    CodeTokenExpired:    "登录已过期，请重新登录",
    CodeTokenInvalid:    "登录状态无效",
    CodeLoginFailed:     "用户名或密码错误",
    CodeForbidden:       "没有权限执行此操作",
    CodeResourceDenied:  "无权访问该资源",
    CodeNotFound:        "请求的资源不存在",
    CodeUserNotFound:    "用户不存在",
    CodeProductNotFound: "商品不存在",
    CodeConflict:        "资源冲突",
    CodeDuplicate:       "数据已存在",
    CodeInternalError:   "系统繁忙，请稍后重试",
    CodeDBError:         "数据操作失败",
    CodeCacheError:      "缓存服务异常",
    CodeExternalError:   "外部服务调用失败",
}

// GetErrorMessage 获取错误码对应的友好提示
func GetErrorMessage(code int) string {
    if msg, ok := ErrorMessage[code]; ok {
        return msg
    }
    return "未知错误"
}
```

---

## 2. 响应结构定义 (controllers/response.go)

```go
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
```

---

## 3. BaseController 改造 (controllers/base_controller.go)

```go
package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "goshopadmin/errors"
)

// BaseController 基础控制器
type BaseController struct{}

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
```

---

## 4. Logger Middleware 改造 (middleware/logger.go)

```go
package middleware

import (
    "bytes"
    "encoding/json"
    "io"
    "time"

    "goshopadmin/controllers"
    "goshopadmin/utils"

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
```

---

## 5. Controller 使用示例

### 改造前（现状）

```go
func (c *ProductController) CreateProduct(ctx *gin.Context) {
    var req CreateProductRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        utils.Info("参数错误: %v", err) // 在 controller 里打日志
        ctx.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的请求数据"})
        return
    }

    if err := c.productService.CreateProduct(&product); err != nil {
        utils.Info("创建商品失败: %v", err) // 在 controller 里打日志
        ctx.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建商品失败"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建商品成功", "data": product})
}
```

### 改造后

```go
func (c *ProductController) CreateProduct(ctx *gin.Context) {
    var req CreateProductRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        // 不需要打日志，中间件会记录
        // 传递原始错误 err，前端只会看到友好提示
        c.ResponseError(ctx, errors.CodeParamInvalid, err)
        return
    }

    if err := c.productService.CreateProduct(&product); err != nil {
        // 根据错误类型返回不同的错误码
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.ResponseError(ctx, errors.CodeProductNotFound, err)
        } else if errors.Is(err, errors.ErrDuplicate) {
            c.ResponseError(ctx, errors.CodeDuplicate, err)
        } else {
            c.ResponseError(ctx, errors.CodeDBError, err)
        }
        return
    }

    c.ResponseSuccess(ctx, product)
}
```

---

## 6. 前端完整改造方案

### 6.1 安装依赖

```bash
npm install axios
```

### 6.2 创建请求工具 (src/utils/request.js)

```javascript
import axios from 'axios';
import { Message, MessageBox } from 'element-ui'; // 或你使用的 UI 库
import router from '@/router';
import store from '@/store';

// 创建 axios 实例
const service = axios.create({
    baseURL: process.env.VUE_APP_BASE_API, // 如：'/api'
    timeout: 15000,
    headers: {
        'Content-Type': 'application/json'
    }
});

// 请求队列（用于取消重复请求）
const pendingRequests = new Map();

// 生成请求 key
const getRequestKey = (config) => {
    return `${config.method}&${config.url}&${JSON.stringify(config.params)}&${JSON.stringify(config.data)}`;
};

// 添加请求到队列
const addPendingRequest = (config) => {
    const requestKey = getRequestKey(config);
    if (pendingRequests.has(requestKey)) {
        config.cancelToken = new axios.CancelToken((cancel) => {
            cancel('重复的请求');
        });
    } else {
        config.cancelToken = new axios.CancelToken((cancel) => {
            pendingRequests.set(requestKey, cancel);
        });
    }
};

// 移除请求从队列
const removePendingRequest = (config) => {
    const requestKey = getRequestKey(config);
    if (pendingRequests.has(requestKey)) {
        const cancel = pendingRequests.get(requestKey);
        cancel(requestKey);
        pendingRequests.delete(requestKey);
    }
};

// 请求拦截器
service.interceptors.request.use(
    config => {
        // 1. 移除重复请求
        removePendingRequest(config);
        addPendingRequest(config);

        // 2. 添加 token
        const token = store.getters.token || localStorage.getItem('token');
        if (token) {
            config.headers['Authorization'] = 'Bearer ' + token;
        }

        // 【单体应用】不需要携带 X-Request-ID
        // 每次请求由服务端生成新的 RequestID

        // 3. 添加时间戳（防止缓存）
        if (config.method === 'get') {
            config.params = {
                ...config.params,
                _t: Date.now()
            };
        }

        return config;
    },
    error => {
        console.error('Request Error:', error);
        return Promise.reject(error);
    }
);

// 响应拦截器
service.interceptors.response.use(
    response => {
        // 1. 移除已完成的请求
        removePendingRequest(response.config);

        // 2. 获取后端返回的数据
        const res = response.data;

        // 3. 【单体应用】获取 RequestID（用于问题排查，可选）
        const requestID = response.headers['x-request-id'];
        // 可以存储在组件中，错误时显示给用户
        response.requestID = requestID;

        // 4. 判断业务状态码
        if (res.code === 0 || res.code === 200) {
            // 成功，直接返回数据
            return res.data;
        }

        // 5. 业务错误处理
        handleBusinessError(res, requestID);

        // 6. 返回拒绝的 Promise
        return Promise.reject(new Error(res.message || '操作失败'));
    },
    error => {
        // 1. 移除失败的请求
        if (error.config) {
            removePendingRequest(error.config);
        }

        // 2. 处理 HTTP 错误（网络层）
        handleHTTPError(error);

        return Promise.reject(error);
    }
);

// 处理业务错误
function handleBusinessError(res, requestID) {
    const { code, message } = res;

    // 根据错误码分类处理
    switch (true) {
        // 认证错误（4010-4019）
        case code >= 4010 && code < 4020:
            handleAuthError(code, message);
            break;

        // 权限错误（4030-4039）
        case code >= 4030 && code < 4040:
            handlePermissionError(code, message);
            break;

        // 资源不存在（4040-4049）
        case code >= 4040 && code < 4050:
            handleNotFoundError(code, message);
            break;

        // 参数错误（4000-4099）
        case code >= 4000 && code < 4100:
            Message.warning(message || '请求参数错误');
            break;

        // 服务器错误（5000+）
        case code >= 5000:
            // 【单体应用】错误时显示 RequestID，方便用户反馈问题
            const errorMsg = requestID 
                ? `${message || '系统繁忙，请稍后重试'} (请求ID: ${requestID})`
                : message || '系统繁忙，请稍后重试';
            
            Message.error({
                message: errorMsg,
                duration: 5000
            });
            
            // 记录错误信息到控制台，方便开发排查
            console.error('Server Error:', {
                code,
                message,
                requestID,
                time: new Date().toISOString()
            });
            break;

        // 其他错误
        default:
            Message.error(message || '操作失败');
    }
}

// 处理认证错误
function handleAuthError(code, message) {
    const errorMap = {
        4010: '请先登录',
        4011: '登录已过期，请重新登录',
        4012: '登录状态无效',
        4013: '用户名或密码错误'
    };

    const msg = message || errorMap[code] || '认证失败';

    // 避免重复弹窗
    if (!store.getters.isLoginDialogShow) {
        store.commit('SET_LOGIN_DIALOG', true);

        MessageBox.confirm(msg, '提示', {
            confirmButtonText: '重新登录',
            cancelButtonText: '取消',
            type: 'warning'
        }).then(() => {
            store.dispatch('user/logout').then(() => {
                router.push('/login');
            });
        }).catch(() => {
            store.commit('SET_LOGIN_DIALOG', false);
        });
    }
}

// 处理权限错误
function handlePermissionError(code, message) {
    Message.error(message || '没有权限执行此操作');

    // 可以跳转到 403 页面
    // router.push('/403');
}

// 处理资源不存在
function handleNotFoundError(code, message) {
    Message.warning(message || '请求的资源不存在');

    // 可以跳转到 404 页面
    // router.push('/404');
}

// 处理 HTTP 错误（网络层）
function handleHTTPError(error) {
    if (axios.isCancel(error)) {
        console.log('Request canceled:', error.message);
        return;
    }

    const { response, request, message } = error;

    if (response) {
        // 服务器返回了错误状态码
        const status = response.status;
        const statusMap = {
            400: '请求参数错误',
            401: '未授权，请重新登录',
            403: '拒绝访问',
            404: '请求的资源不存在',
            408: '请求超时',
            500: '服务器内部错误',
            502: '网关错误',
            503: '服务不可用',
            504: '网关超时'
        };

        Message.error({
            message: statusMap[status] || `服务器错误: ${status}`,
            duration: 5000
        });

        // 401 未授权，跳转到登录页
        if (status === 401) {
            store.dispatch('user/logout').then(() => {
                router.push('/login');
            });
        }
    } else if (request) {
        // 请求发出但没有收到响应
        Message.error('网络连接失败，请检查网络设置');
    } else {
        // 请求配置出错
        Message.error('请求配置错误: ' + message);
    }
}

// 导出请求方法
export default service;

// 导出常用请求方法
export const get = (url, params) => service.get(url, { params });
export const post = (url, data) => service.post(url, data);
export const put = (url, data) => service.put(url, data);
export const del = (url) => service.delete(url);
```

### 6.3 API 封装示例 (src/api/product.js)

```javascript
import request from '@/utils/request';

// 商品列表
export function getProductList(params) {
    return request({
        url: '/api/products',
        method: 'get',
        params
    });
}

// 商品详情
export function getProductDetail(id) {
    return request({
        url: `/api/products/${id}`,
        method: 'get'
    });
}

// 创建商品
export function createProduct(data) {
    return request({
        url: '/api/products',
        method: 'post',
        data
    });
}

// 更新商品
export function updateProduct(id, data) {
    return request({
        url: `/api/products/${id}`,
        method: 'put',
        data
    });
}

// 删除商品
export function deleteProduct(id) {
    return request({
        url: `/api/products/${id}`,
        method: 'delete'
    });
}
```

### 6.4 Vue 组件中使用示例

```vue
<template>
  <div class="product-list">
    <el-table :data="products" v-loading="loading">
      <el-table-column prop="name" label="商品名称" />
      <el-table-column prop="price" label="价格" />
      <el-table-column label="操作">
        <template slot-scope="scope">
          <el-button @click="handleEdit(scope.row)">编辑</el-button>
          <el-button type="danger" @click="handleDelete(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      @current-change="handlePageChange"
      :current-page="page"
      :page-size="limit"
      :total="total"
    />
  </div>
</template>

<script>
import { getProductList, deleteProduct } from '@/api/product';

export default {
  data() {
    return {
      products: [],
      loading: false,
      page: 1,
      limit: 10,
      total: 0
    };
  },

  created() {
    this.fetchProducts();
  },

  methods: {
    // 获取商品列表
    async fetchProducts() {
      this.loading = true;
      try {
        const res = await getProductList({
          page: this.page,
          limit: this.limit
        });
        this.products = res.list;
        this.total = res.total;
      } catch (error) {
        // 错误已被拦截器处理，这里可以执行额外的清理
        console.log('获取列表失败:', error.message);
      } finally {
        this.loading = false;
      }
    },

    // 删除商品
    async handleDelete(row) {
      try {
        await this.$confirm('确认删除该商品？', '提示', {
          type: 'warning'
        });

        await deleteProduct(row.id);
        this.$message.success('删除成功');
        this.fetchProducts(); // 刷新列表
      } catch (error) {
        // 用户取消或删除失败
        if (error !== 'cancel') {
          console.log('删除失败:', error.message);
        }
      }
    },

    handlePageChange(page) {
      this.page = page;
      this.fetchProducts();
    }
  }
};
</script>
```

### 6.5 【可选】在组件中显示 RequestID

```vue
<template>
  <div>
    <!-- 页面内容 -->
    
    <!-- 错误提示区域 -->
    <div v-if="lastError" class="error-info">
      <p>操作失败，请截图联系客服</p>
      <p>错误码: {{ lastError.code }}</p>
      <p>请求ID: {{ lastError.requestID }}</p>
      <el-button @click="copyErrorInfo">复制错误信息</el-button>
    </div>
  </div>
</template>

<script>
import { createProduct } from '@/api/product';

export default {
  data() {
    return {
      lastError: null
    };
  },

  methods: {
    async handleSubmit() {
      try {
        this.lastError = null;
        await createProduct(this.form);
        this.$message.success('创建成功');
      } catch (error) {
        // 保存错误信息，用于显示 RequestID
        this.lastError = {
          code: error.code,
          requestID: error.requestID,
          message: error.message
        };
      }
    },

    copyErrorInfo() {
      const text = `错误码: ${this.lastError.code}\n请求ID: ${this.lastError.requestID}\n时间: ${new Date().toLocaleString()}`;
      navigator.clipboard.writeText(text);
      this.$message.success('已复制到剪贴板');
    }
  }
};
</script>
```

---

## 7. 日志输出示例

### 正常请求

```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "GET",
  "path": "/api/products",
  "status": 200,
  "latency_ms": 45,
  "client_ip": "192.168.1.100",
  "request_body": null,
  "response_body": {
    "code": 0,
    "message": "success",
    "data": [...]
  }
}
```

### 参数错误（HTTP 200，body 中 code 4003）

```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440001",
  "method": "POST",
  "path": "/api/products",
  "status": 200,
  "latency_ms": 12,
  "client_ip": "192.168.1.100",
  "request_body": {
    "name": "",
    "price": -100
  },
  "response_body": {
    "code": 4003,
    "message": "参数格式不正确"
  },
  "error": {
    "code": 4003,
    "message": "参数格式不正确",
    "detail": "Key: 'CreateProductRequest.Price' Error:Field validation for 'Price' failed on the 'gte' tag"
  }
}
```

### 数据库错误（HTTP 200，body 中 code 5001）

```json
{
  "request_id": "550e8400-e29b-41d4-a716-446655440002",
  "method": "POST",
  "path": "/api/products",
  "status": 200,
  "latency_ms": 2345,
  "client_ip": "192.168.1.100",
  "request_body": {...},
  "response_body": {
    "code": 5001,
    "message": "数据操作失败"
  },
  "error": {
    "code": 5001,
    "message": "数据操作失败",
    "detail": "Error 1062: Duplicate entry 'test-product' for key 'idx_name'"
  }
}
```

---

## 8. 问题排查示例

### 用户反馈问题

```
用户：我提交订单失败了，显示"系统繁忙，请稍后重试 (请求ID: abc-123-xyz)"
```

### 开发排查

```bash
# 通过 request_id 查询日志
grep "abc-123-xyz" app.log

# 输出：
{"request_id":"abc-123-xyz","method":"POST","path":"/api/orders","error":{"code":5001,"message":"数据操作失败","detail":"Error 1062: Duplicate entry 'ORD-001' for key 'order_no'"}}

# 问题定位：订单号重复
```

---

## 9. 迁移步骤

### 后端迁移

1. **创建错误码文件** `errors/code.go`
2. **创建响应结构文件** `controllers/response.go`
3. **改造 BaseController** `controllers/base_controller.go`
4. **改造 Logger Middleware** `middleware/logger.go`
5. **更新 main.go 中间件注册：**

```go
func main() {
    r := gin.New()

    // 1. 注册 Logger（包含 RequestID 生成）
    r.Use(middleware.RequestLogger())

    // 2. 其他中间件
    r.Use(middleware.CORS())
    r.Use(gin.Recovery())

    // 路由...
}
```

6. **逐步替换 Controller** 中的 `ctx.JSON` 调用

### 前端迁移

1. **备份现有请求工具**
2. **替换 request.js** 为新版本
3. **更新 API 封装文件**，使用新的请求方法
4. **测试各页面功能**
5. **【可选】添加错误信息显示组件**

---

## 10. 优势总结

| 方面 | 改造前 | 改造后 |
|------|--------|--------|
| **Controller 代码** | 冗长，包含日志和响应组装 | 简洁，只关注业务逻辑 |
| **错误信息** | 前后端一致，可能暴露敏感信息 | 前端友好，后端详细 |
| **日志记录** | 分散在 Controller 各处 | 统一在中间件，结构完整 |
| **错误码管理** | 无规范，随意定义 | 统一规范，便于维护 |
| **前端体验** | 需要处理各种错误格式 | 统一响应格式，易于处理 |
| **问题排查** | 需要查看多处日志 | 通过 RequestID 快速定位 |
| **HTTP 状态码** | 混乱，有时用有时不用 | 始终 200，业务状态由 body.code 表示 |
| **单体应用复杂度** | - | 简单，无需传递 RequestID |
| **前端开发效率** | 需处理各种边界情况 | 统一封装，开箱即用 |
