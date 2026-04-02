<template>
  <div class="home">
    <!-- 搜索栏 -->
    <div class="search-bar">
      <div class="search-input-wrapper">
        <input 
          type="text" 
          v-model="searchKeyword" 
          placeholder="搜索商品..." 
          @keyup.enter="handleSearch"
        />
        <button @click="handleSearch">🔍</button>
      </div>
    </div>
    
    <!-- 活动信息 -->
    <section class="activity-section">
      <h2>活动信息</h2>
      <div v-if="activitiesLoading" class="loading-more">加载中...</div>
      <div v-else-if="activitiesError" class="error-message">{{ activitiesError }}</div>
      <div v-else-if="activities.length === 0" class="empty-message">暂无活动信息</div>
      <div v-else class="activity-list">
        <div v-for="activity in activities" :key="activity.id" class="activity-item">
          <h3>{{ activity.title }}</h3>
          <p>{{ activity.description }}</p>
          <span class="activity-date">{{ activity.date }}</span>
        </div>
      </div>
    </section>
    
    <!-- 热门商品 -->
    <section class="hot-products">
      <h2>热门商品</h2>
      <div class="product-list">
        <div 
          v-for="product in hotProducts" 
          :key="product.id" 
          class="product-card" 
          @click="navigateToProduct(product.id)"
        >
          <div class="product-image">
            <img :src="getProductImage(product)" :alt="product.name" />
          </div>
          <div class="product-info">
            <h3 class="product-name">{{ product.name }}</h3>
            <div class="product-price">
              <span class="price-symbol">¥</span>
              <span class="price-value">{{ formatPrice(product.default_sku_price || product.price) }}</span>
            </div>
            <div class="product-sales">
              <span>已售 {{ product.sales || 0 }}+</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 加载更多 -->
      <div v-if="loading" class="loading-more">加载中...</div>
      <div v-else-if="hasMore" class="load-more" @click="loadMore">加载更多</div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { productAPI, activityAPI } from '../api'

const router = useRouter()
const hotProducts = ref([])
const activities = ref([])
const searchKeyword = ref('')
const loading = ref(false)
const activitiesLoading = ref(false)
const activitiesError = ref('')
const hasMore = ref(true)
const currentPage = ref(1)
const pageSize = 10

const defaultImage = 'https://via.placeholder.com/120x120?text=No+Image'

const formatPrice = (price) => {
  if (!price) return '0.00'
  return parseFloat(price).toFixed(2)
}

const getProductImage = (product) => {
  // 优先使用 images 数组的第一张图片
  if (product.images && product.images.length > 0) {
    return product.images[0]
  }
  // 其次使用 image 字段
  if (product.image) {
    return product.image
  }
  // 最后使用默认图片
  return defaultImage
}

const navigateToProduct = (id) => {
  router.push(`/product/${id}`)
}

const handleSearch = () => {
  if (searchKeyword.value.trim()) {
    router.push(`/products?keyword=${encodeURIComponent(searchKeyword.value.trim())}`)
  }
}

const loadHotProducts = async (page = 1, append = false) => {
  if (loading.value) return
  loading.value = true
  
  try {
    const response = await productAPI.getProducts({ 
      page: page, 
      limit: pageSize 
    })
    
    const products = response.products || []
    const total = response.total || 0
    
    if (append) {
      hotProducts.value = [...hotProducts.value, ...products]
    } else {
      hotProducts.value = products
    }
    
    hasMore.value = hotProducts.value.length < total
    currentPage.value = page
  } catch (error) {
    console.error('加载热门商品失败:', error)
    if (!append) {
      hotProducts.value = []
      hasMore.value = false
    }
  } finally {
    loading.value = false
  }
}

const loadMore = () => {
  if (!loading.value && hasMore.value) {
    loadHotProducts(currentPage.value + 1, true)
  }
}

