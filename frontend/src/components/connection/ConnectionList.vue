<template>
  <div class="connection-list">
    <!-- 页面头部（非 embedded 模式显示） -->
    <div v-if="!embedded" class="list-header">
      <div class="header-left">
        <h2 class="page-title">数据库连接</h2>
        <p class="page-subtitle">管理您的数据库连接配置</p>
      </div>
      <el-button type="primary" @click="$emit('new')">
        <el-icon><Plus /></el-icon>
        新建连接
      </el-button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="filter-tabs">
        <el-button
          :type="!filter.db_type ? 'primary' : 'default'"
          :class="{ active: !filter.db_type }"
          @click="handleFilterType(null)"
        >
          全部
        </el-button>
        <el-button
          v-for="(config, type) in DB_TYPE_CONFIG"
          :key="type"
          :type="filter.db_type === type ? 'primary' : 'default'"
          :class="{ active: filter.db_type === type }"
          @click="handleFilterType(type as DbType)"
        >
          {{ config.label }}
        </el-button>
      </div>
      <el-input
        v-model="searchInput"
        placeholder="搜索连接..."
        clearable
        class="search-input"
        @input="handleSearch"
        @clear="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
    </div>

    <!-- 连接列表表格 -->
    <el-card class="list-card" v-loading="loading">
      <el-table
        :data="connections"
        style="width: 100%"
        :row-class-name="getRowClass"
        @row-click="handleRowClick"
      >
        <!-- 状态指示器 + 名称 -->
        <el-table-column label="连接名称" min-width="200">
          <template #default="{ row }">
            <div class="connection-name">
              <span
                class="status-dot"
                :class="{ active: isConnected(row.id), default: row.is_default }"
              ></span>
              <span class="name-text">{{ row.name }}</span>
              <el-tag v-if="row.is_default" size="small" type="info" class="default-tag">
                默认
              </el-tag>
            </div>
          </template>
        </el-table-column>

        <!-- 数据库类型 -->
        <el-table-column label="类型" width="130">
          <template #default="{ row }">
            <el-tag
              :style="{
                backgroundColor: DB_TYPE_CONFIG[row.db_type].color + '20',
                color: DB_TYPE_CONFIG[row.db_type].color,
                borderColor: DB_TYPE_CONFIG[row.db_type].color + '40',
              }"
              size="small"
            >
              {{ DB_TYPE_CONFIG[row.db_type].label }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- 连接方式 -->
        <el-table-column label="连接方式" width="120">
          <template #default="{ row }">
            <span class="connection-type">
              <el-icon v-if="row.connection_type === 'ssh_tunnel'" class="ssh-icon">
                <Lock />
              </el-icon>
              {{ row.connection_type === 'direct' ? '直连' : 'SSH 隧道' }}
            </span>
          </template>
        </el-table-column>

        <!-- 主机地址 -->
        <el-table-column label="主机地址" min-width="180">
          <template #default="{ row }">
            <span class="host-address">
              {{ row.db_config.host }}:{{ row.db_config.port }}
            </span>
          </template>
        </el-table-column>

        <!-- 最后连接时间 -->
        <el-table-column label="最后连接" width="180">
          <template #default="{ row }">
            <div class="last-connected">
              {{ row.last_used_at ? formatDateTime(row.last_used_at) : '从未连接' }}
            </div>
          </template>
        </el-table-column>

        <!-- 操作按钮 -->
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <div class="action-buttons">
              <el-button
                type="primary"
                link
                @click.stop="handleConnect(row)"
              >
                连接
              </el-button>
              <el-button
                type="primary"
                link
                @click.stop="$emit('edit', row.id)"
              >
                编辑
              </el-button>
              <el-button
                type="danger"
                link
                @click.stop="handleDelete(row)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空状态 -->
      <div v-if="!loading && connections.length === 0" class="empty-state">
        <el-empty description="暂无连接">
          <el-button type="primary" @click="$emit('new')">创建第一个连接</el-button>
        </el-empty>
      </div>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="pagination.total > 0">
        <span class="pagination-info">
          显示 {{ (pagination.page - 1) * pagination.pageSize + 1 }} -
          {{ Math.min(pagination.page * pagination.pageSize, pagination.total) }}
          共 {{ pagination.total }} 条
        </span>
        <el-pagination
          v-model:current-page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          layout="prev, pager, next"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Plus, Search, Lock } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useConnectionStore } from '@/store/connection'
import { DB_TYPE_CONFIG, type DbType, type Connection } from '@/types/connection'

// Props
const props = withDefaults(
  defineProps<{ embedded?: boolean }>(),
  { embedded: false }
)

// Emits
const emit = defineEmits<{
  (e: 'new'): void
  (e: 'edit', id: number): void
  (e: 'connect', connection: Connection): void
}>()

// Store
const store = useConnectionStore()

// 本地状态
const searchInput = ref('')
const searchTimeout = ref<number | null>(null)

