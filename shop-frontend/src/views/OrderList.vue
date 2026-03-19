<template>
  <div class="order-list-page">
    <div class="page-header">
      <h1>我的订单</h1>
    </div>

    <!-- 订单状态筛选 -->
    <div class="filter-tabs">
      <div 
        v-for="tab in tabs" 
        :key="tab.value"
        class="tab-item"
        :class="{ active: currentTab === tab.value }"
        @click="currentTab = tab.value"
      >
        {{ tab.label }}
      </div>
    </div>

    <!-- 订单列表 -->
    <div class="order-list">
      <div v-if="loading" class="loading">加载中...</div>
      <div v-else-if="filteredOrders.length === 0" class="empty-orders">
        <div class="empty-icon">📋</div>
        <p>暂无{{ currentTab === 'all' ? '' : getStatusLabel(currentTab) }}订单</p>
        <button @click="goShopping">去购物</button>
      </div>
      <div v-else>
        <div v-for="order in filteredOrders" :key="order.order_id" class="order-card">
          <div class="order-header">
            <div class="order-info">
              <span class="order-no">订单号：{{ order.order_no }}</span>
              <span class="order-date">{{ formatDate(order.created_at) }}</span>
            </div>
            <span class="order-status" :class="getStatusClass(order.status)">
              {{ getStatusLabel(order.status) }}
            </span>
          </div>

          <div class="order-items" @click="viewOrderDetail(order.order_no)">
            <div v-for="(item, index) in order.items" :key="index" class="order-item">
              <img :src="item.product_image || defaultImage" :alt="item.product_name" />
              <div class="item-info">
                <h4>{{ item.product_name }}</h4>
                <p class="item-sku" v-if="item.sku_attributes && item.sku_attributes !== '{}'">{{ formatSku(item.sku_attributes) }}</p>
                <p class="item-price">¥{{ formatPrice(item.price) }} x {{ item.quantity }}</p>
              </div>
            </div>
          </div>

          <div class="order-footer">
            <div class="order-total">
              <span>共{{ getTotalItems(order.items) }}件商品</span>
              <span class="total-amount">
                合计：<strong>¥{{ formatPrice(order.amount) }}</strong>
              </span>
            </div>
            <div class="order-actions">
              <button
                v-if="order.status === 'pending'"
                class="btn-primary"
                @click="payOrder(order)"
              >
                立即支付
              </button>
              <button 
                v-if="order.status === 'pending'" 
                class="btn-default"
                @click="cancelOrder(order.order_no)"
              >
                取消订单
              </button>
              <button 
                v-if="order.status === 'shipped'" 
                class="btn-primary"
                @click="confirmReceipt(order.order_no)"
              >
                确认收货
              </button>
              <button 
                class="btn-default"
                @click="viewOrderDetail(order.order_no)"
              >
                查看详情
              </button>
            </div>
          </div>
        </div>

        <!-- 加载更多 -->
        <div v-if="hasMore" class="load-more" @click="loadMore">加载更多</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { customerAPI, orderAPI, paymentAPI } from '../api'

const router = useRouter()
const orders = ref([])
const loading = ref(false)
const currentTab = ref('all')
const currentPage = ref(1)
const pageSize = 10
const hasMore = ref(true)

const defaultImage = 'https://via.placeholder.com/80x80?text=No+Image'

const tabs = [
  { label: '全部', value: 'all' },
  { label: '待付款', value: 'pending' },
  { label: '待发货', value: 'paid' },
  { label: '待收货', value: 'shipped' },
  { label: '已完成', value: 'completed' }
]

const filteredOrders = computed(() => {
  if (currentTab.value === 'all') {
    return orders.value
  }
  return orders.value.filter(order => order.status === currentTab.value)
})

const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toFixed(2)
}