const formatActivityDate = (startTime, endTime) => {
  if (!startTime || !endTime) return '长期有效'
  
  const start = new Date(startTime)
  const end = new Date(endTime)
  
  const startStr = start.toLocaleDateString('zh-CN')
  const endStr = end.toLocaleDateString('zh-CN')
  
  return `${startStr} 至 ${endStr}`
}

const loadActivities = async () => {
  activitiesLoading.value = true
  activitiesError.value = ''
  
  try {
    const response = await activityAPI.getActivities()
    // 处理接口返回的数据格式
    const activityList = response || []
    // 转换数据格式，添加 date 字段
    activities.value = activityList.map(activity => ({
      id: activity.id,
      title: activity.name,
      description: activity.description,
      date: formatActivityDate(activity.start_time, activity.end_time)
    }))
  } catch (error) {
    console.error('加载活动失败:', error)
    activitiesError.value = '加载活动失败'
    // 兜底数据
    activities.value = [
      { id: 1, title: '限时折扣', description: '全场商品8折起', date: '2024-01-01 至 2024-01-31' },
      { id: 2, title: '新品上市', description: '新年新品，限时特惠', date: '2024-01-01 至 2024-01-15' },
      { id: 3, title: '满减活动', description: '满199减20，满399减50', date: '长期有效' }
    ]
  } finally {
    activitiesLoading.value = false
  }
}

onMounted(() => {
  loadHotProducts()
  loadActivities()
})
</script>

<style scoped>
.home {
  padding: 12px;
  background-color: #f5f5f5;
  min-height: 100vh;
}

/* 搜索栏 */
.search-bar {
  margin-bottom: 16px;
}

.search-input-wrapper {
  display: flex;
  background-color: white;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.search-input-wrapper input {
  flex: 1;
  padding: 12px 16px;
  border: none;
  outline: none;
  font-size: 14px;
}

.search-input-wrapper button {
  padding: 0 16px;
  border: none;
  background-color: #4CAF50;
  color: white;
  cursor: pointer;
  font-size: 16px;
}

/* 活动信息 */
.activity-section {
  margin-bottom: 20px;
}

.activity-section h2,
.hot-products h2 {
  margin-bottom: 12px;
  color: #333;
  font-size: 18px;
  font-weight: bold;
}

.activity-list {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding: 4px 0;
}

.activity-item {
  flex: 0 0 auto;
  min-width: 260px;
  padding: 16px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  color: white;
}

.activity-item h3 {
  margin-bottom: 8px;
  font-size: 16px;
}

.activity-item p {
  margin-bottom: 12px;
  font-size: 13px;
  opacity: 0.9;
}

.activity-date {
  font-size: 12px;
  opacity: 0.8;
}

/* 商品列表 */
.product-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.product-card {
  display: flex;
  background-color: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  cursor: pointer;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.product-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.product-image {
  width: 120px;
  height: 120px;
  flex-shrink: 0;
  overflow: hidden;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-info {
  flex: 1;
  padding: 12px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.product-name {
  font-size: 14px;
  color: #333;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 8px;
}

.product-price {
  color: #ff4757;
  font-weight: bold;
}

.price-symbol {
  font-size: 12px;
}

.price-value {
  font-size: 18px;
}

.product-sales {
  font-size: 12px;
  color: #999;
}

/* 加载更多 */
.loading-more,
.load-more {
  text-align: center;
  padding: 16px;
  color: #666;
  font-size: 14px;
}

.load-more {
  cursor: pointer;
  color: #4CAF50;
}

.load-more:hover {
  text-decoration: underline;
}

/* 错误消息 */
.error-message {
  text-align: center;
  padding: 16px;
  color: #ff4757;
  font-size: 14px;
  background-color: #fff5f5;
  border-radius: 8px;
  margin-bottom: 12px;
}

/* 空消息 */
.empty-message {
  text-align: center;
  padding: 16px;
  color: #999;
  font-size: 14px;
  background-color: #f9f9f9;
  border-radius: 8px;
  margin-bottom: 12px;
}
</style>
