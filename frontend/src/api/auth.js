import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
});

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response.data;
  },
  error => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

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

export default api;