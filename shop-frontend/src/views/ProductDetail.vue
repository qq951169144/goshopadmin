<template>
  <div class="product-detail">
    <div v-if="loading" class="loading">加载中...</div>
    <div v-else-if="error" class="error">{{ error }}</div>
    <div v-else-if="product" class="product-content">
      <!-- 图片轮播 -->
      <div class="image-carousel">
        <div class="carousel-container">
          <img 
            :src="currentImage || product.images[0] || defaultImage" 
            :alt="product.name" 
            class="main-image"
            @click="previewImage"
          />
          <div class="image-indicator" v-if="product.images && product.images.length > 1">
            {{ currentImageIndex + 1 }} / {{ product.images.length }}
          </div>
        </div>
        <!-- 缩略图 -->
        <div class="thumbnails" v-if="product.images && product.images.length > 1">
          <div 
            v-for="(img, index) in product.images" 
            :key="index"
            class="thumbnail-wrapper"
            :class="{ active: currentImageIndex === index }"
            @click="setCurrentImage(index)"
          >
            <img :src="img" :alt="product.name" class="thumbnail" />
          </div>
        </div>
        <!-- 左右切换按钮 -->
        <button 
          v-if="product.images && product.images.length > 1" 
          class="carousel-btn prev" 
          @click="prevImage"
        >
          ‹
        </button>
        <button 
          v-if="product.images && product.images.length > 1" 
          class="carousel-btn next" 
          @click="nextImage"
        >
          ›
        </button>
      </div>

      <!-- 商品基本信息 -->
      <div class="product-basic-info">
        <h1 class="product-name">{{ product.name }}</h1>
        <p class="product-price">
          <span class="price-symbol">¥</span>
          <span class="price-value">{{ formatPrice(currentPrice) }}</span>
          <span v-if="currentOriginalPrice > 0" class="original-price">
            ¥{{ formatPrice(currentOriginalPrice) }}
          </span>
        </p>

      </div>

      <!-- 规格选择 -->
      <div class="sku-section" v-if="product.specifications && product.specifications.length > 0">
        <h3>选择规格</h3>
        <SpecSelector
          :specifications="product.specifications"
          :selected-specs="selectedSpecs"
          :sku-list="product.sku_list || []"
          @change="handleSpecChange"
          @stock-change="handleStockChange"
        />
      </div>

      <!-- 调试信息：显示是否有规格数据 -->
      <div class="sku-section" v-else style="background: #fff3cd; padding: 10px; color: #856404;">
        <p>暂无规格数据 (specifications: {{ product.specifications }})</p>
        <p>SKU列表: {{ product.sku_list?.length || 0 }} 个</p>
      </div>

      <!-- 旧版SKU选择（兼容） -->
      <div class="sku-section" v-if="product.skus && product.skus.length > 0">
        <h3>选择规格</h3>
        <div class="sku-options">
          <div 
            v-for="sku in product.skus" 
            :key="sku.id" 
            class="sku-option" 
            :class="{ active: selectedSkuId === sku.id }" 
            @click="selectSku(sku)"
          >
            {{ sku.name }}
          </div>
        </div>
      </div>

      <!-- 数量选择 -->
      <div class="quantity-section">
        <label>数量</label>
        <div class="quantity-control">
          <button @click="decreaseQuantity" :disabled="quantity <= 1">-</button>
          <input type="number" v-model.number="quantity" min="1" readonly />
          <button @click="increaseQuantity">+</button>
        </div>
      </div>

      <!-- 商品描述 -->
      <div class="description-section">
        <h3>商品描述</h3>
        <p class="description-text">{{ product.description }}</p>
      </div>

      <!-- 商品详情富文本 -->
      <div class="detail-section" v-if="product.detail">
        <h3>商品详情</h3>
        <div class="detail-content" v-html="product.detail"></div>
      </div>

      <!-- 底部占位，为固定按钮留空间 -->
      <div class="bottom-spacer"></div>
    </div>
    <div v-else class="not-found">商品不存在</div>

    <!-- 库存不足提示 -->
    <div v-if="currentSkuStock !== null && currentSkuStock <= 0" class="stock-warning">
      库存不足，无法购买
    </div>

    <!-- 底部固定操作栏 -->
    <div v-if="product" class="bottom-action-bar">
      <div class="action-buttons">
        <button class="add-to-cart" @click="addToCart" :disabled="!canAddToCart">
          <span class="btn-icon">🛒</span>
          <span class="btn-text">加入购物车</span>
        </button>
        <button class="buy-now" @click="buyNow" :disabled="!canAddToCart">
          <span class="btn-text">立即购买</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { productAPI, cartAPI } from '../api'
import SpecSelector from '../components/SpecSelector.vue'

