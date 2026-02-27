<template>
  <div class="database-page">
    <!-- 页面标题 -->
    <PageHeader 
      :title="pageTitle" 
      :description="pageDescription"
    >
      <template #actions>
        <el-button @click="handleRefresh" :loading="loadingDatabases">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-tooltip content="请先选择一个数据库" :disabled="!!currentDatabase" placement="bottom">
          <span class="header-action-wrap">
            <el-button
              :disabled="!currentDatabase"
              @click="handleGoToSchema"
            >
              <el-icon><List /></el-icon>
              表结构
            </el-button>
          </span>
        </el-tooltip>
        <el-tooltip content="请先选择一个数据库" :disabled="!!currentDatabase" placement="bottom">
          <span class="header-action-wrap">
            <el-button
              :disabled="!currentDatabase"
              @click="handleGoToRowData"
            >
              <el-icon><Grid /></el-icon>
              表数据
            </el-button>
          </span>
        </el-tooltip>
        <el-tooltip content="请先选择一个数据库" :disabled="!!currentDatabase" placement="bottom">
          <span class="header-action-wrap">
            <el-button type="primary" @click="handleGoToQuery" :disabled="!currentDatabase">
              <el-icon><EditPen /></el-icon>
              打开查询编辑器
            </el-button>
          </span>
        </el-tooltip>
      </template>
    </PageHeader>
    
    <!-- 无连接提示 -->
    <div class="empty-state" v-if="!currentConnection">
      <el-empty description="请先选择一个数据库连接">
        <el-button type="primary" @click="handleGoToConnections">
          前往连接管理
        </el-button>
      </el-empty>
    </div>
    
    <!-- Tab切换器 -->
    <div class="database-content" v-else>
      <el-tabs v-model="activeTab" @tab-change="handleTabChange" :key="`tabs-${hasAdminPermission}`">
        <!-- 数据库列表Tab -->
        <el-tab-pane label="数据库列表" name="list">
          <DatabaseListTab />
        </el-tab-pane>
        
        <!-- 管理Tab（仅在有权限时显示） -->
        <el-tab-pane 
          v-if="hasAdminPermission" 
          label="管理" 
          name="admin"
        >
          <DatabaseAdminTabs />
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useConnectionStore } from '@/store/connection'
import { useUrlState } from '@/composables/useUrlState'
import { checkAdminPermissions } from '@/api/databaseAdmin'
import PageHeader from '@/components/layout/PageHeader.vue'
import DatabaseListTab from '@/components/database-admin/DatabaseListTab.vue'
import DatabaseAdminTabs from '@/components/database-admin/DatabaseAdminTabs.vue'
import { 
  Refresh, 
  EditPen,
  List,
  Grid
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const connectionStore = useConnectionStore()
const { updateUrlParams } = useUrlState()

// Tab状态
const activeTab = ref('list')

// 权限状态
const hasAdminPermission = ref(false)
const checkingPermission = ref(false)

// 计算属性
const currentConnection = computed(() => connectionStore.currentConnection)
const databases = computed(() => connectionStore.databases)
const currentDatabase = computed(() => connectionStore.currentDatabase)
const loadingDatabases = computed(() => connectionStore.loadingDatabases)

const pageTitle = computed(() => {
  if (!currentConnection.value) return '数据库'
  return `${currentConnection.value.name} - 数据库`
})

const pageDescription = computed(() => {
  if (!currentConnection.value) return '请选择一个数据库连接'
  return `展示 ${currentConnection.value.name} 服务器上的所有数据库`
})

// 检查管理权限
async function checkAdminPermission() {
  if (!currentConnection.value) {
    hasAdminPermission.value = false
    return
  }
  
  checkingPermission.value = true
  try {
    const permissions = await checkAdminPermissions(currentConnection.value.id)
    console.log('权限检查结果:', permissions)
    console.log('权限详情:', {
      has_database_admin: permissions.has_database_admin,
      has_user_admin: permissions.has_user_admin,
      has_permission_admin: permissions.has_permission_admin
    })
    
    const hasAnyPermission = 
      permissions.has_database_admin || 
      permissions.has_user_admin || 
      permissions.has_permission_admin
    
    hasAdminPermission.value = hasAnyPermission
    console.log('hasAdminPermission设置为:', hasAdminPermission.value)
  } catch (error: any) {
    // 权限检查失败，默认不显示管理Tab
    hasAdminPermission.value = false
    console.error('权限检查失败:', error)
    console.error('错误详情:', error.response?.data || error.message)
  } finally {
    checkingPermission.value = false
  }
}

// Tab切换
function handleTabChange(tabName: string) {
  if (tabName === 'admin') {
    // 切换到管理Tab时重新检查权限
    checkAdminPermission()
  }
}

// 刷新数据库列表
async function handleRefresh() {
  if (currentConnection.value) {
    await connectionStore.fetchDatabases(currentConnection.value.id)
    // 刷新时也检查权限
    await checkAdminPermission()
  }
}

// 打开查询编辑器（使用当前选中的数据库）
function handleGoToQuery() {
  router.push({
    path: '/query',
    query: {
      connection_id: currentConnection.value?.id ? String(currentConnection.value.id) : undefined,
      database: currentDatabase.value || undefined
    }
  })
}

// 打开表结构页（当前库）
function handleGoToSchema() {
  if (!currentConnection.value || !currentDatabase.value) return
  updateUrlParams({})
  router.push({
    path: '/schema',
    query: {
      connection_id: String(currentConnection.value.id),
      database: currentDatabase.value
    }
  })
}

// 打开表数据页（当前库）
function handleGoToRowData() {
  if (!currentConnection.value || !currentDatabase.value) return
  updateUrlParams({})
  router.push({
    path: '/rowdata',
    query: {
      connection_id: String(currentConnection.value.id),
      database: currentDatabase.value
    }
  })
}

// 前往连接管理
function handleGoToConnections() {
  router.push('/')
}

// 从 URL 恢复状态
async function restoreFromUrl() {
  const connectionId = route.query.connection_id
  if (connectionId && !currentConnection.value) {
    await connectionStore.restoreState(Number(connectionId))
  }
  // 恢复状态后检查权限
  if (currentConnection.value) {
    await checkAdminPermission()
  }
}

// 监听连接变化
watch(() => currentConnection.value, async (newConn) => {
  if (newConn) {
    // 等待一下确保连接完全加载
    await new Promise(resolve => setTimeout(resolve, 100))
    await checkAdminPermission()
  } else {
    hasAdminPermission.value = false
  }
})

// 监听路由变化
watch(() => route.query.connection_id, async (newId) => {
  if (newId && (!currentConnection.value || currentConnection.value.id !== Number(newId))) {
    await connectionStore.restoreState(Number(newId))
    // 连接恢复后检查权限
    if (currentConnection.value) {
      await checkAdminPermission()
    }
  }
})

// 从列表 Tab「去建表」跳转时自动切换到管理 Tab
watch(() => route.query.admin_tab, (tab) => {
  if ((tab === 'table' || tab === 'database') && hasAdminPermission.value) {
    activeTab.value = 'admin'
  }
}, { immediate: true })

onMounted(async () => {
  await restoreFromUrl()
  // 如果已经有连接，再次检查权限（确保）
  if (currentConnection.value) {
    await checkAdminPermission()
  }
})
</script>

<style scoped>
.database-page {
  max-width: 1400px;
  margin: 0 auto;
}

/* 空状态 */
.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background-color: var(--color-background);
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-sm);
}

/* 数据库内容区 */
.database-content {
  padding: var(--spacing-md) 0;
}

/* 用于 tooltip 包裹禁用按钮时仍能显示提示 */
.header-action-wrap {
  display: inline-flex;
}
</style>
