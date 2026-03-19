<template>
  <div class="address-list-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>收货地址</h1>
      <button class="add-btn" @click="addAddress">新增</button>
    </div>

    <div class="address-list">
      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="addresses.length === 0" class="empty-addresses">
        <div class="empty-icon">📍</div>
        <p>暂无收货地址</p>
        <button @click="addAddress">添加地址</button>
      </div>
      <div v-else>
        <div 
          v-for="address in addresses" 
          :key="address.id" 
          class="address-card"
          :class="{ 'is-default': address.is_default }"
          @click="selectAddress(address)"
        >
          <div class="address-content">
            <div class="contact">
              <span class="name">{{ address.name }}</span>
              <span class="phone">{{ address.phone }}</span>
              <span v-if="address.is_default" class="default-tag">默认</span>
            </div>
            <div class="address-text">
              {{ address.province }}{{ address.city }}{{ address.district }}{{ address.detail_address }}
            </div>
          </div>
          <div class="address-actions">
            <button class="edit-btn" @click.stop="editAddress(address)">编辑</button>
            <button class="delete-btn" @click.stop="deleteAddress(address.id)">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部添加按钮 -->
    <div class="bottom-bar">
      <button class="add-address-btn" @click="addAddress">
        <span>+</span> 新增收货地址
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { addressAPI } from '../api'

const route = useRoute()
const router = useRouter()
const addresses = ref([])
const loading = ref(false)
const fromCheckout = ref(false)

onMounted(() => {
  // 判断是否从结算页面跳转过来
  fromCheckout.value = route.query.from === 'checkout'
  loadAddresses()
})

const loadAddresses = async () => {
  loading.value = true
  try {
    const response = await addressAPI.getAddresses()
    addresses.value = response.addresses || []
  } catch (error) {
    console.error('加载地址失败:', error)
    addresses.value = []
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.back()
}

const addAddress = () => {
  router.push(`/address/edit${fromCheckout.value ? '?from=checkout' : ''}`)
}

const editAddress = (address) => {
  router.push(`/address/edit/${address.id}${fromCheckout.value ? '?from=checkout' : ''}`)
}

const deleteAddress = async (id) => {
  if (!confirm('确定要删除该地址吗？')) return
  
  try {
    await addressAPI.deleteAddress(id)
    addresses.value = addresses.value.filter(addr => addr.id !== id)
    alert('删除成功')
  } catch (error) {
    console.error('删除地址失败:', error)
    alert('删除失败')
  }
}

const selectAddress = (address) => {
  if (fromCheckout.value) {
    // 从结算页面过来的，选择地址后返回
    localStorage.setItem('selectedAddress', JSON.stringify(address))
    router.back()
  }
}
</script>

<style scoped>
.address-list-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 140px;
}

.page-header {
  display: flex;
  align-items: center;
  background-color: white;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
}

.back-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #333;
  cursor: pointer;
  padding: 4px 8px;
}

.page-header h1 {
  font-size: 18px;
  color: #333;
  margin: 0;
  flex: 1;
  text-align: center;
}

.add-btn {
  background: none;
  border: none;
  color: #4CAF50;
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
}

/* 地址列表 */
.address-list {
  padding: 12px;
}

.loading,
.empty-addresses {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-addresses p {
  margin-bottom: 20px;
  font-size: 14px;
}

.empty-addresses button {
  padding: 10px 30px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  font-size: 14px;
}

/* 地址卡片 */
.address-card {
  background-color: white;
  border-radius: 12px;
  margin-bottom: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.address-card.is-default {
  border: 1px solid #4CAF50;
}

.address-content {
  padding: 16px;
}

.contact {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 8px;
}

.contact .name {
  font-size: 16px;
  font-weight: bold;
  color: #333;
}

.contact .phone {
  font-size: 14px;
  color: #666;
}

.default-tag {
  background-color: #4CAF50;
  color: white;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
}

.address-text {
  font-size: 14px;
  color: #666;
  line-height: 1.5;
}

.address-actions {
  display: flex;
  border-top: 1px solid #f5f5f5;
}

.address-actions button {
  flex: 1;
  padding: 12px;
  border: none;
  background: none;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.3s ease;
}

.edit-btn {
  color: #4CAF50;
  border-right: 1px solid #f5f5f5;
}

.edit-btn:hover {
  background-color: #f5f5f5;
}

.delete-btn {
  color: #ff4757;
}

.delete-btn:hover {
  background-color: #fff5f5;
}

/* 底部添加按钮 */
.bottom-bar {
  position: fixed;
  bottom: 56px;
  left: 0;
  right: 0;
  background-color: white;
  padding: 12px 16px;
  border-top: 1px solid #eee;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.add-address-btn {
  width: 100%;
  padding: 14px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.add-address-btn span {
  font-size: 20px;
}

.add-address-btn:hover {
  background-color: #45a049;
}

/* 隐藏底部TabBar */
:global(.tab-bar) {
  display: none !important;
}
</style>
