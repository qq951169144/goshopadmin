import axios from 'axios'

const API_BASE_URL = '/api'

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    return Promise.reject(error)
  }
)

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
  getProfile: () => api.get('/customer/profile'),
  // 更新个人信息
  updateProfile: (data) => api.put('/customer/profile', data),
  // 获取订单列表
  getOrders: (params) => api.get('/customer/orders', { params })
}

// 订单相关API
export const orderAPI = {
  // 创建订单
  createOrder: (data) => api.post('/orders', data),
  // 获取订单详情
  getOrderDetail: (id) => api.get(`/orders/${id}`)
}

// 支付相关API
export const paymentAPI = {
  // 模拟支付
  fakePay: (params) => api.get('/payment/fake-pay', { params }),
  // 支付回调
  paymentCallback: (data) => api.post('/payment/callback', data)
}

// 验证码相关API
export const captchaAPI = {
  // 获取验证码
  getCaptcha: () => api.get('/captcha'),
  // 验证验证码
  verifyCaptcha: (data) => api.post('/captcha/verify', data)
}

export default api