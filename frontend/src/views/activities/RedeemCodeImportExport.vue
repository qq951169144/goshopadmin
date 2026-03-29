<template>
  <div class="redeem-code-import-export-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>兑换码导入导出</span>
          <el-button @click="handleBack">返回列表</el-button>
        </div>
      </template>
      
      <el-tabs type="border-card">
        <!-- 导出标签页 -->
        <el-tab-pane label="导出兑换码">
          <el-form :model="exportForm" :rules="exportRules" ref="exportFormRef" label-width="120px">
            <el-form-item label="导出数量" prop="count">
              <el-input-number v-model="exportForm.count" :min="1" :max="10000" :step="1" />
            </el-form-item>
            
            <el-form-item label="兑换码状态" prop="status">
              <el-select v-model="exportForm.status" placeholder="选择状态">
                <el-option label="全部" value=""></el-option>
                <el-option label="未使用" value="unused"></el-option>
                <el-option label="已使用" value="used"></el-option>
                <el-option label="已过期" value="expired"></el-option>
              </el-select>
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="handleExport" :loading="exportLoading">导出</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
        
        <!-- 导入标签页 -->
        <el-tab-pane label="导入兑换码">
          <el-form :model="importForm" :rules="importRules" ref="importFormRef" label-width="120px">
            <el-form-item label="导入文件" prop="file">
              <el-upload
                class="upload-demo"
                action=""
                :auto-upload="false"
                :on-change="handleFileChange"
                :show-file-list="false"
                accept=".csv,.txt"
              >
                <el-button type="primary">选择文件</el-button>
                <template #tip>
                  <div class="el-upload__tip">
                    请上传CSV或TXT格式文件，每行一个兑换码
                  </div>
                </template>
              </el-upload>
              <div v-if="importForm.fileName" class="mt-2">
                已选择文件: {{ importForm.fileName }}
              </div>
            </el-form-item>
            
            <el-form-item>
              <el-button type="primary" @click="handleImport" :loading="importLoading">导入</el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { activityApi } from '../../api/auth';

const router = useRouter();
const route = useRoute();
const exportFormRef = ref(null);
const importFormRef = ref(null);

// 活动ID
const activityId = route.params.id;

// 导出表单
const exportForm = reactive({
  count: 100,
  status: ''
});

// 导出表单规则
const exportRules = {
  count: [
    { required: true, message: '请输入导出数量', trigger: 'blur' },
    { type: 'number', min: 1, max: 10000, message: '导出数量必须在 1-10000 之间', trigger: 'blur' }
  ]
};

// 导入表单
const importForm = reactive({
  file: null,
  fileName: ''
});

// 导入表单规则
const importRules = {
  file: [
    { required: true, message: '请选择文件', trigger: 'change' }
  ]
};

// 加载状态
const exportLoading = ref(false);
const importLoading = ref(false);

// 处理文件选择
const handleFileChange = (file) => {
  importForm.file = file.raw;
  importForm.fileName = file.name;
};

// 导出兑换码
const handleExport = async () => {
  if (!exportFormRef.value) return;
  
  await exportFormRef.value.validate(async (valid) => {
    if (valid) {
      exportLoading.value = true;
      try {
        const params = {
          count: exportForm.count,
          status: exportForm.status
        };
        const response = await activityApi.exportRedeemCodes(activityId, params);
        
        // 处理文件下载
        const blob = new Blob([response], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `redeem_codes_${new Date().getTime()}.csv`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        
        ElMessage.success('导出成功');
      } catch (error) {
        console.error('导出兑换码失败:', error);
        ElMessage.error('导出兑换码失败');
      } finally {
        exportLoading.value = false;
      }
    }
  });
};

// 导入兑换码
const handleImport = async () => {
  if (!importFormRef.value) return;
  
  await importFormRef.value.validate(async (valid) => {
    if (valid) {
      importLoading.value = true;
      try {
        const formData = new FormData();
        formData.append('file', importForm.file);
        
        const response = await activityApi.importRedeemCodes(activityId, formData);
        ElMessage.success(`成功导入 ${response.success_count} 个兑换码，失败 ${response.failed_count} 个`);
      } catch (error) {
        console.error('导入兑换码失败:', error);
        ElMessage.error('导入兑换码失败');
      } finally {
        importLoading.value = false;
      }
    }
  });
};

// 返回列表
const handleBack = () => {
  router.push(`/home/activities/${activityId}/redeem-codes`);
};
</script>

<style scoped>
.redeem-code-import-export-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mt-2 {
  margin-top: 10px;
}
</style>