import { createRouter, createWebHistory } from 'vue-router'
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
    redirect: '/login'
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
        component: Users
      },
      {
        path: 'roles',
        name: 'Roles',
        component: Roles
      },
      {
        path: 'permissions',
        name: 'Permissions',
        component: Permissions
      },
      {
        path: 'merchants',
        name: 'Merchants',
        component: Merchants
      },
      {
        path: 'products',
        name: 'Products',
        component: Products
      },
      {
        path: 'product-categories',
        name: 'ProductCategories',
        component: ProductCategories
      },
      {
        path: 'products/:id/specifications',
        name: 'ProductSpecifications',
        component: ProductSpecifications
      },
      {
        path: 'products/:id/skus',
        name: 'ProductSKUs',
        component: ProductSKUs
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path === '/login') {
    next()
  } else {
    if (token) {
      next()
    } else {
      next('/login')
    }
  }
})

export default router
