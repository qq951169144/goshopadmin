<template>
  <el-card>
    <template #header>
      <div class="card-header">
        <span>后台首页</span>
      </div>
    </template>
    <div class="dashboard-content">
      <h3>欢迎回来，{{ user?.username }}！</h3>
      <p>这是商城后台管理系统的后台首页。</p>
    </div>
  </el-card>
</template>

<script setup>
import { computed } from 'vue';

// 计算属性：用户信息
const userInfo = computed(() => {
  const storedUser = localStorage.getItem('user');
  return storedUser ? JSON.parse(storedUser) : null;
});

const user = computed(() => userInfo.value);
</script>

<style scoped>
.dashboard-content {
  padding: 20px 0;
}

.dashboard-content h3 {
  color: #303133;
  margin-bottom: 10px;
}

.dashboard-content p {
  color: #606266;
}
</style>