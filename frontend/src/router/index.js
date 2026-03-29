import { createRouter, createWebHistory } from 'vue-router'
import { usePermissionStore } from '../store/modules/permission'
import Login from '../views/Login.vue'
import Home from '../views/Home.vue'
import Dashboard from '../views/dashboard/Dashboard.vue'
import Users from '../views/users/Users.vue'
import Roles from '../views/roles/Roles.vue'
import Permissions from '../views/permissions/Permissions.vue'
import Merchants from '../views/merchants/Merchants.vue'
import Products from '../views/products/Products.vue'
import ProductCategories from '../views/products/ProductCategories.vue'
import ProductSpecifications from '../views/products/ProductSpecifications.vue'
import ProductSKUs from '../views/products/ProductSKUs.vue'

const routes = [
  {
    path: '/',
    redirect: '/home/dashboard'
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/home',
    name: 'Home',
    component: Home,
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: Dashboard
      },
      {
        path: 'users',
        name: 'Users',
        component: Users,
        meta: { requiresPermission: 'user:manage' }
      },
      {
        path: 'roles',
        name: 'Roles',
        component: Roles,
        meta: { requiresPermission: 'role:manage' }
      },
      {
        path: 'permissions',
        name: 'Permissions',
        component: Permissions,
        meta: { requiresPermission: 'role:manage' }
      },
      {
        path: 'merchants',
        name: 'Merchants',
        component: Merchants,
        meta: { requiresPermission: 'merchant:manage' }
      },
      {
        path: 'products',
        name: 'Products',
        component: Products,
        meta: { requiresPermission: 'product:manage' }
      },
      {
        path: 'product-categories',
        name: 'ProductCategories',
        component: ProductCategories,
        meta: { requiresPermission: 'product:manage' }
      },
      {
        path: 'products/:id/specifications',
        name: 'ProductSpecifications',
        component: ProductSpecifications,
        meta: { requiresPermission: 'product:manage' }
      },
      {
        path: 'products/:id/skus',
        name: 'ProductSKUs',
        component: ProductSKUs,
        meta: { requiresPermission: 'product:manage' }
      },
      {
        path: 'activities',
        name: 'Activities',
        component: () => import('../views/activities/Activities.vue')
      },
      {
        path: 'activities/create',
        name: 'ActivityCreate',
        component: () => import('../views/activities/ActivityForm.vue')
      },
      {
        path: 'activities/:id/edit',
        name: 'ActivityEdit',
        component: () => import('../views/activities/ActivityForm.vue')
      },
      {
        path: 'activities/:id/redeem-codes',
        name: 'RedeemCodes',
        component: () => import('../views/activities/RedeemCodes.vue')
      },
      {
        path: 'activities/:id/redeem-codes/generate',
        name: 'RedeemCodeGenerate',
        component: () => import('../views/activities/RedeemCodeGenerate.vue')
      },
      {
        path: 'activities/:id/redeem-codes/import-export',
        name: 'RedeemCodeImportExport',
        component: () => import('../views/activities/RedeemCodeImportExport.vue')
      },
      {
        path: 'activities/:id',
        name: 'ActivityDetail',
        component: () => import('../views/activities/ActivityDetail.vue')
      },
      {
        path: 'redeem-codes/verify',
        name: 'RedeemCodeVerify',
        component: () => import('../views/activities/RedeemCodeVerify.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const token = localStorage.getItem('token')
  
  if (to.path === '/login') {
    next()
  } else {
    if (!token) {
      next('/login')
      return
    }
    
    const permissionStore = usePermissionStore()
    
    // 确保权限已加载
    if (permissionStore.permissions.length === 0) {
      await permissionStore.fetchPermissions()
    }
    
    // 检查路由权限
    if (to.meta.requiresPermission) {
      if (permissionStore.hasPermission(to.meta.requiresPermission)) {
        next()
      } else {
        // 无权限，重定向到首页
        next('/home/dashboard')
      }
    } else {
      next()
    }
  }
})

export default router
