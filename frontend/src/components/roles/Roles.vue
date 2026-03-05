<template>
  <div>
    <el-card>
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
          <el-table-column prop="status" label="状态">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '活跃' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="240">
            <template #default="scope">
              <el-button type="primary" size="small" @click="editRole(scope.row)">
                编辑
              </el-button>
              <el-button type="success" size="small" @click="assignPermissions(scope.row)">
                分配权限
              </el-button>
              <el-button type="danger" size="small" @click="deleteRole(scope.row.id)">
                禁用
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

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
        <el-form-item label="状态">
          <el-select v-model="roleForm.status" placeholder="请选择状态">
            <el-option label="活跃" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
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

const roles = ref([]);
const permissions = ref([]);
const showRoleDialog = ref(false);
const showPermissionAssignDialog = ref(false);
const roleForm = ref({});
const roleDialogTitle = ref('创建角色');
const currentRole = ref({});
const selectedPermissions = ref([]);

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

// 编辑角色
const editRole = (role) => {
  roleForm.value = {
    id: role.id,
    name: role.name,
    description: role.description,
    status: role.status
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
      emit('refresh');
    }
  } catch (error) {
    ElMessage.error(roleForm.value.id ? '更新角色失败' : '创建角色失败');
  }
};

// 禁用角色
const deleteRole = async (roleId) => {
  try {
    await ElMessageBox.confirm('确定要禁用这个角色吗？', '禁用角色', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    
    const response = await authApi.deleteRole(roleId);
    if (response.code === 200) {
      ElMessage.success('禁用角色成功');
      getRoles();
      emit('refresh');
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('禁用角色失败');
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
      emit('refresh');
    }
  } catch (error) {
    ElMessage.error('分配权限失败');
  }
};

// 初始化
onMounted(() => {
  getRoles();
  getPermissions();
});
</script>

<style scoped>
.roles-content {
  margin-top: 20px;
}
</style>