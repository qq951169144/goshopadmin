<template>
  <div class="order-detail-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>订单详情</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="!order" class="error">订单不存在</div>
    <div v-else class="order-content">
      <div class="status-section" :class="getStatusClass(order.status)">
        <div class="status-icon">{{ getStatusIcon(order.status) }}</div>
        <div class="status-text">{{ getStatusLabel(order.status) }}</div>
        <div class="status-desc">{{ getStatusDesc(order.status) }}</div>
      </div>

      <div class="address-section">
        <div class="section-title">收货地址</div>
        <div class="address-info" v-if="order.address">
          <div class="contact">
            <span class="name">{{ order.address.name }}</span>
            <span class="phone">{{ order.address.phone }}</span>
          </div>
          <div class="address">
            {{ order.address.province }}{{ order.address.city }}{{ order.address.district }}{{ order.address.detail_address }}
          </div>
        </div>
        <div class="address-info" v-else>
          <div class="contact">
            <span class="name">暂无地址信息</span>
          </div>
        </div>
      </div>

      <div class="products-section">
        <div class="section-title">商品信息</div>
        <div class="product-list">
          <div v-for="(item, index) in order.items" :key="index" class="product-item">
            <img :src="item.product_image || item.image || defaultImage" :alt="item.product_name" />
            <div class="product-info">
              <h4>{{ item.product_name }}</h4>
              <p class="sku" v-if="item.sku_code">{{ item.sku_code }}</p>
              <p class="sku" v-else-if="item.sku_attributes && item.sku_attributes !== '{}'">{{ formatSku(item.sku_attributes) }}</p>
              <div class="price-quantity">
                <span class="price">¥{{ formatPrice(item.price) }}</span>
                <span class="quantity">x{{ item.quantity }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="order-info-section">
        <div class="section-title">订单信息</div>
        <div class="info-list">
          <div class="info-item">
            <span class="label">订单编号</span>
            <span class="value">{{ order.order_no }}</span>
          </div>
          <div class="info-item">
            <span class="label">下单时间</span>
            <span class="value">{{ formatDateTime(order.created_at) }}</span>
          </div>
          <div class="info-item">
            <span class="label">支付方式</span>
            <span class="value">{{ order.payment_method || '微信支付' }}</span>
          </div>
        </div>
      </div>

      <div class="amount-section">
        <div class="amount-item">
          <span class="label">商品总额</span>
          <span class="value">¥{{ formatPrice(order.total_amount) }}</span>
        </div>
        <div class="amount-item">
          <span class="label">运费</span>
          <span class="value">¥{{ formatPrice(order.shipping_fee || 0) }}</span>
        </div>
        <div class="amount-item" v-if="order.discount > 0">
          <span class="label">优惠</span>
          <span class="value discount">-¥{{ formatPrice(order.discount) }}</span>
        </div>
        <div class="amount-total">
          <span class="label">实付款</span>
          <span class="value">¥{{ formatPrice(order.total_amount) }}</span>
        </div>
      </div>

      <div class="bottom-actions">
        <button
          v-if="order.status === 'pending'"
          class="btn-primary"
          @click="payOrder"
        >
          立即支付
        </button>
        <button
          v-if="order.status === 'pending'"
          class="btn-default"
          @click="cancelOrder"
        >
          取消订单
        </button>
        <button
          v-if="order.status === 'shipped'"
          class="btn-primary"
          @click="confirmReceipt"
        >
          确认收货
        </button>
        <button
          class="btn-default"
          @click="contactService"
        >
          联系客服
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { activityOrderAPI, paymentAPI } from '../api'

const route = useRoute()
const router = useRouter()
const order = ref(null)
const loading = ref(true)

const defaultImage = 'https://via.placeholder.com/80x80?text=No+Image'

const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toFixed(2)
}