// 计算属性
const connections = computed(() => store.connections)
const loading = computed(() => store.loading)
const filter = computed(() => store.filter)
const pagination = computed(() => store.pagination)

// 格式化日期时间
function formatDateTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

// 判断是否已连接（使用 store 的 currentConnection，跨页面保持状态）
function isConnected(id: number): boolean {
  return store.currentConnection?.id === id
}

// 获取行样式
function getRowClass({ row }: { row: Connection }): string {
  if (isConnected(row.id)) return 'connected-row'
  return ''
}

// 处理类型筛选
function handleFilterType(type: DbType | null) {
  store.setFilter({ db_type: type })
}

// 处理搜索（防抖）
function handleSearch() {
  if (searchTimeout.value) {
    clearTimeout(searchTimeout.value)
  }
  searchTimeout.value = window.setTimeout(() => {
    store.setFilter({ search: searchInput.value })
  }, 300)
}

// 处理分页
function handlePageChange(page: number) {
  store.setPage(page)
}

// 处理行点击
function handleRowClick(row: Connection) {
  // 可以在这里添加行点击逻辑
}

// 处理连接
async function handleConnect(connection: Connection) {
  try {
    ElMessage.info('正在连接...')
    await store.connectTo(connection.id)
    ElMessage.success(`已连接到 ${connection.name}`)
    emit('connect', connection)
  } catch (error: any) {
    ElMessage.error(error.message || '连接失败')
  }
}

// 处理删除
async function handleDelete(connection: Connection) {
  try {
    await ElMessageBox.confirm(
      `确定要删除连接 "${connection.name}" 吗？此操作不可撤销。`,
      '删除确认',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )

    await store.deleteConnection(connection.id)

    // 如果删除的是当前连接，清除连接状态
    if (store.currentConnection?.id === connection.id) {
      store.disconnect()
    }

    ElMessage.success('连接已删除')
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  }
}

// 同步搜索输入
watch(() => filter.value.search, (newVal) => {
  if (searchInput.value !== newVal) {
    searchInput.value = newVal
  }
})

// 初始化加载
onMounted(() => {
  store.fetchConnections()
})
</script>

<style scoped>
.connection-list {
  width: 100%;
}

/* 页面头部 */
.list-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--spacing-lg);
}

.header-left {
  flex: 1;
}

.page-title {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: var(--font-size-h2);
  font-weight: 600;
  color: var(--color-text-primary);
}

.page-subtitle {
  margin: 0;
  font-size: var(--font-size-body);
  color: var(--color-text-secondary);
}

/* 筛选栏 */
.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
  gap: var(--spacing-md);
}

.filter-tabs {
  display: flex;
  gap: var(--spacing-sm);
}

.filter-tabs .el-button {
  border-radius: var(--border-radius-medium);
}

.filter-tabs .el-button.active {
  font-weight: 600;
}

.search-input {
  width: 280px;
}

/* 列表卡片 */
.list-card {
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
}

.list-card :deep(.el-card__body) {
  padding: 0;
}

/* 表格样式 */
.list-card :deep(.el-table) {
  --el-table-border-color: var(--color-border);
  --el-table-header-bg-color: var(--color-background-secondary);
}

.list-card :deep(.el-table th.el-table__cell) {
  font-weight: 600;
  color: var(--color-text-primary);
  background-color: var(--color-background-secondary);
}

.list-card :deep(.el-table td.el-table__cell) {
  padding: 16px 12px;
}

.list-card :deep(.el-table tr.connected-row) {
  background-color: var(--color-nav-active-bg);
}

/* 连接名称 */
.connection-name {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.status-dot.active {
  background-color: var(--color-success);
}

.status-dot.default {
  border: 2px solid var(--color-primary);
  background-color: transparent;
}

.status-dot.active.default {
  background-color: var(--color-success);
  border-color: var(--color-success);
}

.name-text {
  font-weight: 500;
  color: var(--color-text-primary);
}

.default-tag {
  margin-left: var(--spacing-xs);
}

/* 连接类型 */
.connection-type {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  color: var(--color-text-secondary);
  font-size: var(--font-size-small);
}

.ssh-icon {
  color: var(--color-warning);
}

/* 主机地址 */
.host-address {
  font-family: var(--font-family-code);
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

/* 最后连接时间 */
.last-connected {
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: var(--spacing-sm);
}

/* 空状态 */
.empty-state {
  padding: var(--spacing-xxl) 0;
}

/* 分页 */
.pagination-wrapper {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-md) var(--spacing-lg);
  border-top: 1px solid var(--color-border);
}

.pagination-info {
  font-size: var(--font-size-small);
  color: var(--color-text-tertiary);
}

/* 响应式 */
@media (max-width: 768px) {
  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-tabs {
    flex-wrap: wrap;
  }

  .search-input {
    width: 100%;
  }

  .pagination-wrapper {
    flex-direction: column;
    gap: var(--spacing-md);
  }
}
</style>
