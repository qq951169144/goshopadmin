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
            
            <el-card v-else-if="activeMenu === 'users' && hasPermission('user:manage')">
              <template #header>
                <div class="card-header">
                  <span>用户管理</span>
                  <el-button type="primary" size="small" @click="showUserDialog = true">
                    <el-icon><plus /></el-icon> 创建用户
                  </el-button>
                </div>
              </template>
              <div class="users-content">
                <el-table :data="users" style="width: 100%">
                  <el-table-column prop="id" label="ID" width="80" />
                  <el-table-column prop="username" label="用户名" />
                  <el-table-column prop="role_name" label="角色" />
                  <el-table-column prop="status" label="状态">
                    <template #default="scope">
                      <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                        {{ scope.row.status === 'active' ? '活跃' : '禁用' }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column label="操作" width="180">
                    <template #default="scope">
                      <el-button type="primary" size="small" @click="editUser(scope.row)">
                        编辑
                      </el-button>
                      <el-button type="danger" size="small" @click="deleteUser(scope.row.id)">
                        删除
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-card>
            
            <el-card v-else-if="activeMenu === 'roles' && hasPermission('role:manage')">
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
                  <el-table-column label="操作" width="240">
                    <template #default="scope">
                      <el-button type="primary" size="small" @click="editRole(scope.row)">
                        编辑
                      </el-button>
                      <el-button type="success" size="small" @click="assignPermissions(scope.row)">
                        分配权限
                      </el-button>
                      <el-button type="danger" size="small" @click="deleteRole(scope.row.id)">
                        删除
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-card>
            
            <el-card v-else-if="activeMenu === 'permissions' && hasPermission('role:manage')">
              <template #header>
                <div class="card-header">
                  <span>权限管理</span>
                  <el-button type="primary" size="small" @click="showPermissionDialog = true">
                    <el-icon><plus /></el-icon> 创建权限
                  </el-button>
                </div>
              </template>
              <div class="permissions-content">
                <el-table :data="permissions" style="width: 100%">
                  <el-table-column prop="id" label="ID" width="80" />
                  <el-table-column prop="name" label="权限名称" />
                  <el-table-column prop="code" label="权限代码" />
                  <el-table-column prop="description" label="描述" />
                  <el-table-column label="操作" width="180">
                    <template #default="scope">
                      <el-button type="primary" size="small" @click="editPermission(scope.row)">
                        编辑
                      </el-button>
                      <el-button type="danger" size="small" @click="deletePermission(scope.row.id)">
                        删除
                      </el-button>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </el-card>
            
            <!-- 无权限提示 -->
            <el-card v-else-if="(activeMenu === 'users' || activeMenu === 'roles' || activeMenu === 'permissions')">
              <div class="no-permission">
                <el-empty description="您没有权限访问此页面" />
              </div>
            </el-card>
            
            <!-- 用户对话框 -->
            <el-dialog
              v-model="showUserDialog"
              :title="userDialogTitle"
              width="400px"
            >
              <el-form :model="userForm" label-width="80px">
                <el-form-item label="用户名">
                  <el-input v-model="userForm.username" placeholder="请输入用户名" />
                </el-form-item>
                <el-form-item label="密码" v-if="!userForm.id">
                  <el-input v-model="userForm.password" type="password" placeholder="请输入密码" />
                </el-form-item>
                <el-form-item label="角色">
                  <el-select v-model="userForm.role_id" placeholder="请选择角色">
                    <el-option
                      v-for="role in roles"
                      :key="role.id"
                      :label="role.name"
                      :value="role.id"
                    />
                  </el-select>
                </el-form-item>
                <el-form-item label="状态">
                  <el-select v-model="userForm.status" placeholder="请选择状态">
                    <el-option label="活跃" value="active" />
                    <el-option label="禁用" value="inactive" />
                  </el-select>
                </el-form-item>
              </el-form>
              <template #footer>
                <span class="dialog-footer">
                  <el-button @click="showUserDialog = false">取消</el-button>
                  <el-button type="primary" @click="saveUser">保存</el-button>
                </span>
              </template>
            </el-dialog>
            
            <!-- 角色对话框 -->
            <el-dialog
              v-model="showRoleDialog"
              :title="roleDialogTitle"
              width="400px"
            >
              <el-form :model="roleForm" label-width="80px">
                <el-form-item label="角色名称">
                  <el-input v-model="roleForm.name" placeholder="请输入角色名称" />
                </el-form-item>
                <el-form-item label="描述">
                  <el-input v-model="roleForm.description" type="textarea" placeholder="请输入角色描述" />
                </el-form-item>
              </el-form>
              <template #footer>
                <span class="dialog-footer">
                  <el-button @click="showRoleDialog = false">取消</el-button>
                  <el-button type="primary" @click="saveRole">保存</el-button>
                </span>
              </template>
            </el-dialog>
            
            <!-- 权限分配对话框 -->
            <el-dialog
              v-model="showPermissionAssignDialog"
              title="分配权限"
              width="500px"
            >
              <el-form label-width="80px">
                <el-form-item label="角色名称">
                  <el-input v-model="currentRole.name" disabled />
                </el-form-item>
                <el-form-item label="权限列表">
                  <el-checkbox-group v-model="selectedPermissions">
                    <el-checkbox
                      v-for="permission in permissions"
                      :key="permission.id"
                      :label="permission.id"
                    >
                      {{ permission.name }} ({{ permission.code }})
                    </el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
              </el-form>
              <template #footer>
                <span class="dialog-footer">
                  <el-button @click="showPermissionAssignDialog = false">取消</el-button>
                  <el-button type="primary" @click="savePermissions">保存</el-button>
                </span>
              </template>
            </el-dialog>
            
            <!-- 权限对话框 -->
            <el-dialog
              v-model="showPermissionDialog"
              :title="permissionDialogTitle"
              width="400px"
            >
              <el-form :model="permissionForm" label-width="80px">
                <el-form-item label="权限名称">
                  <el-input v-model="permissionForm.name" placeholder="请输入权限名称" />
                </el-form-item>
                <el-form-item label="权限代码">
                  <el-input v-model="permissionForm.code" placeholder="请输入权限代码" />
                </el-form-item>
                <el-form-item label="描述">
                  <el-input v-model="permissionForm.description" type="textarea" placeholder="请输入权限描述" />
                </el-form-item>
              </el-form>
              <template #footer>
                <span class="dialog-footer">
                  <el-button @click="showPermissionDialog = false">取消</el-button>
                  <el-button type="primary" @click="savePermission">保存</el-button>
                </span>
              </template>
            </el-dialog>
          </div>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed, watch } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { ArrowDown, House, User, Position, Lock, Plus } from '@element-plus/icons-vue';
