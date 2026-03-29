<template>
  <div class="redeem-code-verify-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>兑换码核销</span>
        </div>
      </template>
      
      <el-form :model="verifyForm" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="兑换码" prop="code">
          <el-input v-model="verifyForm.code" placeholder="请输入兑换码" />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleVerify" :loading="loading">核销</el-button>
          <el-button @click="resetForm">重置</el-button>
        </el-form-item>
      </el-form>
      
      <!-- 核销结果 -->
      <template v-if="verifyResult">
        <el-divider>核销结果</el-divider>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="兑换码">{{ verifyResult.code }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag v-if="verifyResult.status === 'unused'" type="success">未使用</el-tag>
            <el-tag v-else-if="verifyResult.status === 'used'" type="info">已使用</el-tag>
            <el-tag v-else-if="verifyResult.status === 'expired'" type="warning">已过期</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="活动名称">{{ verifyResult.activity_name }}</el-descriptions-item>
          <el-descriptions-item label="商品信息">
            <div v-for="product in verifyResult.products" :key="product.id">
              {{ product.product_name }} - {{ product.sku_name }}
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="使用时间">{{ verifyResult.used_at || '-' }}</el-descriptions-item>
          <el-descriptions-item label="使用用户">{{ verifyResult.user_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="核销结果">
            <el-tag v-if="verifyResult.verify_success" type="success">核销成功</el-tag>
            <el-tag v-else type="danger">核销失败</el-tag>
            <div v-if="verifyResult.verify_message" class="mt-1">{{ verifyResult.verify_message }}</div>
          </el-descriptions-item>
        </el-descriptions>
      </template>
      
      <!-- 核销记录 -->
      <el-divider>核销记录</el-divider>
      <el-table :data="verifyLogs" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80"></el-table-column>
        <el-table-column prop="code" label="兑换码"></el-table-column>
        <el-table-column prop="activity_name" label="活动名称" width="200"></el-table-column>
        <el-table-column prop="verify_result" label="核销结果" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.verify_result === 'success'" type="success">成功</el-tag>
            <el-tag v-else type="danger">失败</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="verify_message" label="核销消息"></el-table-column>
        <el-table-column prop="verify_time" label="核销时间" width="180"></el-table-column>
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
import { ElMessage } from 'element-plus';
import { activityApi } from '../../api/auth';

const formRef = ref(null);

// 核销表单
const verifyForm = reactive({
  code: ''
});

// 表单规则
const rules = {
  code: [
    { required: true, message: '请输入兑换码', trigger: 'blur' }
  ]
};

// 加载状态
const loading = ref(false);

// 核销结果
const verifyResult = ref(null);

// 核销记录
const verifyLogs = ref([]);

// 分页信息
const pagination = reactive({
  currentPage: 1,
  pageSize: 10,
  total: 0
});

// 核销兑换码
const handleVerify = async () => {
  if (!formRef.value) return;
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true;
      try {
        const data = {
          code: verifyForm.code
        };
        const response = await activityApi.verifyRedeemCode(data);
        verifyResult.value = response;
        
        if (response.verify_success) {
          ElMessage.success('核销成功');
        } else {
          ElMessage.error('核销失败: ' + response.verify_message);
        }
        
        // 刷新核销记录
        getVerifyLogs();
      } catch (error) {
        console.error('核销兑换码失败:', error);
        ElMessage.error('核销兑换码失败');
      } finally {
        loading.value = false;
      }
    }
  });
};

// 重置表单
const resetForm = () => {
  verifyForm.code = '';
  verifyResult.value = null;
};

// 获取核销记录
const getVerifyLogs = async () => {
  try {
    const params = {
      page: pagination.currentPage,
      page_size: pagination.pageSize
    };
    const response = await activityApi.getRedeemCodeLogs(params);
    verifyLogs.value = response.list || [];
    pagination.total = response.total || 0;
  } catch (error) {
    console.error('获取核销记录失败:', error);
    ElMessage.error('获取核销记录失败');
  }
};

// 分页大小变化
const handleSizeChange = (size) => {
  pagination.pageSize = size;
  getVerifyLogs();
};

// 当前页变化
const handleCurrentChange = (current) => {
  pagination.currentPage = current;
  getVerifyLogs();
};

// 初始加载
onMounted(() => {
  getVerifyLogs();
});
</script>

<style scoped>
.redeem-code-verify-container {
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

.mt-1 {
  margin-top: 5px;
}
</style>