const route = useRoute()
const router = useRouter()
const product = ref(null)
const loading = ref(true)
const error = ref('')
const selectedSpecs = ref({})
const selectedSkuId = ref(null)
const quantity = ref(1)
const currentImageIndex = ref(0)
const currentSkuStock = ref(null)

const defaultImage = 'https://via.placeholder.com/400x400?text=No+Image'

const currentImage = computed(() => {
  if (product.value && product.value.images && product.value.images.length > 0) {
    return product.value.images[currentImageIndex.value]
  }
  return ''
})

// 当前选中的SKU
const currentSku = computed(() => {
  if (!product.value) return null

  // 如果有规格选择
  if (product.value.specifications && product.value.specifications.length > 0) {
    const skuList = product.value.sku_list || []
    return skuList.find(sku => {
      const specCombination = sku.spec_combination || {}
      for (const [specId, valueId] of Object.entries(selectedSpecs.value)) {
        // 使用 Number() 进行类型转换后比较
        if (Number(specCombination[specId]) !== Number(valueId)) {
          return false
        }
      }
      return true
    })
  }

  // 旧版SKU选择
  if (product.value.skus && product.value.skus.length > 0) {
    return product.value.skus.find(s => s.id === selectedSkuId.value)
  }

  return null
})

// 当前价格
const currentPrice = computed(() => {
  if (currentSku.value) {
    return currentSku.value.price
  }
  if (product.value?.price_range) {
    return product.value.price_range.min
  }
  return product.value?.price || 0
})

// 当前原价
const currentOriginalPrice = computed(() => {
  if (currentSku.value) {
    return currentSku.value.original_price || 0
  }
  return 0
})

// 是否可以加入购物车
const canAddToCart = computed(() => {
  // 如果有规格，必须选择完整规格
  if (product.value?.specifications && product.value.specifications.length > 0) {
    const selectedCount = Object.keys(selectedSpecs.value).length
    const specCount = product.value.specifications.length
    if (selectedCount !== specCount) {
      return false
    }
  }

  // 检查库存
  if (currentSkuStock.value !== null && currentSkuStock.value <= 0) {
    return false
  }

  return true
})

const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toFixed(2)
}

const setCurrentImage = (index) => {
  currentImageIndex.value = index
}

const nextImage = () => {
  if (!product.value || !product.value.images) return
  currentImageIndex.value = (currentImageIndex.value + 1) % product.value.images.length
}

const prevImage = () => {
  if (!product.value || !product.value.images) return
  currentImageIndex.value = (currentImageIndex.value - 1 + product.value.images.length) % product.value.images.length
}

const previewImage = () => {
  console.log('预览图片:', currentImage.value)
}

// 处理规格选择变化
const handleSpecChange = (specs) => {
  selectedSpecs.value = specs
  quantity.value = 1 // 重置数量
}

// 处理库存变化
const handleStockChange = (stock) => {
  currentSkuStock.value = stock
}

const selectSku = (sku) => {
  selectedSkuId.value = sku.id
  quantity.value = 1
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
    product.value = response || null
    console.log('商品详情数据:', product.value)
    console.log('规格数据:', product.value?.specifications)
    console.log('SKU列表:', product.value?.sku_list)
    if (product.value) {
      currentImageIndex.value = 0
      selectedSpecs.value = {}
      selectedSkuId.value = null
      quantity.value = 1

      // 如果有旧版SKU，默认选择第一个
      if (product.value.skus && product.value.skus.length > 0) {
        selectedSkuId.value = product.value.skus[0].id
      }
    }
  } catch (err) {
    console.error('加载商品详情失败:', err)
    error.value = '加载商品详情失败'
  } finally {
    loading.value = false
  }
}

