<template>
  <div class="search-results">
    <h1>搜索商品</h1>

    <!-- 搜索框 -->
    <div class="search-box">
      <input
        type="text"
        v-model="keyword"
        placeholder="请输入商品名称"
        @keyup.enter="handleSearch"
      />
      <button @click="handleSearch">搜索</button>
    </div>

    <!-- 筛选条件 -->
    <div class="filter-section">
      <div class="filter-item">
        <label>分类：</label>
        <select v-model="categoryId" @change="handleSearch">
          <option value="">全部分类</option>
          <option v-for="category in categories" :key="category.id" :value="category.id">
            {{ category.name }}
          </option>
        </select>
      </div>
      <div class="filter-item">
        <label>价格区间：</label>
        <input type="number" v-model="priceMin" placeholder="最低价" class="price-input" />
        <span class="price-separator">-</span>
        <input type="number" v-model="priceMax" placeholder="最高价" class="price-input" />
        <button class="filter-btn" @click="handleSearch">筛选</button>
      </div>
    </div>

    <!-- 商品网格 -->
    <div class="product-grid" v-if="products.length > 0">
      <div
        v-for="product in products"
        :key="product.id"
        class="product-card"
        @click="navigateToProduct(product.id)"
      >
        <div class="product-image">
          <img :src="getImageUrl(product.main_image)" :alt="stripTags(product.name)" />
        </div>
        <div class="product-info">
          <h3 class="product-name" v-html="product.name"></h3>
          <p class="product-description" v-if="product.description" v-html="product.description"></p>
          <p class="product-price">¥{{ product.min_price }}<span v-if="product.max_price && product.max_price !== product.min_price"> - ¥{{ product.max_price }}</span></p>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-if="hasSearched && products.length === 0 && !loading" class="empty-state">
      <p>未找到相关商品，请尝试其他关键词</p>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading-state">
      <p>搜索中...</p>
    </div>

    <!-- 分页 -->
    <div class="pagination" v-if="total > 0">
      <button @click="changePage(1)" :disabled="currentPage === 1">首页</button>
      <button @click="changePage(currentPage - 1)" :disabled="currentPage === 1">上一页</button>
      <span>{{ currentPage }} / {{ totalPages }}</span>
      <button @click="changePage(currentPage + 1)" :disabled="currentPage === totalPages">下一页</button>
      <button @click="changePage(totalPages)" :disabled="currentPage === totalPages">末页</button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { searchAPI, productAPI } from '../api'

const router = useRouter()
const route = useRoute()

const keyword = ref(route.query.keyword || '')
const categoryId = ref('')
const priceMin = ref('')
const priceMax = ref('')
const products = ref([])
const categories = ref([])
const currentPage = ref(1)
const pageSize = ref(12)
const total = ref(0)
const loading = ref(false)
const hasSearched = ref(false)

const totalPages = computed(() => {
  return Math.ceil(total.value / pageSize.value)
})

// 去除HTML标签（用于alt属性等纯文本场景）
const stripTags = (html) => {
  if (!html) return ''
  return html.replace(/<[^>]*>/g, '')
}

// 获取完整图片URL
const getImageUrl = (imageUrl) => {
  if (!imageUrl) return ''
  if (imageUrl.startsWith('http://') || imageUrl.startsWith('https://')) {
    return imageUrl
  }
  return `/api${imageUrl}`
}

// 跳转到商品详情
const navigateToProduct = (id) => {
  router.push(`/product/${id}`)
}

// 执行搜索
const handleSearch = () => {
  currentPage.value = 1
  searchProducts()
}

// 搜索商品
const searchProducts = async () => {
  if (!keyword.value.trim()) {
    return
  }
  loading.value = true
  hasSearched.value = true
  try {
    const params = {
      keyword: keyword.value,
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (categoryId.value) {
      params.category_id = categoryId.value
    }
    if (priceMin.value) {
      params.price_min = priceMin.value
    }
    if (priceMax.value) {
      params.price_max = priceMax.value
    }
    const data = await searchAPI.searchProducts(params)
    products.value = data.data || data.items || []
    total.value = data.total || 0
  } catch (error) {
    products.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 加载分类列表
const loadCategories = async () => {
  try {
    const data = await productAPI.getProducts({ limit: 0 })
    // 尝试从商品接口获取分类，如果接口不支持则留空
    categories.value = []
  } catch (error) {
    categories.value = []
  }
}

// 翻页
const changePage = (page) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    searchProducts()
  }
}

onMounted(() => {
  // 如果URL中有关键词参数，自动搜索
  if (keyword.value.trim()) {
    searchProducts()
  }
})
</script>

<style scoped>
.search-results {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

h1 {
  margin-bottom: 20px;
  color: #333;
}

.search-box {
  display: flex;
  margin-bottom: 20px;
}

.search-box input {
  flex: 1;
  padding: 12px 16px;
  border: 2px solid #4CAF50;
  border-radius: 4px 0 0 4px;
  font-size: 16px;
  outline: none;
}

.search-box input:focus {
  border-color: #388E3C;
}

.search-box button {
  padding: 12px 30px;
  background-color: #4CAF50;
  color: white;
  border: 2px solid #4CAF50;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
  font-size: 16px;
}

.search-box button:hover {
  background-color: #388E3C;
}

.filter-section {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
  flex-wrap: wrap;
  align-items: center;
}

.filter-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filter-item label {
  color: #666;
  font-size: 14px;
  white-space: nowrap;
}

.filter-item select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  min-width: 150px;
}

.price-input {
  width: 100px;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.price-separator {
  color: #999;
}

.filter-btn {
  padding: 8px 16px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

.filter-btn:hover {
  background-color: #388E3C;
}

.product-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 30px;
}

@media (max-width: 992px) {
  .product-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .product-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.product-card {
  background-color: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.3s ease, box-shadow 0.3s ease;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

.product-image {
  width: 100%;
  height: 200px;
  overflow: hidden;
}

.product-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-info {
  padding: 12px;
}

.product-name {
  font-size: 14px;
  color: #333;
  margin-bottom: 8px;
  height: 40px;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.product-name :deep(em) {
  color: #ff4757;
  font-style: normal;
  font-weight: bold;
}

.product-description {
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
  height: 36px;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.product-description :deep(em) {
  color: #ff4757;
  font-style: normal;
  font-weight: bold;
}

.product-price {
  color: #ff4757;
  font-size: 18px;
  font-weight: bold;
}

.empty-state {
  text-align: center;
  padding: 60px 0;
  color: #999;
  font-size: 16px;
}

.loading-state {
  text-align: center;
  padding: 40px 0;
  color: #666;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 30px;
}

.pagination button {
  padding: 8px 16px;
  border: 1px solid #ddd;
  background-color: white;
  border-radius: 4px;
  cursor: pointer;
}

.pagination button:hover:not(:disabled) {
  background-color: #f0f0f0;
}

.pagination button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.pagination span {
  padding: 0 10px;
  color: #333;
}
</style>
