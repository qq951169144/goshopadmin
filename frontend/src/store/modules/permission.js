import { defineStore } from 'pinia'
import { authApi } from '../../api/auth'

export const usePermissionStore = defineStore('permission', {
  state: () => ({
    permissions: [],
    loading: false,
    error: null
  }),
  
  getters: {
    hasPermission: (state) => (permissionCode) => {
      return state.permissions.includes(permissionCode)
    },
    hasAnyPermission: (state) => (permissionCodes) => {
      return permissionCodes.some(code => state.permissions.includes(code))
    },
    hasAllPermissions: (state) => (permissionCodes) => {
      return permissionCodes.every(code => state.permissions.includes(code))
    }
  },
  
  actions: {
    async fetchPermissions() {
      this.loading = true
      this.error = null
      try {
        const user = JSON.parse(localStorage.getItem('user'))
        if (user) {
          // 从用户信息中获取权限
          // 或者调用API获取权限
          const response = await authApi.getCurrentUser()
          // 提取权限代码并存储为字符串数组
          this.permissions = response.permissions ? response.permissions.map(p => p.code) : []
          localStorage.setItem('permissions', JSON.stringify(this.permissions))
        }
      } catch (error) {
        this.error = error.message
        // 从本地存储恢复权限
        const cachedPermissions = localStorage.getItem('permissions')
        if (cachedPermissions) {
          this.permissions = JSON.parse(cachedPermissions)
        }
      } finally {
        this.loading = false
      }
    },
    
    clearPermissions() {
      this.permissions = []
      localStorage.removeItem('permissions')
    }
  }
})
