<template>
  <div class="database-admin-tabs">
    <!-- 左侧竖栏导航（表管理始终显示；库/用户/权限根据权限显示） -->
    <aside class="admin-sidebar">
      <el-menu
        :default-active="activeAdminTab"
        class="admin-menu"
        @select="handleMenuSelect"
      >
        <div class="menu-group-label">库与表</div>
        <el-menu-item v-if="adminPermissions?.has_database_admin" index="database">
          <el-icon><Folder /></el-icon>
          <span>数据库管理</span>
          <span class="menu-desc">创建、删除数据库</span>
        </el-menu-item>
        <el-menu-item index="table">
          <el-icon><Grid /></el-icon>
          <span>表管理</span>
          <span class="menu-desc">建表、删表</span>
        </el-menu-item>
        <el-menu-item index="backup">
          <el-icon><FolderOpened /></el-icon>
          <span>备份与恢复</span>
          <span class="menu-desc">备份、恢复、管理备份文件</span>
        </el-menu-item>
        <div v-if="adminPermissions?.has_user_admin || adminPermissions?.has_permission_admin" class="menu-group-label">用户与权限</div>
        <el-menu-item v-if="adminPermissions?.has_user_admin" index="user">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
          <span class="menu-desc">创建用户、改密、删用户</span>
        </el-menu-item>
        <el-menu-item v-if="adminPermissions?.has_permission_admin" index="permission">
          <el-icon><Key /></el-icon>
          <span>权限管理</span>
          <span class="menu-desc">授予、回收权限</span>
        </el-menu-item>
      </el-menu>
    </aside>
    <!-- 右侧内容区 -->
    <main class="admin-content">
      <DatabaseManagement v-show="activeAdminTab === 'database'" />
      <TableManagement v-show="activeAdminTab === 'table'" />
      <BackupManagement v-show="activeAdminTab === 'backup'" />
      <UserManagement v-show="activeAdminTab === 'user'" />
      <PermissionManagement v-show="activeAdminTab === 'permission'" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Folder, FolderOpened, Grid, User, Key } from '@element-plus/icons-vue'
import type { AdminPermission } from '@/types/databaseAdmin'
import DatabaseManagement from './DatabaseManagement.vue'
import TableManagement from './TableManagement.vue'
import BackupManagement from './BackupManagement.vue'
import UserManagement from './UserManagement.vue'
import PermissionManagement from './PermissionManagement.vue'

const props = defineProps<{
  adminPermissions?: AdminPermission | null
}>()

const route = useRoute()
const router = useRouter()
type TabType = 'database' | 'table' | 'backup' | 'user' | 'permission'
const activeAdminTab = ref<TabType>('table')

function resolveTab(): TabType {
  const tab = route.query.admin_tab as string
  if (tab === 'table') return 'table'
  if (tab === 'backup') return 'backup'
  if (tab === 'database' && props.adminPermissions?.has_database_admin) return 'database'
  if (tab === 'user' && props.adminPermissions?.has_user_admin) return 'user'
  if (tab === 'permission' && props.adminPermissions?.has_permission_admin) return 'permission'
  return 'table'
}

function handleMenuSelect(index: string) {
  activeAdminTab.value = index as TabType
  router.replace({
    path: route.path,
    query: { ...route.query, admin_tab: index }
  })
}

watch(
  () => [route.query.admin_tab, props.adminPermissions] as const,
  () => {
    activeAdminTab.value = resolveTab()
  },
  { immediate: true }
)
</script>

<style scoped>
.database-admin-tabs {
  display: flex;
  gap: var(--spacing-lg);
  padding: var(--spacing-md) 0;
  min-height: 360px;
}

.admin-sidebar {
  flex-shrink: 0;
  width: 220px;
  background-color: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-large);
  padding: var(--spacing-sm) 0;
  height: fit-content;
  position: sticky;
  top: var(--spacing-md);
}

.admin-menu {
  border: none;
}

.admin-menu .el-menu-item {
  height: auto;
  min-height: 44px;
  padding: var(--spacing-sm) var(--spacing-md);
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 2px;
}

.admin-menu .el-menu-item .menu-desc {
  font-size: 11px;
  color: var(--color-text-tertiary);
  font-weight: normal;
  margin-left: 28px;
}

.menu-group-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-tertiary);
  letter-spacing: 0.5px;
  padding: var(--spacing-md) var(--spacing-md) var(--spacing-xs);
  margin-top: var(--spacing-xs);
}

.menu-group-label:first-child {
  margin-top: 0;
}

.admin-content {
  flex: 1;
  min-width: 0;
  background-color: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-large);
  padding: 0;
}
</style>
