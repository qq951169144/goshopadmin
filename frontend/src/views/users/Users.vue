<template>
  <div>
    <el-card>
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
                禁用
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 用户对话框 -->
    <el-dialog
      v-model="showUserDialog"
      :title="userDialogTitle"
      width="400px"
      @closed="resetUserForm"
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
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus } from '@element-plus/icons-vue';
import { authApi } from '../../api/auth';

const props = defineProps({
  hasPermission: {
    type: Function,
    required: true
  }
});

const emit = defineEmits(['refresh']);

const users = ref([]);
const roles = ref([]);
const showUserDialog = ref(false);
const userForm = ref({});
const userDialogTitle = ref('创建用户');

// 获取角色列表
const getRoles = async () => {
  try {
    roles.value = await authApi.getRoles();
  } catch (error) {
    ElMessage.error('获取角色列表失败');
  }
};

// 获取用户列表
const getUsers = async () => {
  try {
    const data = await authApi.getUsers();
    // 为每个用户添加role_name字段
    users.value = data.map(user => ({
      ...user,
      role_name: user.role ? user.role.name : ''
    }));
  } catch (error) {
    ElMessage.error('获取用户列表失败');
  }
};

// 重置用户表单
const resetUserForm = () => {
  userForm.value = {};
  userDialogTitle.value = '创建用户';
};

// 编辑用户
const editUser = async (user) => {
  // 确保角色列表已加载
  if (roles.value.length === 0) {
    await getRoles();
  }

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
    if (userForm.value.id) {
      // 更新用户
      await authApi.updateUser(userForm.value.id, userForm.value);
    } else {
      // 创建用户
      await authApi.createUser(userForm.value);
    }
    ElMessage.success(userForm.value.id ? '更新用户成功' : '创建用户成功');
    showUserDialog.value = false;
    getUsers();
    emit('refresh');
  } catch (error) {
    ElMessage.error(userForm.value.id ? '更新用户失败' : '创建用户失败');
  }
};

// 禁用用户
const deleteUser = async (userId) => {
  try {
    await ElMessageBox.confirm('确定要禁用这个用户吗？', '禁用用户', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    
    await authApi.deleteUser(userId);
    ElMessage.success('禁用用户成功');
    getUsers();
    emit('refresh');
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('禁用用户失败');
    }
  }
};

// 初始化
onMounted(() => {
  getUsers();
  getRoles();
});
</script>

<style scoped>
.users-content {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>