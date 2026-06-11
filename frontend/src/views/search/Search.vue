<template>
  <div class="search-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>综合搜索</span>
        </div>
      </template>

      <!-- 搜索输入框 -->
      <div class="search-bar">
        <el-input
          v-model="keyword"
          placeholder="请输入搜索关键词"
          clearable
          size="large"
          @keyup.enter="handleSearch"
          @clear="handleClear"
        >
          <template #append>
            <el-button @click="handleSearch" :loading="loading">搜索</el-button>
          </template>
        </el-input>
      </div>

      <!-- Tab 切换 -->
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane label="商品" name="products" />
        <el-tab-pane label="订单" name="orders" />
        <el-tab-pane label="用户" name="users" />
        <el-tab-pane label="客户" name="customers" />
      </el-tabs>

      <!-- 搜索结果 -->
      <div v-if="hasSearched">
        <!-- 商品表格 -->
        <el-table v-if="activeTab === 'products'" :data="results" style="width: 100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column label="商品名称" min-width="200">
            <template #default="scope">
              <span>{{ scope.row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="category_name" label="分类" width="120" />
          <el-table-column label="价格" width="100">
            <template #default="scope">
              ¥{{ scope.row.price }}
            </template>
          </el-table-column>
          <el-table-column prop="stock" label="库存" width="80" />
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '激活' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <!-- 订单表格 -->
        <el-table v-if="activeTab === 'orders'" :data="results" style="width: 100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="order_no" label="订单号" min-width="200" />
          <el-table-column prop="customer_name" label="客户" min-width="120" />
          <el-table-column label="商品" min-width="200">
            <template #default="scope">
              <div v-if="scope.row.items && scope.row.items.length > 0" class="order-products-cell">
                <div v-for="(item, idx) in scope.row.items.slice(0, 3)" :key="idx" class="order-product-item">
                  <img
                    v-if="item.product_image"
                    :src="item.product_image"
                    :alt="item.product_name"
                    class="product-thumb"
                  />
                  <el-icon v-else :size="24" color="#c0c4cc"><Picture /></el-icon>
                  <span class="product-name">{{ item.product_name }}</span>
                </div>
                <span v-if="scope.row.items.length > 3" class="more-items">等{{ scope.row.items.length }}件商品</span>
              </div>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="金额" width="120">
            <template #default="scope">
              ¥{{ scope.row.total_amount }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="orderStatusType(scope.row.status)">
                {{ orderStatusText(scope.row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180" />
        </el-table>

        <!-- 用户表格 -->
        <el-table v-if="activeTab === 'users'" :data="results" style="width: 100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" width="150" />
          <el-table-column prop="email" label="邮箱" width="200" />
          <el-table-column prop="role_name" label="角色" width="120" />
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '激活' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <!-- 客户表格 -->
        <el-table v-if="activeTab === 'customers'" :data="results" style="width: 100%" v-loading="loading">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="username" label="用户名" min-width="150" />
          <el-table-column prop="nickname" label="昵称" min-width="150" />
          <el-table-column prop="phone" label="手机号" min-width="150" />
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '激活' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>

        <!-- 空状态 -->
        <el-empty v-if="!loading && results.length === 0" description="未找到相关结果" />

        <!-- 分页 -->
        <div class="pagination-wrapper" v-if="total > 0">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handleSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Picture } from '@element-plus/icons-vue'
import { searchAPI } from '@/api/auth'

// 搜索关键词
const keyword = ref('')
// 当前激活的 Tab
const activeTab = ref('products')
// 搜索结果列表
const results = ref([])
// 加载状态
const loading = ref(false)
// 是否已执行过搜索
const hasSearched = ref(false)
// 当前页码
const currentPage = ref(1)
// 每页条数
const pageSize = ref(10)
// 总记录数
const total = ref(0)

// 执行搜索
const handleSearch = () => {
  if (!keyword.value.trim()) {
    ElMessage.warning('请输入搜索关键词')
    return
  }
  currentPage.value = 1
  fetchResults()
}

// 清除搜索
const handleClear = () => {
  keyword.value = ''
  results.value = []
  total.value = 0
  hasSearched.value = false
  currentPage.value = 1
}

// 切换Tab
const handleTabChange = () => {
  if (hasSearched.value) {
    currentPage.value = 1
    fetchResults()
  }
}

// 获取搜索结果
const fetchResults = async () => {
  loading.value = true
  try {
    const params = {
      keyword: keyword.value,
      page: currentPage.value,
      page_size: pageSize.value
    }
    const apiMap = {
      products: searchAPI.searchProducts,
      orders: searchAPI.searchOrders,
      users: searchAPI.searchUsers,
      customers: searchAPI.searchCustomers
    }
    const apiFn = apiMap[activeTab.value]
    const data = await apiFn(params)
    results.value = data.data || data.items || []
    total.value = data.total || 0
    hasSearched.value = true
  } catch (error) {
    results.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

// 每页条数变化
const handleSizeChange = () => {
  currentPage.value = 1
  fetchResults()
}

// 页码变化
const handlePageChange = () => {
  fetchResults()
}

// 订单状态标签类型
const orderStatusType = (status) => {
  const map = {
    pending: 'warning',
    paid: 'primary',
    shipped: '',
    completed: 'success',
    cancelled: 'danger'
  }
  return map[status] || ''
}

// 订单状态文本
const orderStatusText = (status) => {
  const map = {
    pending: '待支付',
    paid: '已支付',
    shipped: '已发货',
    completed: '已完成',
    cancelled: '已取消'
  }
  return map[status] || status
}
</script>

<style scoped>
.search-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.search-bar {
  margin-bottom: 20px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.order-products-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.order-product-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.product-thumb {
  width: 40px;
  height: 40px;
  object-fit: cover;
  border-radius: 4px;
  border: 1px solid #eee;
}

.product-name {
  font-size: 12px;
  color: #333;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 140px;
}

.more-items {
  font-size: 11px;
  color: #999;
}
</style>
