<template>
  <div class="cart">
    <h1>购物车</h1>
    
    <div v-if="cartItems.length === 0" class="empty-cart">
      <div class="empty-icon">🛒</div>
      <p>购物车是空的</p>
      <button @click="goShopping">去购物</button>
    </div>
    
    <div v-else>
      <div class="cart-items">
        <div v-for="item in cartItems" :key="item.id" class="cart-item">
          <img :src="item.main_image || defaultImage" :alt="item.product_name" />
          <div class="item-details">
            <h3>{{ item.product_name }}</h3>
            <p class="sku-code" v-if="item.sku_code">规格: {{ item.sku_code }}</p>
            <p class="price">¥{{ formatPrice(item.price) }}</p>
            <div class="quantity-control">
              <button @click="decreaseQuantity(item)" :disabled="item.quantity <= 1">-</button>
              <span>{{ item.quantity }}</span>
              <button @click="increaseQuantity(item)">+</button>
            </div>
          </div>
          <button class="remove-btn" @click="removeItem(item.id)">删除</button>
        </div>
      </div>
      
      <div class="cart-summary">
        <div class="total">
          <span>合计：</span>
          <span class="total-price">¥{{ formatPrice(totalPrice) }}</span>
        </div>
        <button class="checkout-btn" @click="checkout">去结算</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { cartAPI } from '../api'

const router = useRouter()
const cartItems = ref([])

const defaultImage = 'https://via.placeholder.com/80x80?text=No+Image'

const totalPrice = computed(() => {
  return cartItems.value.reduce((sum, item) => sum + (item.price * item.quantity), 0)
})

const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toFixed(2)
}

const loadCart = async () => {
  try {
    const response = await cartAPI.getCart()
    cartItems.value = response.items || []
  } catch (error) {
    console.error('加载购物车失败:', error)
    cartItems.value = []
  }
}

const increaseQuantity = async (item) => {
  try {
    await cartAPI.updateCartItem(item.id, { quantity: item.quantity + 1 })
    item.quantity++
  } catch (error) {
    console.error('更新数量失败:', error)
    item.quantity++
  }
}

const decreaseQuantity = async (item) => {
  if (item.quantity <= 1) return
  try {
    await cartAPI.updateCartItem(item.id, { quantity: item.quantity - 1 })
    item.quantity--
  } catch (error) {
    console.error('更新数量失败:', error)
    item.quantity--
  }
}

const removeItem = async (id) => {
  if (!confirm('确定要删除该商品吗？')) return
  try {
    await cartAPI.removeCartItem(id)
    cartItems.value = cartItems.value.filter(item => item.id !== id)
  } catch (error) {
    console.error('删除商品失败:', error)
    cartItems.value = cartItems.value.filter(item => item.id !== id)
  }
}

const goShopping = () => {
  router.push('/')
}

const checkout = () => {
  if (cartItems.value.length === 0) {
    alert('购物车为空')
    return
  }
  router.push('/checkout')
}

onMounted(() => {
  loadCart()
})
</script>

<style scoped>
.cart {
  padding: 16px;
  background-color: #f5f5f5;
  min-height: 100vh;
  padding-bottom: 100px;
}

h1 {
  margin-bottom: 20px;
  color: #333;
  font-size: 20px;
}

/* 空购物车 */
.empty-cart {
  text-align: center;
  padding: 60px 20px;
  color: #999;
}

.empty-icon {
  font-size: 64px;
  margin-bottom: 16px;
}

.empty-cart p {
  margin-bottom: 20px;
  font-size: 16px;
}

.empty-cart button {
  padding: 12px 30px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 24px;
  cursor: pointer;
  font-size: 16px;
}

/* 购物车商品列表 */
.cart-items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cart-item {
  display: flex;
  gap: 12px;
  background-color: white;
  padding: 16px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.cart-item img {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 8px;
  background-color: #f5f5f5;
}

.item-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.item-details h3 {
  font-size: 14px;
  color: #333;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin: 0;
}

.price {
  color: #ff4757;
  font-weight: bold;
  font-size: 16px;
  margin: 8px 0;
}

.sku-code {
  color: #999;
  font-size: 12px;
  margin: 4px 0;
}

.quantity-control {
  display: flex;
  align-items: center;
  gap: 12px;
}

.quantity-control button {
  width: 28px;
  height: 28px;
  border: 1px solid #ddd;
  background-color: white;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.quantity-control button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.quantity-control span {
  font-size: 14px;
  min-width: 30px;
  text-align: center;
}

.remove-btn {
  background: none;
  border: none;
  color: #999;
  font-size: 13px;
  cursor: pointer;
  padding: 4px 8px;
  align-self: flex-start;
}

.remove-btn:hover {
  color: #ff4757;
}

/* 购物车结算栏 */
.cart-summary {
  position: fixed;
  bottom: 60px;
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

.total {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.total span:first-child {
  font-size: 14px;
  color: #666;
}

.total-price {
  font-size: 20px;
  color: #ff4757;
  font-weight: bold;
}

.checkout-btn {
  padding: 12px 32px;
  background-color: #ff4757;
  color: white;
  border: none;
  border-radius: 24px;
  font-size: 16px;
  cursor: pointer;
}

.checkout-btn:hover {
  background-color: #e84118;
}
</style>
