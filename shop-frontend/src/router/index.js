import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('../views/Home.vue')
    },
    {
      path: '/products',
      name: 'products',
      component: () => import('../views/Products.vue')
    },
    {
      path: '/product/:id',
      name: 'product',
      component: () => import('../views/ProductDetail.vue')
    },
    {
      path: '/cart',
      name: 'cart',
      component: () => import('../views/Cart.vue')
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/Login.vue')
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/Register.vue')
    },
    {
      path: '/customer/profile',
      name: 'customer-profile',
      component: () => import('../views/CustomerProfile.vue')
    },
    {
      path: '/orders',
      name: 'orders',
      component: () => import('../views/OrderList.vue')
    },
    {
      path: '/order/:id',
      name: 'order-detail',
      component: () => import('../views/OrderDetail.vue')
    },
    {
      path: '/addresses',
      name: 'addresses',
      component: () => import('../views/AddressList.vue')
    },
    {
      path: '/address/edit/:id?',
      name: 'address-edit',
      component: () => import('../views/AddressEdit.vue')
    },
    {
      path: '/checkout',
      name: 'checkout',
      component: () => import('../views/OrderConfirm.vue')
    },
    {
      path: '/activity/:id',
      name: 'activity-detail',
      component: () => import('../views/ActivityDetail.vue')
    },
    {
      path: '/activity/order/confirm',
      name: 'activity-order-confirm',
      component: () => import('../views/ActivityOrderConfirm.vue')
    },
    {
      path: '/activity/orders',
      name: 'activity-orders',
      component: () => import('../views/ActivityOrderList.vue')
    },
    {
      path: '/activity/order/:id',
      name: 'activity-order-detail',
      component: () => import('../views/ActivityOrderDetail.vue')
    },
    {
      path: '/messages',
      name: 'messages',
      component: () => import('../views/MessageList.vue')
    },
    {
      path: '/customer/service',
      name: 'customer-service',
      component: () => import('../views/CustomerService.vue')
    }
  ]
})

export default router
