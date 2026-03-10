<template>
  <div class="products-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>商品管理</span>
          <el-button type="primary" @click="handleCreateProduct">添加商品</el-button>
        </div>
      </template>
      
      <el-table :data="products" style="width: 100%">
        <el-table-column prop="id" label="商品ID" width="80" />
        <el-table-column prop="name" label="商品名称" />
        <el-table-column prop="price" label="价格" width="100" />
        <el-table-column prop="stock" label="库存" width="80" />
        <el-table-column prop="category.name" label="分类" width="120" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
              {{ scope.row.status === 'active' ? '激活' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="300">
          <template #default="scope">
            <el-button size="small" @click="handleEditProduct(scope.row)">编辑</el-button>
            <el-button size="small" type="danger" @click="handleDeleteProduct(scope.row.id)">删除</el-button>
            <el-button size="small" @click="handleViewProduct(scope.row)">预览</el-button>
            <el-button size="small" @click="handleManageImages(scope.row)">图片管理</el-button>
            <el-button size="small" @click="handleManageSKUs(scope.row)">SKU管理</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建/编辑商品对话框 -->
    <el-dialog
      v-model="productDialogVisible"
      :title="productDialogTitle"
      width="800px"
    >
      <el-form :model="productForm" :rules="productRules" ref="productFormRef">
        <el-form-item label="商品名称" prop="name">
          <el-input v-model="productForm.name" placeholder="请输入商品名称" />
        </el-form-item>
        <el-form-item label="商品描述" prop="description">
          <el-input
            v-model="productForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入商品描述"
          />
        </el-form-item>
        <el-form-item label="商品价格" prop="price">
          <el-input-number
            v-model="productForm.price"
            :min="0"
            :step="0.01"
            :precision="2"
            placeholder="请输入商品价格"
          />
        </el-form-item>
        <el-form-item label="商品库存" prop="stock">
          <el-input-number
            v-model="productForm.stock"
            :min="0"
            placeholder="请输入商品库存"
          />
        </el-form-item>
        <el-form-item label="商品分类" prop="category_id">
          <el-select v-model="productForm.category_id" placeholder="请选择商品分类">
            <el-option
              v-for="category in categories"
              :key="category.id"
              :label="category.name"
              :value="category.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="商品状态" prop="status">
          <el-select v-model="productForm.status" placeholder="请选择商品状态">
            <el-option label="激活" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="商品详情" prop="detail">
          <div class="quill-editor">
            <div ref="quillEditor" style="height: 300px;"></div>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="productDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveProduct">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 商品图片管理对话框 -->
    <el-dialog
      v-model="imageDialogVisible"
      title="商品图片管理"
      width="600px"
    >
      <div class="image-upload-section">
        <el-upload
          class="image-uploader"
          action="#"
          :auto-upload="false"
          :on-change="handleImageUpload"
          :show-file-list="false"
        >
          <el-button type="primary">上传图片</el-button>
        </el-upload>
      </div>
      <div class="image-list">
        <div
          v-for="image in productImages"
          :key="image.id"
          class="image-item"
        >
          <img :src="image.image_url" alt="商品图片" class="image-preview" />
          <div class="image-actions">
            <el-checkbox v-model="image.is_main" @change="handleSetMainImage(image)">
              主图
            </el-checkbox>
            <el-input-number
              v-model="image.sort"
              :min="0"
              @change="handleUpdateImageSort(image)"
              style="width: 80px; margin: 5px 0"
            />
            <el-button size="small" type="danger" @click="handleDeleteImage(image.id)">
              删除
            </el-button>
          </div>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="imageDialogVisible = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 商品SKU管理对话框 -->
    <el-dialog
      v-model="skuDialogVisible"
      title="商品SKU管理"
      width="600px"
    >
      <div class="sku-section">
        <el-button type="primary" @click="handleAddSKU">添加SKU</el-button>
        <el-table :data="productSKUs" style="margin-top: 20px">
          <el-table-column prop="sku_code" label="SKU编码" />
          <el-table-column prop="attributes" label="属性" />
          <el-table-column prop="price" label="价格" width="100" />
          <el-table-column prop="stock" label="库存" width="80" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '激活' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="scope">
              <el-button size="small" @click="handleEditSKU(scope.row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDeleteSKU(scope.row.id)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="skuDialogVisible = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- SKU编辑对话框 -->
    <el-dialog
      v-model="skuEditDialogVisible"
      :title="skuForm.id ? '编辑SKU' : '添加SKU'"
      width="500px"
    >
      <el-form :model="skuForm" ref="skuFormRef">
        <el-form-item label="SKU编码" prop="sku_code">
          <el-input v-model="skuForm.sku_code" placeholder="请输入SKU编码" />
        </el-form-item>
        <el-form-item label="属性" prop="attributes">
          <el-input v-model="skuForm.attributes" placeholder='请输入SKU属性，如{"color": "red", "size": "M"}' />
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number
            v-model="skuForm.price"
            :min="0"
            :step="0.01"
            :precision="2"
            placeholder="请输入价格"
          />
        </el-form-item>
        <el-form-item label="库存" prop="stock">
          <el-input-number
            v-model="skuForm.stock"
            :min="0"
            placeholder="请输入库存"
          />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="skuForm.status" placeholder="请选择状态">
            <el-option label="激活" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="skuEditDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSaveSKU">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 商品预览对话框 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="商品预览"
      width="800px"
    >
      <div class="product-preview">
        <div class="preview-images">
          <img
            v-for="image in previewProduct.images"
            :key="image.id"
            :src="image.image_url"
            alt="商品图片"
            class="preview-image"
          />
        </div>
        <div class="preview-info">
          <h2>{{ previewProduct.name }}</h2>
          <div class="preview-price">¥{{ previewProduct.price }}</div>
          <div class="preview-stock">库存：{{ previewProduct.stock }}</div>
          <div class="preview-category">分类：{{ previewProduct.category?.name }}</div>
          <div class="preview-description" v-html="previewProduct.description"></div>
          <div class="preview-detail" v-if="previewProduct.detail" v-html="previewProduct.detail"></div>
        </div>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="previewDialogVisible = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { productApi } from '@/api/auth'
import Quill from 'quill'
import 'quill/dist/quill.snow.css'

export default {
  name: 'Products',
  data() {
    return {
      products: [],
      categories: [],
      productDialogVisible: false,
      productDialogTitle: '添加商品',
      productForm: {
        name: '',
        description: '',
        detail: '',
        price: 0,
        stock: 0,
        category_id: 0,
        status: 'active'
      },
      quillEditor: null,
      productRules: {
        name: [{ required: true, message: '请输入商品名称', trigger: 'blur' }],
        price: [{ required: true, message: '请输入商品价格', trigger: 'blur' }],
        stock: [{ required: true, message: '请输入商品库存', trigger: 'blur' }],
        category_id: [{ required: true, message: '请选择商品分类', trigger: 'blur' }]
      },
      imageDialogVisible: false,
      productImages: [],
      skuDialogVisible: false,
      productSKUs: [],
      skuForm: {
        id: 0,
        product_id: 0,
        sku_code: '',
        attributes: '',
        price: 0,
        stock: 0,
        status: 'active'
      },
      skuEditDialogVisible: false,
      previewDialogVisible: false,
      previewProduct: {}
    }
  },
  mounted() {
    this.getProducts()
    this.getCategories()
  },
  watch: {
    productDialogVisible(val) {
      if (val) {
        this.$nextTick(() => {
          this.initQuillEditor()
        })
      }
    }
  },
  methods: {
    // 初始化Quill编辑器
    initQuillEditor() {
      if (!this.quillEditor) {
        this.quillEditor = new Quill(this.$refs.quillEditor, {
          theme: 'snow',
          modules: {
            toolbar: [
              ['bold', 'italic', 'underline', 'strike'],
              ['blockquote', 'code-block'],
              [{ 'header': 1 }, { 'header': 2 }],
              [{ 'list': 'ordered' }, { 'list': 'bullet' }],
              [{ 'script': 'sub' }, { 'script': 'super' }],
              [{ 'indent': '-1' }, { 'indent': '+1' }],
              [{ 'direction': 'rtl' }],
              [{ 'size': ['small', false, 'large', 'huge'] }],
              [{ 'header': [1, 2, 3, 4, 5, 6, false] }],
              [{ 'color': [] }, { 'background': [] }],
              [{ 'font': [] }],
              [{ 'align': [] }],
              ['clean']
            ]
          }
        })
        
        // 设置编辑器内容
        if (this.productForm.detail) {
          this.quillEditor.root.innerHTML = this.productForm.detail
        }
        
        // 监听内容变化
        this.quillEditor.on('text-change', () => {
          this.productForm.detail = this.quillEditor.root.innerHTML
        })
      } else {
        // 更新编辑器内容
        if (this.productForm.detail) {
          this.quillEditor.root.innerHTML = this.productForm.detail
        } else {
          this.quillEditor.root.innerHTML = ''
        }
      }
    },
    // 获取商品列表
    getProducts() {
      productApi.getProducts().then(res => {
        if (res.code === 200) {
          this.products = res.data
        }
      })
    },
    // 获取商品分类列表
    getCategories() {
      productApi.getCategories().then(res => {
        if (res.code === 200) {
          this.categories = res.data
        }
      })
    },
    // 处理添加商品
    handleCreateProduct() {
      this.productDialogTitle = '添加商品'
      this.productForm = {
        name: '',
        description: '',
        detail: '',
        price: 0,
        stock: 0,
        category_id: 0,
        status: 'active'
      }
      this.productDialogVisible = true
    },
    // 处理编辑商品
    handleEditProduct(product) {
      this.productDialogTitle = '编辑商品'
      this.productForm = {
        id: product.id,
        name: product.name,
        description: product.description,
        detail: product.detail || '',
        price: product.price,
        stock: product.stock,
        category_id: product.category_id,
        status: product.status
      }
      this.productDialogVisible = true
    },
    // 处理保存商品
    handleSaveProduct() {
      this.$refs.productFormRef.validate((valid) => {
        if (valid) {
          if (this.productForm.id) {
            // 更新商品
            productApi.updateProduct(this.productForm.id, this.productForm).then(res => {
              if (res.code === 200) {
                this.$message.success('更新商品成功')
                this.productDialogVisible = false
                this.getProducts()
              } else {
                this.$message.error(res.message)
              }
            })
          } else {
            // 创建商品
            productApi.createProduct(this.productForm).then(res => {
              if (res.code === 200) {
                this.$message.success('创建商品成功')
                this.productDialogVisible = false
                this.getProducts()
              } else {
                this.$message.error(res.message)
              }
            })
          }
        }
      })
    },
    // 处理删除商品
    handleDeleteProduct(id) {
      this.$confirm('确定要删除这个商品吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        productApi.deleteProduct(id).then(res => {
          if (res.code === 200) {
            this.$message.success('删除商品成功')
            this.getProducts()
          } else {
            this.$message.error(res.message)
          }
        })
      }).catch(() => {
        // 取消删除
      })
    },
    // 处理查看商品
    handleViewProduct(product) {
      // 获取商品详情，包括SKU和图片
      productApi.getProduct(product.id).then(res => {
        if (res.code === 200) {
          this.previewProduct = res.data
          this.previewDialogVisible = true
        }
      })
    },
    // 打开图片管理对话框
    handleManageImages(product) {
      this.productForm.id = product.id
      // 获取商品图片列表
      productApi.getProduct(product.id).then(res => {
        if (res.code === 200) {
          this.productImages = res.data.images || []
          this.imageDialogVisible = true
        }
      })
    },
    // 打开SKU管理对话框
    handleManageSKUs(product) {
      this.productForm.id = product.id
      // 获取商品SKU列表
      this.getProductSKUs(product.id)
      this.skuDialogVisible = true
    },
    // 处理图片上传
    handleImageUpload(file) {
      // 这里应该实现图片上传逻辑，返回图片URL
      const imageUrl = URL.createObjectURL(file.raw)
      const newImage = {
        id: Date.now(),
        product_id: this.productForm.id,
        image_url: imageUrl,
        is_main: false,
        sort: this.productImages.length
      }
      this.productImages.push(newImage)
      // 调用API添加图片
      productApi.addProductImage({
        product_id: this.productForm.id,
        image_url: imageUrl,
        is_main: false,
        sort: this.productImages.length
      })
    },
    // 处理删除图片
    handleDeleteImage(id) {
      productApi.deleteProductImage(id).then(res => {
        if (res.code === 200) {
          this.$message.success('删除图片成功')
          this.productImages = this.productImages.filter(image => image.id !== id)
        } else {
          this.$message.error(res.message)
        }
      })
    },
    // 处理设置主图
    handleSetMainImage(image) {
      productApi.updateProductImage(image.id, {
        product_id: image.product_id,
        is_main: image.is_main,
        sort: image.sort
      }).then(res => {
        if (res.code === 200) {
          this.$message.success('设置主图成功')
          // 更新其他图片为非主图
          this.productImages.forEach(img => {
            if (img.id !== image.id) {
              img.is_main = false
            }
          })
        } else {
          this.$message.error(res.message)
          image.is_main = !image.is_main
        }
      })
    },
    // 处理更新图片排序
    handleUpdateImageSort(image) {
      productApi.updateProductImage(image.id, {
        product_id: image.product_id,
        is_main: image.is_main,
        sort: image.sort
      }).then(res => {
        if (res.code !== 200) {
          this.$message.error(res.message)
        }
      })
    },
    // 处理添加SKU
    handleAddSKU() {
      this.skuForm = {
        id: 0,
        product_id: this.productForm.id,
        sku_code: `SKU${Date.now()}`,
        attributes: '{"color": "red", "size": "M"}',
        price: 0,
        stock: 0,
        status: 'active'
      }
      this.skuEditDialogVisible = true
    },
    // 处理编辑SKU
    handleEditSKU(sku) {
      this.skuForm = {
        id: sku.id,
        product_id: sku.product_id,
        sku_code: sku.sku_code,
        attributes: sku.attributes,
        price: sku.price,
        stock: sku.stock,
        status: sku.status
      }
      this.skuEditDialogVisible = true
    },
    // 处理保存SKU
    handleSaveSKU() {
      if (this.skuForm.id) {
        // 更新SKU
        productApi.updateProductSKU(this.skuForm.id, this.skuForm).then(res => {
          if (res.code === 200) {
            this.$message.success('更新SKU成功')
            this.skuEditDialogVisible = false
            this.getProductSKUs(this.skuForm.product_id)
          } else {
            this.$message.error(res.message)
          }
        })
      } else {
        // 创建SKU
        productApi.addProductSKU(this.skuForm).then(res => {
          if (res.code === 200) {
            this.$message.success('添加SKU成功')
            this.skuEditDialogVisible = false
            this.getProductSKUs(this.skuForm.product_id)
          } else {
            this.$message.error(res.message)
          }
        })
      }
    },
    // 处理删除SKU
    handleDeleteSKU(id) {
      this.$confirm('确定要删除这个SKU吗？', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        productApi.deleteProductSKU(id).then(res => {
          if (res.code === 200) {
            this.$message.success('删除SKU成功')
            this.productSKUs = this.productSKUs.filter(sku => sku.id !== id)
          } else {
            this.$message.error(res.message)
          }
        })
      }).catch(() => {
        // 取消删除
      })
    },
    // 获取商品SKU列表
    getProductSKUs(productId) {
      productApi.getProduct(productId).then(res => {
        if (res.code === 200) {
          this.productSKUs = res.data.skus || []
        }
      })
    }
  }
}
</script>

<style scoped>
.products-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.image-upload-section {
  margin-bottom: 20px;
}

.image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.image-item {
  width: 120px;
  border: 1px solid #e4e7ed;
  padding: 10px;
  border-radius: 4px;
}

.image-preview {
  width: 100%;
  height: 100px;
  object-fit: cover;
  margin-bottom: 10px;
}

.image-actions {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.sku-section {
  margin-top: 20px;
}

.product-preview {
  display: flex;
  gap: 20px;
}

.preview-images {
  flex: 1;
  max-width: 300px;
}

.preview-image {
  width: 100%;
  height: 300px;
  object-fit: cover;
  border-radius: 4px;
}

.preview-info {
  flex: 1;
}

.preview-price {
  font-size: 24px;
  font-weight: bold;
  color: #f56c6c;
  margin: 10px 0;
}

.preview-stock,
.preview-category {
  margin: 10px 0;
  color: #606266;
}

.preview-description {
  margin-top: 20px;
  line-height: 1.5;
}

.preview-detail {
  margin-top: 20px;
  line-height: 1.5;
  padding: 10px;
  border-top: 1px solid #e4e7ed;
}

.quill-editor {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 5px;
}

.dialog-footer {
  text-align: right;
}
</style>
