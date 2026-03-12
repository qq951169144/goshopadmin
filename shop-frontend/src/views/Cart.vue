<template>
  <div class="cart">
    <h1>购物车</h1>
    
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="cartItems.length === 0" class="empty-cart">
      <p>购物车为空</p>
      <button @click="goShopping">去购物</button>
    </div>
    <div v-else class="cart-content">
      <div class="cart-items">
        <div v-for="item in cartItems" :key="item.product_id" class="cart-item">
          <div class="item-image">
            <img :src="item.image" :alt="item.name" />
          </div>
          <div class="item-info">
            <h3>{{ item.name }}</h3>
            <p class="item-sku" v-if="item.sku">{{ item.sku }}</p>
            <p class="item-price">¥{{ item.price }}</p>
          </div>
          <div class="item-quantity">
            <button @click="updateQuantity(item.product_id, item.quantity - 1)" :disabled="item.quantity <= 1">-</button>
            <input type="number" v-model.number="item.quantity" min="1" @change="updateQuantity(item.product_id, item.quantity)" />
            <button @click="updateQuantity(item.product_id, item.quantity + 1)">+</button>
          </div>
          <div class="item-subtotal">
            ¥{{ (item.price * item.quantity).toFixed(2) }}
          </div>
          <div class="item-actions">
            <button class="remove-btn" @click="removeItem(item.product_id)">删除</button>
          </div>
        </div>
      </div>
      
      <div class="cart-summary">
        <h2>订单 summary</h2>
        <div class="summary-item">
          <span>商品总价：</span>
          <span>¥{{ totalPrice.toFixed(2) }}</span>
        </div>
        <div class="summary-item">
          <span>运费：</span>
          <span>¥{{ shippingFee.toFixed(2) }}</span>
        </div>
        <div class="summary-total">
          <span>合计：</span>
          <span>¥{{ (totalPrice + shippingFee).toFixed(2) }}</span>
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
const loading = ref(true)
const error = ref('')
const shippingFee = ref(10)

const totalPrice = computed(() => {
  return cartItems.value.reduce((total, item) => {
    return total + (item.price * item.quantity)
  }, 0)
})

const goShopping = () => {
  router.push('/')
}

const loadCart = async () => {
  try {
    const response = await cartAPI.getCart()
    cartItems.value = response.items || []
  } catch (err) {
    console.error('加载购物车失败:', err)
    error.value = '加载购物车失败'
    // 使用模拟数据
    cartItems.value = [
      {
        product_id: 1,
        quantity: 2,
        price: 99.99,
        sku: 'red-medium',
        name: '商品1',
        image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%201&image_size=square'
      },
      {
        product_id: 2,
        quantity: 1,
        price: 199.99,
        sku: 'blue-large',
        name: '商品2',
        image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%202&image_size=square'
      }
    ]
  } finally {
    loading.value = false
  }
}

const updateQuantity = async (productId, quantity) => {
  if (quantity < 1) return
  
  try {
    await cartAPI.updateCartItem(productId, { quantity })
    // 本地更新
    const item = cartItems.value.find(item => item.product_id === productId)
    if (item) {
      item.quantity = quantity
    }
  } catch (err) {
    console.error('更新数量失败:', err)
    alert('更新数量失败')
  }
}

const removeItem = async (productId) => {
  try {
    await cartAPI.removeCartItem(productId)
    // 本地更新
    cartItems.value = cartItems.value.filter(item => item.product_id !== productId)
  } catch (err) {
    console.error('删除商品失败:', err)
    alert('删除商品失败')
  }
}

const checkout = () => {
  // 跳转到订单确认页面
  alert('跳转到订单确认页面')
  // 实际项目中应该跳转到订单确认页面
  // router.push('/checkout')
}

onMounted(() => {
  loadCart()
})
</script>

<style scoped>
.cart {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

h1 {
  margin-bottom: 30px;
  color: #333;
}

.loading,
.error {
  text-align: center;
  padding: 50px 0;
  color: #666;
}

.empty-cart {
  text-align: center;
  padding: 100px 0;
  color: #666;
}

.empty-cart button {
  margin-top: 20px;
  padding: 10px 30px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.cart-content {
  display: flex;
  gap: 40px;
  flex-wrap: wrap;
}

.cart-items {
  flex: 1;
  min-width: 600px;
}

.cart-item {
  display: flex;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #eee;
  gap: 20px;
}

.item-image {
  width: 100px;
  height: 100px;
}

.item-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 4px;
}

.item-info {
  flex: 1;
}

.item-info h3 {
  margin-bottom: 10px;
  color: #333;
}

.item-sku {
  color: #999;
  font-size: 14px;
  margin-bottom: 10px;
}

.item-price {
  color: #ff4757;
  font-weight: bold;
}

.item-quantity {
  display: flex;
  align-items: center;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.item-quantity button {
  padding: 8px 15px;
  border: none;
  background-color: #f9f9f9;
  cursor: pointer;
}

.item-quantity input {
  width: 60px;
  text-align: center;
  border: none;
  outline: none;
}

.item-subtotal {
  font-weight: bold;
  color: #333;
  min-width: 100px;
  text-align: right;
}

.item-actions {
  min-width: 80px;
}

.remove-btn {
  padding: 8px 16px;
  background-color: #ff4757;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.cart-summary {
  width: 300px;
  background-color: #f9f9f9;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 20px;
}

.cart-summary h2 {
  margin-bottom: 20px;
  color: #333;
  font-size: 18px;
}

.summary-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 15px;
  color: #666;
}

.summary-total {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
  font-weight: bold;
  font-size: 18px;
  color: #333;
  padding-top: 15px;
  border-top: 1px solid #ddd;
}

.checkout-btn {
  width: 100%;
  padding: 15px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.3s ease;
}

.checkout-btn:hover {
  background-color: #45a049;
}

@media (max-width: 768px) {
  .cart-content {
    flex-direction: column;
  }
  
  .cart-items {
    min-width: 100%;
  }
  
  .cart-summary {
    width: 100%;
    position: static;
  }
  
  .cart-item {
    flex-wrap: wrap;
  }
  
  .item-image {
    flex-shrink: 0;
  }
  
  .item-info {
    flex: 1;
    min-width: 0;
  }
  
  .item-quantity,
  .item-subtotal,
  .item-actions {
    min-width: auto;
  }
}
</style>