<template>
  <div class="redeem-code-generate-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>生成兑换码</span>
          <el-button @click="handleBack">返回列表</el-button>
        </div>
      </template>
      
      <el-form :model="generateForm" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="生成数量" prop="count">
          <el-input-number v-model="generateForm.count" :min="1" :max="10000" :step="1" />
        </el-form-item>
        
        <el-form-item label="兑换码类型" prop="type">
          <el-select v-model="generateForm.type" placeholder="选择类型">
            <el-option label="数字" value="numeric"></el-option>
            <el-option label="字母" value="alpha"></el-option>
            <el-option label="混合" value="alphanumeric"></el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="兑换码长度" prop="length">
          <el-input-number v-model="generateForm.length" :min="6" :max="20" :step="1" />
        </el-form-item>
        
        <el-form-item label="排除字符">
          <el-input v-model="generateForm.exclude_chars" placeholder="例如：01IOl" />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleGenerate" :loading="loading">生成兑换码</el-button>
          <el-button @click="handleBack">取消</el-button>
        </el-form-item>
      </el-form>
      
      <!-- 生成结果 -->
      <template v-if="generateResult.length > 0">
        <el-divider>生成结果</el-divider>
        <div class="result-container">
          <el-button @click="handleExportCodes">导出兑换码</el-button>
          <el-tag class="ml-2" size="small">共生成 {{ generateResult.length }} 个兑换码</el-tag>
          
          <el-table :data="generateResult" style="width: 100%; margin-top: 10px">
            <el-table-column prop="code" label="兑换码"></el-table-column>
          </el-table>
        </div>
      </template>
    </el-card>
  </div>
</template>

<script>
import { activityApi } from '../../api/auth'

export default {
  name: 'RedeemCodeGenerate',
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
      generateForm: {
        count: 100,
        type: 'alphanumeric',
        length: 12,
        exclude_chars: ''
      },
      rules: {
        count: [
          { required: true, message: '请输入生成数量', trigger: 'blur' },
          { type: 'number', min: 1, max: 10000, message: '生成数量必须在 1-10000 之间', trigger: 'blur' }
        ],
        type: [
          { required: true, message: '请选择兑换码类型', trigger: 'change' }
        ],
        length: [
          { required: true, message: '请输入兑换码长度', trigger: 'blur' },
          { type: 'number', min: 6, max: 20, message: '兑换码长度必须在 6-20 之间', trigger: 'blur' }
        ]
      },
      loading: false,
      generateResult: []
    }
  },
  mounted() {
    // 可以在这里加载活动信息，获取默认的兑换码规则
  },
  methods: {
    // 生成兑换码
    handleGenerate() {
      this.$refs.formRef.validate((valid) => {
        if (valid) {
          this.loading = true;
          const data = {
            activity_id: this.activityIdValue,
            quantity: this.generateForm.count,
            code_type: this.generateForm.type,
            code_length: this.generateForm.length,
            exclude_chars: this.generateForm.exclude_chars
          };
          activityApi.generateRedeemCodes(this.activityIdValue, data).then(response => {
            // 后端返回的是字符串数组，需要映射为对象数组格式以适配表格和复制功能
            this.generateResult = (response.codes || []).map(code => ({ code }));
            this.$message.success(`成功生成 ${this.generateResult.length} 个兑换码`);
          }).catch(() => {
            this.$message.error('生成兑换码失败');
          }).finally(() => {
            this.loading = false;
          });
        }
      });
    },
    
    // 导出兑换码
    handleExportCodes() {
      if (this.generateResult.length === 0) {
        this.$message.warning('没有可导出的兑换码');
        return;
      }
      
      // 这里应该调用导出API，暂时使用前端导出
      const codes = this.generateResult.map(item => item.code).join('\n');
      const blob = new Blob([codes], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `redeem_codes_${new Date().getTime()}.txt`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      this.$message.success('导出成功');
    },
    
    // 取消
    handleBack() {
      this.$emit('back');
    }
  }
}
</script>

<style scoped>
.redeem-code-generate-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.result-container {
  margin-top: 20px;
}

.ml-2 {
  margin-left: 10px;
}
</style>