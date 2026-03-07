<template>
  <aside class="sidebar" :class="{ collapsed }">
    <!-- 连接信息卡片 -->
    <div class="connection-card" v-if="currentConnection">
      <div class="card-header">
        <span class="label">ACTIVE CONNECTION</span>
        <button class="collapse-btn" @click="$emit('collapse')">
          <el-icon :size="16">
            <ArrowLeft v-if="!collapsed" />
            <ArrowRight v-else />
          </el-icon>
        </button>
      </div>
      <div class="connection-info">
        <span class="status-dot connected"></span>
        <span class="name" v-if="!collapsed">{{ currentConnection.name }}</span>
        <span class="connection-initial" v-else :title="currentConnection.name">{{ connectionInitial }}</span>
      </div>
      <div class="connection-meta" v-if="!collapsed">
        {{ currentConnection.db_type }} • {{ currentConnection.db_config?.host }}
      </div>
      <!-- 数据库选择器 -->
      <div class="database-selector" v-if="!collapsed">
        <el-select
          v-model="selectedDatabase"
          placeholder="选择数据库"
          size="small"
          :loading="loadingDatabases"
          @change="handleDatabaseChange"
        >
          <el-option
            v-for="db in databases"
            :key="db"
            :label="db"
            :value="db"
          />
        </el-select>
      </div>
    </div>

    <!-- 表列表区域 -->
    <div class="tables-panel" v-if="currentConnection && selectedDatabase && !collapsed">
      <!-- 搜索输入框 -->
      <div class="tables-search">
        <el-input
          v-model="tableSearchKeyword"
          placeholder="搜索表名..."
          size="small"
          clearable
          :prefix-icon="Search"
        />
      </div>
      
      <!-- 表列表 -->
      <div class="tables-list" v-loading="loadingTables">
        <!-- 加载中 -->
        <template v-if="loadingTables">
          <div class="table-skeleton" v-for="i in 5" :key="i">
            <el-skeleton :rows="0" animated />
          </div>
        </template>
        
        <!-- 表列表 -->
        <template v-else-if="filteredTables.length > 0">
          <div
            v-for="table in filteredTables"
            :key="table.name"
            class="table-item"
            :class="{ selected: selectedTable === table.name }"
            @click="handleTableClick(table.name)"
            :title="table.name"
          >
            <el-icon :size="14" class="table-icon"><Grid /></el-icon>
            <span class="table-name">{{ table.name }}</span>
            <span class="table-rows" v-if="table.rows !== undefined">{{ formatNumber(table.rows) }}</span>
          </div>
        </template>
        
        <!-- 空状态 -->
        <div class="tables-empty" v-else-if="!loadingTables && tables.length === 0">
          <el-icon :size="32"><FolderOpened /></el-icon>
          <span>暂无表</span>
        </div>
        
        <!-- 搜索无结果 -->
        <div class="tables-empty" v-else-if="!loadingTables && filteredTables.length === 0">
          <span>未找到匹配的表</span>
        </div>
      </div>
    </div>
    
    <!-- 表详情面板 -->
    <div class="table-info-panel" v-if="selectedTable && !collapsed" v-loading="loadingTableInfo">
      <div class="panel-header">
        <span class="label">TABLE INFO</span>
      </div>
      <div class="table-info-name">{{ selectedTable }}</div>
      
      <template v-if="tableDetailInfo">
        <div class="info-list">
          <div class="info-item">
            <span class="info-label">Rows</span>
            <span class="info-value">{{ formatNumber(tableDetailInfo.rows) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Size</span>
            <span class="info-value">{{ formatSize(tableDetailInfo.data_length + tableDetailInfo.index_length) }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Engine</span>
            <span class="info-value">{{ tableDetailInfo.engine || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Charset</span>
            <span class="info-value">{{ tableDetailInfo.collation || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Auto Inc</span>
            <span class="info-value">{{ tableDetailInfo.auto_increment !== null ? formatNumber(tableDetailInfo.auto_increment) : '-' }}</span>
          </div>
          <div class="info-item" v-if="tableDetailInfo.comment">
            <span class="info-label">Comment</span>
            <span class="info-value comment" :title="tableDetailInfo.comment">{{ tableDetailInfo.comment }}</span>
          </div>
          <div class="info-item" v-if="tableDetailInfo.create_time">
            <span class="info-label">Created</span>
            <span class="info-value">{{ tableDetailInfo.create_time }}</span>
          </div>
        </div>
      </template>
    </div>
    
    <!-- 无连接时显示 -->
    <div class="connection-card empty" v-if="!currentConnection">
      <div class="card-header">
        <span class="label">NO CONNECTION</span>
        <button class="collapse-btn" @click="$emit('collapse')">
          <el-icon :size="16">
            <ArrowLeft v-if="!collapsed" />
            <ArrowRight v-else />
          </el-icon>
        </button>
      </div>
      <div class="connection-info">
        <span class="status-dot disconnected"></span>
        <span class="name" v-if="!collapsed">未选择连接</span>
        <span class="connection-initial no-connection" v-else title="未选择连接">—</span>
      </div>
    </div>
    
    <!-- 快捷操作按钮 -->
    <!-- <div class="sidebar-actions">
      <el-button 
        type="primary" 
        :class="{ 'icon-only': collapsed }"
        @click="handleNewQuery"
      >
        <el-icon><Plus /></el-icon>
        <span v-if="!collapsed">New Query</span>
      </el-button>
      <el-button 
        v-if="!collapsed"
        @click="handleNewConnection"
      >
        <el-icon><Connection /></el-icon>
        <span>New Connection</span>
      </el-button>
    </div> -->
    
    <!-- 存储池指示器（暂无真实数据，暂隐藏） -->
    <div class="storage-indicator" v-if="false && !collapsed">
      <div class="indicator-header">
        <span class="label">STORAGE POOL</span>
        <span class="value">{{ storageUsage }}%</span>
      </div>
      <el-progress 
        :percentage="storageUsage" 
        :show-text="false"
        :stroke-width="6"
      />
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useConnectionStore } from '@/store/connection'
import { useUrlState } from '@/composables/useUrlState'
import { getTableInfo, type TableDetailInfo } from '@/api/table'
import { 
  ArrowLeft, 
  ArrowRight, 
  FolderOpened, 
  Search,
  Grid
} from '@element-plus/icons-vue'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  collapse: []
}>()

const route = useRoute()
const router = useRouter()
const connectionStore = useConnectionStore()
const { updateUrlParams } = useUrlState()

// 当前连接
const currentConnection = computed(() => connectionStore.currentConnection)

// 折叠态下显示连接名首字
const connectionInitial = computed(() => {
  const name = currentConnection.value?.name
  if (!name || !name.trim()) return '?'
  const first = name.trim()[0]
  return /[\u4e00-\u9fa5a-zA-Z0-9]/.test(first) ? first.toUpperCase() : '?'
})

// 数据库列表
const databases = computed(() => connectionStore.databases)
const loadingDatabases = computed(() => connectionStore.loadingDatabases)

// 当前选择的数据库（双向绑定）
const selectedDatabase = computed({
  get: () => connectionStore.currentDatabase,
  set: (value) => connectionStore.setCurrentDatabase(value)
})

// 数据库切换处理
function handleDatabaseChange(database: string) {
  connectionStore.setCurrentDatabase(database)
  // 切换数据库时清空选中的表
  selectedTable.value = null
  tableDetailInfo.value = null
  // 更新 URL 参数
  updateUrlParams({ database, table: null })
}

// 表列表
const tables = computed(() => connectionStore.tables)
const loadingTables = computed(() => connectionStore.loadingTables)

// 表搜索关键词
const tableSearchKeyword = ref('')

// 选中的表名
const selectedTable = ref<string | null>(null)

// 表详情信息
const tableDetailInfo = ref<TableDetailInfo | null>(null)
const loadingTableInfo = ref(false)
let tableInfoDebounceTimer: ReturnType<typeof setTimeout> | null = null

// 过滤后的表列表
const filteredTables = computed(() => {
  if (!tableSearchKeyword.value) {
    return tables.value
  }
  const keyword = tableSearchKeyword.value.toLowerCase()
  return tables.value.filter(table => 
    table.name.toLowerCase().includes(keyword)
  )
})

// 格式化数字
function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

// 格式化文件大小
function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return (bytes / Math.pow(1024, i)).toFixed(1) + ' ' + units[i]
}

// 点击表名
async function handleTableClick(tableName: string) {
  // 如果点击的是同一个表，则取消选中
  if (selectedTable.value === tableName) {
    selectedTable.value = null
    tableDetailInfo.value = null
    if (tableInfoDebounceTimer) {
      clearTimeout(tableInfoDebounceTimer)
      tableInfoDebounceTimer = null
    }
    updateUrlParams({ table: null })
    return
  }

  selectedTable.value = tableName

  // 根据当前页面决定行为
  const currentPath = route.path
  
  if (currentPath === '/schema' || currentPath.startsWith('/schema')) {
    // 在 Schema 页面，更新 URL 参数，页面会自动响应
    updateUrlParams({ table: tableName })
  } else if (currentPath === '/rowdata' || currentPath.startsWith('/rowdata')) {
    // 在 RowData 页面，更新 URL 参数，页面会自动响应
    updateUrlParams({ table: tableName })
  } else if (currentPath === '/query' || currentPath.startsWith('/query')) {
    // 在 Query 页面，保持当前行为，只更新 URL 参数
    updateUrlParams({ table: tableName })
  } else {
    // 在其他页面（如 Dashboard、Connections），跳转到 RowData 页面查看表数据
    router.push({
      path: '/rowdata',
      query: {
        connection_id: currentConnection.value?.id ? String(currentConnection.value.id) : undefined,
        database: selectedDatabase.value || undefined,
        table: tableName,
      }
    })
  }

  // 获取表详情（400ms 防抖，晚于 RowData 的 150ms，错峰减少并发）
  if (!currentConnection.value || !selectedDatabase.value) return
  if (tableInfoDebounceTimer) clearTimeout(tableInfoDebounceTimer)
  tableInfoDebounceTimer = setTimeout(async () => {
    tableInfoDebounceTimer = null
    if (selectedTable.value !== tableName) return
    loadingTableInfo.value = true
    try {
      tableDetailInfo.value = await getTableInfo(
        currentConnection.value!.id,
        selectedDatabase.value!,
        tableName
      )
    } catch (error) {
      console.error('获取表详情失败:', error)
      tableDetailInfo.value = null
    } finally {
      loadingTableInfo.value = false
    }
  }, 400)
}

// 监听数据库变化，清空选中的表
watch(() => connectionStore.currentDatabase, () => {
  selectedTable.value = null
  tableDetailInfo.value = null
})

// 监听连接变化，清空选中的表
watch(() => connectionStore.currentConnection, () => {
  selectedTable.value = null
  tableDetailInfo.value = null
})

// 设置选中的表（供外部调用，用于 URL 状态恢复）
async function setSelectedTable(tableName: string | null) {
  if (!tableName) {
    selectedTable.value = null
    tableDetailInfo.value = null
    return
  }

  // 检查表是否在当前表列表中
  const tableExists = tables.value.some(t => t.name === tableName)
  if (!tableExists) {
    console.warn('表不存在:', tableName)
    return
  }

  selectedTable.value = tableName

  // 获取表详情
  if (!currentConnection.value || !selectedDatabase.value) return

  loadingTableInfo.value = true
  try {
    tableDetailInfo.value = await getTableInfo(
      currentConnection.value.id,
      selectedDatabase.value,
      tableName
    )
  } catch (error) {
    console.error('获取表详情失败:', error)
    tableDetailInfo.value = null
  } finally {
    loadingTableInfo.value = false
  }
}

// 暴露方法给父组件
defineExpose({
  setSelectedTable,
})

// 存储使用率（模拟数据）
const storageUsage = ref(82)
</script>

<style scoped>
.sidebar {
  position: fixed;
  top: 56px;
  left: 0;
  bottom: 32px;
  width: 240px;
  background-color: var(--color-background);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  z-index: 100;
  overflow: hidden;
}

.sidebar.collapsed {
  width: 64px;
}

/* 连接信息卡片 */
.connection-card {
  padding: var(--spacing-md);
  border-bottom: 1px solid var(--color-border);
}

.connection-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
}

.connection-card .label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-tertiary);
  letter-spacing: 0.5px;
}

.collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  border-radius: var(--border-radius-small);
  cursor: pointer;
  color: var(--color-text-tertiary);
  transition: all 0.2s;
}

