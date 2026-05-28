<template>
  <div id="app">
    <!-- 顶部导航栏 -->
    <header class="top-nav" v-if="showTopNav">
      <router-link to="/messages" class="nav-item">
        <span class="nav-icon">📬</span>
        <span v-if="unreadMessageCount > 0" class="message-badge">{{ unreadMessageCount > 99 ? '99+' : unreadMessageCount }}</span>
      </router-link>
      <span class="nav-title">GoShop</span>
      <router-link to="/customer/service" class="nav-item">
        <span class="nav-icon">📞</span>
      </router-link>
    </header>

    <div class="main-content">
      <router-view />
    </div>

    <!-- 底部导航栏 -->
    <nav class="bottom-nav" v-if="showBottomNav">
      <router-link to="/" class="nav-item" :class="{ active: $route.path === '/' }">
        <span class="nav-icon">🏠</span>
        <span class="nav-text">首页</span>
      </router-link>
      <router-link to="/orders" class="nav-item" :class="{ active: $route.path === '/orders' || $route.path.startsWith('/order/') }">
        <span class="nav-icon">📋</span>
        <span class="nav-text">订单</span>
      </router-link>
      <router-link to="/activity/orders" class="nav-item" :class="{ active: $route.path.startsWith('/activity/order') }">
        <span class="nav-icon">⚡</span>
        <span class="nav-text">活动</span>
      </router-link>
      <router-link to="/cart" class="nav-item" :class="{ active: $route.path === '/cart' }">
        <span class="nav-icon">🛒</span>
        <span class="nav-text">购物车</span>
        <span v-if="cartCount > 0" class="cart-badge">{{ cartCount }}</span>
      </router-link>
      <div class="nav-item" :class="{ active: $route.path === '/customer/profile' }" @click="handleProfileClick">
        <span class="nav-icon">👤</span>
        <span class="nav-text">我的</span>
      </div>
    </nav>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { cartAPI } from './api'
import { useMessageStore } from './store/message'

const route = useRoute()
const router = useRouter()
const cartCount = ref(0)

const messageStore = useMessageStore()

const hideNavPaths = ['/login', '/register', '/checkout', '/address/edit', '/activity/order/confirm']

const showTopNav = computed(() => {
  return !['/login', '/register'].some(path => route.path.startsWith(path))
})

const showBottomNav = computed(() => {
  return !hideNavPaths.some(path => route.path.startsWith(path))
})

const unreadMessageCount = computed(() => messageStore.unreadCount)

const isLoggedIn = computed(() => {
  return !!localStorage.getItem('token')
})

const handleProfileClick = () => {
  const token = localStorage.getItem('token')
  if (token) {
    router.push('/customer/profile')
  } else {
    router.push('/login')
  }
}

const loadCartCount = async () => {
  if (!isLoggedIn.value) {
    cartCount.value = 0
    return
  }
  try {
    const response = await cartAPI.getCart()
    const items = response.items || []
    cartCount.value = items.reduce((total, item) => total + item.quantity, 0)
  } catch (error) {
    console.error('加载购物车数量失败:', error)
    cartCount.value = 0
  }
}

watch(() => localStorage.getItem('token'), () => {
  loadCartCount()
})

onMounted(() => {
  messageStore.loadFromLocal()
  loadCartCount()
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

#app {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 顶部导航栏样式 */
.top-nav {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 44px;
  background-color: #fff;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  z-index: 1001;
}

.top-nav .nav-title {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.top-nav .nav-item {
  display: flex;
  align-items: center;
  position: relative;
  text-decoration: none;
}

.top-nav .nav-icon {
  font-size: 18px;
}

.top-nav .message-badge {
  position: absolute;
  top: -5px;
  right: -8px;
  background-color: #ff4757;
  color: white;
  font-size: 10px;
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}

.main-content {
  flex: 1;
  padding-top: 44px;
  padding-bottom: 60px;
}

/* 底部导航栏样式 */
.bottom-nav {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 60px;
  background-color: #fff;
  border-top: 1px solid #eee;
  display: flex;
  justify-content: space-around;
  align-items: center;
  z-index: 1000;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  height: 100%;
  text-decoration: none;
  color: #999;
  cursor: pointer;
  position: relative;
  transition: all 0.3s ease;
}

.nav-item:hover,
.nav-item.active {
  color: #4CAF50;
}

.nav-icon {
  font-size: 20px;
  margin-bottom: 2px;
}

.nav-text {
  font-size: 11px;
}

/* 购物车角标 */
.cart-badge {
  position: absolute;
  top: 5px;
  right: calc(50% - 20px);
  background-color: #ff4757;
  color: white;
  font-size: 10px;
  min-width: 16px;
  height: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 4px;
}

/* 响应式适配 */
@media (min-width: 768px) {
  .bottom-nav {
    max-width: 768px;
    left: 50%;
    transform: translateX(-50%);
  }
}
</style>
