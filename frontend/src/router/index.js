import { createRouter, createWebHistory } from 'vue-router'
import Login from '../components/Login.vue'
import Home from '../components/Home.vue'
import Dashboard from '../components/dashboard/Dashboard.vue'
import Users from '../components/users/Users.vue'
import Roles from '../components/roles/Roles.vue'
import Permissions from '../components/permissions/Permissions.vue'
import Merchants from '../components/merchants/Merchants.vue'
import Products from '../components/products/Products.vue'
import ProductCategories from '../components/products/ProductCategories.vue'

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
