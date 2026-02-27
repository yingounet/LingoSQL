<template>
  <div class="permission-management">
    <!-- 用户选择器 -->
    <div class="user-selector">
      <el-select
        v-model="selectedUser"
        placeholder="选择用户"
        @change="handleUserChange"
        style="width: 300px"
      >
        <el-option
          v-for="user in users"
          :key="getUserKey(user)"
          :label="getUserLabel(user)"
          :value="getUserKey(user)"
        />
      </el-select>
    </div>
    
    <!-- 权限管理内容 -->
    <div class="permission-content" v-if="selectedUser">
      <div class="permission-tree-panel">
        <el-tree
          :data="permissionTree"
          :props="treeProps"
          node-key="path"
          default-expand-all
          :expand-on-click-node="false"
        >
          <template #default="{ node, data }">
            <div class="tree-node">
              <span class="node-label">{{ node.label }}</span>
              <span class="node-privileges" v-if="data.privileges && data.privileges.length > 0">
                {{ data.privileges.join(', ') }}
              </span>
            </div>
          </template>
        </el-tree>
      </div>
      
      <div class="permission-action-panel">
        <PermissionGrantDialog
          v-model="showGrantDialog"
          :target="selectedNode"
          :user="selectedUserInfo"
          :db-type="dbType"
          @grant="handleGrant"
        />
      </div>
    </div>
    
    <!-- 未选择用户提示 -->
    <el-empty v-else description="请先选择一个用户" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useConnectionStore } from '@/store/connection'
import { getUserList } from '@/api/userAdmin'
import { getPermissionTree, grantPermission, revokePermission } from '@/api/permissionAdmin'
import type { DatabaseUser } from '@/types/userAdmin'
import type { PermissionNode, GrantPermissionRequest } from '@/types/permissionAdmin'
import PermissionGrantDialog from './PermissionGrantDialog.vue'

const connectionStore = useConnectionStore()

const currentConnection = computed(() => connectionStore.currentConnection)
const dbType = computed(() => currentConnection.value?.db_type || 'mysql')

const users = ref<DatabaseUser[]>([])
const selectedUser = ref<string>('')
const selectedUserInfo = ref<DatabaseUser | null>(null)
const permissionTree = ref<PermissionNode[]>([])
const showGrantDialog = ref(false)
const selectedNode = ref<PermissionNode | null>(null)

const treeProps = {
  children: 'children',
  label: 'name'
}

// 加载用户列表
async function loadUsers() {
  if (!currentConnection.value) return
  
  try {
    users.value = await getUserList(currentConnection.value.id)
  } catch (error: any) {
    ElMessage.error(error.message || '加载用户列表失败')
  }
}

// 获取用户键值
function getUserKey(user: DatabaseUser): string {
  if (dbType.value === 'mysql') {
    return `${user.username}@${user.host || 'localhost'}`
  }
  return user.username
}

// 获取用户标签
function getUserLabel(user: DatabaseUser): string {
  return getUserKey(user)
}

// 用户选择变化
async function handleUserChange(value: string) {
  if (!value) {
    selectedUserInfo.value = null
    permissionTree.value = []
    return
  }
  
  // 找到选中的用户信息
  selectedUserInfo.value = users.value.find(u => getUserKey(u) === value) || null
  
  if (!selectedUserInfo.value || !currentConnection.value) return
  
  // 加载权限树
  try {
    const [username, host] = value.split('@')
    permissionTree.value = await getPermissionTree(
      currentConnection.value.id,
      username,
      dbType.value === 'mysql' ? host : undefined
    )
  } catch (error: any) {
    ElMessage.error(error.message || '加载权限树失败')
  }
}

// 授予权限
async function handleGrant(data: { privileges: string[], isGrant: boolean }) {
  if (!selectedUserInfo.value || !selectedNode.value || !currentConnection.value) return
  
  const [username, host] = selectedUser.value.split('@')
  const request: GrantPermissionRequest = {
    username,
    host: dbType.value === 'mysql' ? host : undefined,
    target_type: selectedNode.value.type,
    target_name: selectedNode.value.name,
    privileges: data.privileges as any[]
  }
  
  // 根据类型设置database和table
  if (selectedNode.value.type === 'table' || selectedNode.value.type === 'column') {
    const pathParts = selectedNode.value.path.split('.')
    request.database = pathParts[0]
    if (selectedNode.value.type === 'table') {
      request.table = pathParts[1]
    }
  }
  
  try {
    if (data.isGrant) {
      await grantPermission(currentConnection.value.id, request)
      ElMessage.success('权限授予成功')
    } else {
      await revokePermission(currentConnection.value.id, request as any)
      ElMessage.success('权限撤销成功')
    }
    // 重新加载权限树
    await handleUserChange(selectedUser.value)
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  }
}

onMounted(() => {
  loadUsers()
})
</script>

<style scoped>
.permission-management {
  padding: var(--spacing-md);
}

.user-selector {
  margin-bottom: var(--spacing-lg);
}

.permission-content {
  display: grid;
  grid-template-columns: 1fr 400px;
  gap: var(--spacing-lg);
}

.permission-tree-panel {
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-medium);
  padding: var(--spacing-md);
  max-height: 600px;
  overflow-y: auto;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  width: 100%;
}

.node-label {
  flex: 1;
  font-weight: 500;
}

.node-privileges {
  font-size: 12px;
  color: var(--color-text-tertiary);
  background-color: var(--color-background-secondary);
  padding: 2px 8px;
  border-radius: var(--border-radius-small);
}

.permission-action-panel {
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-medium);
  padding: var(--spacing-md);
}

/* 响应式 */
@media (max-width: 1024px) {
  .permission-content {
    grid-template-columns: 1fr;
  }
}
</style>
