<template>
  <div class="skus-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-button @click="handleBack" size="small">返回</el-button>
            <span class="title">SKU管理 - {{ displayProductName }}</span>
          </div>
          <div class="header-right">
            <el-button type="success" @click="handleGenerateSKUs">自动生成SKU组合</el-button>
            <el-button type="primary" @click="handleAddSKU">手动添加SKU</el-button>
          </div>
        </div>
      </template>

      <!-- 规格预览 -->
      <div class="specs-preview" v-if="specifications.length > 0">
        <h4>已配置规格：</h4>
        <div v-for="spec in specifications" :key="spec.id" class="spec-row">
          <span class="spec-label">{{ spec.name }}:</span>
          <el-tag
            v-for="value in spec.values"
            :key="value.id"
            size="small"
            class="spec-value-tag"
          >
            {{ value.value }}
          </el-tag>
        </div>
      </div>
      <el-alert v-else title="该商品尚未配置规格，请先配置规格维度" type="warning" :closable="false" style="margin-bottom: 20px;" />

      <!-- SKU列表 -->
      <el-table :data="skus" style="width: 100%" v-loading="loading">
        <el-table-column prop="sku_code" label="SKU编码" min-width="200" />
        <el-table-column prop="price" label="售价" width="120">
          <template #default="scope">
            <span class="price">¥{{ scope.row.price }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="original_price" label="原价" width="120">
          <template #default="scope">
            <span v-if="scope.row.original_price > 0" class="original-price">¥{{ scope.row.original_price }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="stock" label="库存" width="100" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
              {{ scope.row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="is_activity" label="活动专用" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.is_activity" type="warning">是</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="handleEditSKU(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDeleteSKU(scope.row)">禁用</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <el-empty v-if="skus.length === 0 && !loading" description="暂无SKU，请先生成或手动添加" />
    </el-card>

    <!-- 自动生成SKU对话框 -->
    <el-dialog
      v-model="generateDialogVisible"
      title="自动生成SKU组合"
      width="600px"
    >
      <el-form :model="generateForm">
        <el-form-item label="基础价格">
          <el-input-number v-model="generateForm.base_price" :min="0" :precision="2" />
          <span class="form-tip">所有生成的SKU将使用此价格，可在生成后单独修改</span>
        </el-form-item>
      </el-form>
      
      <div class="preview-section" v-if="generatedSKUs.length > 0">
        <h4>预览生成的SKU组合：</h4>
        <el-table :data="generatedSKUs" size="small" max-height="300">
          <el-table-column prop="sku_code" label="SKU编码" />
          <el-table-column prop="price" label="价格" width="100" />
        </el-table>
      </div>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="generateDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handlePreviewGenerate">预览组合</el-button>
          <el-button type="success" @click="handleConfirmGenerate" :disabled="generatedSKUs.length === 0">确认生成</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- SKU编辑对话框 -->
    <el-dialog
      v-model="skuDialogVisible"
      :title="skuForm.id ? '编辑SKU' : '添加SKU'"
      width="500px"
    >
      <el-form :model="skuForm" :rules="skuRules" ref="skuFormRef">
        <el-form-item label="SKU编码" prop="sku_code">
          <el-input v-model="skuForm.sku_code" placeholder="请输入SKU编码" />
        </el-form-item>
        <el-form-item label="售价" prop="price">
          <el-input-number v-model="skuForm.price" :min="0" :precision="2" />
        </el-form-item>
        <el-form-item label="原价">
          <el-input-number v-model="skuForm.original_price" :min="0" :precision="2" />
        </el-form-item>
        <el-form-item label="库存" prop="stock">
          <el-input-number v-model="skuForm.stock" :min="0" :precision="0" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="skuForm.status">
            <el-radio label="active">启用</el-radio>
            <el-radio label="inactive">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <!-- 活动专用SKU -->
        <el-form-item label="活动专用SKU">
          <el-checkbox v-model="skuForm.is_activity" @change="handleActivityChange">是</el-checkbox>
        </el-form-item>
        
        <!-- 活动选择器 -->
        <el-form-item label="关联活动" v-if="skuForm.is_activity">
          <el-select v-model="skuForm.activity_id" placeholder="选择活动">
            <el-option
              v-for="activity in activities"
              :key="activity.id"
              :label="activity.name"
              :value="activity.id"
            />
          </el-select>
        </el-form-item>
        
        <!-- 规格组合选择 -->
        <el-form-item label="规格组合" v-if="!skuForm.id && specifications.length > 0">
          <div v-for="spec in specifications" :key="spec.id" class="spec-select-row">
            <span class="spec-name">{{ spec.name }}:</span>
            <el-select v-model="skuForm.spec_combinations[spec.id]" placeholder="请选择" size="small">
              <el-option
                v-for="value in spec.values"
                :key="value.id"
                :label="value.value"
                :value="value.id"
              />
            </el-select>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="skuDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveSKU">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { productApi, activityApi } from '@/api/auth'

export default {
  name: 'ProductSKUs',
  props: {
    productId: {
      type: Number,
      default: null
    },
    productName: {
      type: String,
      default: ''
    }
  },
  data() {
    return {
      localProductId: null,
      localProductName: '',
      specifications: [],
      skus: [],
      loading: false,
      generateDialogVisible: false,
      generateForm: {
        base_price: 0
      },
      generatedSKUs: [],
      skuDialogVisible: false,
      activities: [],
      skuForm: {
        id: null,
        sku_code: '',
        price: 0,
        original_price: 0,
        stock: 0,
        status: 'active',
        is_activity: false,
        activity_id: null,
        spec_combinations: {}
      },
      skuRules: {
        sku_code: [{ required: true, message: '请输入SKU编码', trigger: 'blur' }],
        price: [{ required: true, message: '请输入售价', trigger: 'blur' }],
        stock: [{ required: true, message: '请输入库存', trigger: 'blur' }]
      }
    }
  },
  computed: {
    displayProductName() {
      return this.productName || this.localProductName
    },
    effectiveProductId() {
      return this.productId || this.localProductId
    }
  },
  mounted() {
    this.initProduct()
  },
  watch: {
    productId: {
      handler() {
        this.initProduct()
      },
      immediate: true
    }
  },
  methods: {
    // 初始化产品信息
    initProduct() {
      if (this.productId) {
        // 通过 props 传入
        this.localProductId = this.productId
        this.localProductName = this.productName
        this.getSpecifications()
        this.getSKUs()
        this.getActivities()
      } else if (this.$route.params.id) {
        // 通过路由参数传入（兼容旧方式）
        this.localProductId = parseInt(this.$route.params.id)
        this.getProductInfo()
        this.getSpecifications()
        this.getSKUs()
        this.getActivities()
      }
    },
    // 获取活动列表
    getActivities() {
      activityApi.getActivities().then(data => {
        this.activities = data.list || []
      }).catch(() => {})
    },
    // 处理活动专用SKU变化
    handleActivityChange() {
      if (!this.skuForm.is_activity) {
        this.skuForm.activity_id = null
      }
    },
    // 返回
    handleBack() {
      if (this.productId) {
        // 通过 props 传入，使用事件返回
        this.$emit('back')
      } else {
        // 通过路由跳转，使用路由返回
        this.$router.back()
      }
    },
    // 获取商品信息
    getProductInfo() {
      productApi.getProduct(this.effectiveProductId).then(data => {
        this.localProductName = data.name
      }).catch(() => {})
    },
    // 获取规格列表
    getSpecifications() {
      productApi.getProductSpecifications(this.effectiveProductId).then(data => {
        this.specifications = data || []
      }).catch(() => {})
    },
    // 获取SKU列表
    getSKUs() {
      this.loading = true
      productApi.getProductSKUs(this.effectiveProductId).then(data => {
        this.skus = data || []
        this.loading = false
      }).catch(() => {
        this.loading = false
      })
    },
    // 获取规格组合显示
    getSpecCombination(sku) {
      // 从spec_combinations字段解析规格组合
      if (!sku.spec_combinations) return {}
      
      const result = {}
      for (const [specId, valueId] of Object.entries(sku.spec_combinations)) {
        const spec = this.specifications.find(s => s.id.toString() === specId)
        if (spec) {
          const value = spec.values.find(v => v.id === valueId)
          if (value) {
            result[spec.name] = value.value
          }
        }
      }
      return result
    },
    // 打开自动生成对话框
    handleGenerateSKUs() {
      if (this.specifications.length === 0) {
        this.$message.warning('请先配置商品规格')
        return
      }
      
      // 检查是否所有规格都有规格值
      for (const spec of this.specifications) {
        if (!spec.values || spec.values.length === 0) {
          this.$message.warning(`规格 "${spec.name}" 没有配置规格值`)
          return
        }
      }
      
      this.generateForm.base_price = 0
      this.generatedSKUs = []
      this.generateDialogVisible = true
    },
    // 预览生成SKU
    handlePreviewGenerate() {
      productApi.generateSKUs(this.effectiveProductId, this.generateForm).then(data => {
        this.generatedSKUs = data || []
        if (this.generatedSKUs.length === 0) {
          this.$message.warning('无法生成SKU组合，请检查规格配置')
        }
      }).catch(() => {})
    },
    // 确认生成SKU
    handleConfirmGenerate() {
      if (this.generatedSKUs.length === 0) {
        this.$message.warning('没有可生成的SKU')
        return
      }

      productApi.batchCreateSKUs(this.effectiveProductId, { skus: this.generatedSKUs }).then(() => {
        this.$message.success('SKU生成成功')
        this.generateDialogVisible = false
        this.getSKUs()
      }).catch(() => {})
    },
    // 添加SKU
    handleAddSKU() {
      this.skuForm = {
        id: null,
        sku_code: '',
        price: 0,
        original_price: 0,
        stock: 0,
        status: 'active',
        is_activity: false,
        activity_id: null,
        spec_combinations: {}
      }
      this.skuDialogVisible = true
    },
    // 编辑SKU
    handleEditSKU(sku) {
      this.skuForm = {
        id: sku.id,
        sku_code: sku.sku_code,
        price: sku.price,
        original_price: sku.original_price || 0,
        stock: sku.stock,
        status: sku.status,
        is_activity: sku.is_activity || false,
        activity_id: sku.activity_id || null
      }
      this.skuDialogVisible = true
    },
    // 保存SKU
    handleSaveSKU() {
      this.$refs.skuFormRef.validate(valid => {
        if (valid) {
          if (this.skuForm.id) {
            // 更新
            productApi.updateSKU(this.skuForm.id, this.skuForm).then(() => {
              this.$message.success('更新SKU成功')
              this.skuDialogVisible = false
              this.getSKUs()
            }).catch(() => {})
          } else {
            // 创建
            // 构建规格组合
            const specCombinations = []
            for (const [specId, valueId] of Object.entries(this.skuForm.spec_combinations)) {
              if (valueId) {
                specCombinations.push({
                  spec_id: parseInt(specId),
                  spec_value_id: valueId
                })
              }
            }
            
            const data = {
              ...this.skuForm,
              spec_combinations: specCombinations
            }
            
            productApi.createProductSKU(this.effectiveProductId, data).then(() => {
              this.$message.success('添加SKU成功')
              this.skuDialogVisible = false
              this.getSKUs()
            }).catch(() => {})
          }
        }
      })
    },
    // 删除SKU
    handleDeleteSKU(sku) {
      this.$confirm(`确定要禁用SKU "${sku.sku_code}" 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        productApi.deleteSKU(sku.id).then(() => {
          this.$message.success('禁用SKU成功')
          this.getSKUs()
        }).catch(() => {})
      }).catch(() => {})
    }
  }
}
</script>

<style scoped>
.skus-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 15px;
}

.header-right {
  display: flex;
  gap: 10px;
}

.title {
  font-size: 16px;
  font-weight: bold;
}

.specs-preview {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
}

.specs-preview h4 {
  margin: 0 0 10px 0;
}

.spec-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.spec-label {
  font-weight: bold;
  min-width: 60px;
}

.spec-value-tag {
  margin-right: 5px;
}

.spec-combo-tag {
  margin-right: 5px;
  margin-bottom: 5px;
}

.price {
  color: #f56c6c;
  font-weight: bold;
}

.original-price {
  color: #909399;
  text-decoration: line-through;
}

.form-tip {
  margin-left: 10px;
  color: #909399;
  font-size: 12px;
}

.preview-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e4e7ed;
}

.preview-section h4 {
  margin: 0 0 10px 0;
}

.spec-select-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.spec-name {
  min-width: 80px;
  font-weight: bold;
}

.dialog-footer {
  text-align: right;
}
</style>
