import axios from 'axios'
import { ElMessage, ElMessageBox } from 'element-plus'
import router from '../router'

const API_BASE_URL = '/api'

// 创建 axios 实例
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求队列（用于取消重复请求）
const pendingRequests = new Map()

// 生成请求 key
const getRequestKey = (config) => {
  return `${config.method}&${config.url}&${JSON.stringify(config.params)}&${JSON.stringify(config.data)}`
}

// 添加请求到队列
const addPendingRequest = (config) => {
  const requestKey = getRequestKey(config)
  if (pendingRequests.has(requestKey)) {
    config.cancelToken = new axios.CancelToken((cancel) => {
      cancel('重复的请求')
    })
  } else {
    config.cancelToken = new axios.CancelToken((cancel) => {
      pendingRequests.set(requestKey, cancel)
    })
  }
}

// 移除请求从队列
const removePendingRequest = (config) => {
  const requestKey = getRequestKey(config)
  if (pendingRequests.has(requestKey)) {
    const cancel = pendingRequests.get(requestKey)
    cancel(requestKey)
    pendingRequests.delete(requestKey)
  }
}

// 请求拦截器
api.interceptors.request.use(
  config => {
    // 1. 移除重复请求
    removePendingRequest(config)
    addPendingRequest(config)

    // 2. 添加 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // 3. 添加时间戳（防止缓存）
    if (config.method === 'get') {
      config.params = {
        ...config.params,
        _t: Date.now()
      }
    }

    return config
  },
  error => {
    console.error('Request Error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    // 1. 移除已完成的请求
    removePendingRequest(response.config)

    // 2. 如果是 blob 响应类型（如验证码图片），直接返回完整响应
    if (response.config.responseType === 'blob') {
      return response
    }

    // 3. 获取后端返回的数据
    const res = response.data

    // 4. 【单体应用】获取 RequestID（用于问题排查，可选）
    const requestID = response.headers['x-request-id']
    response.requestID = requestID

    // 5. 判断业务状态码
    if (res.code === 0 || res.code === 200) {
      // 成功，直接返回数据
      return res.data
    }

    // 6. 业务错误处理
    handleBusinessError(res, requestID)

    // 7. 返回拒绝的 Promise
    return Promise.reject(new Error(res.message || '操作失败'))
  },
  error => {
    // 1. 移除失败的请求
    if (error.config) {
      removePendingRequest(error.config)
    }

    // 2. 处理 HTTP 错误（网络层）
    handleHTTPError(error)

    return Promise.reject(error)
  }
)

// 处理业务错误
function handleBusinessError(res, requestID) {
  const { code, message } = res

  // 根据错误码分类处理
  switch (true) {
    // 认证错误（4010-4019）
    case code >= 4010 && code < 4020:
      handleAuthError(code, message)
      break

    // 权限错误（4030-4039）
    case code >= 4030 && code < 4040:
      handlePermissionError(code, message)
      break

    // 资源不存在（4040-4049）
    case code >= 4040 && code < 4050:
      handleNotFoundError(code, message)
      break

    // 参数错误（4000-4099）
    case code >= 4000 && code < 4100:
      ElMessage.warning(message || '请求参数错误')
      break

    // 服务器错误（5000+）
    case code >= 5000:
      // 【单体应用】错误时显示 RequestID，方便用户反馈问题
      const errorMsg = requestID
        ? `${message || '系统繁忙，请稍后重试'} (请求ID: ${requestID})`
        : message || '系统繁忙，请稍后重试'

      ElMessage.error({
        message: errorMsg,
        duration: 5000
      })

      // 记录错误信息到控制台，方便开发排查
      console.error('Server Error:', {
        code,
        message,
        requestID,
        time: new Date().toISOString()
      })
      break

    // 其他错误
    default:
      ElMessage.error(message || '操作失败')
  }
}

// 处理认证错误
function handleAuthError(code, message) {
  const errorMap = {
    4010: '请先登录',
    4011: '登录已过期，请重新登录',
    4012: '登录状态无效',
    4013: '用户名或密码错误'
  }

  const msg = message || errorMap[code] || '认证失败'

  // 清除登录状态
  localStorage.removeItem('token')
  localStorage.removeItem('user')

  // 显示提示并跳转
  ElMessageBox.confirm(msg, '提示', {
    confirmButtonText: '重新登录',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    router.push('/login')
  }).catch(() => {
    // 用户取消
  })
}

// 处理权限错误
function handlePermissionError(code, message) {
  ElMessage.error(message || '没有权限执行此操作')
}

// 处理资源不存在
function handleNotFoundError(code, message) {
  ElMessage.warning(message || '请求的资源不存在')
}

// 处理 HTTP 错误（网络层）
function handleHTTPError(error) {
  if (axios.isCancel(error)) {
    console.log('Request canceled:', error.message)
    return
  }

  const { response, request, message } = error

  if (response) {
    // 服务器返回了错误状态码
    const status = response.status
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
    }

    ElMessage.error({
      message: statusMap[status] || `服务器错误: ${status}`,
      duration: 5000
    })

    // 401 未授权，跳转到登录页
    if (status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      router.push('/login')
    }
  } else if (request) {
    // 请求发出但没有收到响应
    ElMessage.error('网络连接失败，请检查网络设置')
  } else {
    // 请求配置出错
    ElMessage.error('请求配置错误: ' + message)
  }
}

