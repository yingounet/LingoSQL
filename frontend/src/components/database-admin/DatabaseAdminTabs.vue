<template>
  <div class="database-admin-tabs">
    <!-- 左侧竖栏导航 -->
    <aside class="admin-sidebar">
      <el-menu
        :default-active="activeAdminTab"
        class="admin-menu"
        @select="handleMenuSelect"
      >
        <div class="menu-group-label">库与表</div>
        <el-menu-item index="database">
          <el-icon><Folder /></el-icon>
          <span>数据库管理</span>
          <span class="menu-desc">创建、删除数据库</span>
        </el-menu-item>
        <el-menu-item index="table">
          <el-icon><Grid /></el-icon>
          <span>表管理</span>
          <span class="menu-desc">建表、删表</span>
        </el-menu-item>
        <div class="menu-group-label">用户与权限</div>
        <el-menu-item index="user">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
          <span class="menu-desc">创建用户、改密、删用户</span>
        </el-menu-item>
        <el-menu-item index="permission">
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
      <UserManagement v-show="activeAdminTab === 'user'" />
      <PermissionManagement v-show="activeAdminTab === 'permission'" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Folder, Grid, User, Key } from '@element-plus/icons-vue'
import DatabaseManagement from './DatabaseManagement.vue'
import TableManagement from './TableManagement.vue'
import UserManagement from './UserManagement.vue'
import PermissionManagement from './PermissionManagement.vue'

const route = useRoute()
const activeAdminTab = ref<'database' | 'table' | 'user' | 'permission'>('database')

function handleMenuSelect(index: string) {
  activeAdminTab.value = index as typeof activeAdminTab.value
}

// 从 URL admin_tab 恢复（如「去建表」跳转）
watch(
  () => route.query.admin_tab,
  (tab) => {
    if (tab === 'table' || tab === 'database' || tab === 'user' || tab === 'permission') {
      activeAdminTab.value = tab
    }
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
