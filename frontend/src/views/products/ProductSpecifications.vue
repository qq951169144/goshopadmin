<template>
  <div class="specifications-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-button @click="handleBack" size="small">返回</el-button>
            <span class="title">商品规格管理 - {{ displayProductName }}</span>
          </div>
          <el-button type="primary" @click="handleAddSpec">添加规格维度</el-button>
        </div>
      </template>

      <div class="specs-list">
        <el-empty v-if="specifications.length === 0" description="暂无规格维度" />
        
        <div v-for="spec in specifications" :key="spec.id" class="spec-item">
          <div class="spec-header">
            <div class="spec-info">
              <span class="spec-name">{{ spec.name }}</span>
              <span class="spec-sort">排序: {{ spec.sort }}</span>
            </div>
            <div class="spec-actions">
              <el-button size="small" @click="handleEditSpec(spec)">编辑</el-button>
              <el-button size="small" type="primary" @click="handleAddValue(spec)">添加规格值</el-button>
              <el-button size="small" type="danger" @click="handleDeleteSpec(spec)">删除</el-button>
            </div>
          </div>
          
          <div class="spec-values">
            <el-tag
              v-for="value in spec.values"
              :key="value.id"
              :type="value.status === 'active' ? 'primary' : 'info'"
              class="value-tag"
              closable
              @close="handleDeleteValue(value)"
              @click="handleEditValue(value, spec)"
            >
              {{ value.value }}
            </el-tag>
            <el-tag v-if="!spec.values || spec.values.length === 0" type="info">暂无规格值</el-tag>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 规格维度对话框 -->
    <el-dialog
      v-model="specDialogVisible"
      :title="specForm.id ? '编辑规格维度' : '添加规格维度'"
      width="500px"
    >
      <el-form :model="specForm" :rules="specRules" ref="specFormRef">
        <el-form-item label="规格名称" prop="name">
          <el-input v-model="specForm.name" placeholder="如：颜色、尺寸、版本" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="specForm.sort" :min="0" :max="999" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="specDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveSpec">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 规格值对话框 -->
    <el-dialog
      v-model="valueDialogVisible"
      :title="valueForm.id ? '编辑规格值' : '添加规格值'"
      width="500px"
    >
      <el-form :model="valueForm" :rules="valueRules" ref="valueFormRef">
        <el-form-item label="规格值" prop="value">
          <el-input v-model="valueForm.value" placeholder="如：红色、M码" />
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number v-model="valueForm.sort" :min="0" :max="999" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="valueForm.status">
            <el-radio label="active">启用</el-radio>
            <el-radio label="inactive">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="valueDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveValue">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>

import { productApi } from '@/api/auth'

export default {
  name: 'ProductSpecifications',

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
      specDialogVisible: false,
      valueDialogVisible: false,
      specForm: {
        id: null,
        name: '',
        sort: 0
      },
      valueForm: {
        id: null,
        spec_id: null,
        value: '',
        sort: 0,
        status: 'active'
      },
      currentSpec: null,
      specRules: {
        name: [{ required: true, message: '请输入规格名称', trigger: 'blur' }]
      },
      valueRules: {
        value: [{ required: true, message: '请输入规格值', trigger: 'blur' }]
      },

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
      } else if (this.$route.params.id) {
        // 通过路由参数传入（兼容旧方式）
        this.localProductId = parseInt(this.$route.params.id)
        this.getProductInfo()
        this.getSpecifications()
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
    // 获取完整的图片URL
    getImageUrl(imageUrl) {
      if (!imageUrl) return ''
      if (imageUrl.startsWith('http://') || imageUrl.startsWith('https://')) {
        return imageUrl
      }
      return `/api${imageUrl}`
    },
    // 添加规格维度
    handleAddSpec() {
      this.specForm = {
        id: null,
        name: '',
        sort: 0
      }
      this.specDialogVisible = true
    },
    // 编辑规格维度
    handleEditSpec(spec) {
      this.specForm = {
        id: spec.id,
        name: spec.name,
        sort: spec.sort
      }
      this.specDialogVisible = true
    },
    // 保存规格维度
    handleSaveSpec() {
      this.$refs.specFormRef.validate(valid => {
        if (valid) {
          if (this.specForm.id) {
            // 更新
            productApi.updateSpecification(this.specForm.id, this.specForm).then(() => {
              this.$message.success('更新规格成功')
              this.specDialogVisible = false
              this.getSpecifications()
            }).catch(() => {})
          } else {
            // 创建
            productApi.createProductSpecification(this.effectiveProductId, this.specForm).then(() => {
              this.$message.success('添加规格成功')
              this.specDialogVisible = false
              this.getSpecifications()
            }).catch(() => {})
          }
        }
      })
    },
    // 删除规格维度
    handleDeleteSpec(spec) {
      this.$confirm(`确定要删除规格 "${spec.name}" 吗？删除后将同时删除该规格下的所有规格值。`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        productApi.deleteSpecification(spec.id).then(() => {
          this.$message.success('删除规格成功')
          this.getSpecifications()
        }).catch(() => {})
      }).catch(() => {})
    },
    // 添加规格值
    handleAddValue(spec) {
      this.currentSpec = spec
      this.valueForm = {
        id: null,
        spec_id: spec.id,
        value: '',
        sort: 0,
        status: 'active'
      }
      this.valueDialogVisible = true
    },
    // 编辑规格值
    handleEditValue(value, spec) {
      this.currentSpec = spec
      this.valueForm = {
        id: value.id,
        spec_id: spec.id,
        value: value.value,
        sort: value.sort,
        status: value.status
      }
      this.valueDialogVisible = true
    },
    // 保存规格值
    handleSaveValue() {
      this.$refs.valueFormRef.validate(valid => {
        if (valid) {
          if (this.valueForm.id) {
            // 更新
            productApi.updateSpecificationValue(this.valueForm.id, this.valueForm).then(() => {
              this.$message.success('更新规格值成功')
              this.valueDialogVisible = false
              this.getSpecifications()
            }).catch(() => {})
          } else {
            // 创建
            productApi.createSpecificationValue(this.currentSpec.id, this.valueForm).then(() => {
              this.$message.success('添加规格值成功')
              this.valueDialogVisible = false
              this.getSpecifications()
            }).catch(() => {})
          }
        }
      })
    },
    // 删除规格值
    handleDeleteValue(value) {
      this.$confirm(`确定要删除规格值 "${value.value}" 吗？`, '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        productApi.deleteSpecificationValue(value.id).then(() => {
          this.$message.success('删除规格值成功')
          this.getSpecifications()
        }).catch(() => {})
      }).catch(() => {})
    },

  }
}
</script>

<style scoped>
.specifications-container {
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

.title {
  font-size: 16px;
  font-weight: bold;
}

.specs-list {
  margin-top: 20px;
}

.spec-item {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 15px;
  margin-bottom: 15px;
}

.spec-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
}

.spec-info {
  display: flex;
  align-items: center;
  gap: 15px;
}

.spec-name {
  font-size: 16px;
  font-weight: bold;
}

.spec-sort {
  font-size: 12px;
  color: #909399;
}

.spec-actions {
  display: flex;
  gap: 10px;
}

.spec-values {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.value-tag {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
}

.dialog-footer {
  text-align: right;
}
</style>
