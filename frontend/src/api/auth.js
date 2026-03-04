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
  
  // 获取角色列表
  getRoles: () => api.get('/roles'),
  
  // 创建角色
  createRole: (data) => api.post('/roles', data),
  
  // 更新角色
  updateRole: (id, data) => api.put(`/roles/${id}`, data),
  
  // 删除角色
  deleteRole: (id) => api.delete(`/roles/${id}`),
  
  // 获取权限列表
  getPermissions: () => api.get('/permissions'),
  
  // 为角色分配权限
  assignPermissions: (roleId, data) => api.post(`/roles/${roleId}/permissions`, data)
};

export default api;