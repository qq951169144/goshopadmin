<template>
  <div class="home">
    <h1>欢迎来到C端商城</h1>
    
    <!-- 活动信息 -->
    <section class="activity-section">
      <h2>活动信息</h2>
      <div class="activity-list">
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
        <div v-for="product in hotProducts" :key="product.id" class="product-item" @click="navigateToProduct(product.id)">
          <img :src="product.image" :alt="product.name" />
          <h3>{{ product.name }}</h3>
          <p class="price">¥{{ product.price }}</p>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { productAPI } from '../api'

const router = useRouter()
const hotProducts = ref([])
const activities = ref([])

const navigateToProduct = (id) => {
  router.push(`/product/${id}`)
}

const loadHotProducts = async () => {
  try {
    const response = await productAPI.getProducts({ hot: 1, limit: 6 })
    hotProducts.value = response.data || []
  } catch (error) {
    console.error('加载热门商品失败:', error)
    // 使用模拟数据
    hotProducts.value = [
      { id: 1, name: '商品1', price: 99.99, image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%201&image_size=square' },
      { id: 2, name: '商品2', price: 199.99, image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%202&image_size=square' },
      { id: 3, name: '商品3', price: 299.99, image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%203&image_size=square' },
      { id: 4, name: '商品4', price: 399.99, image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%204&image_size=square' },
      { id: 5, name: '商品5', price: 499.99, image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%205&image_size=square' },
      { id: 6, name: '商品6', price: 599.99, image: 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=product%20image%206&image_size=square' }
    ]
  }
}

const loadActivities = () => {
  // 模拟活动数据
  activities.value = [
    { id: 1, title: '限时折扣', description: '全场商品8折起', date: '2024-01-01 至 2024-01-31' },
    { id: 2, title: '新品上市', description: '新年新品，限时特惠', date: '2024-01-01 至 2024-01-15' }
  ]
}

onMounted(() => {
  loadHotProducts()
  loadActivities()
})
</script>

<style scoped>
.home {
  padding: 20px;
}

h1 {
  margin-bottom: 30px;
  color: #333;
  text-align: center;
}

.activity-section {
  margin-bottom: 40px;
}

.activity-section h2,
.hot-products h2 {
  margin-bottom: 20px;
  color: #333;
  border-bottom: 2px solid #4CAF50;
  padding-bottom: 10px;
}

.activity-list {
  display: flex;
  gap: 20px;
  overflow-x: auto;
  padding: 10px 0;
}

.activity-item {
  flex: 1;
  min-width: 300px;
  padding: 20px;
  background-color: #f9f9f9;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.activity-item h3 {
  margin-bottom: 10px;
  color: #333;
}

.activity-item p {
  margin-bottom: 15px;
  color: #666;
}

.activity-date {
  font-size: 14px;
  color: #999;
}

.product-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
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
  height: 150px;
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
}
</style>