<template>
  <div class="categories-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>商品分类管理</span>
          <el-button type="primary" @click="handleCreateCategory">添加分类</el-button>
        </div>
      </template>
      
      <el-table :data="categories" style="width: 100%">
        <el-table-column prop="id" label="分类ID" width="80" />
        <el-table-column prop="name" label="分类名称" />
        <el-table-column prop="parent_id" label="父分类ID" width="100" />
        <el-table-column prop="level" label="层级" width="80" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
              {{ scope.row.status === 'active' ? '激活' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <el-button size="small" @click="handleEditCategory(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDeleteCategory(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑分类对话框 -->
    <el-dialog
      v-model="categoryDialogVisible"
      :title="categoryDialogTitle"
      width="500px"
    >
      <el-form :model="categoryForm" :rules="categoryRules" ref="categoryFormRef">
        <el-form-item label="分类名称" prop="name">
          <el-input v-model="categoryForm.name" placeholder="请输入分类名称" />
        </el-form-item>
        <el-form-item label="父分类" prop="parent_id">
          <el-select v-model="categoryForm.parent_id" placeholder="请选择父分类">
            <el-option label="顶级分类" value="0" />
            <el-option
              v-for="category in categories"
              :key="category.id"
              :label="category.name"
              :value="category.id"
              :disabled="category.id === categoryForm.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="排序" prop="sort">
          <el-input-number
            v-model="categoryForm.sort"
            :min="0"
            placeholder="请输入排序"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="categoryForm.status" placeholder="请选择状态">
            <el-option label="激活" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="categoryDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveCategory">保存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { productApi } from '@/api/auth'

export default {
  name: 'ProductCategories',
  data() {
    return {
      categories: [],
      categoryDialogVisible: false,
      categoryDialogTitle: '添加分类',
      categoryForm: {
        id: '',
        name: '',
        parent_id: 0,
        level: 1,
        sort: 0,
        status: 'active'
      },
      categoryRules: {
        name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }]
      }
    }
  },
  mounted() {
    this.getCategories()
  },
  methods: {
    // 获取分类列表
    getCategories() {
      productApi.getCategories().then(res => {
        if (res.code === 200) {
          this.categories = res.data
        }
      })
    },
    // 处理添加分类
    handleCreateCategory() {
      this.categoryDialogTitle = '添加分类'
      this.categoryForm = {
        id: '',
        name: '',
        parent_id: 0,
        level: 1,
        sort: 0,
        status: 'active'
      }
      this.categoryDialogVisible = true
    },
    // 处理编辑分类
    handleEditCategory(category) {
      this.categoryDialogTitle = '编辑分类'
      this.categoryForm = {
        id: category.id,
        name: category.name,
        parent_id: category.parent_id,
        level: category.level,
        sort: category.sort,
        status: category.status
      }
      this.categoryDialogVisible = true
    },
    // 处理保存分类
    handleSaveCategory() {
      this.$refs.categoryFormRef.validate((valid) => {
        if (valid) {
          if (this.categoryForm.id) {
            // 更新分类
            productApi.updateCategory(this.categoryForm.id, this.categoryForm).then(res => {
              if (res.code === 200) {
                this.$message.success('更新分类成功')
                this.categoryDialogVisible = false
                this.getCategories()
              } else {
                this.$message.error(res.message)
              }
            })
          } else {
            // 创建分类
            productApi.createCategory(this.categoryForm).then(res => {
              if (res.code === 200) {
                this.$message.success('创建分类成功')
                this.categoryDialogVisible = false
                this.getCategories()
              } else {
                this.$message.error(res.message)
              }
            })
          }
        }
      })
    },
    // 处理删除分类
    handleDeleteCategory(id) {
      this.$confirm('确定要删除这个分类吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        productApi.deleteCategory(id).then(res => {
          if (res.code === 200) {
            this.$message.success('删除分类成功')
            this.getCategories()
          } else {
            this.$message.error(res.message)
          }
        })
      }).catch(() => {
        // 取消删除
      })
    }
  }
}
</script>

<style scoped>
.categories-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-footer {
  text-align: right;
}
</style>
