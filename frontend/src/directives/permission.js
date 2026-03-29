import { usePermissionStore } from '../store/modules/permission'

// 权限指令
export const permission = {
  mounted(el, binding) {
    const permissionStore = usePermissionStore()
    const { value } = binding
    
    if (value && !permissionStore.hasPermission(value)) {
      el.style.display = 'none'
    }
  },
  updated(el, binding) {
    const permissionStore = usePermissionStore()
    const { value } = binding
    
    if (value && !permissionStore.hasPermission(value)) {
      el.style.display = 'none'
    } else {
      el.style.display = ''
    }
  }
}

// 权限验证函数
export function checkPermission(permissionCode) {
  const permissionStore = usePermissionStore()
  return permissionStore.hasPermission(permissionCode)
}

export function checkAnyPermission(permissionCodes) {
  const permissionStore = usePermissionStore()
  return permissionStore.hasAnyPermission(permissionCodes)
}

export function checkAllPermissions(permissionCodes) {
  const permissionStore = usePermissionStore()
  return permissionStore.hasAllPermissions(permissionCodes)
}
