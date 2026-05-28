<template>
  <div class="activity-order-confirm-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>确认订单</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else class="order-content">
      <div v-if="!productInfo" class="loading">未找到商品信息</div>
      <template v-else>
        <div class="address-section" @click="selectAddress">
          <div v-if="selectedAddress" class="address-info">
            <div class="contact">
              <span class="name">{{ selectedAddress.name }}</span>
              <span class="phone">{{ selectedAddress.phone }}</span>
            </div>
            <div class="address-text">
              {{ selectedAddress.province }}{{ selectedAddress.city }}{{ selectedAddress.district }}{{ selectedAddress.detail_address }}
            </div>
          </div>
          <div v-else class="no-address">
            <span>请选择收货地址</span>
          </div>
          <span class="arrow">›</span>
        </div>

        <div class="products-section">
          <div class="section-title">商品信息</div>
          <div class="product-list">
            <div class="product-item">
              <img :src="productInfo.main_image || defaultImage" :alt="productInfo.product_name" />
              <div class="product-info">
                <h4>{{ productInfo.product_name }}</h4>
                <p class="sku" v-if="productInfo.sku_code">{{ productInfo.sku_code }}</p>
                <p class="activity-tag" v-if="productInfo.activity_type">限时活动</p>
                <div class="price-quantity">
                  <span class="price">¥{{ formatPrice(productInfo.price) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="delivery-section">
          <div class="section-title">配送方式</div>
          <div class="delivery-info">
            <span class="delivery-text">包邮</span>
            <span class="delivery-price">¥0.00</span>
          </div>
        </div>

        <div class="amount-section">
          <div class="amount-item">
            <span>商品总额</span>
            <span>¥{{ formatPrice(totalAmount) }}</span>
          </div>
          <div class="amount-item">
            <span>运费</span>
            <span>¥{{ formatPrice(shippingFee) }}</span>
          </div>
          <div class="amount-item total">
            <span>应付总额</span>
            <span class="total-amount">¥{{ formatPrice(finalAmount) }}</span>
          </div>
        </div>
      </template>
    </div>

    <div class="bottom-bar">
      <div class="total-info">
        <span class="total-label">合计：</span>
        <span class="total-price">¥{{ formatPrice(finalAmount) }}</span>
      </div>
      <button class="submit-btn" @click="submitOrder" :disabled="loading || !productInfo">提交订单</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { activityAPI, activityOrderAPI, addressAPI } from '../api'

const route = useRoute()
const router = useRouter()

const productInfo = ref(null)
const selectedAddress = ref(null)
const loading = ref(false)

const activityId = ref(0)
const skuId = ref(0)
const addressId = ref(0)

const defaultImage = 'https://via.placeholder.com/80x80?text=No+Image'

const shippingFee = 0

const totalAmount = computed(() => {
  if (!productInfo.value) return 0
  return productInfo.value.price || 0
})

const finalAmount = computed(() => {
  return totalAmount.value + shippingFee
})

onMounted(async () => {
  activityId.value = parseInt(route.query.activity_id) || 0
  skuId.value = parseInt(route.query.sku_id) || 0

  if (!activityId.value || !skuId.value) {
    return
  }

  await loadProductInfo()
  await loadAddress()
})

const loadProductInfo = async () => {
  loading.value = true
  try {
    const skuDetail = await activityAPI.getActivitySkuDetail(activityId.value, skuId.value)
    if (skuDetail) {
      productInfo.value = {
        product_id: skuDetail.product_id,
        product_name: skuDetail.product_name || '活动商品',
        sku_code: skuDetail.sku_code,
        price: skuDetail.price || 0,
        main_image: skuDetail.main_image || defaultImage,
        stock: skuDetail.stock || 0
      }
    }
  } catch (error) {
    console.error('获取活动商品SKU详情失败:', error)
    productInfo.value = null
  } finally {
    loading.value = false
  }
}

const loadAddress = async () => {
  const savedAddress = localStorage.getItem('selectedAddress')
  if (savedAddress) {
    const address = JSON.parse(savedAddress)
    selectedAddress.value = address
    addressId.value = address.id
    localStorage.removeItem('selectedAddress')
  } else {
    try {
      const response = await addressAPI.getDefaultAddress()
      if (response && response.address) {
        selectedAddress.value = response.address
        addressId.value = response.address.id
      }
    } catch (error) {
      console.error('加载默认地址失败:', error)
    }
  }
}

const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toFixed(2)
}

const selectAddress = () => {
  router.push('/addresses?from=activity-checkout')
}

const goBack = () => {
  router.back()
}

const submitOrder = async () => {
  if (!selectedAddress.value) {
    alert('请选择收货地址')
    return
  }

  if (!productInfo.value) {
    alert('商品信息加载失败')
    return
  }

  try {
    const orderData = {
      activity_id: activityId.value,
      address_id: addressId.value,
      items: [
        {
          product_id: productInfo.value.product_id,
          sku_id: skuId.value,
          quantity: 1
        }
      ]
    }

    await activityOrderAPI.createActivityOrder(orderData)

    alert('订单已提交，正在处理中...')
    router.push('/activity/orders')
  } catch (error) {
    console.error('提交订单失败:', error)
    alert('提交订单失败，请重试')
  }
}
</script>

<style scoped>
.activity-order-confirm-page {
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
}

.loading {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.address-section {
  display: flex;
  align-items: center;
  background-color: white;
  padding: 16px;
  margin-bottom: 12px;
  cursor: pointer;
}

.address-info {
  flex: 1;
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

.address-text {
  font-size: 14px;
  color: #666;
  line-height: 1.5;
}

.no-address {
  flex: 1;
  color: #999;
  font-size: 14px;
}

.arrow {
  color: #999;
  font-size: 18px;
  margin-left: 8px;
}

.products-section,
.delivery-section,
.amount-section {
  background-color: white;
  padding: 16px;
  margin-bottom: 12px;
}

.section-title {
  font-size: 15px;
  font-weight: bold;
  color: #333;
  margin-bottom: 12px;
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

.activity-tag {
  font-size: 12px;
  color: #ff4757;
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

.quantity-control {
  display: flex;
  align-items: center;
  gap: 8px;
}

.quantity-control button {
  width: 28px;
  height: 28px;
  border: 1px solid #ddd;
  background: white;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.quantity-control button:disabled {
  color: #ccc;
  cursor: not-allowed;
}

.quantity-control .quantity {
  min-width: 24px;
  text-align: center;
  font-size: 14px;
  color: #333;
}

.delivery-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background-color: #f9f9f9;
  border-radius: 8px;
}

.delivery-text {
  font-size: 14px;
  color: #333;
}

.delivery-price {
  font-size: 14px;
  color: #4CAF50;
  font-weight: bold;
}

.amount-item {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.amount-item.total {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #eee;
  margin-bottom: 0;
}

.total-amount {
  color: #ff4757;
  font-weight: bold;
}

.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  height: 60px;
  background-color: white;
  padding: 8px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid #eee;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  box-sizing: border-box;
}

.total-info {
  display: flex;
  align-items: baseline;
}

.total-label {
  font-size: 14px;
  color: #333;
}

.total-price {
  font-size: 20px;
  color: #ff4757;
  font-weight: bold;
}

.submit-btn {
  padding: 12px 32px;
  background-color: #ff4757;
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  cursor: pointer;
}

.submit-btn:hover:not(:disabled) {
  background-color: #e84118;
}

.submit-btn:disabled {
  background-color: #ccc;
  cursor: not-allowed;
}
</style>
