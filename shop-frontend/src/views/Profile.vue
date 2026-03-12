<template>
  <div class="profile">
    <h1>个人中心</h1>
    
    <div v-if="!isLoggedIn" class="login-prompt">
      <p>请先登录</p>
      <button @click="goLogin">去登录</button>
    </div>
    
    <div v-else>
      <div class="profile-content">
        <div class="profile-sidebar">
          <div class="user-info">
            <div class="avatar">
              <img :src="userInfo.avatar || 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=user%20avatar&image_size=square'" alt="用户头像" />
            </div>
            <h2>{{ userInfo.username || '用户' }}</h2>
          </div>
          <div class="nav-menu">
            <button :class="{ active: activeTab === 'info' }" @click="activeTab = 'info'">个人信息</button>
            <button :class="{ active: activeTab === 'orders' }" @click="activeTab = 'orders'">我的订单</button>
            <button @click="logout">退出登录</button>
          </div>
        </div>
        
        <div class="profile-main">
          <!-- 个人信息 -->
          <div v-if="activeTab === 'info'" class="info-section">
            <h3>个人信息</h3>
            <form @submit.prevent="updateProfile">
              <div class="form-group">
                <label>用户名</label>
                <input type="text" v-model="userInfo.username" required />
              </div>
              <div class="form-group">
                <label>邮箱</label>
                <input type="email" v-model="userInfo.email" />
              </div>
              <div class="form-group">
                <label>手机号</label>
                <input type="tel" v-model="userInfo.phone" />
              </div>
              <button type="submit" :disabled="loading">{{ loading ? '保存中...' : '保存修改' }}</button>
            </form>
          </div>
          
          <!-- 订单列表 -->
          <div v-if="activeTab === 'orders'" class="orders-section">
            <h3>我的订单</h3>
            <div v-if="loading" class="loading">加载中...</div>
            <div v-else-if="orders.length === 0" class="empty-orders">
              <p>暂无订单</p>
              <button @click="goShopping">去购物</button>
            </div>
            <div v-else class="order-list">
              <div v-for="order in orders" :key="order.id" class="order-item">
                <div class="order-header">
                  <span class="order-id">订单号：{{ order.id }}</span>
                  <span class="order-status">{{ order.status }}</span>
                </div>
                <div class="order-items">
                  <div v-for="item in order.items" :key="item.product_id" class="order-item-product">
                    <img :src="item.image" :alt="item.name" />
                    <div class="product-info">
                      <h4>{{ item.name }}</h4>
                      <p>¥{{ item.price }} x {{ item.quantity }}</p>
                    </div>
                  </div>
                </div>
                <div class="order-footer">
                  <span class="order-total">合计：¥{{ order.total }}</span>
                  <div class="order-actions">
                    <button v-if="order.status === '待支付'" @click="payOrder(order.id)">去支付</button>
                    <button v-if="order.status === '待发货'" @click="cancelOrder(order.id)">取消订单</button>
                    <button v-if="order.status === '待收货'" @click="confirmReceipt(order.id)">确认收货</button>
                    <button @click="viewOrderDetail(order.id)">查看详情</button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { userAPI, orderAPI } from '../api'

const router = useRouter()
const activeTab = ref('info')
const loading = ref(false)
const userInfo = ref({
  username: '',
  email: '',
  phone: '',
  avatar: ''
})
const orders = ref([])

const isLoggedIn = computed(() => {
  return !!localStorage.getItem('token')
})

const goLogin = () => {
  router.push('/login')
}

const goShopping = () => {
  router.push('/')
}

const loadUserInfo = async () => {
  if (!isLoggedIn.value) return
  
  try {
    const response = await userAPI.getProfile()
    userInfo.value = response.data || userInfo.value
  } catch (error) {
    console.error('加载个人信息失败:', error)
    // 使用模拟数据
    userInfo.value = {
      username: '测试用户',
      email: 'test@example.com',
      phone: '13800138000',
      avatar: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=user%20avatar&image_size=square'
    }
  }
}

const loadOrders = async () => {
  if (!isLoggedIn.value) return
  
  loading.value = true
  try {
    const response = await userAPI.getOrders()
    orders.value = response.data || []
  } catch (error) {
    console.error('加载订单失败:', error)
    // 使用模拟数据
    orders.value = [
      {
        id: '202401010001',
        status: '待支付',
        total: 299.97,
        items: [
          {
            product_id: 1,
            name: '商品1',
            price: 99.99,
            quantity: 3,
            image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%201&image_size=square'
          }
        ]
      },
      {
        id: '202401010002',
        status: '已完成',
        total: 199.99,
        items: [
          {
            product_id: 2,
            name: '商品2',
            price: 199.99,
            quantity: 1,
            image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%202&image_size=square'
          }
        ]
      }
    ]
  } finally {
    loading.value = false
  }
}

