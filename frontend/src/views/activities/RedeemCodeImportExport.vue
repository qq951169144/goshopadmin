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

<script>
import { activityApi } from '../../api/auth'

export default {
  name: 'RedeemCodeImportExport',
  props: {
    activityId: {
      type: Number,
      default: null
    }
  },
  computed: {
    activityIdValue() {
      if (this.activityId) {
        return this.activityId;
      }
      return parseInt(this.$route.params.id);
    }
  },
  data() {
    return {
      exportForm: {
        count: 100,
        status: ''
      },
      exportRules: {
        count: [
          { required: true, message: '请输入导出数量', trigger: 'blur' },
          { type: 'number', min: 1, max: 10000, message: '导出数量必须在 1-10000 之间', trigger: 'blur' }
        ]
      },
      importForm: {
        file: null,
        fileName: ''
      },
      importRules: {
        file: [
          { required: true, message: '请选择文件', trigger: 'change' }
        ]
      },
      exportLoading: false,
      importLoading: false
    }
  },
  methods: {
    // 处理文件选择
    handleFileChange(file) {
      this.importForm.file = file.raw;
      this.importForm.fileName = file.name;
    },
    
    // 导出兑换码
    handleExport() {
      this.$refs.exportFormRef.validate((valid) => {
        if (valid) {
          this.exportLoading = true;
          const params = {
            count: this.exportForm.count,
            status: this.exportForm.status
          };
          activityApi.exportRedeemCodes(this.activityIdValue, params).then(response => {
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
            
            this.$message.success('导出成功');
          }).catch(() => {
            this.$message.error('导出兑换码失败');
          }).finally(() => {
            this.exportLoading = false;
          });
        }
      });
    },
    
    // 导入兑换码
    handleImport() {
      this.$refs.importFormRef.validate((valid) => {
        if (valid) {
          this.importLoading = true;
          const formData = new FormData();
          formData.append('file', this.importForm.file);
          
          activityApi.importRedeemCodes(this.activityIdValue, formData).then(response => {
            this.$message.success(`成功导入 ${response.success_count} 个兑换码，失败 ${response.failed_count} 个`);
          }).catch(() => {
            this.$message.error('导入兑换码失败');
          }).finally(() => {
            this.importLoading = false;
          });
        }
      });
    },
    
    // 返回列表
    handleBack() {
      this.$emit('back');
    }
  }
}
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