import { authApi } from '../api/auth';

const activeMenu = ref('dashboard');
const user = ref(null);
const users = ref([]);
const roles = ref([]);
const permissions = ref([]);
const showUserDialog = ref(false);
const showRoleDialog = ref(false);
const showPermissionDialog = ref(false);
const showPermissionAssignDialog = ref(false);
const userForm = ref({});
const userDialogTitle = ref('创建用户');
const roleForm = ref({});
const roleDialogTitle = ref('创建角色');
const currentRole = ref({});
const selectedPermissions = ref([]);
const permissionForm = ref({});
const permissionDialogTitle = ref('创建权限');

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

// 获取用户列表
const getUsers = async () => {
  try {
    const response = await authApi.getUsers();
    if (response.code === 200) {
      // 为每个用户添加role_name字段
      users.value = response.data.map(user => ({
        ...user,
        role_name: user.role ? user.role.name : ''
      }));
    }
  } catch (error) {
    ElMessage.error('获取用户列表失败');
  }
};

// 打开创建用户对话框
const openCreateUserDialog = () => {
  userForm.value = {
    username: '',
    password: '',
    role_id: '',
    status: 'active'
  };
  userDialogTitle.value = '创建用户';
  showUserDialog.value = true;
};

// 编辑用户
const editUser = (user) => {
  userForm.value = {
    id: user.id,
    username: user.username,
    role_id: user.role_id,
    status: user.status
  };
  userDialogTitle.value = '编辑用户';
  showUserDialog.value = true;
};

// 保存用户
const saveUser = async () => {
  try {
    let response;
    if (userForm.value.id) {
      // 更新用户
      response = await authApi.updateUser(userForm.value.id, userForm.value);
    } else {
      // 创建用户
      response = await authApi.createUser(userForm.value);
    }
    if (response.code === 200) {
      ElMessage.success(userForm.value.id ? '更新用户成功' : '创建用户成功');
      showUserDialog.value = false;
      getUsers();
    }
  } catch (error) {
    ElMessage.error(userForm.value.id ? '更新用户失败' : '创建用户失败');
  }
};

