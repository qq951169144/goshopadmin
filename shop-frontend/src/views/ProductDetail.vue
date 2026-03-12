<template>
  <div class="product-detail">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="product" class="product-content">
      <div class="product-info">
        <div class="product-images">
          <img :src="product.image" :alt="product.name" class="main-image" />
          <div class="thumbnails">
            <img 
              v-for="(img, index) in product.images" 
              :key="index" 
              :src="img" 
              :alt="product.name" 
              class="thumbnail" 
              @click="setMainImage(img)"
            />
          </div>
        </div>
        <div class="product-details">
          <h1>{{ product.name }}</h1>
          <p class="price">¥{{ product.price }}</p>
          <p class="description">{{ product.description }}</p>
          
          <!-- SKU选择 -->
          <div class="sku-section" v-if="product.skus && product.skus.length > 0">
            <h3>选择规格</h3>
            <div class="sku-options">
              <div 
                v-for="sku in product.skus" 
                :key="sku.id" 
                class="sku-option" 
                :class="{ active: selectedSku === sku.id }" 
                @click="selectSku(sku)"
              >
                {{ sku.name }} - ¥{{ sku.price }}
              </div>
            </div>
          </div>
          
          <!-- 数量选择 -->
          <div class="quantity-section">
            <label>数量：</label>
            <div class="quantity-control">
              <button @click="decreaseQuantity" :disabled="quantity <= 1">-</button>
              <input type="number" v-model.number="quantity" min="1" />
              <button @click="increaseQuantity">+</button>
            </div>
          </div>
          
          <!-- 操作按钮 -->
          <div class="action-buttons">
            <button class="add-to-cart" @click="addToCart">添加到购物车</button>
            <button class="buy-now" @click="buyNow">立即购买</button>
          </div>
        </div>
      </div>
      
      <!-- 商品详情 -->
      <div class="product-description">
        <h2>商品详情</h2>
        <div v-html="product.detail"></div>
      </div>
    </div>
    <div v-else class="not-found">商品不存在</div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { productAPI, cartAPI } from '../api'

const route = useRoute()
const router = useRouter()
const product = ref(null)
const loading = ref(true)
const error = ref('')
const selectedSku = ref(null)
const quantity = ref(1)
const mainImage = ref('')

const setMainImage = (img) => {
  mainImage.value = img
}

const selectSku = (sku) => {
  selectedSku.value = sku.id
}

const decreaseQuantity = () => {
  if (quantity.value > 1) {
    quantity.value--
  }
}

const increaseQuantity = () => {
  quantity.value++
}

const loadProductDetail = async () => {
  try {
    const id = route.params.id
    const response = await productAPI.getProductDetail(id)
    product.value = response.data || null
    if (product.value) {
      mainImage.value = product.value.image
      if (product.value.skus && product.value.skus.length > 0) {
        selectedSku.value = product.value.skus[0].id
      }
    }
  } catch (err) {
    console.error('加载商品详情失败:', err)
    error.value = '加载商品详情失败'
    // 使用模拟数据
    product.value = {
      id: route.params.id,
      name: '商品详情',
      price: 99.99,
      description: '这是商品描述',
      image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20detail%20image&image_size=landscape_4_3',
      images: [
        'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20detail%20image%201&image_size=square',
        'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20detail%20image%202&image_size=square',
        'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20detail%20image%203&image_size=square'
      ],
      skus: [
        { id: 1, name: '红色-中号', price: 99.99 },
        { id: 2, name: '蓝色-大号', price: 119.99 }
      ],
      detail: '<h3>商品详情</h3><p>这是商品的详细信息，包括材质、尺寸、使用方法等。</p><p>商品质量保证，欢迎购买！</p>'
    }
    mainImage.value = product.value.image
    selectedSku.value = product.value.skus[0].id
  } finally {
    loading.value = false
  }
}

const addToCart = async () => {
  if (!product.value) return
  
  try {
    const cartItem = {
      product_id: product.value.id,
      quantity: quantity.value,
      price: product.value.price,
      sku: selectedSku.value
    }
    await cartAPI.addToCart(cartItem)
    alert('添加到购物车成功')
  } catch (err) {
    console.error('添加到购物车失败:', err)
    alert('添加到购物车失败')
  }
}

const buyNow = async () => {
  if (!product.value) return
  
  // 先添加到购物车，然后跳转到购物车页面
  try {
    const cartItem = {
      product_id: product.value.id,
      quantity: quantity.value,
      price: product.value.price,
      sku: selectedSku.value
    }
    await cartAPI.addToCart(cartItem)
    router.push('/cart')
  } catch (err) {
    console.error('添加到购物车失败:', err)
    alert('添加到购物车失败')
  }
}

onMounted(() => {
  loadProductDetail()
})
</script>

<style scoped>
.product-detail {
  padding: 20px;
}

.loading,
.error,
.not-found {
  text-align: center;
  padding: 50px 0;
  color: #666;
}

.product-content {
  max-width: 1200px;
  margin: 0 auto;
}

.product-info {
  display: flex;
  gap: 40px;
  margin-bottom: 40px;
  flex-wrap: wrap;
}

.product-images {
  flex: 1;
  min-width: 400px;
}

.main-image {
  width: 100%;
  height: 400px;
  object-fit: cover;
  border-radius: 8px;
  margin-bottom: 20px;
}

.thumbnails {
  display: flex;
  gap: 10px;
}

.thumbnail {
  width: 80px;
  height: 80px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
  border: 2px solid transparent;
}

.thumbnail:hover {
  border-color: #4CAF50;
}

.product-details {
  flex: 1;
  min-width: 300px;
}

.product-details h1 {
  margin-bottom: 20px;
  color: #333;
  font-size: 24px;
}

.price {
  color: #ff4757;
  font-size: 28px;
  font-weight: bold;
  margin-bottom: 20px;
}

.description {
  color: #666;
  margin-bottom: 30px;
  line-height: 1.6;
}

.sku-section {
  margin-bottom: 30px;
}

.sku-section h3 {
  margin-bottom: 15px;
  color: #333;
}

.sku-options {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.sku-option {
  padding: 10px 20px;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.sku-option:hover {
  border-color: #4CAF50;
}

.sku-option.active {
  border-color: #4CAF50;
  background-color: #f0f9f0;
  color: #4CAF50;
}

.quantity-section {
  margin-bottom: 30px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.quantity-control {
  display: flex;
  align-items: center;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.quantity-control button {
  padding: 8px 15px;
  border: none;
  background-color: #f9f9f9;
  cursor: pointer;
}

.quantity-control input {
  width: 60px;
  text-align: center;
  border: none;
  outline: none;
}

.action-buttons {
  display: flex;
  gap: 20px;
}

.add-to-cart,
.buy-now {
  flex: 1;
  padding: 15px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.3s ease;
}

.add-to-cart {
  background-color: #ff9800;
  color: white;
}

.add-to-cart:hover {
  background-color: #f57c00;
}

.buy-now {
  background-color: #ff4757;
  color: white;
}

.buy-now:hover {
  background-color: #e84118;
}

.product-description {
  margin-top: 40px;
  padding-top: 40px;
  border-top: 1px solid #eee;
}

.product-description h2 {
  margin-bottom: 20px;
  color: #333;
}

.product-description div {
  color: #666;
  line-height: 1.6;
}

@media (max-width: 768px) {
  .product-info {
    flex-direction: column;
  }
  
  .product-images,
  .product-details {
    min-width: 100%;
  }
  
  .main-image {
    height: 300px;
  }
}
</style>