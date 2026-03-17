<template>
  <div class="home-container">
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header height="60px" class="header">
        <div class="header-left">
          <h1>商城后台管理系统-1.0</h1>
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
            <el-menu-item index="users" v-if="hasPermission('user:manage')">
              <el-icon><user /></el-icon>
              <span>用户管理</span>
            </el-menu-item>
            <el-menu-item index="roles" v-if="hasPermission('role:manage')">
              <el-icon><position /></el-icon>
              <span>角色管理</span>
            </el-menu-item>
            <el-menu-item index="permissions" v-if="hasPermission('role:manage')">
              <el-icon><lock /></el-icon>
              <span>权限管理</span>
            </el-menu-item>
            <el-menu-item index="merchants" v-if="hasPermission('merchant:manage')">
              <el-icon><shop /></el-icon>
              <span>商户管理</span>
            </el-menu-item>
            <el-menu-item index="product-categories" v-if="hasPermission('product:category')">
              <el-icon><grid /></el-icon>
              <span>商品分类</span>
            </el-menu-item>
            <el-menu-item index="products" v-if="hasPermission('product:manage')">
              <el-icon><goods /></el-icon>
              <span>商品管理</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        
        <!-- 主内容区 -->
        <el-main class="main">
          <div class="content">
            <!-- 仪表盘 -->
            <Dashboard v-if="currentView === 'dashboard'" />
            
            <!-- 用户管理 -->
            <Users v-else-if="currentView === 'users' && hasPermission('user:manage')" :has-permission="hasPermission" @refresh="handleRefresh" />
            
            <!-- 角色管理 -->
            <Roles v-else-if="currentView === 'roles' && hasPermission('role:manage')" :has-permission="hasPermission" @refresh="handleRefresh" />
            
            <!-- 权限管理 -->
            <Permissions v-else-if="currentView === 'permissions' && hasPermission('role:manage')" :has-permission="hasPermission" @refresh="handleRefresh" />
            
            <!-- 商户管理 -->
            <Merchants v-else-if="currentView === 'merchants' && hasPermission('merchant:manage')" :has-permission="hasPermission" @refresh="handleRefresh" />
            
            <!-- 商品管理 -->
            <Products 
              v-else-if="currentView === 'products' && hasPermission('product:manage')" 
              :has-permission="hasPermission" 
              @refresh="handleRefresh"
              @manage-specifications="handleManageSpecifications"
              @manage-skus="handleManageSKUs"
            />
            
            <!-- 商品分类管理 -->
            <ProductCategories v-else-if="currentView === 'product-categories' && hasPermission('product:category')" :has-permission="hasPermission" @refresh="handleRefresh" />
            
            <!-- 规格管理 -->
            <ProductSpecifications 
              v-else-if="currentView === 'specifications' && hasPermission('product:manage')" 
              :product-id="currentProduct?.id"
              :product-name="currentProduct?.name"
              @back="handleBackToProducts"
            />
            
            <!-- SKU管理 -->
            <ProductSKUs 
              v-else-if="currentView === 'skus' && hasPermission('product:manage')" 
              :product-id="currentProduct?.id"
              :product-name="currentProduct?.name"
              @back="handleBackToProducts"
            />
            
            <!-- 无权限提示 -->
            <el-card v-else-if="(currentView === 'users' || currentView === 'roles' || currentView === 'permissions' || currentView === 'merchants' || currentView === 'products' || currentView === 'product-categories')">
              <div class="no-permission">
                <el-empty description="您没有权限访问此页面" />
              </div>
            </el-card>
          </div>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { ArrowDown, House, User, Position, Lock, Shop, Goods, Grid } from '@element-plus/icons-vue';
import { authApi } from '../api/auth';

// 导入子组件
import Dashboard from './dashboard/Dashboard.vue';
import Users from './users/Users.vue';
import Roles from './roles/Roles.vue';
import Permissions from './permissions/Permissions.vue';
import Merchants from './merchants/Merchants.vue';
import Products from './products/Products.vue';
import ProductCategories from './products/ProductCategories.vue';
import ProductSpecifications from './products/ProductSpecifications.vue';
import ProductSKUs from './products/ProductSKUs.vue';

const activeMenu = ref('dashboard');
const currentView = ref('dashboard'); // 当前视图：dashboard, users, roles, permissions, merchants, products, product-categories, specifications, skus
const currentProduct = ref(null); // 当前选中的商品
const user = ref(null);

// 计算属性：用户信息
const userInfo = computed(() => {
  const storedUser = localStorage.getItem('user');
  return storedUser ? JSON.parse(storedUser) : null;
});

// 计算属性：用户权限
const userPermissions = computed(() => {
  const storedPermissions = localStorage.getItem('permissions');
  return storedPermissions ? JSON.parse(storedPermissions) : [];
});

// 检查用户是否有指定权限
const hasPermission = (permissionCode) => {
  return userPermissions.value.some(p => p.code === permissionCode);
};

// 处理菜单选择
const handleMenuSelect = (key) => {
  activeMenu.value = key;
  currentView.value = key;
  currentProduct.value = null;
};

// 处理管理规格
const handleManageSpecifications = (product) => {
  currentProduct.value = product;
  currentView.value = 'specifications';
};

// 处理管理SKU
const handleManageSKUs = (product) => {
  currentProduct.value = product;
  currentView.value = 'skus';
};

// 返回商品列表
const handleBackToProducts = () => {
  currentView.value = 'products';
  currentProduct.value = null;
};

// 处理登出
const handleLogout = async () => {
  try {
    await authApi.logout();
    // 清除所有localStorage数据
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('permissions');
    ElMessage.success('退出登录成功');
    window.location.href = '/login';
  } catch (error) {
    ElMessage.error('退出登录失败');
  }
};

// 处理刷新
const handleRefresh = () => {
  // 可以在这里添加刷新逻辑，例如重新加载用户信息等
  console.log('刷新数据');
};

// 监听用户信息变化
watch(() => userInfo.value, async (newUser) => {
  if (newUser) {
    user.value = newUser;
    // 获取用户权限信息
    const token = localStorage.getItem('token');
    if (token) {
      try {
        const data = await authApi.getCurrentUser();
        // 保存权限信息到localStorage
        localStorage.setItem('permissions', JSON.stringify(data.permissions || []));
      } catch (error) {
        ElMessage.warning('权限信息加载失败，部分功能可能受限');
      }
    }
  }
}, { immediate: true });
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

.no-permission {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
}
</style>