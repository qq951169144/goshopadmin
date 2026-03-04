<template>
  <div class="home-container">
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header height="60px" class="header">
        <div class="header-left">
          <h1>商城后台管理系统</h1>
        </div>
        <div class="header-right">
          <el-dropdown>
            <span class="user-info">
              {{ user?.username || '用户' }}
              <el-icon class="el-icon--right"><arrow-down /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      
      <el-container>
        <!-- 左侧菜单 -->
        <el-aside width="200px" class="aside">
          <el-menu
            :default-active="activeMenu"
            class="el-menu-vertical-demo"
            @select="handleMenuSelect"
          >
            <el-menu-item index="dashboard">
              <el-icon><house /></el-icon>
              <span>仪表盘</span>
            </el-menu-item>
            <el-menu-item index="users">
              <el-icon><user /></el-icon>
              <span>用户管理</span>
            </el-menu-item>
            <el-menu-item index="roles">
              <el-icon><position /></el-icon>
              <span>角色管理</span>
            </el-menu-item>
            <el-menu-item index="permissions">
              <el-icon><lock /></el-icon>
              <span>权限管理</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        
        <!-- 主内容区 -->
        <el-main class="main">
          <div class="content">
            <el-card v-if="activeMenu === 'dashboard'">
              <template #header>
                <div class="card-header">
                  <span>仪表盘</span>
                </div>
              </template>
              <div class="dashboard-content">
                <h3>欢迎回来，{{ user?.username }}！</h3>
                <p>这是商城后台管理系统的仪表盘页面。</p>
              </div>
            </el-card>
            
            <el-card v-else-if="activeMenu === 'users'">
              <template #header>
                <div class="card-header">
                  <span>用户管理</span>
                </div>
              </template>
              <div class="users-content">
                <p>用户管理功能开发中...</p>
              </div>
            </el-card>
            
            <el-card v-else-if="activeMenu === 'roles'">
              <template #header>
                <div class="card-header">
                  <span>角色管理</span>
                  <el-button type="primary" size="small" @click="showRoleDialog = true">
                    <el-icon><plus /></el-icon> 创建角色
                  </el-button>
                </div>
              </template>
              <div class="roles-content">
                <el-table :data="roles" style="width: 100%">
                  <el-table-column prop="id" label="ID" width="80" />
                  <el-table-column prop="name" label="角色名称" />
                  <el-table-column prop="description" label="描述" />
                  <el-table-column label="操作" width="150">
                    <template #default="scope">
                      <el-button type="primary" size="small" @click="editRole(scope.row)">
                        编辑
                      </el-button>
                      <el-button type="danger" size="small" @click="deleteRole(scope.row.id)">
                        删除
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-card>
            
            <el-card v-else-if="activeMenu === 'permissions'">
              <template #header>
                <div class="card-header">
                  <span>权限管理</span>
                </div>
              </template>
              <div class="permissions-content">
                <p>权限管理功能开发中...</p>
              </div>
            </el-card>
          </div>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { ArrowDown, House, User, Position, Lock, Plus } from '@element-plus/icons-vue';
import { authApi } from '../api/auth';

const activeMenu = ref('dashboard');
const user = ref(null);
const roles = ref([]);
const showRoleDialog = ref(false);

// 计算属性：用户信息
const userInfo = computed(() => {
  const storedUser = localStorage.getItem('user');
  return storedUser ? JSON.parse(storedUser) : null;
});

// 处理菜单选择
const handleMenuSelect = (key) => {
  activeMenu.value = key;
};

// 处理登出
const handleLogout = async () => {
  try {
    await authApi.logout();
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    ElMessage.success('退出登录成功');
    window.location.href = '/login';
  } catch (error) {
    ElMessage.error('退出登录失败');
  }
};

// 获取角色列表
const getRoles = async () => {
  try {
    const response = await authApi.getRoles();
    if (response.code === 200) {
      roles.value = response.data;
    }
  } catch (error) {
    ElMessage.error('获取角色列表失败');
  }
};

// 编辑角色
const editRole = (role) => {
  ElMessage.info('编辑角色功能开发中');
};

// 删除角色
const deleteRole = async (roleId) => {
  try {
    await ElMessageBox.confirm('确定要删除这个角色吗？', '删除角色', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    
    const response = await authApi.deleteRole(roleId);
    if (response.code === 200) {
      ElMessage.success('删除角色成功');
      getRoles();
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除角色失败');
    }
  }
};

// 初始化
onMounted(() => {
  user.value = userInfo.value;
  getRoles();
});
</script>

<style scoped>
.home-container {
  height: 100vh;
  overflow: hidden;
}

.header {
  background: #409EFF;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header h1 {
  font-size: 20px;
  margin: 0;
}

.user-info {
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
}

.aside {
  background: #f5f7fa;
  border-right: 1px solid #e4e7ed;
}

.main {
  background: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
}

.content {
  min-height: calc(100vh - 100px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

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