.collapse-btn:hover {
  background-color: var(--color-background-secondary);
  color: var(--color-text-primary);
}

.connection-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-dot.connected {
  background-color: var(--color-success);
}

.status-dot.disconnected {
  background-color: var(--color-text-tertiary);
}

.connection-info .name {
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.connection-initial {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: var(--border-radius-small);
  background-color: var(--color-background-secondary);
  font-size: 12px;
  font-weight: 600;
  color: var(--color-primary);
}

.connection-initial.no-connection {
  color: var(--color-text-tertiary);
  font-weight: 400;
}

.connection-meta {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin-top: 4px;
  margin-left: 16px;
}

/* 数据库选择器 */
.database-selector {
  margin-top: var(--spacing-sm);
}

.database-selector .el-select {
  width: 100%;
}

/* 表列表区域 */
.tables-panel {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  border-bottom: 1px solid var(--color-border);
}

.tables-search {
  padding: var(--spacing-sm) var(--spacing-md);
  border-bottom: 1px solid var(--color-border);
}

.tables-search .el-input {
  width: 100%;
}

.tables-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-xs) 0;
}

.table-skeleton {
  padding: var(--spacing-xs) var(--spacing-md);
}

.table-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-xs) var(--spacing-md);
  cursor: pointer;
  transition: background-color 0.2s;
}

