<template>
  <div class="order-confirm-page">
    <div class="page-header">
      <button class="back-btn" @click="goBack">←</button>
      <h1>确认订单</h1>
    </div>

    <div v-if="loading" class="loading">加载中...</div>
    <div v-else class="order-content">
      <!-- 收货地址 -->
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

      <!-- 商品列表 -->
      <div class="products-section">
        <div class="section-title">商品信息</div>
        <div class="product-list">
          <div v-for="item in cartItems" :key="item.product_id" class="product-item">
            <img :src="item.main_image || defaultImage" :alt="item.product_name" />
            <div class="product-info">
              <h4>{{ item.product_name }}</h4>
              <p class="sku" v-if="item.sku_code">{{ item.sku_code }}</p>
              <div class="price-quantity">
                <span class="price">¥{{ formatPrice(item.price) }}</span>
                <span class="quantity">x{{ item.quantity }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 配送方式 - 默认包邮 -->
      <div class="delivery-section">
        <div class="section-title">配送方式</div>
        <div class="delivery-info">
          <span class="delivery-text">包邮</span>
          <span class="delivery-price">¥0.00</span>
        </div>
      </div>

      <!-- 订单备注 -->
      <div class="remark-section">
        <div class="section-title">订单备注</div>
        <textarea 
          v-model="remark" 
          placeholder="请输入订单备注（选填）"
          rows="2"
        ></textarea>
      </div>

      <!-- 金额明细 -->
      <div class="amount-section">
        <div class="amount-item">
          <span>商品总额</span>
          <span>¥{{ formatPrice(totalAmount) }}</span>
        </div>
        <div class="amount-item">
          <span>运费</span>
          <span>¥{{ formatPrice(shippingFee) }}</span>
        </div>
        <div class="amount-item" v-if="discount > 0">
          <span>优惠</span>
          <span class="discount">-¥{{ formatPrice(discount) }}</span>
        </div>
        <div class="amount-item total">
          <span>应付总额</span>
          <span class="total-amount">¥{{ formatPrice(finalAmount) }}</span>
        </div>
      </div>
    </div>

    <!-- 底部结算栏 -->
    <div class="bottom-bar">
      <div class="total-info">
        <span class="total-label">合计：</span>
        <span class="total-price">¥{{ formatPrice(finalAmount) }}</span>
      </div>
      <button class="submit-btn" @click="submitOrder">提交订单</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { cartAPI, addressAPI, orderAPI } from '../api'

const route = useRoute()
const router = useRouter()

const cartItems = ref([])
const selectedAddress = ref(null)
const remark = ref('')
const loading = ref(false)
const discount = ref(0)

const defaultImage = 'https://via.placeholder.com/80x80?text=No+Image'

// 运费固定为0（包邮）
const shippingFee = 0

const totalAmount = computed(() => {
  return cartItems.value.reduce((sum, item) => sum + (item.price * item.quantity), 0)
})

const finalAmount = computed(() => {
  return totalAmount.value + shippingFee - discount.value
})

onMounted(async () => {
  loadCartItems()
  
  // 检查是否有从地址页面选择的地址
  const savedAddress = localStorage.getItem('selectedAddress')
  if (savedAddress) {
    selectedAddress.value = JSON.parse(savedAddress)
    localStorage.removeItem('selectedAddress')
  } else {
    // 没有选择的地址，加载默认地址
    await loadDefaultAddress()
  }
})

const loadCartItems = async () => {
  loading.value = true
  try {
    const response = await cartAPI.getCart()
    cartItems.value = response.items || []
    
    if (cartItems.value.length === 0) {
      alert('购物车为空')
      router.push('/cart')
    }
  } catch (error) {
    console.error('加载购物车失败:', error)
    cartItems.value = []
    alert('加载购物车失败')
    router.push('/cart')
  } finally {
    loading.value = false
  }
}

const loadDefaultAddress = async () => {
  try {
    const response = await addressAPI.getDefaultAddress()
    if (response && response.address) {
      selectedAddress.value = response.address
    }
  } catch (error) {
    console.error('加载默认地址失败:', error)
  }
}

const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toFixed(2)
}

const selectAddress = () => {
  router.push('/addresses?from=checkout')
}

const goBack = () => {
  router.back()
}

const submitOrder = async () => {
  if (!selectedAddress.value) {
    alert('请选择收货地址')
    return
  }

  if (cartItems.value.length === 0) {
    alert('购物车为空')
    return
  }

  try {
    const orderData = {
      address_id: selectedAddress.value.id,
      items: cartItems.value.map(item => ({
        product_id: item.product_id,
        sku_id: item.sku_id || 0,
        quantity: item.quantity
      })),
      remark: remark.value
    }

    const response = await orderAPI.createOrder(orderData)
    alert('订单提交成功')
    
    // 跳转到支付页面或订单详情
    if (response && response.order_id) {
      router.push(`/order/${response.order_id}?action=pay`)
    } else {
      router.push('/orders')
    }
  } catch (error) {
    console.error('提交订单失败:', error)
    alert('提交订单失败，请重试')
  }
}
</script>

<style scoped>
.order-confirm-page {
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

/* 地址区域 */
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

/* 商品区域 */
.products-section,
.delivery-section,
.remark-section,
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

/* 配送方式 */
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

/* 备注 */
.remark-section textarea {
  width: 100%;
  padding: 12px;
  border: 1px solid #eee;
  border-radius: 8px;
  font-size: 14px;
  resize: none;
  outline: none;
}

.remark-section textarea:focus {
  border-color: #4CAF50;
}

/* 金额明细 */
.amount-item {
  display: flex;
  justify-content: space-between;
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.amount-item .discount {
  color: #ff4757;
}

/* 底部结算栏 */
.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  background-color: white;
  padding: 12px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-top: 1px solid #eee;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
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

.submit-btn:hover {
  background-color: #e84118;
}
</style>
