<template>
  <div class="order-search-results">
    <div class="page-header">
      <h1>订单搜索</h1>
      <div class="search-box">
        <input type="text" v-model="keyword" placeholder="搜索订单号/商品名称" @keyup.enter="handleSearch" />
        <button @click="handleSearch">搜索</button>
      </div>
    </div>

    <!-- 订单状态筛选 -->
    <div class="status-tabs">
      <span :class="['tab', activeStatus === '' ? 'active' : '']" @click="changeStatus('')">全部</span>
      <span :class="['tab', activeStatus === 'pending' ? 'active' : '']" @click="changeStatus('pending')">待支付</span>
      <span :class="['tab', activeStatus === 'paid' ? 'active' : '']" @click="changeStatus('paid')">已支付</span>
      <span :class="['tab', activeStatus === 'shipped' ? 'active' : '']" @click="changeStatus('shipped')">已发货</span>
      <span :class="['tab', activeStatus === 'completed' ? 'active' : '']" @click="changeStatus('completed')">已完成</span>
      <span :class="['tab', activeStatus === 'cancelled' ? 'active' : '']" @click="changeStatus('cancelled')">已取消</span>
    </div>

    <!-- 搜索结果列表 -->
    <div class="order-list">
      <div v-if="loading" class="loading">搜索中...</div>

      <div v-else-if="orders.length > 0">
        <div v-for="order in orders" :key="order.order_id" class="order-card">
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

      <div v-else-if="hasSearched" class="empty">
        <div class="empty-icon">📋</div>
        <p>未找到相关订单</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { searchAPI, orderAPI, paymentAPI } from '../api'

const route = useRoute()
const router = useRouter()
const keyword = ref('')
const activeStatus = ref('')
const orders = ref([])
const loading = ref(false)
const hasSearched = ref(false)
const currentPage = ref(1)
const hasMore = ref(false)

const defaultImage = 'https://via.placeholder.com/80x80?text=No+Image'

onMounted(() => {
  keyword.value = route.query.keyword || ''
  if (keyword.value) {
    handleSearch()
  }
})

const handleSearch = () => {
  currentPage.value = 1
  orders.value = []
  fetchOrders()
}

const changeStatus = (status) => {
  activeStatus.value = status
  handleSearch()
}

const fetchOrders = async () => {
  if (!keyword.value.trim()) return
  loading.value = true
  try {
    const params = {
      keyword: keyword.value,
      page: currentPage.value,
      page_size: 10
    }
    if (activeStatus.value) params.status = activeStatus.value
    const response = await searchAPI.searchOrders(params)
    const items = response.orders || response.items || response.data || []
    if (currentPage.value === 1) {
      orders.value = items
    } else {
      orders.value = [...orders.value, ...items]
    }
    hasMore.value = items.length >= 10
    hasSearched.value = true
  } catch (error) {
    console.error('搜索订单失败:', error)
    hasSearched.value = true
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  currentPage.value++
  fetchOrders()
}

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

const viewOrderDetail = (orderNo) => {
  router.push(`/order/${orderNo}`)
}

const payOrder = async (order) => {
  try {
    await paymentAPI.fakePay(order.order_no)
    router.push(`/order/${order.order_no}`)
  } catch (error) {
    console.error('支付失败:', error)
    alert('支付失败: ' + (error.message || '请稍后重试'))
    handleSearch()
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
</script>

<style scoped>
.order-search-results {
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
  margin: 0 0 12px 0;
}

.search-box {
  display: flex;
  gap: 8px;
}

.search-box input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 20px;
  font-size: 14px;
  outline: none;
}

.search-box input:focus {
  border-color: #4CAF50;
}

.search-box button {
  padding: 8px 20px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  font-size: 14px;
}

/* 状态筛选标签 */
.status-tabs {
  display: flex;
  gap: 12px;
  background-color: white;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  overflow-x: auto;
}

.tab {
  padding: 6px 16px;
  border-radius: 20px;
  cursor: pointer;
  font-size: 14px;
  white-space: nowrap;
  background: #f5f5f5;
  color: #666;
}

.tab.active {
  background: #4CAF50;
  color: white;
}

/* 订单列表 */
.order-list {
  padding: 12px;
}

.loading,
.empty {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.empty p {
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
  cursor: pointer;
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