.table-item:hover {
  background-color: var(--color-background-secondary);
}

.table-icon {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.table-name {
  flex: 1;
  font-size: 13px;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.table-item:hover .table-name {
  color: var(--color-text-primary);
}

.table-item.selected {
  background-color: var(--color-nav-active-bg);
}

.table-item.selected .table-icon {
  color: var(--color-primary);
}

.table-item.selected .table-name {
  color: var(--color-primary);
  font-weight: 500;
}

.table-rows {
  font-size: 11px;
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.tables-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-lg);
  color: var(--color-text-tertiary);
  font-size: 12px;
  gap: var(--spacing-sm);
}

/* 表详情面板 */
.table-info-panel {
  padding: var(--spacing-md);
  border-bottom: 1px solid var(--color-border);
  background-color: var(--color-background);
}

.table-info-panel .panel-header {
  margin-bottom: var(--spacing-sm);
}

.table-info-panel .label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-tertiary);
  letter-spacing: 0.5px;
}

.table-info-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: var(--spacing-sm);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--color-border);
  word-break: break-all;
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  font-size: 12px;
  padding: 2px 0;
}

.info-label {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
  min-width: 70px;
}

.info-value {
  color: var(--color-text-secondary);
  text-align: right;
  word-break: break-all;
}

.info-value.comment {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 快捷操作 */
.sidebar-actions {
  padding: var(--spacing-md);
  border-top: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.sidebar-actions .el-button {
  width: 100%;
  justify-content: center;
}

.sidebar-actions .el-button.icon-only {
  width: 40px;
  padding: 8px;
}

.collapsed .sidebar-actions {
  align-items: center;
}

/* 存储指示器 */
.storage-indicator {
  padding: var(--spacing-md);
  border-top: 1px solid var(--color-border);
}

.indicator-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-sm);
}

.storage-indicator .label {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-tertiary);
  letter-spacing: 0.5px;
}

.storage-indicator .value {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-primary);
}

/* 响应式 */
@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
  }
  
  .sidebar:not(.collapsed) {
    transform: translateX(0);
    box-shadow: var(--shadow-lg);
  }
}
</style>
