<template>
  <div class="customers-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>C端客户管理</span>
        </div>
      </template>

      <div class="filter-section">
        <el-row :gutter="16">
          <el-col :span="8">
            <el-input v-model="keyword" placeholder="搜索用户名/昵称/邮箱/手机号" clearable @keyup.enter="handleSearch" />
          </el-col>
          <el-col :span="4">
            <el-select v-model="status" placeholder="状态" clearable @change="handleSearch">
              <el-option label="活跃" value="active" />
              <el-option label="禁用" value="inactive" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-button type="primary" @click="handleSearch" :loading="loading">搜索</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-col>
        </el-row>
      </div>

      <el-table :data="customers" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="username" label="用户名" width="150" />
        <el-table-column prop="nickname" label="昵称" width="150" />
        <el-table-column prop="email" label="邮箱" width="200" />
        <el-table-column prop="phone" label="手机号" width="150" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
              {{ scope.row.status === 'active' ? '活跃' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="180" />
      </el-table>

      <el-empty v-if="!loading && hasSearched && customers.length === 0" description="未找到相关客户" />

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
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { searchAPI } from '@/api/auth'

const keyword = ref('')
const status = ref('')
const customers = ref([])
const loading = ref(false)
const hasSearched = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const handleSearch = () => { currentPage.value = 1; fetchCustomers() }
const handleReset = () => {
  keyword.value = ''
  status.value = ''
  customers.value = []
  total.value = 0
  hasSearched.value = false
}

const fetchCustomers = async () => {
  loading.value = true
  try {
    const params = { keyword: keyword.value, page: currentPage.value, page_size: pageSize.value }
    if (status.value) params.status = status.value
    const response = await searchAPI.searchCustomers(params)
    const data = response.data || response
    customers.value = data.items || data.data || []
    total.value = data.total || 0
    hasSearched.value = true
  } catch (error) {
    console.error('搜索客户失败:', error)
    customers.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleSizeChange = () => { currentPage.value = 1; fetchCustomers() }
const handlePageChange = () => { fetchCustomers() }
</script>

<style scoped>
.customers-container { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.filter-section { margin-bottom: 20px; }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 20px; }
</style>
