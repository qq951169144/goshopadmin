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
	CodeCartNotFound    = 4043 // 购物车不存在
	CodeOrderNotFound   = 4044 // 订单不存在

	// 4090 - 业务冲突
	CodeConflict       = 4090 // 资源冲突
	CodeDuplicate      = 4091 // 数据重复
	CodeStockInsufficient = 4092 // 库存不足

	// 5000 - 服务器错误 (5xx 服务端错误)
	CodeInternalError = 5000 // 内部错误
	CodeDBError       = 5001 // 数据库错误
	CodeCacheError    = 5002 // 缓存错误
	CodeExternalError = 5003 // 外部服务错误
)

// ErrorMessage 错误码对应的前端友好提示
var ErrorMessage = map[int]string{
	CodeSuccess:           "操作成功",
	CodeParamError:        "请求参数错误",
	CodeParamMissing:      "缺少必要参数",
	CodeParamInvalid:      "参数格式不正确",
	CodeParamOutOfRange:   "参数超出有效范围",
	CodeUnauthorized:      "请先登录",
	CodeTokenExpired:      "登录已过期，请重新登录",
	CodeTokenInvalid:      "登录状态无效",
	CodeLoginFailed:       "用户名或密码错误",
	CodeForbidden:         "没有权限执行此操作",
	CodeResourceDenied:    "无权访问该资源",
	CodeNotFound:          "请求的资源不存在",
	CodeUserNotFound:      "用户不存在",
	CodeProductNotFound:   "商品不存在",
	CodeCartNotFound:      "购物车不存在",
	CodeOrderNotFound:     "订单不存在",
	CodeConflict:          "资源冲突",
	CodeDuplicate:         "数据已存在",
	CodeStockInsufficient: "库存不足",
	CodeInternalError:     "系统繁忙，请稍后重试",
	CodeDBError:           "数据操作失败",
	CodeCacheError:        "缓存服务异常",
	CodeExternalError:     "外部服务调用失败",
}

// GetErrorMessage 获取错误码对应的友好提示
func GetErrorMessage(code int) string {
	if msg, ok := ErrorMessage[code]; ok {
		return msg
	}
	return "未知错误"
}
