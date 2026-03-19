<template>
  <div class="products">
    <h1>商品列表</h1>
    
    <!-- 搜索和筛选 -->
    <div class="filter-section">
      <div class="search-box">
        <input type="text" v-model="searchKeyword" placeholder="请输入商品名称" />
        <button @click="searchProducts">搜索</button>
      </div>
      <div class="category-filter">
        <select v-model="selectedCategory" @change="filterProducts">
          <option value="">全部分类</option>
          <option v-for="category in categories" :key="category.id" :value="category.id">
            {{ category.name }}
          </option>
        </select>
      </div>
    </div>
    
    <!-- 商品列表 -->
    <div class="product-list">
      <div v-for="product in products" :key="product.id" class="product-item" @click="navigateToProduct(product.id)">
        <img :src="product.image" :alt="product.name" />
        <h3>{{ product.name }}</h3>
        <p class="price">¥{{ product.price }}</p>
        <p class="description">{{ product.description }}</p>
      </div>
    </div>
    
    <!-- 无商品提示 -->
    <div v-if="products.length === 0" class="empty-state">
      <p>暂无商品</p>
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
import { useRouter } from 'vue-router'
import { productAPI } from '../api'

const router = useRouter()
const products = ref([])
const categories = ref([])
const searchKeyword = ref('')
const selectedCategory = ref('')
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const totalPages = computed(() => {
  return Math.ceil(total.value / pageSize.value)
})

const navigateToProduct = (id) => {
  router.push(`/product/${id}`)
}

const loadProducts = async () => {
  try {
    const params = {
      page: currentPage.value,
      limit: pageSize.value,
      keyword: searchKeyword.value,
      category_id: selectedCategory.value
    }
    const response = await productAPI.getProducts(params)
    products.value = response.data || []
    total.value = response.total || 0
  } catch (error) {
    console.error('加载商品列表失败:', error)
    products.value = []
    total.value = 0
  }
}

const loadCategories = async () => {
  try {
    const response = await productAPI.getCategories()
    categories.value = response || []
  } catch (error) {
    console.error('加载分类失败:', error)
    categories.value = []
  }
}

const searchProducts = () => {
  currentPage.value = 1
  loadProducts()
}

const filterProducts = () => {
  currentPage.value = 1
  loadProducts()
}

const changePage = (page) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    loadProducts()
  }
}

onMounted(() => {
  loadCategories()
  loadProducts()
})
</script>

<style scoped>
.products {
  padding: 20px;
}

h1 {
  margin-bottom: 30px;
  color: #333;
}

.filter-section {
  display: flex;
  gap: 20px;
  margin-bottom: 30px;
  flex-wrap: wrap;
}

.search-box {
  flex: 1;
  min-width: 300px;
  display: flex;
}

.search-box input {
  flex: 1;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px 0 0 4px;
}

.search-box button {
  padding: 10px 20px;
  background-color: #4CAF50;
  color: white;
  border: none;
  border-radius: 0 4px 4px 0;
  cursor: pointer;
}

.category-filter select {
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  min-width: 150px;
}

.product-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.product-item {
  padding: 15px;
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  cursor: pointer;
  transition: transform 0.3s ease;
}

.product-item:hover {
  transform: translateY(-5px);
}

.product-item img {
  width: 100%;
  height: 180px;
  object-fit: cover;
  border-radius: 4px;
  margin-bottom: 10px;
}

.product-item h3 {
  margin-bottom: 10px;
  color: #333;
  font-size: 16px;
  height: 48px;
  overflow: hidden;
}

.price {
  color: #ff4757;
  font-size: 18px;
  font-weight: bold;
  margin-bottom: 10px;
}

.description {
  color: #666;
  font-size: 14px;
  height: 40px;
  overflow: hidden;
  margin-bottom: 10px;
}

.empty-state {
  text-align: center;
  padding: 50px 0;
  color: #999;
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