<template>
  <div class="user-management">
    <!-- 操作栏 -->
    <div class="action-bar">
      <el-button type="primary" @click="handleCreateUser">
        <el-icon><Plus /></el-icon>
        {{ t('userAdmin.createUser') }}
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
      <el-table-column prop="username" :label="t('auth.username')" sortable />
      <el-table-column 
        :prop="dbType === 'mysql' ? 'host' : 'role'" 
        :label="dbType === 'mysql' ? t('userAdmin.hostCol') : t('userAdmin.roleCol')"
        v-if="dbType === 'mysql'"
      />
      <el-table-column prop="role" :label="t('userAdmin.roles')" v-if="dbType === 'postgresql'" />
      <el-table-column prop="privileges" :label="t('userAdmin.privilegeSummary')" :show-overflow-tooltip="true" />
      <el-table-column prop="created_at" :label="t('dbAdmin.createdAt')" />
      <el-table-column :label="t('common.actions')" width="280" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleChangePassword(row)">
            {{ t('userAdmin.changePassword') }}
          </el-button>
          <el-button size="small" @click="handleViewGrants(row)">
            {{ t('userAdmin.viewPermissions') }}
          </el-button>
          <el-button 
            size="small" 
            type="danger" 
            @click="handleDelete(row)"
          >
            {{ t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 统计信息 -->
    <div class="table-footer">
      <span>{{ t('userAdmin.totalNUsers', { n: userList.length }) }}</span>
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
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useConnectionStore } from '@/store/connection'
import { getUserList, deleteUser } from '@/api/userAdmin'
import type { DatabaseUser } from '@/types/userAdmin'
import UserCreateDialog from './UserCreateDialog.vue'
import UserPasswordDialog from './UserPasswordDialog.vue'
import UserDeleteConfirm from './UserDeleteConfirm.vue'

const { t } = useI18n()
const connectionStore = useConnectionStore()

const currentConnection = computed(() => connectionStore.currentConnection)
const dbType = computed(() => currentConnection.value?.db_type || 'mysql')

const userList = ref<DatabaseUser[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const showPasswordDialog = ref(false)
const showDeleteDialog = ref(false)
const selectedUser = ref<DatabaseUser | null>(null)

async function loadUserList() {
  if (!currentConnection.value) return
  
  loading.value = true
  try {
    userList.value = await getUserList(currentConnection.value.id)
  } catch (error: any) {
    ElMessage.error(error.message || t('userAdmin.loadUsersFailed'))
  } finally {
    loading.value = false
  }
}

function handleCreateUser() {
  showCreateDialog.value = true
}

async function handleCreateSuccess() {
  await loadUserList()
  ElMessage.success(t('userAdmin.userCreated'))
}

function handleChangePassword(row: DatabaseUser) {
  selectedUser.value = row
  showPasswordDialog.value = true
}

async function handlePasswordSuccess() {
  await loadUserList()
  ElMessage.success(t('userAdmin.passwordChanged'))
}

function handleViewGrants(row: DatabaseUser) {
  ElMessage.info(t('userAdmin.viewPermTodo'))
}

function handleDelete(row: DatabaseUser) {
  selectedUser.value = row
  showDeleteDialog.value = true
}

async function handleDeleteConfirm() {
  if (!selectedUser.value || !currentConnection.value) return
  
  try {
    await deleteUser(currentConnection.value.id, {
      username: selectedUser.value.username,
      host: selectedUser.value.host
    })
    ElMessage.success(t('userAdmin.userDeleted'))
    await loadUserList()
  } catch (error: any) {
    ElMessage.error(error.message || t('userAdmin.deleteUserFailed'))
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
