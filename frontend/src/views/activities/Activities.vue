<template>
  <div class="activities-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>活动管理</span>
          <el-button type="primary" @click="handleCreateActivity">创建活动</el-button>
        </div>
      </template>
      
      <!-- 筛选表单 -->
      <el-form :inline="true" :model="searchForm" class="mb-4">
        <el-form-item label="活动类型">
          <el-select v-model="searchForm.type" placeholder="选择活动类型" clearable style="width: 150px;">

            <el-option label="秒杀活动1" value="seckill"></el-option>
            <el-option label="兑换码活动" value="redeem_code"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="活动状态">
          <el-select v-model="searchForm.status" placeholder="选择活动状态" clearable style="width: 120px;">
            <el-option label="激活" value="active"></el-option>
            <el-option label="禁用" value="inactive"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
      
      <!-- 活动列表 -->
      <el-table :data="activitiesList" style="width: 100%">
        <el-table-column prop="id" label="活动ID" width="100"></el-table-column>
        <el-table-column prop="name" label="活动名称"></el-table-column>
        <el-table-column prop="type" label="活动类型" width="120">
          <template #default="scope">
            <el-tag v-if="scope.row.type === 'seckill'" type="primary">秒杀活动</el-tag>
            <el-tag v-else-if="scope.row.type === 'redeem_code'" type="success">兑换码活动</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="开始时间" width="180"></el-table-column>
        <el-table-column prop="end_time" label="结束时间" width="180"></el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.status === 'active'" type="success">激活</el-tag>
            <el-tag v-else-if="scope.row.status === 'inactive'" type="info">禁用</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="handleViewDetail(scope.row.id)">查看</el-button>
            <el-button size="small" type="primary" @click="handleEditActivity(scope.row.id)">编辑</el-button>
            <el-button 
              v-if="scope.row.status === 'inactive'" 
              size="small" 
              type="success" 
              @click="handleUpdateStatus(scope.row.id, 'active')"
            >
              激活
            </el-button>
            <el-button 
              v-else-if="scope.row.status === 'active'" 
              size="small" 
              type="warning" 
              @click="handleUpdateStatus(scope.row.id, 'inactive')"
            >
              禁用
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
import { activityApi } from '../../api/auth';

// 搜索表单 - 使用 reactive 确保响应式
const searchForm = reactive({
  type: '',
  status: ''
});

// 活动列表
const activitiesList = ref([]);

// 分页信息
const pagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

// 获取活动列表
const getActivities = async () => {
  try {
    const params = {
      page: pagination.currentPage,
      page_size: pagination.pageSize
    };
    // 只添加有值的筛选条件
    if (searchForm.type) {
      params.type = searchForm.type;
    }
    if (searchForm.status) {
      params.status = searchForm.status;
    }
    const response = await activityApi.getActivities(params);
    activitiesList.value = response.list || [];
    pagination.total = response.total || 0;
  } catch (error) {
    console.error('获取活动列表失败:', error);
    ElMessage.error('获取活动列表失败');
  }
};

// 搜索
const handleSearch = () => {
  pagination.currentPage = 1;
  getActivities();
};

// 重置表单
const resetForm = () => {
  searchForm.type = '';
  searchForm.status = '';
  pagination.currentPage = 1;
  getActivities();
};

// 分页大小变化
const handleSizeChange = (size) => {
  pagination.pageSize = size;
  getActivities();
};

// 当前页变化
const handleCurrentChange = (current) => {
  pagination.currentPage = current;
  getActivities();
};

// 创建活动
const handleCreateActivity = () => {
  emit('create-activity');
};

// 编辑活动
const handleEditActivity = (id) => {
  const activity = activitiesList.value.find(item => item.id === id);
  emit('edit-activity', activity || { id });
};

// 查看活动详情
const handleViewDetail = (id) => {
  const activity = activitiesList.value.find(item => item.id === id);
  emit('view-activity', activity || { id });
};

// 删除活动
const handleDeleteActivity = (id) => {
  ElMessageBox.confirm('确定要删除这个活动吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await activityApi.deleteActivity(id);
      ElMessage.success('删除成功');
      getActivities();
    } catch (error) {
      console.error('删除活动失败:', error);
      ElMessage.error('删除活动失败');
    }
  }).catch(() => {
    // 取消删除
  });
};

// 更新活动状态
const handleUpdateStatus = async (id, status) => {
  try {
    await activityApi.updateActivityStatus(id, { status });
    ElMessage.success('状态更新成功');
    // 添加短暂延迟确保后端数据已更新
    setTimeout(() => {
      getActivities();
    }, 300);
  } catch (error) {
    console.error('更新活动状态失败:', error);
    ElMessage.error('更新活动状态失败');
  }
};

// 定义 emit
const emit = defineEmits(['create-activity', 'edit-activity', 'view-activity']);

onMounted(() => {
  getActivities();
});
</script>

<style scoped>
.activities-container {
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>
