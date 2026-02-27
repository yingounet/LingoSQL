<template>
  <div class="user-management">
    <!-- 操作栏 -->
    <div class="action-bar">
      <el-button type="primary" @click="handleCreateUser">
        <el-icon><Plus /></el-icon>
        创建用户
      </el-button>
    </div>
    
    <!-- 用户列表表格 -->
    <el-table
      :data="userList"
      v-loading="loading"
      stripe
      border
      style="width: 100%"
    >
      <el-table-column prop="username" label="用户名" sortable />
      <el-table-column 
        :prop="dbType === 'mysql' ? 'host' : 'role'" 
        :label="dbType === 'mysql' ? '主机' : '角色'"
        v-if="dbType === 'mysql'"
      />
      <el-table-column prop="role" label="角色" v-if="dbType === 'postgresql'" />
      <el-table-column prop="privileges" label="权限摘要" :show-overflow-tooltip="true" />
      <el-table-column prop="created_at" label="创建时间" />
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleChangePassword(row)">
            修改密码
          </el-button>
          <el-button size="small" @click="handleViewGrants(row)">
            查看权限
          </el-button>
          <el-button 
            size="small" 
            type="danger" 
            @click="handleDelete(row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 统计信息 -->
    <div class="table-footer">
      <span>共 {{ userList.length }} 个用户</span>
    </div>
    
    <!-- 对话框组件 -->
    <UserCreateDialog
      v-model="showCreateDialog"
      :db-type="dbType"
      @success="handleCreateSuccess"
    />
    <UserPasswordDialog
      v-model="showPasswordDialog"
      :user="selectedUser"
      :db-type="dbType"
      @success="handlePasswordSuccess"
    />
    <UserDeleteConfirm
      v-model="showDeleteDialog"
      :user="selectedUser"
      :db-type="dbType"
      @confirm="handleDeleteConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useConnectionStore } from '@/store/connection'
import { getUserList, deleteUser } from '@/api/userAdmin'
import type { DatabaseUser } from '@/types/userAdmin'
import UserCreateDialog from './UserCreateDialog.vue'
import UserPasswordDialog from './UserPasswordDialog.vue'
import UserDeleteConfirm from './UserDeleteConfirm.vue'

const connectionStore = useConnectionStore()

const currentConnection = computed(() => connectionStore.currentConnection)
const dbType = computed(() => currentConnection.value?.db_type || 'mysql')

const userList = ref<DatabaseUser[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const showPasswordDialog = ref(false)
const showDeleteDialog = ref(false)
const selectedUser = ref<DatabaseUser | null>(null)

// 加载用户列表
async function loadUserList() {
  if (!currentConnection.value) return
  
  loading.value = true
  try {
    userList.value = await getUserList(currentConnection.value.id)
  } catch (error: any) {
    ElMessage.error(error.message || '加载用户列表失败')
  } finally {
    loading.value = false
  }
}

// 创建用户
function handleCreateUser() {
  showCreateDialog.value = true
}

// 创建成功
async function handleCreateSuccess() {
  await loadUserList()
  ElMessage.success('用户创建成功')
}

// 修改密码
function handleChangePassword(row: DatabaseUser) {
  selectedUser.value = row
  showPasswordDialog.value = true
}

// 密码修改成功
async function handlePasswordSuccess() {
  await loadUserList()
  ElMessage.success('密码修改成功')
}

// 查看权限
function handleViewGrants(row: DatabaseUser) {
  // TODO: 跳转到权限管理Tab并选中该用户
  ElMessage.info('查看权限功能待实现')
}

// 删除
function handleDelete(row: DatabaseUser) {
  selectedUser.value = row
  showDeleteDialog.value = true
}

// 确认删除
async function handleDeleteConfirm() {
  if (!selectedUser.value || !currentConnection.value) return
  
  try {
    await deleteUser(currentConnection.value.id, {
      username: selectedUser.value.username,
      host: selectedUser.value.host
    })
    ElMessage.success('用户删除成功')
    await loadUserList()
  } catch (error: any) {
    ElMessage.error(error.message || '删除用户失败')
  } finally {
    showDeleteDialog.value = false
    selectedUser.value = null
  }
}

onMounted(() => {
  loadUserList()
})
</script>

<style scoped>
.user-management {
  padding: var(--spacing-md);
}

.action-bar {
  margin-bottom: var(--spacing-md);
}

.table-footer {
  margin-top: var(--spacing-md);
  padding: var(--spacing-sm);
  text-align: right;
  color: var(--color-text-secondary);
  font-size: 14px;
}
</style>
