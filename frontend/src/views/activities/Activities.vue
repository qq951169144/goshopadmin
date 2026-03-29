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
          <el-select v-model="searchForm.type" placeholder="选择活动类型">
            <el-option label="秒杀活动" value="seckill"></el-option>
            <el-option label="兑换码活动" value="redeem_code"></el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="活动状态">
          <el-select v-model="searchForm.status" placeholder="选择活动状态">
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

<script>
import { activityApi } from '../../api/auth';

export default {
  data() {
    return {
      // 搜索表单
      searchForm: {
        type: '',
        status: ''
      },
      // 活动列表
      activitiesList: [],
      // 分页信息
      pagination: {
        currentPage: 1,
        pageSize: 10,
        total: 0
      }
    };
  },

  methods: {
    // 获取活动列表
    async getActivities() {
      try {
        const params = {
          page: this.pagination.currentPage,
          page_size: this.pagination.pageSize
        };
        // 只添加有值的筛选条件
        if (this.searchForm.type) {
          params.type = this.searchForm.type;
        }
        if (this.searchForm.status) {
          params.status = this.searchForm.status;
        }
        const response = await activityApi.getActivities(params);
        this.activitiesList = response.list || [];
        this.pagination.total = response.total || 0;
      } catch (error) {
        console.error('获取活动列表失败:', error);
        this.$message.error('获取活动列表失败');
      }
    },
    // 搜索
    handleSearch() {
      this.pagination.currentPage = 1;
      this.getActivities();
    },
    // 重置表单
    resetForm() {
      this.searchForm.type = '';
      this.searchForm.status = '';
      this.pagination.currentPage = 1;
      this.getActivities();
    },
    // 分页大小变化
    handleSizeChange(size) {
      this.pagination.pageSize = size;
      this.getActivities();
    },
    // 当前页变化
    handleCurrentChange(current) {
      this.pagination.currentPage = current;
      this.getActivities();
    },
    // 创建活动
    handleCreateActivity() {
      this.$emit('create-activity');
    },
    // 编辑活动
    handleEditActivity(id) {
      this.$emit('edit-activity', { id });
    },
    // 查看活动详情
    handleViewDetail(id) {
      this.$emit('view-activity', { id });
    },
    // 删除活动
    handleDeleteActivity(id) {
      this.$confirm('确定要删除这个活动吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(async () => {
        try {
          await activityApi.deleteActivity(id);
          this.$message.success('删除成功');
          this.getActivities();
        } catch (error) {
          console.error('删除活动失败:', error);
          this.$message.error('删除活动失败');
        }
      }).catch(() => {
        // 取消删除
      });
    },
    // 更新活动状态
    async handleUpdateStatus(id, status) {
      try {
        await activityApi.updateActivityStatus(id, { status });
        this.$message.success('状态更新成功');
        this.getActivities();
      } catch (error) {
        console.error('更新活动状态失败:', error);
        this.$message.error('更新活动状态失败');
      }
    }
  },
  mounted() {
    this.getActivities();
  }
};
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