<template>
  <div class="history-page">
    <!-- 页面标题 -->
    <PageHeader 
      title="查询历史" 
      description="查看用户执行和系统执行的 SQL 历史记录"
    />

    <!-- 未选择连接时的提示 -->
    <div v-if="!connectionId" class="empty-state">
      <el-empty description="请先选择连接">
        <template #image>
          <el-icon :size="64" class="empty-icon"><Connection /></el-icon>
        </template>
        <el-button type="primary" @click="handleGoToConnections">
          前往连接管理
        </el-button>
      </el-empty>
    </div>

    <!-- 主内容区域 -->
    <template v-else>
      <!-- 标签页 -->
      <el-card class="content-card">
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <!-- 用户执行标签 -->
          <el-tab-pane label="用户执行" name="user">
            <HistoryList
              :connection-id="connectionId"
              type="user"
              :loading="loadingUser"
              :data="userHistoryList"
              :total="userTotal"
              :page="userPage"
              :page-size="userPageSize"
              @page-change="handleUserPageChange"
            />
          </el-tab-pane>

          <!-- 系统执行标签 -->
          <el-tab-pane label="系统执行" name="system">
            <HistoryList
              :connection-id="connectionId"
              type="system"
              :loading="loadingSystem"
              :data="systemHistoryList"
              :total="systemTotal"
              :page="systemPage"
              :page-size="systemPageSize"
              @page-change="handleSystemPageChange"
            />
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Connection } from '@element-plus/icons-vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import HistoryList from '@/components/history/HistoryList.vue'
import { getUserHistory, getSystemHistory } from '@/api/history'
import type { QueryHistory, QueryHistoryListResponse } from '@/types/history'

const route = useRoute()
const router = useRouter()

// 当前标签页
const activeTab = ref<'user' | 'system'>('user')

// 连接 ID（从 URL 参数获取）
const connectionId = computed(() => {
  const id = route.query.connection_id
  return id ? Number(id) : null
})

// 用户执行历史
const loadingUser = ref(false)
const userHistoryList = ref<QueryHistory[]>([])
const userTotal = ref(0)
const userPage = ref(1)
const userPageSize = ref(50)

// 系统执行历史
const loadingSystem = ref(false)
const systemHistoryList = ref<QueryHistory[]>([])
const systemTotal = ref(0)
const systemPage = ref(1)
const systemPageSize = ref(50)

// 加载用户执行历史
async function loadUserHistory() {
  if (!connectionId.value) return

  loadingUser.value = true
  try {
    const response: QueryHistoryListResponse = await getUserHistory({
      connection_id: connectionId.value,
      page: userPage.value,
      page_size: userPageSize.value,
    })
    userHistoryList.value = response.list
    userTotal.value = response.total
  } catch (error: any) {
    ElMessage.error(error.message || '加载用户执行历史失败')
  } finally {
    loadingUser.value = false
  }
}

// 加载系统执行历史
async function loadSystemHistory() {
  if (!connectionId.value) return

  loadingSystem.value = true
  try {
    const response: QueryHistoryListResponse = await getSystemHistory({
      connection_id: connectionId.value,
      page: systemPage.value,
      page_size: systemPageSize.value,
    })
    systemHistoryList.value = response.list
    systemTotal.value = response.total
  } catch (error: any) {
    ElMessage.error(error.message || '加载系统执行历史失败')
  } finally {
    loadingSystem.value = false
  }
}

// 标签页切换
function handleTabChange(tabName: string) {
  if (tabName === 'user' && userHistoryList.value.length === 0) {
    loadUserHistory()
  } else if (tabName === 'system' && systemHistoryList.value.length === 0) {
    loadSystemHistory()
  }
}

// 用户历史分页变化
function handleUserPageChange(page: number) {
  userPage.value = page
  loadUserHistory()
}

// 系统历史分页变化
function handleSystemPageChange(page: number) {
  systemPage.value = page
  loadSystemHistory()
}


// 前往连接管理
function handleGoToConnections() {
  router.push('/')
}

// 监听 connectionId 变化
watch(connectionId, (newId) => {
  if (newId) {
    if (activeTab.value === 'user') {
      loadUserHistory()
    } else {
      loadSystemHistory()
    }
  }
})

// 初始化加载
onMounted(() => {
  if (connectionId.value && activeTab.value === 'user') {
    loadUserHistory()
  }
})
</script>

<style scoped>
.history-page {
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

.empty-state {
  margin-top: var(--spacing-xl);
}

.empty-icon {
  color: var(--color-text-tertiary);
}

.content-card {
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
  background: var(--color-background);
  margin-top: var(--spacing-lg);
}

.content-card :deep(.el-card__body) {
  padding: var(--spacing-xl);
}

.content-card :deep(.el-tabs__content) {
  padding-top: var(--spacing-lg);
}
</style>
