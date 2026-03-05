<template>
  <div>
    <el-card>
      <template #header>
        <div class="card-header">
          <span>商户管理</span>
          <el-button type="primary" size="small" @click="showMerchantDialog = true">
            <el-icon><plus /></el-icon> 创建商户
          </el-button>
        </div>
      </template>
      <div class="merchants-content">
        <el-table :data="merchants" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column prop="name" label="商户名称" />
          <el-table-column prop="contact_person" label="联系人" />
          <el-table-column prop="contact_phone" label="联系电话" />
          <el-table-column prop="audit_status" label="审核状态">
            <template #default="scope">
              <el-tag :type="getAuditStatusType(scope.row.audit_status)">
                {{ getAuditStatusText(scope.row.audit_status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '活跃' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="300">
            <template #default="scope">
              <el-button type="primary" size="small" @click="editMerchant(scope.row)">
                编辑
              </el-button>
              <el-button type="success" size="small" @click="auditMerchant(scope.row)">
                审核
              </el-button>
              <el-button type="info" size="small" @click="manageMerchantUsers(scope.row)">
                管理用户
              </el-button>
              <el-button type="danger" size="small" @click="deleteMerchant(scope.row.id)">
                禁用
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 商户对话框 -->
    <el-dialog
      v-model="showMerchantDialog"
      :title="merchantDialogTitle"
      width="500px"
    >
      <el-form :model="merchantForm" label-width="100px">
        <el-form-item label="商户名称">
          <el-input v-model="merchantForm.name" placeholder="请输入商户名称" />
        </el-form-item>
        <el-form-item label="联系人">
          <el-input v-model="merchantForm.contact_person" placeholder="请输入联系人" />
        </el-form-item>
        <el-form-item label="联系电话">
          <el-input v-model="merchantForm.contact_phone" placeholder="请输入联系电话" />
        </el-form-item>
        <el-form-item label="地址">
          <el-input v-model="merchantForm.address" type="textarea" placeholder="请输入地址" />
        </el-form-item>
        <el-form-item label="状态" v-if="merchantForm.id">
          <el-select v-model="merchantForm.status" placeholder="请选择状态">
            <el-option label="活跃" value="active" />
            <el-option label="禁用" value="inactive" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showMerchantDialog = false">取消</el-button>
          <el-button type="primary" @click="saveMerchant">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 商户审核对话框 -->
    <el-dialog
      v-model="showMerchantAuditDialog"
      title="审核商户"
      width="400px"
    >
      <el-form :model="merchantAuditForm" label-width="80px">
        <el-form-item label="商户名称">
          <el-input v-model="currentMerchant.name" disabled />
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="merchantAuditForm.audit_status" placeholder="请选择审核状态">
            <el-option label="待审核" value="pending" />
            <el-option label="已通过" value="approved" />
            <el-option label="已拒绝" value="rejected" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input v-model="merchantAuditForm.audit_note" type="textarea" placeholder="请输入审核备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showMerchantAuditDialog = false">取消</el-button>
          <el-button type="primary" @click="saveMerchantAudit">保存</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 商户用户管理对话框 -->
    <el-dialog
      v-model="showMerchantUserDialog"
      title="管理商户用户"
      width="600px"
    >
      <div>
        <h4>商户：{{ currentMerchant.name }}</h4>
        <el-button type="primary" size="small" @click="showAddMerchantUserDialog = true" style="margin-bottom: 10px">
          <el-icon><plus /></el-icon> 添加用户
        </el-button>
        <el-table :data="merchantUsers" style="width: 100%">
          <el-table-column prop="user.id" label="用户ID" width="80" />
          <el-table-column prop="user.username" label="用户名" />
          <el-table-column label="操作" width="100">
            <template #default="scope">
              <el-button type="danger" size="small" @click="removeMerchantUser(scope.row.user.id)">
                移除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showMerchantUserDialog = false">关闭</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 添加商户用户对话框 -->
    <el-dialog
      v-model="showAddMerchantUserDialog"
      title="添加商户用户"
      width="400px"
    >
      <el-form :model="addMerchantUserForm" label-width="80px">
        <el-form-item label="选择用户">
          <el-select v-model="addMerchantUserForm.user_id" placeholder="请选择用户">
            <el-option
              v-for="user in allUsers"
              :key="user.id"
              :label="user.username"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showAddMerchantUserDialog = false">取消</el-button>
          <el-button type="primary" @click="addMerchantUser">添加</el-button>
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

const merchants = ref([]);
const merchantUsers = ref([]);
const allUsers = ref([]);
const showMerchantDialog = ref(false);
const showMerchantAuditDialog = ref(false);
const showMerchantUserDialog = ref(false);
const showAddMerchantUserDialog = ref(false);
const merchantForm = ref({});
const merchantDialogTitle = ref('创建商户');
const merchantAuditForm = ref({});
const currentMerchant = ref({});
const addMerchantUserForm = ref({});

// 获取商户列表
const getMerchants = async () => {
  try {
    const response = await authApi.getMerchants();
    if (response.code === 200) {
      merchants.value = response.data;
    }
  } catch (error) {
    ElMessage.error('获取商户列表失败');
  }
};

// 获取商户用户列表
const getMerchantUsers = async (merchantId) => {
  try {
    const response = await authApi.getMerchantUsers(merchantId);
    if (response.code === 200) {
      merchantUsers.value = response.data;
    }
  } catch (error) {
    ElMessage.error('获取商户用户列表失败');
  }
};

// 获取所有用户列表
const getAllUsers = async () => {
  try {
    const response = await authApi.getUsers();
    if (response.code === 200) {
      allUsers.value = response.data;
    }
  } catch (error) {
    ElMessage.error('获取用户列表失败');
  }
};

// 编辑商户
const editMerchant = (merchant) => {
  merchantForm.value = {
    id: merchant.id,
    name: merchant.name,
    contact_person: merchant.contact_person,
    contact_phone: merchant.contact_phone,
    address: merchant.address,
    status: merchant.status
  };
  merchantDialogTitle.value = '编辑商户';
  showMerchantDialog.value = true;
};

// 保存商户
const saveMerchant = async () => {
  try {
    let response;
    if (merchantForm.value.id) {
      // 更新商户
      response = await authApi.updateMerchant(merchantForm.value.id, merchantForm.value);
    } else {
      // 创建商户
      response = await authApi.createMerchant(merchantForm.value);
    }
    if (response.code === 200) {
      ElMessage.success(merchantForm.value.id ? '更新商户成功' : '创建商户成功');
      showMerchantDialog.value = false;
      getMerchants();
      emit('refresh');
    }
  } catch (error) {
    ElMessage.error(merchantForm.value.id ? '更新商户失败' : '创建商户失败');
  }
};

// 禁用商户
const deleteMerchant = async (merchantId) => {
  try {
    await ElMessageBox.confirm('确定要禁用这个商户吗？', '禁用商户', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    
    const response = await authApi.deleteMerchant(merchantId);
    if (response.code === 200) {
      ElMessage.success('禁用商户成功');
      getMerchants();
      emit('refresh');
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('禁用商户失败');
    }
  }
};

// 审核商户
const auditMerchant = (merchant) => {
  currentMerchant.value = merchant;
  merchantAuditForm.value = {
    audit_status: merchant.audit_status,
    audit_note: ''
  };
  showMerchantAuditDialog.value = true;
};

// 保存商户审核
const saveMerchantAudit = async () => {
  try {
    const response = await authApi.auditMerchant(currentMerchant.value.id, merchantAuditForm.value);
    if (response.code === 200) {
      ElMessage.success('审核商户成功');
      showMerchantAuditDialog.value = false;
      getMerchants();
      emit('refresh');
    }
  } catch (error) {
    ElMessage.error('审核商户失败');
  }
};

// 管理商户用户
const manageMerchantUsers = async (merchant) => {
  currentMerchant.value = merchant;
  await getMerchantUsers(merchant.id);
  await getAllUsers();
  showMerchantUserDialog.value = true;
};

// 添加商户用户
const addMerchantUser = async () => {
  try {
    const response = await authApi.addMerchantUser(currentMerchant.value.id, addMerchantUserForm.value);
    if (response.code === 200) {
      ElMessage.success('添加商户用户成功');
      showAddMerchantUserDialog.value = false;
      getMerchantUsers(currentMerchant.value.id);
    }
  } catch (error) {
    ElMessage.error('添加商户用户失败');
  }
};

// 移除商户用户
const removeMerchantUser = async (userId) => {
  try {
    await ElMessageBox.confirm('确定要移除这个用户吗？', '移除用户', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    });
    
    const response = await authApi.removeMerchantUser(currentMerchant.value.id, userId);
    if (response.code === 200) {
      ElMessage.success('移除商户用户成功');
      getMerchantUsers(currentMerchant.value.id);
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('移除商户用户失败');
    }
  }
};

// 获取审核状态类型
const getAuditStatusType = (status) => {
  switch (status) {
    case 'pending':
      return 'warning';
    case 'approved':
      return 'success';
    case 'rejected':
      return 'danger';
    default:
      return '';
  }
};

// 获取审核状态文本
const getAuditStatusText = (status) => {
  switch (status) {
    case 'pending':
      return '待审核';
    case 'approved':
      return '已通过';
    case 'rejected':
      return '已拒绝';
    default:
      return '';
  }
};

// 初始化
onMounted(() => {
  getMerchants();
});
</script>

<style scoped>
.merchants-content {
  margin-top: 20px;
}
</style>