<template>
  <div class="orders-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>订单管理</span>
        </div>
      </template>

      <div class="filter-section">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-input v-model="keyword" placeholder="搜索订单号/商品名称" clearable @keyup.enter="handleSearch" />
          </el-col>
          <el-col :span="4">
            <el-select v-model="status" placeholder="订单状态" clearable @change="handleSearch">
              <el-option label="待支付" value="pending" />
              <el-option label="已支付" value="paid" />
              <el-option label="已发货" value="shipped" />
              <el-option label="已完成" value="completed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-col>
          <el-col :span="4">
            <el-select v-model="paymentStatus" placeholder="支付状态" clearable @change="handleSearch">
              <el-option label="待支付" value="pending" />
              <el-option label="支付成功" value="success" />
              <el-option label="支付失败" value="failed" />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
              @change="handleSearch"
            />
          </el-col>
          <el-col :span="4">
            <el-button type="primary" @click="handleSearch" :loading="loading">搜索</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-col>
        </el-row>
      </div>

      <el-table :data="orders" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="order_no" label="订单号" width="200" />
        <el-table-column prop="customer_name" label="客户" width="120" />
        <el-table-column label="金额" width="120">
          <template #default="scope">¥{{ scope.row.total_amount }}</template>
        </el-table-column>
        <el-table-column label="订单状态" width="100">
          <template #default="scope">
            <el-tag :type="orderStatusType(scope.row.status)">{{ orderStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="支付状态" width="100">
          <template #default="scope">
            <el-tag :type="paymentStatusType(scope.row.payment_status)">{{ paymentStatusText(scope.row.payment_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180" />
      </el-table>

      <el-empty v-if="!loading && hasSearched && orders.length === 0" description="未找到相关订单" />

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
import { ref, onMounted } from 'vue'
import { searchAPI } from '@/api/auth'

const keyword = ref('')
const status = ref('')
const paymentStatus = ref('')
const dateRange = ref(null)
const orders = ref([])
const loading = ref(false)
const hasSearched = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const handleSearch = () => { currentPage.value = 1; fetchOrders() }
const handleReset = () => {
  keyword.value = ''
  status.value = ''
  paymentStatus.value = ''
  dateRange.value = null
  currentPage.value = 1
  fetchOrders()
}

const fetchOrders = async () => {
  loading.value = true
  try {
    const params = { keyword: keyword.value, page: currentPage.value, page_size: pageSize.value }
    if (status.value) params.status = status.value
    if (paymentStatus.value) params.payment_status = paymentStatus.value
    if (dateRange.value && dateRange.value[0]) params.start_date = dateRange.value[0]
    if (dateRange.value && dateRange.value[1]) params.end_date = dateRange.value[1]
    const response = await searchAPI.searchOrders(params)
    const data = response.data || response
    orders.value = data.items || data.data || []
    total.value = data.total || 0
    hasSearched.value = true
  } catch (error) {
    ElMessage.error('搜索订单失败，请稍后重试')
    orders.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleSizeChange = () => { currentPage.value = 1; fetchOrders() }
const handlePageChange = () => { fetchOrders() }

onMounted(() => {
  fetchOrders()
})

const orderStatusType = (s) => ({ pending: 'warning', paid: 'primary', shipped: '', completed: 'success', cancelled: 'danger' }[s] || '')
const orderStatusText = (s) => ({ pending: '待支付', paid: '已支付', shipped: '已发货', completed: '已完成', cancelled: '已取消' }[s] || s)
const paymentStatusType = (s) => ({ pending: 'warning', success: 'success', failed: 'danger' }[s] || '')
const paymentStatusText = (s) => ({ pending: '待支付', success: '支付成功', failed: '支付失败' }[s] || s)
</script>

<style scoped>
.orders-container { padding: 20px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.filter-section { margin-bottom: 20px; }
.pagination-wrapper { display: flex; justify-content: flex-end; margin-top: 20px; }
</style>