const formatDateTime = (date) => {
  if (!date) return ''
  const d = new Date(date)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
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

const getStatusIcon = (status) => {
  const iconMap = {
    'pending': '⏰',
    'paid': '📦',
    'shipped': '🚚',
    'completed': '✅',
    'cancelled': '❌'
  }
  return iconMap[status] || '📋'
}

const getStatusDesc = (status) => {
  const descMap = {
    'pending': '请在30分钟内完成支付，超时订单将自动关闭',
    'paid': '商家正在备货中，请耐心等待',
    'shipped': '商品正在运输中，请注意查收',
    'completed': '交易已完成，感谢您的购买',
    'cancelled': '订单已取消'
  }
  return descMap[status] || ''
}

const goBack = () => {
  router.back()
}

const payOrder = async () => {
  if (!order.value || !order.value.order_no) {
    alert('订单信息不完整')
    return
  }
  try {
    await paymentAPI.fakePay(order.value.order_no)
    alert('支付成功')
    loadOrderDetail()
  } catch (error) {
    console.error('支付失败:', error)
    alert('支付失败')
  }
}

const cancelOrder = async () => {
  if (!confirm('确定要取消该订单吗？')) return
  try {
    await activityOrderAPI.cancelActivityOrder(order.value.id)
    order.value.status = 'cancelled'
    alert('订单已取消')
  } catch (error) {
    console.error('取消订单失败:', error)
    alert('取消订单失败')
  }
}

const confirmReceipt = async () => {
  if (!confirm('确认已收到商品？')) return
  try {
    await activityOrderAPI.confirmActivityOrder(order.value.id)
    order.value.status = 'completed'
    alert('确认收货成功')
  } catch (error) {
    console.error('确认收货失败:', error)
    alert('确认收货失败')
  }
}

const contactService = () => {
  alert('客服功能开发中...')
}

const loadOrderDetail = async () => {
  loading.value = true
  try {
    const id = route.params.id
    const response = await activityOrderAPI.getActivityOrderDetail(id)
    order.value = response
  } catch (error) {
    console.error('加载订单详情失败:', error)
    alert('加载订单详情失败')
    router.back()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadOrderDetail()
})
</script>

<style scoped>
.order-detail-page {
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 80px;
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
.error {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.status-section {
  padding: 24px 16px;
  text-align: center;
  color: white;
}

.status-pending {
  background: linear-gradient(135deg, #ff9800 0%, #f57c00 100%);
}

.status-paid {
  background: linear-gradient(135deg, #2196F3 0%, #1976D2 100%);
}

.status-shipped {
  background: linear-gradient(135deg, #9C27B0 0%, #7B1FA2 100%);
}

.status-completed {
  background: linear-gradient(135deg, #4CAF50 0%, #388E3C 100%);
}

.status-cancelled {
  background: linear-gradient(135deg, #9E9E9E 0%, #616161 100%);
}

.status-icon {
  font-size: 48px;
  margin-bottom: 8px;
}

.status-text {
  font-size: 20px;
  font-weight: bold;
  margin-bottom: 8px;
}

.status-desc {
  font-size: 13px;
  opacity: 0.9;
}

.address-section,
.products-section,
.order-info-section,
.amount-section {
  background-color: white;
  margin-top: 12px;
  padding: 16px;
}

.section-title {
  font-size: 15px;
  font-weight: bold;
  color: #333;
  margin-bottom: 12px;
}

.address-info {
  font-size: 14px;
}

.contact {
  margin-bottom: 8px;
}

.contact .name {
  font-weight: bold;
  color: #333;
  margin-right: 12px;
}

.contact .phone {
  color: #666;
}

.address {
  color: #666;
  line-height: 1.5;
}

.product-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.product-item {
  display: flex;
  gap: 12px;
}

.product-item img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 8px;
  background-color: #f5f5f5;
}

.product-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.product-info h4 {
  font-size: 14px;
  color: #333;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 0;
}

.sku {
  font-size: 12px;
  color: #999;
  margin: 4px 0;
}

.price-quantity {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.price {
  color: #ff4757;
  font-weight: bold;
}

.quantity {
  color: #999;
  font-size: 13px;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-item {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
}

.info-item .label {
  color: #999;
}

.info-item .value {
  color: #333;
}

.amount-item {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  margin-bottom: 8px;
}

.amount-item .label {
  color: #666;
}

.amount-item .value {
  color: #333;
}

.amount-item .discount {
  color: #ff4757;
}

.amount-total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #eee;
}

.amount-total .label {
  font-size: 14px;
  color: #333;
}

.amount-total .value {
  font-size: 20px;
  color: #ff4757;
  font-weight: bold;
}

.bottom-actions {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: white;
  padding: 12px 16px;
  display: flex;
  gap: 12px;
  justify-content: flex-end;
  border-top: 1px solid #eee;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
}

.bottom-actions button {
  padding: 10px 20px;
  border-radius: 20px;
  font-size: 14px;
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
</style>