const addToCart = async () => {
  if (!product.value) return
  
  if (!canAddToCart.value) {
    alert('请选择完整的规格')
    return
  }
  
  try {
    const cartItem = {
      product_id: product.value.id,
      quantity: quantity.value,
      price: currentPrice.value,
      sku_id: currentSku.value?.id || selectedSku.value
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
  
  if (!canAddToCart.value) {
    alert('请选择完整的规格')
    return
  }
  
  try {
    const cartItem = {
      product_id: product.value.id,
      quantity: quantity.value,
      price: currentPrice.value,
      sku_id: currentSku.value?.id || selectedSku.value
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
  min-height: 100vh;
  background-color: #f5f5f5;
  padding-bottom: 80px;
}

.loading,
.error,
.not-found {
  text-align: center;
  padding: 100px 20px;
  color: #666;
  font-size: 16px;
}

/* 图片轮播 */
.image-carousel {
  position: relative;
  background-color: white;
}

.carousel-container {
  position: relative;
  width: 100%;
  height: 375px;
  overflow: hidden;
}

.main-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.image-indicator {
  position: absolute;
  bottom: 16px;
  right: 16px;
  background-color: rgba(0, 0, 0, 0.5);
  color: white;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
}

/* 缩略图 */
.thumbnails {
  display: flex;
  gap: 8px;
  padding: 12px 16px;
  overflow-x: auto;
  background-color: white;
}

.thumbnail-wrapper {
  flex-shrink: 0;
  width: 60px;
  height: 60px;
  border-radius: 4px;
  overflow: hidden;
  border: 2px solid transparent;
  cursor: pointer;
}

.thumbnail-wrapper.active {
  border-color: #ff4757;
}

.thumbnail {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 轮播按钮 */
.carousel-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  width: 40px;
  height: 40px;
  background-color: rgba(0, 0, 0, 0.3);
  border: none;
  border-radius: 50%;
  color: white;
  font-size: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

.carousel-btn.prev {
  left: 12px;
}

.carousel-btn.next {
  right: 12px;
}

/* 商品基本信息 */
.product-basic-info {
  background-color: white;
  padding: 16px;
  margin-bottom: 8px;
}

.product-name {
  font-size: 18px;
  color: #333;
  line-height: 1.4;
  margin-bottom: 12px;
}

.product-price {
  color: #ff4757;
  display: flex;
  align-items: baseline;
  gap: 10px;
}

.price-symbol {
  font-size: 14px;
}

.price-value {
  font-size: 28px;
  font-weight: bold;
}

.original-price {
  font-size: 14px;
  color: #999;
  text-decoration: line-through;
}

.product-stock {
  font-size: 14px;
  color: #666;
  margin-top: 8px;
}

/* SKU选择 */
.sku-section {
  background-color: white;
  padding: 16px;
  margin-bottom: 8px;
}

.sku-section h3 {
  font-size: 14px;
  color: #666;
  margin-bottom: 12px;
}

.sku-options {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.sku-option {
  padding: 8px 16px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  color: #333;
  cursor: pointer;
  transition: all 0.3s ease;
}

.sku-option:hover {
  border-color: #ff4757;
}

.sku-option.active {
  border-color: #ff4757;
  background-color: #fff5f5;
  color: #ff4757;
}

/* 数量选择 */
.quantity-section {
  background-color: white;
  padding: 16px;
  margin-bottom: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.quantity-section label {
  font-size: 14px;
  color: #666;
}

.quantity-control {
  display: flex;
  align-items: center;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.quantity-control button {
  width: 32px;
  height: 32px;
  border: none;
  background-color: #f5f5f5;
  font-size: 16px;
  cursor: pointer;
}

.quantity-control button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.quantity-control input {
  width: 50px;
  height: 32px;
  border: none;
  text-align: center;
  font-size: 14px;
}

/* 商品描述 */
.description-section {
  background-color: white;
  padding: 16px;
  margin-bottom: 8px;
}

.description-section h3 {
  font-size: 14px;
  color: #666;
  margin-bottom: 12px;
}

.description-text {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
}

/* 商品详情富文本 */
.detail-section {
  background-color: white;
  padding: 16px;
  margin-bottom: 8px;
}

.detail-section h3 {
  font-size: 14px;
  color: #666;
  margin-bottom: 12px;
}

.detail-content {
  font-size: 14px;
  color: #333;
  line-height: 1.6;
}

.detail-content :deep(img) {
  max-width: 100%;
  height: auto;
}

/* 底部占位 */
.bottom-spacer {
  height: 80px;
}

/* 底部固定操作栏 */
.bottom-action-bar {
  position: fixed;
  bottom: 60px;
  left: 0;
  right: 0;
  background-color: white;
  padding: 12px 16px;
  border-top: 1px solid #eee;
  box-shadow: 0 -2px 10px rgba(0, 0, 0, 0.05);
  z-index: 100;
}

.action-buttons {
  display: flex;
  gap: 12px;
}

.add-to-cart,
.buy-now {
  flex: 1;
  padding: 12px;
  border: none;
  border-radius: 20px;
  font-size: 15px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.add-to-cart {
  background-color: #ff9800;
  color: white;
}

.add-to-cart:hover:not(:disabled) {
  background-color: #f57c00;
}

.add-to-cart:disabled,
.buy-now:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.buy-now {
  background-color: #ff4757;
  color: white;
}

.buy-now:hover:not(:disabled) {
  background-color: #e84118;
}

.btn-icon {
  font-size: 16px;
}

/* 库存不足提示 */
.stock-warning {
  position: fixed;
  bottom: 120px;
  left: 0;
  right: 0;
  text-align: center;
  padding: 8px;
  background-color: #fff2f0;
  color: #ff4d4f;
  font-size: 14px;
  z-index: 99;
}
</style>
