import axios from 'axios';
import { ElMessage, ElMessageBox } from 'element-plus';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api';

// 创建 axios 实例
const api = axios.create({
  baseURL: API_BASE_URL,
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
api.interceptors.request.use(
  config => {
    // 1. 移除重复请求
    removePendingRequest(config);
    addPendingRequest(config);

    // 2. 添加 token
    const token = localStorage.getItem('token');
    if (token) {
      config.headers['Authorization'] = 'Bearer ' + token;
    }

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
api.interceptors.response.use(
  response => {
    // 1. 移除已完成的请求
    removePendingRequest(response.config);

    // 2. 获取后端返回的数据
    const res = response.data;

    // 3. 【单体应用】获取 RequestID（用于问题排查，可选）
    const requestID = response.headers['x-request-id'];
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
      ElMessage.warning(message || '请求参数错误');
      break;

    // 服务器错误（5000+）
    case code >= 5000:
      // 【单体应用】错误时显示 RequestID，方便用户反馈问题
      const errorMsg = requestID
        ? `${message || '系统繁忙，请稍后重试'} (请求ID: ${requestID})`
        : message || '系统繁忙，请稍后重试';

      ElMessage.error({
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
      ElMessage.error(message || '操作失败');
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

  // 清除登录状态
  localStorage.removeItem('token');
  localStorage.removeItem('user');

  // 显示提示消息
  ElMessage.warning(msg);

  // 直接跳转到登录页面
  window.location.href = '/login';
}

// 处理权限错误
function handlePermissionError(code, message) {
  ElMessage.error(message || '没有权限执行此操作');
}

// 处理资源不存在
function handleNotFoundError(code, message) {
  ElMessage.warning(message || '请求的资源不存在');
}

// 处理 HTTP 错误（网络层）
async function handleHTTPError(error) {
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

    ElMessage.error({
      message: statusMap[status] || `服务器错误: ${status}`,
      duration: 5000
    });

    // 401 未授权，跳转到登录页
    if (status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      const { default: router } = await import('../router');
      router.push('/login');
    }
  } else if (request) {
    // 请求发出但没有收到响应
    ElMessage.error('网络连接失败，请检查网络设置');
  } else {
    // 请求配置出错
    ElMessage.error('请求配置错误: ' + message);
  }
}

// 认证相关 API
export const authApi = {
  // 登录
  login: (data) => api.post('/auth/login', data),

  // 登出
  logout: () => api.post('/auth/logout'),

  // 刷新 token
  refresh: () => api.post('/auth/refresh'),

  // 获取当前用户信息
  getCurrentUser: () => api.get('/auth/me'),

  // 获取验证码
  getCaptcha: () => api.get('/auth/captcha'),

  // 验证验证码
  verifyCaptcha: (data) => api.post('/auth/captcha/verify', data),

  // 用户管理
  getUsers: () => api.get('/users'),
  getUser: (id) => api.get(`/users/${id}`),
  createUser: (data) => api.post('/users', data),
  updateUser: (id, data) => api.put(`/users/${id}`, data),
  deleteUser: (id) => api.delete(`/users/${id}`),

  // 角色管理
  getRoles: () => api.get('/roles'),
  getRole: (id) => api.get(`/roles/${id}`),
  createRole: (data) => api.post('/roles', data),
  updateRole: (id, data) => api.put(`/roles/${id}`, data),
  deleteRole: (id) => api.delete(`/roles/${id}`),
  assignPermissions: (roleId, data) => api.post(`/roles/${roleId}/permissions`, data),

  // 权限管理
  getPermissions: () => api.get('/permissions'),
  getPermission: (id) => api.get(`/permissions/${id}`),
  createPermission: (data) => api.post('/permissions', data),
  updatePermission: (id, data) => api.put(`/permissions/${id}`, data),
  deletePermission: (id) => api.delete(`/permissions/${id}`),

  // 商户管理
  getMerchants: () => api.get('/merchants'),
  getMerchant: (id) => api.get(`/merchants/${id}`),
  createMerchant: (data) => api.post('/merchants', data),
  updateMerchant: (id, data) => api.put(`/merchants/${id}`, data),
  deleteMerchant: (id) => api.delete(`/merchants/${id}`),
  auditMerchant: (id, data) => api.put(`/merchants/${id}/audit`, data),
  getMerchantUsers: (merchantId) => api.get(`/merchants/${merchantId}/users`),
  addMerchantUser: (merchantId, data) => api.post(`/merchants/${merchantId}/users`, data),
  removeMerchantUser: (merchantId, userId) => api.delete(`/merchants/${merchantId}/users/${userId}`)
};

// 商品相关API
export const productApi = {
  // 获取商品列表
  getProducts: () => api.get('/products'),
  // 获取商品详情
  getProduct: (id) => api.get(`/products/${id}`),
  // 创建商品
  createProduct: (data) => api.post('/products', data),
  // 更新商品
  updateProduct: (id, data) => api.put(`/products/${id}`, data),
  // 删除商品
  deleteProduct: (id) => api.delete(`/products/${id}`),
  // 获取商品分类列表
  getCategories: () => api.get('/product-categories'),
  // 获取商品分类详情
  getCategory: (id) => api.get(`/product-categories/${id}`),
  // 创建商品分类
  createCategory: (data) => api.post('/product-categories', data),
  // 更新商品分类
  updateCategory: (id, data) => api.put(`/product-categories/${id}`, data),
  // 删除商品分类
  deleteCategory: (id) => api.delete(`/product-categories/${id}`),
  // 添加商品图片
  addProductImage: (data) => api.post('/product-images', data),
  // 删除商品图片
  deleteProductImage: (id) => api.delete(`/product-images/${id}`),
  // 更新商品图片
  updateProductImage: (id, data) => api.put(`/product-images/${id}`, data),

  // ========== SKU多规格管理API ==========
  // 获取商品规格列表
  getProductSpecifications: (productId) => api.get(`/products/${productId}/specifications`),
  // 创建商品规格
  createProductSpecification: (productId, data) => api.post(`/products/${productId}/specifications`, data),
  // 更新规格
  updateSpecification: (id, data) => api.put(`/specifications/${id}`, data),
  // 删除规格
  deleteSpecification: (id) => api.delete(`/specifications/${id}`),
  // 创建规格值
  createSpecificationValue: (specId, data) => api.post(`/specifications/${specId}/values`, data),
  // 更新规格值
  updateSpecificationValue: (id, data) => api.put(`/specification-values/${id}`, data),
  // 删除规格值
  deleteSpecificationValue: (id) => api.delete(`/specification-values/${id}`),
  // 获取商品SKU列表
  getProductSKUs: (productId) => api.get(`/products/${productId}/skus`),
  // 创建SKU
  createProductSKU: (productId, data) => api.post(`/products/${productId}/skus`, data),
  // 批量创建SKU
  batchCreateSKUs: (productId, data) => api.post(`/products/${productId}/skus/batch`, data),
  // 更新SKU
  updateSKU: (id, data) => api.put(`/skus/${id}`, data),
  // 删除SKU
  deleteSKU: (id) => api.delete(`/skus/${id}`),
  // 自动生成SKU组合
  generateSKUs: (productId, data) => api.post(`/products/${productId}/skus/generate`, data)
};

// 活动相关API
export const activityApi = {
  // 活动管理
  getActivities: (params) => api.get('/activities', { params }),
  getActivity: (id) => api.get(`/activities/${id}`),
  createActivity: (data) => api.post('/activities', data),
  updateActivity: (id, data) => api.put(`/activities/${id}`, data),
  deleteActivity: (id) => api.delete(`/activities/${id}`),
  updateActivityStatus: (id, data) => api.put(`/activities/${id}/status`, data),
  
  // 兑换码管理
  generateRedeemCodes: (activityId, data) => api.post(`/activities/${activityId}/redeem-codes/generate`, data),
  getRedeemCodes: (activityId, params) => api.get(`/activities/${activityId}/redeem-codes`, { params }),
  exportRedeemCodes: (activityId, params) => api.get(`/activities/${activityId}/redeem-codes/export`, { params, responseType: 'blob' }),
  importRedeemCodes: (activityId, formData) => api.post(`/activities/${activityId}/redeem-codes/import`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  }),
  updateRedeemCodeStatus: (id, data) => api.put(`/redeem-codes/${id}/status`, data),
  
  // 兑换码核销
  verifyRedeemCode: (data) => api.post('/redeem-codes/verify', data),
  getRedeemCodeLogs: (params) => api.get('/redeem-codes/logs', { params })
};

export default api;
