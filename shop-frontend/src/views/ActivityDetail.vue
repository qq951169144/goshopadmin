<template>
  <div class="activity-detail-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>活动详情</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error-message">{{ error }}</div>
    <template v-else-if="activity">
      <div class="activity-header">
        <div class="activity-name">{{ activity.name }}</div>
        <div class="activity-time">
          {{ formatDateTime(activity.start_time) }} 至 {{ formatDateTime(activity.end_time) }}
        </div>
        <div class="activity-status" :class="getStatusClass(activity.status)">
          {{ getStatusLabel(activity.status) }}
        </div>
      </div>

      <div class="products-section">
        <div class="section-title">活动商品</div>
        <div v-if="productsLoading" class="loading-more">加载中...</div>
        <div v-else-if="productsError" class="error-message">{{ productsError }}</div>
        <div v-else-if="skuList.length === 0" class="empty-message">暂无活动商品</div>
        <div v-else class="product-list">
          <div v-for="sku in skuList" :key="sku.sku_id" class="product-card">
            <div class="product-image">
              <img :src="sku.main_image || defaultImage" :alt="sku.product_name" />
            </div>
            <div class="product-info">
              <h3 class="product-name">{{ sku.product_name }}</h3>
              <p class="product-desc">{{ sku.description || '暂无描述' }}</p>
              <div class="product-meta">
                <span class="sku-code">SKU: {{ sku.sku_code }}</span>
              </div>
              <div class="price-stock">
                <div class="price">
                  <span class="price-symbol">¥</span>
                  <span class="price-value">{{ formatPrice(sku.price) }}</span>
                </div>
                <div class="stock">库存: {{ sku.stock }}</div>
              </div>
              <button
                class="buy-btn"
                :disabled="!canBuy(activity.status) || sku.stock <= 0"
                @click="handleBuy(sku)"
              >
                {{ getBuyBtnText(activity.status, sku.stock) }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { activityAPI } from '../api'

const route = useRoute()
const router = useRouter()
const activity = ref(null)
const skuList = ref([])
const loading = ref(true)
const productsLoading = ref(false)
const error = ref('')
const productsError = ref('')

const defaultImage = 'https://via.placeholder.com/120x120?text=No+Image'

const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toFixed(2)
}

const formatDateTime = (date) => {
  if (!date) return ''
  const d = new Date(date)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const getStatusLabel = (status) => {
  const statusMap = {
    'pending': '未开始',
    'ongoing': '进行中',
    'ended': '已结束'
  }
  return statusMap[status] || status
}

const getStatusClass = (status) => {
  const classMap = {
    'pending': 'status-pending',
    'ongoing': 'status-ongoing',
    'ended': 'status-ended'
  }
  return classMap[status] || ''
}

const canBuy = (status) => {
  return status === 'active'
}

const getBuyBtnText = (status, stock) => {
  if (stock <= 0) {
    return '已售罄'
  }
  if (status === 'ended') {
    return '已结束'
  }
  if (status === 'pending') {
    return '未开始'
  }
  return '抢购'
}

const goBack = () => {
  router.back()
}

const handleBuy = (sku) => {
  if (!canBuy(activity.value.status) || sku.stock <= 0) {
    return
  }
  router.push(`/activity/order/confirm?activity_id=${activity.value.id}&sku_id=${sku.sku_id}`)
}

const loadActivityDetail = async () => {
  loading.value = true
  error.value = ''
  try {
    const activityId = route.params.id
    const response = await activityAPI.getActivityDetail(activityId)
    activity.value = response
    if (response && response.id) {
      loadActivityProducts(response.id)
    }
  } catch (err) {
    console.error('加载活动详情失败:', err)
    error.value = '加载活动详情失败'
  } finally {
    loading.value = false
  }
}

const loadActivityProducts = async (activityId) => {
  productsLoading.value = true
  productsError.value = ''
  try {
    const skus = await activityAPI.getActivityProductSkus(activityId)
    skuList.value = skus || []
  } catch (err) {
    console.error('加载活动商品SKU失败:', err)
    productsError.value = '加载活动商品失败'
  } finally {
    productsLoading.value = false
  }
}

onMounted(() => {
  loadActivityDetail()
})
</script>

<style scoped>
.activity-detail-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 20px;
}

.page-header {
  display: flex;
  align-items: center;
  background-color: white;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  position: sticky;
  top: 0;
  z-index: 100;
}

.back-btn {
  background: none;
  border: none;
  font-size: 20px;
  color: #333;
  cursor: pointer;
  padding: 4px 8px;
  margin-right: 12px;
}

.page-header h1 {
  font-size: 18px;
  color: #333;
  margin: 0;
  flex: 1;
  text-align: center;
  margin-right: 40px;
}

.loading,
.loading-more {
  text-align: center;
  padding: 60px 20px;
  color: #999;
  font-size: 14px;
}

.loading-more {
  padding: 20px;
}

.error-message {
  text-align: center;
  padding: 20px;
  color: #ff4757;
  font-size: 14px;
  background-color: #fff5f5;
  border-radius: 8px;
  margin: 20px 16px;
}

.empty-message {
  text-align: center;
  padding: 40px 20px;
  color: #999;
  font-size: 14px;
  background-color: white;
  border-radius: 12px;
  margin: 0 16px;
}

.activity-header {
  background: linear-gradient(135deg, #ff4757 0%, #ff6b81 100%);
  color: white;
  padding: 20px 16px;
  margin-bottom: 12px;
}

.activity-name {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 12px;
}

.activity-time {
  font-size: 13px;
  opacity: 0.9;
  margin-bottom: 12px;
}

.activity-status {
  display: inline-block;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  background-color: rgba(255, 255, 255, 0.2);
}

.status-ongoing {
  background-color: #4CAF50;
  color: white;
}

.status-pending {
  background-color: #FF9800;
  color: white;
}

.status-ended {
  background-color: #9E9E9E;
  color: white;
}

.products-section {
  background-color: white;
  margin: 0 16px;
  border-radius: 12px;
  padding: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  color: #333;
  margin-bottom: 16px;
}

.product-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.product-card {
  display: flex;
  gap: 12px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.product-card:last-child {
  padding-bottom: 0;
  border-bottom: none;
}

.product-image {
  width: 100px;
  height: 100px;
  flex-shrink: 0;
  border-radius: 8px;
  overflow: hidden;
  background-color: #f5f5f5;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.product-name {
  font-size: 14px;
  color: #333;
  line-height: 1.4;
  margin: 0 0 4px 0;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.product-desc {
  font-size: 12px;
  color: #999;
  margin: 0 0 4px 0;
  display: -webkit-box;
  -webkit-line-clamp: 1;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.product-meta {
  margin-bottom: 4px;
}

.sku-code {
  font-size: 11px;
  color: #999;
}

.price-stock {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.price {
  color: #ff4757;
  font-weight: bold;
}

.price-symbol {
  font-size: 12px;
}

.price-value {
  font-size: 16px;
}

.stock {
  font-size: 12px;
  color: #666;
}

.buy-btn {
  width: 100%;
  padding: 10px;
  border-radius: 20px;
  border: none;
  font-size: 14px;
  cursor: pointer;
  background-color: #ff4757;
  color: white;
  transition: all 0.3s ease;
}

.buy-btn:hover:not(:disabled) {
  background-color: #e84118;
}

.buy-btn:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}
</style>