const updateProfile = async () => {
  loading.value = true
  try {
    await userAPI.updateProfile(userInfo.value)
    alert('个人信息更新成功')
  } catch (error) {
    console.error('更新个人信息失败:', error)
    alert('更新个人信息失败，请稍后重试')
  } finally {
    loading.value = false
  }
}

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user_id')
  router.push('/')
}

const payOrder = (orderId) => {
  alert(`支付订单 ${orderId}`)
}

const cancelOrder = (orderId) => {
  alert(`取消订单 ${orderId}`)
}

const confirmReceipt = (orderId) => {
  alert(`确认收货 ${orderId}`)
}

const viewOrderDetail = (orderId) => {
  alert(`查看订单详情 ${orderId}`)
}

onMounted(() => {
  loadUserInfo()
  loadOrders()
})
</script>

<style scoped>
.profile {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

h1 {
  margin-bottom: 30px;
  color: #333;
  text-align: center;
}

.login-prompt {
  text-align: center;
  padding: 100px 0;
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.login-prompt p {
  margin-bottom: 20px;
  color: #666;
  font-size: 18px;
}

.login-prompt button {
  padding: 10px 30px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
}

.profile-content {
  display: flex;
  gap: 40px;
  flex-wrap: wrap;
}

.profile-sidebar {
  width: 250px;
  background-color: #f9f9f9;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.user-info {
  text-align: center;
  margin-bottom: 30px;
}

.avatar {
  width: 100px;
  height: 100px;
  margin: 0 auto 20px;
  border-radius: 50%;
  overflow: hidden;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.user-info h2 {
  color: #333;
  margin-bottom: 10px;
}

.nav-menu {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.nav-menu button {
  padding: 12px;
  border: none;
  background-color: white;
  border-radius: 4px;
  cursor: pointer;
  text-align: left;
  transition: all 0.3s ease;
}

.nav-menu button:hover {
  background-color: #f0f0f0;
}

.nav-menu button.active {
  background-color: #4CAF50;
  color: white;
}

.nav-menu button:last-child {
  margin-top: 20px;
  background-color: #ff4757;
  color: white;
}

.profile-main {
  flex: 1;
  min-width: 600px;
  background-color: white;
  border-radius: 8px;
  padding: 30px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.info-section h3,
.orders-section h3 {
  margin-bottom: 20px;
  color: #333;
  border-bottom: 2px solid #4CAF50;
  padding-bottom: 10px;
}

.form-group {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 5px;
  color: #666;
  font-size: 14px;
}

input {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 16px;
}

input:focus {
  outline: none;
  border-color: #4CAF50;
  box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.1);
}

.info-section button {
  padding: 12px 30px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  margin-top: 20px;
}

.info-section button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.loading,
.empty-orders {
  text-align: center;
  padding: 50px 0;
  color: #666;
}

.empty-orders button {
  margin-top: 20px;
  padding: 10px 30px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.order-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.order-item {
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 20px;
}

.order-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 15px;
  padding-bottom: 15px;
  border-bottom: 1px solid #eee;
}

.order-id {
  color: #666;
}

.order-status {
  color: #4CAF50;
  font-weight: bold;
}

.order-items {
  margin-bottom: 15px;
}

.order-item-product {
  display: flex;
  gap: 15px;
  margin-bottom: 10px;
}

.order-item-product img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
}

.product-info h4 {
  margin-bottom: 5px;
  color: #333;
}

.product-info p {
  color: #666;
  font-size: 14px;
}

.order-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 15px;
  border-top: 1px solid #eee;
}

.order-total {
  font-weight: bold;
  color: #333;
}

.order-actions {
  display: flex;
  gap: 10px;
}

.order-actions button {
  padding: 6px 12px;
  border: 1px solid #ddd;
  background-color: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: all 0.3s ease;
}

.order-actions button:hover {
  background-color: #f0f0f0;
}

@media (max-width: 768px) {
  .profile-content {
    flex-direction: column;
  }
  
  .profile-sidebar {
    width: 100%;
  }
  
  .profile-main {
    min-width: 100%;
  }
}
</style>