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
              <span v-html="scope.row.highlight_name || scope.row.name"></span>
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
          <el-table-column prop="order_no" label="订单号" width="200" />
          <el-table-column prop="customer_name" label="客户" width="120" />
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
          <el-table-column prop="username" label="用户名" width="150" />
          <el-table-column prop="nickname" label="昵称" width="150" />
          <el-table-column prop="phone" label="手机号" width="150" />
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

<script>
import { searchAPI } from '@/api/auth'

export default {
  name: 'Search',
  data() {
    return {
      keyword: '',
      activeTab: 'products',
      results: [],
      loading: false,
      hasSearched: false,
      currentPage: 1,
      pageSize: 10,
      total: 0
    }
  },
  methods: {
    // 执行搜索
    handleSearch() {
      if (!this.keyword.trim()) {
        this.$message.warning('请输入搜索关键词')
        return
      }
      this.currentPage = 1
      this.fetchResults()
    },
    // 切换Tab
    handleTabChange() {
      if (this.hasSearched) {
        this.currentPage = 1
        this.fetchResults()
      }
    },
    // 获取搜索结果
    async fetchResults() {
      this.loading = true
      try {
        const params = {
          keyword: this.keyword,
          page: this.currentPage,
          page_size: this.pageSize
        }
        const apiMap = {
          products: searchAPI.searchProducts,
          orders: searchAPI.searchOrders,
          users: searchAPI.searchUsers,
          customers: searchAPI.searchCustomers
        }
        const apiFn = apiMap[this.activeTab]
        const data = await apiFn(params)
        this.results = data.data || data.items || []
        this.total = data.total || 0
        this.hasSearched = true
      } catch (error) {
        this.results = []
        this.total = 0
      } finally {
        this.loading = false
      }
    },
    // 每页条数变化
    handleSizeChange() {
      this.currentPage = 1
      this.fetchResults()
    },
    // 页码变化
    handlePageChange() {
      this.fetchResults()
    },
    // 订单状态标签类型
    orderStatusType(status) {
      const map = {
        pending: 'warning',
        paid: 'primary',
        shipped: '',
        completed: 'success',
        cancelled: 'danger'
      }
      return map[status] || ''
    },
    // 订单状态文本
    orderStatusText(status) {
      const map = {
        pending: '待支付',
        paid: '已支付',
        shipped: '已发货',
        completed: '已完成',
        cancelled: '已取消'
      }
      return map[status] || status
    }
  }
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
</style>