// 认证相关API
export const authAPI = {
  // 注册
  register: (data) => api.post('/auth/register', data),
  // 登录
  login: (data) => api.post('/auth/login', data),
  // 登出
  logout: () => api.post('/auth/logout')
}

// 商品相关API
export const productAPI = {
  // 获取商品列表
  getProducts: (params) => api.get('/products', { params }),
  // 获取商品详情
  getProductDetail: (id) => api.get(`/products/${id}`)
}

// 购物车相关API
export const cartAPI = {
  // 获取购物车
  getCart: () => api.get('/cart'),
  // 添加商品到购物车
  addToCart: (data) => api.post('/cart/items', data),
  // 更新购物车项
  updateCartItem: (id, data) => api.put(`/cart/items/${id}`, data),
  // 移除购物车项
  removeCartItem: (id) => api.delete(`/cart/items/${id}`),
  // 同步购物车
  syncCart: (data) => api.post('/cart/sync', data)
}

// 客户相关API
export const customerAPI = {
  // 获取个人信息
  getProfile: () => api.get('/user/profile'),
  // 更新个人信息
  updateProfile: (data) => api.put('/user/profile', data),
  // 获取订单列表
  getOrders: (params) => api.get('/user/orders', { params })
}

// 地址相关API
export const addressAPI = {
  // 获取地址列表
  getAddresses: () => api.get('/customer/addresses'),
  // 获取单个地址
  getAddress: (id) => api.get(`/customer/addresses/${id}`),
  // 创建地址
  createAddress: (data) => api.post('/customer/addresses', data),
  // 更新地址
  updateAddress: (id, data) => api.put(`/customer/addresses/${id}`, data),
  // 删除地址
  deleteAddress: (id) => api.delete(`/customer/addresses/${id}`),
  // 设置默认地址
  setDefaultAddress: (id) => api.put(`/customer/addresses/${id}/default`),
  // 获取默认地址
  getDefaultAddress: () => api.get('/customer/addresses/default')
}

// 订单相关API
export const orderAPI = {
  // 创建订单
  createOrder: (data) => api.post('/orders', data),
  // 获取订单详情
  getOrderDetail: (orderNo) => api.get(`/orders/${orderNo}`),
  // 取消订单
  cancelOrder: (orderNo) => api.put(`/orders/${orderNo}/cancel`),
  // 确认收货
  confirmReceipt: (orderNo) => api.put(`/orders/${orderNo}/confirm`)
}

// 支付相关API
export const paymentAPI = {
  // 模拟支付
  fakePay: (orderNo) => api.get(`/payment/fake-pay?orderNo=${orderNo}`),
  // 支付回调
  paymentCallback: (data) => api.post('/payment/callback', data)
}

// 验证码相关API
export const captchaAPI = {
  // 获取验证码 - 返回原始响应以获取 headers
  getCaptcha: () => api.get('/captcha', { responseType: 'blob' }),
  // 验证验证码
  verifyCaptcha: (data) => api.post('/captcha/verify', data)
}

export default api
