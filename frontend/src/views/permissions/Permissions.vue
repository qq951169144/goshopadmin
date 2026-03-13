<template>
  <div>
    <el-card>
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
          <el-table-column prop="status" label="状态">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '活跃' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180">
            <template #default="scope">
              <el-button type="primary" size="small" @click="editPermission(scope.row)">
                编辑
              </el-button>
              <el-button type="danger" size="small" @click="deletePermission(scope.row.id)">
                禁用
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

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
        <el-form-item label="状态">
          <el-select v-model="permissionForm.status" placeholder="请选择状态">
            <el-option label="活跃" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
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

const permissions = ref([]);
const showPermissionDialog = ref(false);
const permissionForm = ref({});
const permissionDialogTitle = ref('创建权限');

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

// 编辑权限
const editPermission = (permission) => {
  permissionForm.value = {
    id: permission.id,
    name: permission.name,
    code: permission.code,
    description: permission.description,
    status: permission.status
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
      emit('refresh');
    }
  } catch (error) {
    ElMessage.error(permissionForm.value.id ? '更新权限失败' : '创建权限失败');
  }
};

// 禁用权限
const deletePermission = async (permissionId) => {
  try {
    await ElMessageBox.confirm('确定要禁用这个权限吗？', '禁用权限', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    
    const response = await authApi.deletePermission(permissionId);
    if (response.code === 200) {
      ElMessage.success('禁用权限成功');
      getPermissions();
      emit('refresh');
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('禁用权限失败');
    }
  }
};

// 初始化
onMounted(() => {
  getPermissions();
});
</script>

<style scoped>
.permissions-content {
  margin-top: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>