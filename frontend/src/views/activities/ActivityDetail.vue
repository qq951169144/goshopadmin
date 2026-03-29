<template>
  <div class="activity-detail-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>活动详情</span>
          <el-button type="primary" @click="handleEditActivity">编辑活动</el-button>
          <el-button @click="handleBack">返回列表</el-button>
        </div>
      </template>
      
      <!-- 活动基本信息 -->
      <el-descriptions :column="1" border>
        <el-descriptions-item label="活动ID">{{ activityDetail.id }}</el-descriptions-item>
        <el-descriptions-item label="活动名称">{{ activityDetail.name }}</el-descriptions-item>
        <el-descriptions-item label="活动类型">
          <el-tag v-if="activityDetail.type === 'flash_sale'" type="primary">秒杀活动</el-tag>
          <el-tag v-else-if="activityDetail.type === 'redeem_code'" type="success">兑换码活动</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="活动时间">{{ activityDetail.start_time }} 至 {{ activityDetail.end_time }}</el-descriptions-item>
        <el-descriptions-item label="活动状态">
          <el-tag v-if="activityDetail.status === 'pending'" type="info">未开始</el-tag>
          <el-tag v-else-if="activityDetail.status === 'active'" type="success">进行中</el-tag>
          <el-tag v-else-if="activityDetail.status === 'ended'" type="warning">已结束</el-tag>
          <el-tag v-else-if="activityDetail.status === 'cancelled'" type="danger">已取消</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="活动描述">{{ activityDetail.description || '-' }}</el-descriptions-item>
      </el-descriptions>
      
      <!-- 兑换码规则（兑换码活动） -->
      <template v-if="activityDetail.type === 'redeem_code' && activityDetail.redeem_code_rules">
        <el-divider>兑换码规则</el-divider>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="兑换码类型">{{ activityDetail.redeem_code_rules.type === 'number' ? '数字' : activityDetail.redeem_code_rules.type === 'letter' ? '字母' : '混合' }}</el-descriptions-item>
          <el-descriptions-item label="兑换码长度">{{ activityDetail.redeem_code_rules.length }}</el-descriptions-item>
          <el-descriptions-item label="排除字符">{{ activityDetail.redeem_code_rules.exclude_chars || '-' }}</el-descriptions-item>
        </el-descriptions>
        
        <!-- 兑换码统计 -->
        <el-divider>兑换码统计</el-divider>
        <el-row :gutter="20">
          <el-col :span="6">
            <el-card shadow="hover">
              <template #header>
                <div class="card-header">
                  <span>总兑换码数</span>
                </div>
              </template>
              <div class="stat-value">{{ redeemCodeStats.total || 0 }}</div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <template #header>
                <div class="card-header">
                  <span>已使用</span>
                </div>
              </template>
              <div class="stat-value">{{ redeemCodeStats.used || 0 }}</div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <template #header>
                <div class="card-header">
                  <span>未使用</span>
                </div>
              </template>
              <div class="stat-value">{{ redeemCodeStats.unused || 0 }}</div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card shadow="hover">
              <template #header>
                <div class="card-header">
                  <span>已过期</span>
                </div>
              </template>
              <div class="stat-value">{{ redeemCodeStats.expired || 0 }}</div>
            </el-card>
          </el-col>
        </el-row>
        
        <!-- 兑换码管理操作 -->
        <div class="mt-4">
          <el-button type="primary" @click="handleGenerateCodes">生成兑换码</el-button>
          <el-button @click="handleImportExportCodes">导入导出兑换码</el-button>
          <el-button @click="handleViewCodes">查看兑换码列表</el-button>
        </div>
      </template>
      
      <!-- 关联商品 -->
      <el-divider>关联商品</el-divider>
      <el-table :data="activityDetail.products || []" style="width: 100%">
        <el-table-column prop="product_id" label="商品ID" width="100"></el-table-column>
        <el-table-column prop="product_name" label="商品名称"></el-table-column>
        <el-table-column prop="sku_id" label="SKU ID" width="100"></el-table-column>
        <el-table-column prop="sku_name" label="SKU名称"></el-table-column>
        <el-table-column prop="activity_price" label="活动价格" width="120">
          <template #default="scope">
            ¥{{ scope.row.activity_price || scope.row.price }}
          </template>
        </el-table-column>
        <el-table-column prop="activity_stock" label="活动库存" width="120">
          <template #default="scope">
            {{ scope.row.activity_stock || scope.row.stock }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import { activityApi } from '../../api/auth';

const router = useRouter();
const route = useRoute();

// 活动详情
const activityDetail = ref({});

// 兑换码统计
const redeemCodeStats = ref({
  total: 0,
  used: 0,
  unused: 0,
  expired: 0
});

// 获取活动详情
const getActivityDetail = async () => {
  try {
    const activityId = route.params.id;
    const activity = await activityApi.getActivity(activityId);
    activityDetail.value = activity;
    
    // 如果是兑换码活动，获取兑换码统计
    if (activity.type === 'redeem_code') {
      getRedeemCodeStats(activityId);
    }
  } catch (error) {
    console.error('获取活动详情失败:', error);
    ElMessage.error('获取活动详情失败');
  }
};

// 获取兑换码统计
const getRedeemCodeStats = async (activityId) => {
  try {
    // 这里应该调用获取兑换码统计的API，暂时使用模拟数据
    // const stats = await activityApi.getRedeemCodeStats(activityId);
    // redeemCodeStats.value = stats;
    
    // 模拟数据
    redeemCodeStats.value = {
      total: 100,
      used: 20,
      unused: 75,
      expired: 5
    };
  } catch (error) {
    console.error('获取兑换码统计失败:', error);
  }
};

// 编辑活动
const handleEditActivity = () => {
  const activityId = route.params.id;
  router.push(`/home/activities/${activityId}/edit`);
};

// 返回列表
const handleBack = () => {
  router.push('/home/activities');
};

// 生成兑换码
const handleGenerateCodes = () => {
  const activityId = route.params.id;
  router.push(`/home/activities/${activityId}/redeem-codes/generate`);
};

// 导入导出兑换码
const handleImportExportCodes = () => {
  const activityId = route.params.id;
  router.push(`/home/activities/${activityId}/redeem-codes/import-export`);
};

// 查看兑换码列表
const handleViewCodes = () => {
  const activityId = route.params.id;
  router.push(`/home/activities/${activityId}/redeem-codes`);
};

// 初始加载
onMounted(() => {
  getActivityDetail();
});
</script>

<style scoped>
.activity-detail-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  text-align: center;
  margin-top: 10px;
}

.mt-4 {
  margin-top: 20px;
}
</style>