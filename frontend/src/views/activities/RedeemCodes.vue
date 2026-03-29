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

<script setup>
import { ref, reactive, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { activityApi } from '../../api/auth';

const router = useRouter();
const route = useRoute();

// 活动ID
const activityId = route.params.id;

// 搜索表单
const searchForm = reactive({
  status: '',
  code: ''
});

// 兑换码列表
const redeemCodesList = ref([]);

// 分页信息
const pagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

// 获取兑换码列表
const getRedeemCodes = async () => {
  try {
    const params = {
      page: pagination.currentPage,
      page_size: pagination.pageSize,
      status: searchForm.status,
      code: searchForm.code
    };
    const response = await activityApi.getRedeemCodes(activityId, params);
    redeemCodesList.value = response.list || [];
    pagination.total = response.total || 0;
  } catch (error) {
    console.error('获取兑换码列表失败:', error);
    ElMessage.error('获取兑换码列表失败');
  }
};

// 搜索
const handleSearch = () => {
  pagination.currentPage = 1;
  getRedeemCodes();
};

// 重置表单
const resetForm = () => {
  searchForm.status = '';
  searchForm.code = '';
  pagination.currentPage = 1;
  getRedeemCodes();
};

// 分页大小变化
const handleSizeChange = (size) => {
  pagination.pageSize = size;
  getRedeemCodes();
};

// 当前页变化
const handleCurrentChange = (current) => {
  pagination.currentPage = current;
  getRedeemCodes();
};

// 生成兑换码
const handleGenerateCodes = () => {
  router.push(`/home/activities/${activityId}/redeem-codes/generate`);
};

// 导入导出兑换码
const handleImportExportCodes = () => {
  router.push(`/home/activities/${activityId}/redeem-codes/import-export`);
};

// 更新兑换码状态
const handleUpdateStatus = async (id, status) => {
  ElMessageBox.confirm('确定要作废这个兑换码吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await activityApi.updateRedeemCodeStatus(id, { status });
      ElMessage.success('操作成功');
      getRedeemCodes();
    } catch (error) {
      console.error('更新兑换码状态失败:', error);
      ElMessage.error('更新兑换码状态失败');
    }
  }).catch(() => {
    // 取消操作
  });
};

// 初始加载
onMounted(() => {
  getRedeemCodes();
});
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