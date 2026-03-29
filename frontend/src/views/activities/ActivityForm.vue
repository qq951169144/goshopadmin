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
            <el-option label="秒杀活动" value="flash_sale"></el-option>
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
            <el-option label="未开始" value="pending"></el-option>
            <el-option label="进行中" value="active"></el-option>
            <el-option label="已结束" value="ended"></el-option>
            <el-option label="已取消" value="cancelled"></el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="活动描述">
          <el-input
            v-model="activityForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入活动描述"
          />
        </el-form-item>
        
        <!-- 秒杀活动配置 -->
        <template v-if="activityForm.type === 'flash_sale'">
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
                <el-select v-model="redeemCodeRules.type" placeholder="选择类型">
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
      width="80%"
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
        <el-table-column type="selection" width="55"></el-table-column>
        <el-table-column prop="id" label="商品ID" width="100"></el-table-column>
        <el-table-column prop="name" label="商品名称"></el-table-column>
        <el-table-column prop="sku_count" label="SKU数量" width="100"></el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-button size="small" @click="showSKUSelector(scope.row)">选择SKU</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="dialog-footer">
        <el-button @click="showProductSelector = false">取消</el-button>
      </div>
    </el-dialog>
    
    <!-- SKU选择器 -->
    <el-dialog
      v-model="showSKUSelectorDialog"
      :title="`选择${currentProduct?.name}的SKU`"
      width="60%"
    >
      <el-table :data="skuList" style="width: 100%" @selection-change="handleSKUSelectChange">
        <el-table-column type="selection" width="55"></el-table-column>
        <el-table-column prop="id" label="SKU ID" width="100"></el-table-column>
        <el-table-column prop="name" label="SKU名称"></el-table-column>
        <el-table-column prop="price" label="原价" width="100"></el-table-column>
        <el-table-column prop="stock" label="库存" width="100"></el-table-column>
      </el-table>
      
      <div class="dialog-footer">
        <el-button @click="showSKUSelectorDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmSKUSelect">确定</el-button>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { activityApi, productApi } from '../../api/auth';

const router = useRouter();
const route = useRoute();
const formRef = ref(null);

// 判断是否为编辑模式
const isEdit = computed(() => route.path.includes('/edit'));
const activityId = computed(() => route.params.id);

// 活动表单
const activityForm = reactive({
  name: '',
  type: '',
  timeRange: [],
  status: 'pending',
  description: ''
});

// 表单规则
const rules = {
  name: [
    { required: true, message: '请输入活动名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择活动类型', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择活动状态', trigger: 'change' }
  ]
};

// 秒杀活动商品
const selectedProducts = ref([]);

// 兑换码规则
const redeemCodeRules = reactive({
  type: 'mixed',
  length: 12,
  exclude_chars: ''
});

// 商品选择器
const showProductSelector = ref(false);
const productSearchForm = reactive({
  name: ''
});
const productsList = ref([]);

// SKU选择器
const showSKUSelectorDialog = ref(false);
const currentProduct = ref(null);
const skuList = ref([]);

// 获取活动详情（编辑模式）
const getActivityDetail = async () => {
  if (!isEdit.value) return;
  
  try {
    const activity = await activityApi.getActivity(activityId.value);
    activityForm.name = activity.name;
    activityForm.type = activity.type;
    activityForm.timeRange = [activity.start_time, activity.end_time];
    activityForm.status = activity.status;
    activityForm.description = activity.description;
    
    // 处理关联商品
    if (activity.products) {
      selectedProducts.value = activity.products;
    }
    
    // 处理兑换码规则
    if (activity.type === 'redeem_code' && activity.redeem_code_rules) {
      redeemCodeRules.type = activity.redeem_code_rules.type;
      redeemCodeRules.length = activity.redeem_code_rules.length;
      redeemCodeRules.exclude_chars = activity.redeem_code_rules.exclude_chars;
    }
  } catch (error) {
    console.error('获取活动详情失败:', error);
    ElMessage.error('获取活动详情失败');
  }
};

// 搜索商品
const searchProducts = async () => {
  try {
    const response = await productApi.getProducts();
    productsList.value = response || [];
  } catch (error) {
    console.error('搜索商品失败:', error);
    ElMessage.error('搜索商品失败');
  }
};

// 显示SKU选择器
const showSKUSelector = async (product) => {
  currentProduct.value = product;
  try {
    const response = await productApi.getProductSKUs(product.id);
    skuList.value = response || [];
    showSKUSelectorDialog.value = true;
  } catch (error) {
    console.error('获取SKU列表失败:', error);
    ElMessage.error('获取SKU列表失败');
  }
};

// 确认选择SKU
const selectedSKUIds = ref([]);

// 处理SKU选择变化
const handleSKUSelectChange = (selection) => {
  selectedSKUIds.value = selection.map(sku => sku.id);
};

// 确认选择SKU
const confirmSKUSelect = () => {
  if (selectedSKUIds.value.length === 0) {
    ElMessage.warning('请选择至少一个SKU');
    return;
  }
  
  const selectedRows = skuList.value.filter(sku => selectedSKUIds.value.includes(sku.id));
  selectedRows.forEach(sku => {
    selectedProducts.value.push({
      product_id: currentProduct.value.id,
      product_name: currentProduct.value.name,
      sku_id: sku.id,
      sku_name: sku.name,
      activity_price: sku.price,
      activity_stock: sku.stock
    });
  });
  
  showSKUSelectorDialog.value = false;
  currentProduct.value = null;
  skuList.value = [];
  selectedSKUIds.value = [];
};

// 移除商品
const removeProduct = (index) => {
  selectedProducts.value.splice(index, 1);
};

// 活动类型变化
const handleTypeChange = () => {
  selectedProducts.value = [];
};

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return;
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        // 构建请求数据
        const data = {
          name: activityForm.name,
          type: activityForm.type,
          start_time: activityForm.timeRange[0],
          end_time: activityForm.timeRange[1],
          status: activityForm.status,
          description: activityForm.description,
          products: selectedProducts.value
        };
        
        // 如果是兑换码活动，添加兑换码规则
        if (activityForm.type === 'redeem_code') {
          data.redeem_code_rules = redeemCodeRules;
        }
        
        let response;
        if (isEdit.value) {
          response = await activityApi.updateActivity(activityId.value, data);
        } else {
          response = await activityApi.createActivity(data);
        }
        
        ElMessage.success(isEdit.value ? '编辑成功' : '创建成功');
        router.push('/home/activities');
      } catch (error) {
        console.error(isEdit.value ? '编辑活动失败:' : '创建活动失败:', error);
        ElMessage.error(isEdit.value ? '编辑活动失败' : '创建活动失败');
      }
    }
  });
};

// 取消
const handleCancel = () => {
  router.push('/home/activities');
};

// 初始加载
onMounted(() => {
  getActivityDetail();
  searchProducts();
});
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