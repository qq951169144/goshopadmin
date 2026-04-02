<template>
  <div class="activity-form-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑活动' : '创建活动' }}</span>
          <el-button @click="handleCancel">取消</el-button>
        </div>
      </template>
      
      <el-form :model="activityForm" :rules="rules" ref="formRef" label-width="120px">
        <!-- 基本信息 -->
        <el-form-item label="活动名称" prop="name">
          <el-input v-model="activityForm.name" placeholder="请输入活动名称" />
        </el-form-item>
        
        <el-form-item label="活动类型" prop="type">
          <el-select v-model="activityForm.type" placeholder="选择活动类型" @change="handleTypeChange">
            <el-option label="秒杀活动" value="seckill"></el-option>
            <el-option label="兑换码活动" value="redeem_code"></el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="活动时间" required>
          <el-date-picker
            v-model="activityForm.timeRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD HH:mm:ss"
            value-format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </el-form-item>
        
        <el-form-item label="活动状态" prop="status">
          <el-select v-model="activityForm.status" placeholder="选择活动状态">
            <el-option label="激活" value="active"></el-option>
            <el-option label="禁用" value="inactive"></el-option>
          </el-select>
        </el-form-item>
        

        
        <!-- 秒杀活动配置 -->
        <template v-if="activityForm.type === 'seckill'">
          <el-form-item label="关联商品" required>
            <el-button type="primary" @click="showProductSelector = true">选择商品</el-button>
            <el-table :data="selectedProducts" style="width: 100%; margin-top: 10px">
              <el-table-column prop="product_name" label="商品名称"></el-table-column>
              <el-table-column prop="sku_name" label="SKU名称"></el-table-column>
              <el-table-column prop="activity_price" label="活动价格">
                <template #default="scope">
                  <el-input-number 
                    v-model="scope.row.activity_price" 
                    :min="0" 
                    :step="0.01" 
                    :precision="2"
                  />
                </template>
              </el-table-column>
              <el-table-column prop="activity_stock" label="活动库存">
                <template #default="scope">
                  <el-input-number 
                    v-model="scope.row.activity_stock" 
                    :min="0" 
                    :step="1"
                  />
                </template>
              </el-table-column>
              <el-table-column label="操作" width="100">
                <template #default="scope">
                  <el-button size="small" type="danger" @click="removeProduct(scope.$index)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-form-item>
        </template>
        
        <!-- 兑换码活动配置 -->
        <template v-if="activityForm.type === 'redeem_code'">
          <el-form-item label="关联商品" required>
            <el-button type="primary" @click="showProductSelector = true">选择商品</el-button>
            <el-table :data="selectedProducts" style="width: 100%; margin-top: 10px">
              <el-table-column prop="product_name" label="商品名称"></el-table-column>
              <el-table-column prop="sku_name" label="SKU名称"></el-table-column>
              <el-table-column label="操作" width="100">
                <template #default="scope">
                  <el-button size="small" type="danger" @click="removeProduct(scope.$index)">删除</el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-form-item>
          
          <el-form-item label="兑换码规则" required>
            <el-form :inline="true" :model="redeemCodeRules" class="mb-2">
              <el-form-item label="兑换码类型">
                <el-select v-model="redeemCodeRules.type" placeholder="选择类型" style="width: 120px">
                  <el-option label="数字" value="number"></el-option>
                  <el-option label="字母" value="letter"></el-option>
                  <el-option label="混合" value="mixed"></el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="兑换码长度">
                <el-input-number v-model="redeemCodeRules.length" :min="6" :max="20" :step="1" />
              </el-form-item>
              <el-form-item label="排除字符">
                <el-input v-model="redeemCodeRules.exclude_chars" placeholder="例如：01IOl" />
              </el-form-item>
              <el-form-item label="总数量">
                <el-input-number v-model="redeemCodeRules.total_quantity" :min="1" :max="10000" :step="1" />
              </el-form-item>
              <el-form-item label="每用户限制">
                <el-input-number v-model="redeemCodeRules.limit_per_user" :min="1" :max="10" :step="1" />
              </el-form-item>
              <el-form-item label="需要验证">
                <el-switch v-model="redeemCodeRules.need_verify" />
              </el-form-item>
            </el-form>
          </el-form-item>
        </template>
        
        <el-form-item>
          <el-button type="primary" @click="handleSubmit">提交</el-button>
          <el-button @click="handleCancel">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
    
    <!-- 商品选择器 -->
    <el-dialog
      v-model="showProductSelector"
      title="选择商品"
      width="60%"
    >
      <el-form :inline="true" :model="productSearchForm" class="mb-4">
        <el-form-item label="商品名称">
          <el-input v-model="productSearchForm.name" placeholder="请输入商品名称" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="searchProducts">查询</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="productsList" style="width: 100%">
        <el-table-column prop="id" label="商品ID" width="100"></el-table-column>
        <el-table-column prop="name" label="商品名称"></el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button size="small" @click="showSkuSelector(scope.row)">选择SKU</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="dialog-footer">
        <el-button @click="showProductSelector = false">取消</el-button>
      </div>
    </el-dialog>
    
    <!-- SKU选择器 -->
    <el-dialog
      v-model="showSkuSelectorDialog"
      :title="`选择${currentProduct?.name}的SKU`"
      width="600px"
    >
      <el-table :data="skuList" style="width: 100%" @selection-change="handleSkuSelectChange">
        <el-table-column type="selection" width="55"></el-table-column>
        <el-table-column prop="id" label="SKU ID" width="100"></el-table-column>
        <el-table-column prop="sku_name" label="SKU名称">
          <template #default="scope">
            {{ scope.row.sku_name || scope.row.sku_code }}
          </template>
        </el-table-column>
        <el-table-column prop="sku_code" label="SKU编码"></el-table-column>
        <el-table-column prop="price" label="价格" width="100"></el-table-column>
        <el-table-column prop="stock" label="库存" width="100"></el-table-column>
      </el-table>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showSkuSelectorDialog = false">取消</el-button>
          <el-button type="primary" @click="confirmSkuSelect">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { activityApi, productApi } from '../../api/auth'