const formatDate = (date) => {
  if (!date) return ''
  const d = new Date(date)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

const formatSku = (skuAttrs) => {
  if (!skuAttrs) return ''
  if (typeof skuAttrs === 'string') {
    try {
      skuAttrs = JSON.parse(skuAttrs)
    } catch {
      return skuAttrs
    }
  }
  return Object.entries(skuAttrs).map(([k, v]) => `${k}: ${v}`).join(' ')
}

const getTotalItems = (items) => {
  if (!items || !Array.isArray(items)) return 0
  return items.reduce((sum, item) => sum + (item.quantity || 0), 0)
}

const getStatusLabel = (status) => {
  const statusMap = {
    'pending': '待付款',
    'paid': '待发货',
    'shipped': '待收货',
    'completed': '已完成',
    'cancelled': '已取消'
  }
  return statusMap[status] || status
}

const getStatusClass = (status) => {
  const classMap = {
    'pending': 'status-pending',
    'paid': 'status-paid',
    'shipped': 'status-shipped',
    'completed': 'status-completed',
    'cancelled': 'status-cancelled'
  }
  return classMap[status] || ''
}

const goShopping = () => {
  router.push('/')
}

const viewOrderDetail = (orderNo) => {
  router.push(`/order/${orderNo}`)
}

const payOrder = async (order) => {
  try {
    await paymentAPI.fakePay(order.order_no)
    // 支付成功，跳转到订单详情页
    router.push(`/order/${order.order_no}`)
  } catch (error) {
    // 支付失败，显示弹窗
    console.error('支付失败:', error)
    alert('支付失败: ' + (error.message || '请稍后重试'))
    // 刷新订单列表
    loadOrders()
  }
}

const cancelOrder = async (orderNo) => {
  if (!confirm('确定要取消该订单吗？')) return
  try {
    await orderAPI.cancelOrder(orderNo)
    const order = orders.value.find(o => o.order_no === orderNo)
    if (order) {
      order.status = 'cancelled'
    }
    alert('订单已取消')
  } catch (error) {
    console.error('取消订单失败:', error)
    alert('取消订单失败')
  }
}

const confirmReceipt = async (orderNo) => {
  if (!confirm('确认已收到商品？')) return
  try {
    await orderAPI.confirmReceipt(orderNo)
    const order = orders.value.find(o => o.order_no === orderNo)
    if (order) {
      order.status = 'completed'
    }
    alert('确认收货成功')
  } catch (error) {
    console.error('确认收货失败:', error)
    alert('确认收货失败')
  }
}

const loadOrders = async (page = 1, append = false) => {
  if (loading.value) return
  loading.value = true

  try {
    // 构建查询参数
    const params = { page, limit: pageSize }
    if (currentTab.value !== 'all') {
      params.status = currentTab.value
    }

    const response = await customerAPI.getOrders(params)
    const newOrders = response.orders || []
    const total = response.total || 0

    if (append) {
      orders.value = [...orders.value, ...newOrders]
    } else {
      orders.value = newOrders
    }

    hasMore.value = orders.value.length < total
    currentPage.value = page
  } catch (error) {
    console.error('加载订单失败:', error)
    if (!append) {
      orders.value = []
      hasMore.value = false
    }
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  if (!loading.value && hasMore.value) {
    loadOrders(currentPage.value + 1, true)
  }
}

onMounted(() => {
  loadOrders()
})

watch(currentTab, () => {
  currentPage.value = 1
  hasMore.value = true
  loadOrders(1, false)
})
</script>

<style scoped>
.order-list-page {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.page-header {
  background-color: white;
  padding: 16px;
  text-align: center;
  border-bottom: 1px solid #eee;
}

.page-header h1 {
  font-size: 18px;
  color: #333;
  margin: 0;
}

/* 筛选标签 */
.filter-tabs {
  display: flex;
  background-color: white;
  padding: 0 12px;
  border-bottom: 1px solid #eee;
  overflow-x: auto;
}

.tab-item {
  flex: 1;
  padding: 14px 8px;
  text-align: center;
  font-size: 14px;
  color: #666;
  cursor: pointer;
  white-space: nowrap;
  border-bottom: 2px solid transparent;
  transition: all 0.3s ease;
}

.tab-item.active {
  color: #4CAF50;
  border-bottom-color: #4CAF50;
}

/* 订单列表 */
.order-list {
  padding: 12px;
}

.loading,
.empty-orders {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty-orders p {
  margin-bottom: 20px;
  font-size: 14px;
}

.empty-orders button {
  padding: 10px 30px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  font-size: 14px;
}

/* 订单卡片 */
.order-card {
  background-color: white;
  border-radius: 12px;
  margin-bottom: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #f5f5f5;
}

.order-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.order-no {
  font-size: 13px;
  color: #666;
}

.order-date {
  font-size: 12px;
  color: #999;
}

.order-status {
  font-size: 13px;
  font-weight: bold;
}

.status-pending {
  color: #ff9800;
}

.status-paid {
  color: #2196F3;
}

.status-shipped {
  color: #9C27B0;
}

.status-completed {
  color: #4CAF50;
}

.status-cancelled {
  color: #999;
}

/* 订单商品 */
.order-items {
  padding: 12px 16px;
}

.order-item {
  display: flex;
  gap: 12px;
  padding: 8px 0;
}

.order-item img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 8px;
  background-color: #f5f5f5;
}

.item-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.item-info h4 {
  font-size: 14px;
  color: #333;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 0;
}

.item-sku {
  font-size: 12px;
  color: #999;
  margin: 4px 0;
}

.item-price {
  font-size: 13px;
  color: #666;
}

/* 订单底部 */
.order-footer {
  padding: 12px 16px;
  border-top: 1px solid #f5f5f5;
  background-color: #fafafa;
}

.order-total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  font-size: 13px;
  color: #666;
}

.total-amount {
  font-size: 14px;
}

.total-amount strong {
  color: #ff4757;
  font-size: 16px;
}

.order-actions {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
}

.order-actions button {
  padding: 8px 16px;
  border-radius: 16px;
  font-size: 13px;
  cursor: pointer;
  border: none;
  transition: all 0.3s ease;
}

.btn-primary {
  background-color: #ff4757;
  color: white;
}

.btn-primary:hover {
  background-color: #e84118;
}

.btn-default {
  background-color: white;
  color: #666;
  border: 1px solid #ddd;
}

.btn-default:hover {
  background-color: #f5f5f5;
}

/* 加载更多 */
.load-more {
  text-align: center;
  padding: 16px;
  color: #4CAF50;
  font-size: 14px;
  cursor: pointer;
}

.load-more:hover {
  text-decoration: underline;
}
</style>