// 删除用户
const deleteUser = async (userId) => {
  try {
    await ElMessageBox.confirm('确定要删除这个用户吗？', '删除用户', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    
    const response = await authApi.deleteUser(userId);
    if (response.code === 200) {
      ElMessage.success('删除用户成功');
      getUsers();
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除用户失败');
    }
  }
};

// 获取权限列表
const getPermissions = async () => {
  try {
    const response = await authApi.getPermissions();
    if (response.code === 200) {
      permissions.value = response.data;
    }
  } catch (error) {
    ElMessage.error('获取权限列表失败');
  }
};

// 打开创建角色对话框
const openCreateRoleDialog = () => {
  roleForm.value = {
    name: '',
    description: ''
  };
  roleDialogTitle.value = '创建角色';
  showRoleDialog.value = true;
};

// 编辑角色
const editRole = (role) => {
  roleForm.value = {
    id: role.id,
    name: role.name,
    description: role.description
  };
  roleDialogTitle.value = '编辑角色';
  showRoleDialog.value = true;
};

// 保存角色
const saveRole = async () => {
  try {
    let response;
    if (roleForm.value.id) {
      // 更新角色
      response = await authApi.updateRole(roleForm.value.id, roleForm.value);
    } else {
      // 创建角色
      response = await authApi.createRole(roleForm.value);
    }
    if (response.code === 200) {
      ElMessage.success(roleForm.value.id ? '更新角色成功' : '创建角色成功');
      showRoleDialog.value = false;
      getRoles();
    }
  } catch (error) {
    ElMessage.error(roleForm.value.id ? '更新角色失败' : '创建角色失败');
  }
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

// 分配权限
const assignPermissions = async (role) => {
  currentRole.value = role;
  // 获取所有权限
  await getPermissions();
  // 设置已选权限
  selectedPermissions.value = role.permissions ? role.permissions.map(p => p.id) : [];
  showPermissionAssignDialog.value = true;
};

// 保存权限
const savePermissions = async () => {
  try {
    const response = await authApi.assignPermissions(currentRole.value.id, {
      permission_ids: selectedPermissions.value
    });
    if (response.code === 200) {
      ElMessage.success('分配权限成功');
      showPermissionAssignDialog.value = false;
      getRoles();
    }
  } catch (error) {
    ElMessage.error('分配权限失败');
  }
};

// 打开创建权限对话框
const openCreatePermissionDialog = () => {
  permissionForm.value = {
    name: '',
    code: '',
    description: ''
  };
  permissionDialogTitle.value = '创建权限';
  showPermissionDialog.value = true;
};

// 编辑权限
const editPermission = (permission) => {
  permissionForm.value = {
    id: permission.id,
    name: permission.name,
    code: permission.code,
    description: permission.description
  };
  permissionDialogTitle.value = '编辑权限';
  showPermissionDialog.value = true;
};

// 保存权限
const savePermission = async () => {
  try {
    let response;
    if (permissionForm.value.id) {
      // 更新权限
      response = await authApi.updatePermission(permissionForm.value.id, permissionForm.value);
    } else {
      // 创建权限
      response = await authApi.createPermission(permissionForm.value);
    }
    if (response.code === 200) {
      ElMessage.success(permissionForm.value.id ? '更新权限成功' : '创建权限成功');
      showPermissionDialog.value = false;
      getPermissions();
    }
  } catch (error) {
    ElMessage.error(permissionForm.value.id ? '更新权限失败' : '创建权限失败');
  }
};

// 删除权限
const deletePermission = async (permissionId) => {
  try {
    await ElMessageBox.confirm('确定要删除这个权限吗？', '删除权限', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    
    const response = await authApi.deletePermission(permissionId);
    if (response.code === 200) {
      ElMessage.success('删除权限成功');
      getPermissions();
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除权限失败');
    }
  }
};

// 监听用户信息变化
watch(() => userInfo.value, async (newUser) => {
  if (newUser) {
    user.value = newUser;
    // 获取用户权限信息
    try {
      const response = await authApi.getCurrentUser();
      if (response.code === 200) {
        // 保存权限信息到localStorage
        localStorage.setItem('permissions', JSON.stringify(response.data.permissions || []));
        // 重新获取数据
        getRoles();
        getUsers();
        getPermissions();
      }
    } catch (error) {
      console.error('获取用户权限失败', error);
    }
  }
}, { immediate: true });

// 初始化
onMounted(() => {
  // 只在用户已登录时获取数据
  if (userInfo.value) {
    getRoles();
    getUsers();
    getPermissions();
  }
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