export default {
  name: 'ActivityForm',
  props: {
    activity: {
      type: Object,
      default: null
    }
  },
  computed: {
    activityId() {
      if (this.activity && this.activity.id) {
        return this.activity.id;
      }
      const id = this.$route.params.id;
      const parsedId = parseInt(id);
      return isNaN(parsedId) ? null : parsedId;
    },
    isEdit() {
      return this.activityId !== null;
    }
  },
  data() {
    return {
      activityForm: {
        name: '',
        type: '',
        timeRange: [],
        status: 'active'
      },
      rules: {
        name: [
          { required: true, message: '请输入活动名称', trigger: 'blur' }
        ],
        type: [
          { required: true, message: '请选择活动类型', trigger: 'change' }
        ],
        status: [
          { required: true, message: '请选择活动状态', trigger: 'change' }
        ]
      },
      selectedProducts: [],
      redeemCodeRules: {
        type: 'mixed',
        length: 12,
        exclude_chars: '',
        total_quantity: 100,
        limit_per_user: 1,
        need_verify: false
      },
      showProductSelector: false,
      productSearchForm: {
        name: ''
      },
      productsList: [],
      showSkuSelectorDialog: false,
      currentProduct: null,
      skuList: [],
      selectedSkuIds: []
    }
  },
  mounted() {
    if (this.isEdit) {
      this.getActivityDetail();
    }
  },
  watch: {
    showProductSelector(newVal) {
      if (newVal) {
        this.productSearchForm.name = ''
        this.searchProducts()
      }
    }
  },
  methods: {
    // 设置表单数据
    setFormData(activity) {
      this.activityForm.name = activity.name;
      this.activityForm.type = activity.type;
      this.activityForm.timeRange = [activity.start_time, activity.end_time];
      this.activityForm.status = activity.status;
      
      // 处理关联商品
      if (activity.products) {
        this.selectedProducts = activity.products;
      }
      
      // 处理兑换码规则
      if (activity.type === 'redeem_code' && activity.redeem_code_rules) {
        this.redeemCodeRules.type = activity.redeem_code_rules.type;
        this.redeemCodeRules.length = activity.redeem_code_rules.length;
        this.redeemCodeRules.exclude_chars = activity.redeem_code_rules.exclude_chars;
        this.redeemCodeRules.total_quantity = activity.redeem_code_rules.total_quantity || 100;
        this.redeemCodeRules.limit_per_user = activity.redeem_code_rules.limit_per_user || 1;
        this.redeemCodeRules.need_verify = activity.redeem_code_rules.need_verify === 1;
      }
    },
    
    // 获取活动详情（编辑模式）
    getActivityDetail() {
      if (!this.isEdit) return;
      
      try {
        activityApi.getActivity(this.activityId).then(activity => {
          this.setFormData(activity);
        }).catch(error => {
          console.error('获取活动详情失败:', error);
          this.$message.error('获取活动详情失败');
        });
      } catch (error) {
        console.error('获取活动详情失败:', error);
        this.$message.error('获取活动详情失败');
      }
    },
    
    // 搜索商品
    searchProducts() {
      try {
        const params = {}
        if (this.productSearchForm.name) {
          params.name = this.productSearchForm.name
        }
        productApi.getProducts(params).then(response => {
          this.productsList = response || [];
        }).catch(error => {
          console.error('搜索商品失败:', error);
          this.$message.error('搜索商品失败');
        });
      } catch (error) {
        console.error('搜索商品失败:', error);
        this.$message.error('搜索商品失败');
      }
    },
    
    // 显示SKU选择器
    showSkuSelector(product) {
      this.currentProduct = product;
      try {
        productApi.getProductSkus(product.id).then(response => {
          this.skuList = response || [];
          this.showSkuSelectorDialog = true;
        }).catch(error => {
          console.error('获取SKU列表失败:', error);
          this.$message.error('获取SKU列表失败');
        });
      } catch (error) {
        console.error('获取SKU列表失败:', error);
        this.$message.error('获取SKU列表失败');
      }
    },
    
    // 处理SKU选择变化
    handleSkuSelectChange(selection) {
      this.selectedSkuIds = selection.map(sku => sku.id);
    },
    
    // 确认选择SKU
    confirmSkuSelect() {
      if (this.selectedSkuIds.length === 0) {
        this.$message.warning('请选择至少一个SKU');
        return;
      }
      
      const selectedRows = this.skuList.filter(sku => this.selectedSkuIds.includes(sku.id));
      selectedRows.forEach(sku => {
        this.selectedProducts.push({
          product_id: this.currentProduct.id,
          product_name: this.currentProduct.name,
          sku_id: sku.id,
          sku_name: sku.sku_name || sku.sku_code,
          activity_price: sku.price,
          activity_stock: sku.stock
        });
      });
      
      this.showSkuSelectorDialog = false;
      this.currentProduct = null;
      this.skuList = [];
      this.selectedSkuIds = [];
    },
    
    // 移除商品
    removeProduct(index) {
      this.selectedProducts.splice(index, 1);
    },
    
    // 活动类型变化
    handleTypeChange() {
      this.selectedProducts = [];
    },
    
    // 提交表单
    handleSubmit() {
      this.$refs.formRef.validate((valid) => {
        if (valid) {
          try {
            // 构建请求数据
            const data = {
              name: this.activityForm.name,
              type: this.activityForm.type,
              start_time: this.activityForm.timeRange[0],
              end_time: this.activityForm.timeRange[1],
              status: this.activityForm.status
            };
            
            // 处理关联商品
            if (this.selectedProducts.length > 0) {
              data.products = this.selectedProducts.map(product => ({
                product_id: product.product_id,
                sku_id: product.sku_id,
                original_price: product.activity_price,
                activity_price: product.activity_price,
                stock: product.activity_stock,
                product_type: this.activityForm.type
              }));
            }
            
            // 如果是兑换码活动，添加兑换码规则
            if (this.activityForm.type === 'redeem_code') {
              data.redeem_setting = {
                code_type: this.redeemCodeRules.type,
                code_length: this.redeemCodeRules.length,
                exclude_chars: this.redeemCodeRules.exclude_chars,
                total_quantity: this.redeemCodeRules.total_quantity,
                limit_per_user: this.redeemCodeRules.limit_per_user,
                need_verify: this.redeemCodeRules.need_verify ? 1 : 0
              };
            }
            
            if (this.isEdit) {
              activityApi.updateActivity(this.activityId, data).then(() => {
                this.$message.success('编辑成功');
                this.$emit('back');
              }).catch(error => {
                console.error('编辑活动失败:', error);
                this.$message.error('编辑活动失败');
              });
            } else {
              activityApi.createActivity(data).then(() => {
                this.$message.success('创建成功');
                this.$emit('back');
              }).catch(error => {
                console.error('创建活动失败:', error);
                this.$message.error('创建活动失败');
              });
            }
          } catch (error) {
            console.error('提交表单失败:', error);
            this.$message.error('提交失败');
          }
        }
      });
    },
    // 取消
    handleCancel() {
      this.$emit('back');
    }
  }
}
</script>

<style scoped>
.activity-form-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}
</style>