<template>
  <div class="redeem-codes-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>兑换码管理</span>
          <el-button type="primary" @click="handleGenerateCodes">生成兑换码</el-button>
          <el-button @click="handleImportExportCodes">导入导出兑换码</el-button>
        </div>
      </template>
      
      <!-- 筛选表单 -->
      <el-form :inline="true" :model="searchForm" class="mb-4">
        <el-form-item label="兑换码状态">
          <el-select v-model="searchForm.status" placeholder="选择状态">
            <el-option label="未使用" value="unused"></el-option>
            <el-option label="已使用" value="used"></el-option>
            <el-option label="已过期" value="expired"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="兑换码">
          <el-input v-model="searchForm.code" placeholder="请输入兑换码" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
      
      <!-- 兑换码列表 -->
      <el-table :data="redeemCodesList" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="code" label="兑换码"></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.status === 'unused'" type="success">未使用</el-tag>
            <el-tag v-else-if="scope.row.status === 'used'" type="info">已使用</el-tag>
            <el-tag v-else-if="scope.row.status === 'expired'" type="warning">已过期</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180"></el-table-column>
        <el-table-column prop="used_at" label="使用时间" width="180">
          <template #default="scope">
            {{ scope.row.used_at || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="user_id" label="使用用户" width="100">
          <template #default="scope">
            {{ scope.row.user_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="scope">
            <el-button 
              v-if="scope.row.status === 'unused'" 
              size="small" 
              type="danger" 
              @click="handleUpdateStatus(scope.row.id, 'expired')"
            >
              作废
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="pagination.currentPage"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          :total="pagination.total"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script>
import { activityApi } from '../../api/auth'

export default {
  name: 'RedeemCodes',
  computed: {
    activityId() {
      return parseInt(this.$route.params.id);
    }
  },
  data() {
    return {
      searchForm: {
        status: '',
        code: ''
      },
      redeemCodesList: [],
      pagination: {
        currentPage: 1,
        pageSize: 10,
        total: 0
      }
    }
  },
  mounted() {
    this.getRedeemCodes()
  },
  methods: {
    // 获取兑换码列表
    getRedeemCodes() {
      try {
        const params = {
          page: this.pagination.currentPage,
          page_size: this.pagination.pageSize,
          status: this.searchForm.status,
          code: this.searchForm.code
        };
        activityApi.getRedeemCodes(this.activityId, params).then(response => {
          this.redeemCodesList = response.list || [];
          this.pagination.total = response.total || 0;
        });
      } catch (error) {
        console.error('获取兑换码列表失败:', error);
        this.$message.error('获取兑换码列表失败');
      }
    },
    
    // 搜索
    handleSearch() {
      this.pagination.currentPage = 1;
      this.getRedeemCodes();
    },
    
    // 重置表单
    resetForm() {
      this.searchForm.status = '';
      this.searchForm.code = '';
      this.pagination.currentPage = 1;
      this.getRedeemCodes();
    },
    
    // 分页大小变化
    handleSizeChange(size) {
      this.pagination.pageSize = size;
      this.getRedeemCodes();
    },
    
    // 当前页变化
    handleCurrentChange(current) {
      this.pagination.currentPage = current;
      this.getRedeemCodes();
    },
    
    // 生成兑换码
    handleGenerateCodes() {
      this.$emit('generate-redeem-codes', { id: this.activityId });
    },
    
    // 导入导出兑换码
    handleImportExportCodes() {
      this.$emit('import-export-redeem-codes', { id: this.activityId });
    },
    
    // 更新兑换码状态
    handleUpdateStatus(id, status) {
      this.$confirm('确定要作废这个兑换码吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        activityApi.updateRedeemCodeStatus(id, { status }).then(() => {
          this.$message.success('操作成功');
          this.getRedeemCodes();
        }).catch(error => {
          console.error('更新兑换码状态失败:', error);
          this.$message.error('更新兑换码状态失败');
        });
      }).catch(() => {
        // 取消操作
      });
    }
  }
}
</script>

<style scoped>
.redeem-codes-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>