<template>
  <div class="customer-profile">
    <div class="profile-header">
      <div class="avatar" @click="triggerUpload">
        <img :src="user.avatar || defaultAvatar" alt="头像" />
        <div class="avatar-overlay">更换</div>
        <input
          ref="fileInput"
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp"
          style="display: none"
          @change="handleFileChange"
        />
      </div>
      <div class="user-info">
        <h2>{{ user.username || '未登录' }}</h2>
        <p v-if="user.phone">{{ user.phone }}</p>
      </div>
    </div>

    <!-- 订单状态快捷入口 -->
    <div class="order-status">
      <div class="status-item" @click="goToOrders('pending')">
        <span class="icon">⏰</span>
        <span class="label">待付款</span>
      </div>
      <div class="status-item" @click="goToOrders('paid')">
        <span class="icon">📦</span>
        <span class="label">待发货</span>
      </div>
      <div class="status-item" @click="goToOrders('shipped')">
        <span class="icon">🚚</span>
        <span class="label">待收货</span>
      </div>
      <div class="status-item" @click="goToOrders('completed')">
        <span class="icon">✅</span>
        <span class="label">已完成</span>
      </div>
    </div>

    <!-- 功能菜单 -->
    <div class="menu-list">
      <div class="menu-item" @click="goToOrders('all')">
        <span class="menu-icon">📋</span>
        <span class="menu-text">我的订单</span>
        <span class="arrow">›</span>
      </div>
      <div class="menu-item" @click="goToAddresses">
        <span class="menu-icon">📍</span>
        <span class="menu-text">收货地址</span>
        <span class="arrow">›</span>
      </div>
      <div class="menu-item" @click="contactService">
        <span class="menu-icon">💬</span>
        <span class="menu-text">联系客服</span>
        <span class="arrow">›</span>
      </div>
      <div class="menu-item" @click="goToSettings">
        <span class="menu-icon">⚙️</span>
        <span class="menu-text">设置</span>
        <span class="arrow">›</span>
      </div>
    </div>

    <!-- 退出登录按钮 -->
    <div class="logout-section">
      <button class="logout-btn" @click="logout">退出登录</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { customerAPI } from '../api'

const router = useRouter()
const user = ref({})
const fileInput = ref(null)
const uploading = ref(false)

const defaultAvatar = 'https://via.placeholder.com/80x80?text=Avatar'

const loadProfile = async () => {
  try {
    const response = await customerAPI.getProfile()
    user.value = response
  } catch (error) {
    console.error('加载用户信息失败:', error)
    user.value = null
  }
}

const triggerUpload = () => {
  if (uploading.value) return
  fileInput.value.click()
}

const handleFileChange = async (event) => {
  const file = event.target.files[0]
  if (!file) return

  // 验证文件大小（2MB）
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.warning('头像文件大小不能超过2MB')
    return
  }

  uploading.value = true
  try {
    const formData = new FormData()
    formData.append('avatar', file)
    const response = await customerAPI.uploadAvatar(formData)
    user.value.avatar = response.avatar
    ElMessage.success('头像上传成功')
  } catch (error) {
    console.error('头像上传失败:', error)
    ElMessage.error('头像上传失败')
  } finally {
    uploading.value = false
    // 清空 input，允许重复选择同一文件
    event.target.value = ''
  }
}

const goToOrders = (status) => {
  router.push('/orders')
}

const goToAddresses = () => {
  router.push('/addresses')
}

const contactService = () => {
  alert('客服功能开发中...')
}

const goToSettings = () => {
  alert('设置功能开发中...')
}

const logout = () => {
  if (!confirm('确定要退出登录吗？')) return
  
  localStorage.removeItem('token')
  router.push('/login')
}

onMounted(() => {
  loadProfile()
})
</script>

<style scoped>
.customer-profile {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 20px;
}

/* 用户信息头部 */
.profile-header {
  display: flex;
  align-items: center;
  padding: 24px 16px;
  background: linear-gradient(135deg, #4CAF50 0%, #45a049 100%);
  color: white;
}

.avatar {
  position: relative;
  width: 70px;
  height: 70px;
  border-radius: 50%;
  overflow: hidden;
  margin-right: 16px;
  border: 3px solid rgba(255, 255, 255, 0.3);
  cursor: pointer;
}

.avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-overlay {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.5);
  color: white;
  text-align: center;
  font-size: 12px;
  padding: 2px 0;
  opacity: 0;
  transition: opacity 0.3s ease;
}

.avatar:hover .avatar-overlay {
  opacity: 1;
}

.user-info h2 {
  font-size: 20px;
  margin-bottom: 4px;
}

.user-info p {
  font-size: 14px;
  opacity: 0.9;
}

/* 订单状态快捷入口 */
.order-status {
  display: flex;
  justify-content: space-around;
  background-color: white;
  padding: 20px 16px;
  margin-bottom: 12px;
}

.status-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
}

.status-item .icon {
  font-size: 24px;
  margin-bottom: 8px;
}

.status-item .label {
  font-size: 12px;
  color: #666;
}

/* 功能菜单 */
.menu-list {
  background-color: white;
  margin-bottom: 12px;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 16px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-item:hover {
  background-color: #fafafa;
}

.menu-icon {
  font-size: 20px;
  margin-right: 12px;
}

.menu-text {
  flex: 1;
  font-size: 15px;
  color: #333;
}

.arrow {
  color: #999;
  font-size: 18px;
}

/* 退出登录 */
.logout-section {
  padding: 20px 16px;
}

.logout-btn {
  width: 100%;
  padding: 14px;
  background-color: white;
  color: #ff4757;
  border: 1px solid #ff4757;
  border-radius: 24px;
  font-size: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.logout-btn:hover {
  background-color: #ff4757;
  color: white;
}
</style>
