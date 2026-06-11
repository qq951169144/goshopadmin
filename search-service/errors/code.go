package errors

// ============================================================
// 错误码定义
// 格式：HTTP状态码(1位) + 模块(1位) + 序号(2位)
// 4xxx 表示客户端错误，5xxx 表示服务端错误
// ============================================================

const (
	// CodeSuccess 成功
	CodeSuccess = 0

	// 4000 - 参数错误 (4xx 客户端错误)
	// CodeParamError 通用参数错误
	CodeParamError = 4001
	// CodeParamMissing 参数缺失
	CodeParamMissing = 4002
	// CodeParamInvalid 参数格式无效
	CodeParamInvalid = 4003
	// CodeParamOutOfRange 参数超出范围
	CodeParamOutOfRange = 4004

	// 4010 - 认证错误
	// CodeUnauthorized 未认证
	CodeUnauthorized = 4010
	// CodeTokenExpired Token 过期
	CodeTokenExpired = 4011
	// CodeTokenInvalid Token 无效
	CodeTokenInvalid = 4012

	// 4040 - 资源错误
	// CodeNotFound 资源不存在
	CodeNotFound = 4040

	// 4080 - 搜索相关错误
	// CodeSearchError 通用搜索错误
	CodeSearchError = 4080
	// CodeSearchTimeout 搜索请求超时
	CodeSearchTimeout = 4081
	// CodeSearchRateLimited 搜索请求被限流
	CodeSearchRateLimited = 4082
	// CodeESUnavailable Elasticsearch 服务不可用
	CodeESUnavailable = 4083

	// CodeSyncInProgress 全量同步正在进行中
	CodeSyncInProgress = 4090

	// 5000 - 服务器错误 (5xx 服务端错误)
	// CodeInternalError 内部错误
	CodeInternalError = 5000
	// CodeDBError 数据库错误
	CodeDBError = 5001
	// CodeESError Elasticsearch 错误
	CodeESError = 5002
)

// ErrorMessage 错误码对应的友好提示信息映射表
// 用于将数字错误码转换为用户可读的中文提示
var ErrorMessage = map[int]string{
	CodeSuccess:           "成功",
	CodeParamError:        "参数错误",
	CodeParamMissing:      "参数缺失",
	CodeParamInvalid:      "参数格式无效",
	CodeParamOutOfRange:   "参数超出范围",
	CodeUnauthorized:      "未认证",
	CodeTokenExpired:      "Token 过期",
	CodeTokenInvalid:      "Token 无效",
	CodeNotFound:          "资源不存在",
	CodeSearchError:       "搜索服务错误",
	CodeSearchTimeout:     "搜索请求超时",
	CodeSearchRateLimited: "搜索请求过于频繁",
	CodeESUnavailable:     "搜索服务暂不可用",
	CodeSyncInProgress:     "全量同步正在进行中",
	CodeInternalError:     "内部错误",
	CodeDBError:           "数据库错误",
	CodeESError:           "搜索引擎错误",
}

// GetErrorMessage 根据错误码获取对应的友好提示信息
// 如果错误码不在映射表中，返回 "未知错误"
func GetErrorMessage(code int) string {
	msg, ok := ErrorMessage[code]
	if !ok {
		return "未知错误"
	}
	return msg
}
