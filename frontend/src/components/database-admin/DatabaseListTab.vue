<template>
  <div class="database-list-tab">
    <!-- 左侧：数据库网格 + 当前库下表 -->
    <div class="main-column">
      <!-- 数据库网格 -->
      <div class="databases-grid" v-loading="loadingDatabases">
        <!-- 数据库卡片 -->
        <div
          v-for="db in databases"
          :key="db"
          class="database-card"
          :class="{ selected: currentDatabase === db }"
          @click="handleSelectDatabase(db)"
        >
          <div class="card-icon">
            <el-icon :size="32"><Coin /></el-icon>
          </div>
          <div class="card-content">
            <span class="db-name">{{ db }}</span>
            <span class="db-status" v-if="currentDatabase === db">
              <el-icon :size="12"><Select /></el-icon>
              {{ t('dbAdmin.currentSelected') }}
            </span>
          </div>
          <div class="card-actions">
            <el-button 
              type="primary" 
              size="small"
              @click.stop="handleOpenQuery(db)"
            >
              {{ t('nav.query') }}
            </el-button>
          </div>
        </div>
        
        <!-- 空状态 -->
        <div class="empty-databases" v-if="!loadingDatabases && databases.length === 0">
          <el-empty :description="t('dbAdmin.noDatabases')" />
        </div>
      </div>

      <!-- 当前数据库下的表（已选库时展示） -->
      <div class="current-db-tables" v-if="currentDatabase">
        <div class="tables-section-header">
          <div class="tables-section-left">
            <el-icon :size="18"><Grid /></el-icon>
            <span class="tables-section-title">{{ t('dbAdmin.tablesOfDb') }}{{ currentDatabase }}</span>
          </div>
          <el-button link type="primary" size="small" class="new-db-entry" @click="handleGoToCreateDatabase">
            <el-icon :size="14"><Plus /></el-icon>
            {{ t('dbAdmin.createDb') }}
          </el-button>
        </div>
        <div class="tables-block" v-loading="loadingTables">
          <template v-if="tables.length > 0">
            <el-table :data="tables" stripe size="small" class="tables-table">
              <el-table-column prop="name" :label="t('dbAdmin.tableName')" min-width="140" show-overflow-tooltip />
              <el-table-column prop="rows" :label="t('sidebar.rows')" width="100">
                <template #default="{ row }">
                  {{ row.rows !== undefined ? formatNumber(row.rows) : '-' }}
                </template>
              </el-table-column>
              <el-table-column :label="t('common.actions')" width="220" fixed="right">
                <template #default="{ row }">
                  <el-button link type="primary" size="small" @click="handleTableQuery(row.name)">
                    {{ t('nav.query') }}
                  </el-button>
                  <el-button link type="primary" size="small" @click="handleTableSchema(row.name)">
                    {{ t('database.tableSchema') }}
                  </el-button>
                  <el-button link type="primary" size="small" @click="handleTableRowData(row.name)">
                    {{ t('database.tableData') }}
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </template>
          <div v-else-if="!loadingTables" class="tables-empty">
            <el-empty :description="t('dbAdmin.noTables')">
              <template #extra>
                <el-button type="primary" size="small" @click="handleGoToCreateTable">{{ t('dbAdmin.goCreateTable') }}</el-button>
              </template>
            </el-empty>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 连接信息卡片 -->
    <div class="connection-info-card">
      <div class="info-header">
        <el-icon :size="20"><Connection /></el-icon>
        <span class="info-title">{{ t('dbAdmin.connInfo') }}</span>
      </div>
      <div class="info-list">
        <div class="info-item">
          <span class="info-label">{{ t('dbAdmin.connName') }}</span>
          <span class="info-value">{{ currentConnection?.name }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('dbAdmin.dbTypeLabel') }}</span>
          <span class="info-value">
            <el-tag size="small" :type="getDbTypeTag(currentConnection?.db_type || '')">
              {{ currentConnection?.db_type }}
            </el-tag>
          </span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('dbAdmin.hostAddress') }}</span>
          <span class="info-value code">{{ currentConnection?.db_config?.host }}:{{ currentConnection?.db_config?.port }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">{{ t('dbAdmin.dbCount') }}</span>
          <span class="info-value">{{ databases.length }} {{ t('dbAdmin.unit') }}</span>
        </div>
        <div class="info-item" v-if="currentDatabase">
          <span class="info-label">{{ t('dbAdmin.currentDatabase') }}</span>
          <span class="info-value highlight">{{ currentDatabase }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useConnectionStore } from '@/store/connection'
import { useUrlState } from '@/composables/useUrlState'
import { 
  Coin, 
  Connection, 
  Grid,
  Plus,
  Select
} from '@element-plus/icons-vue'

const { t } = useI18n()
const router = useRouter()
const connectionStore = useConnectionStore()
const { updateUrlParams } = useUrlState()

// 计算属性
const currentConnection = computed(() => connectionStore.currentConnection)
const databases = computed(() => connectionStore.databases)
const currentDatabase = computed(() => connectionStore.currentDatabase)
const loadingDatabases = computed(() => connectionStore.loadingDatabases)
const tables = computed(() => connectionStore.tables)
const loadingTables = computed(() => connectionStore.loadingTables)

function formatNumber(num: number): string {
  if (num >= 1000000) return (num / 1000000).toFixed(1) + 'M'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'K'
  return String(num)
}

// 获取数据库类型标签颜色
function getDbTypeTag(dbType: string): string {
  const typeMap: Record<string, string> = {
    mysql: '',
    postgresql: 'success',
    sqlite: 'info',
    mariadb: 'warning'
  }
  return typeMap[dbType] || ''
}

// 选择数据库
async function handleSelectDatabase(db: string) {
  await connectionStore.setCurrentDatabase(db)
  updateUrlParams({ database: db })
}

// 打开指定数据库的查询编辑器
function handleOpenQuery(db: string) {
  connectionStore.setCurrentDatabase(db)
  router.push({
    path: '/query',
    query: {
      connection_id: currentConnection.value?.id ? String(currentConnection.value.id) : undefined,
      database: db
    }
  })
}

// 当前库下表的快捷操作：查询
function handleTableQuery(tableName: string) {
  if (!currentConnection.value || !currentDatabase.value) return
  router.push({
    path: '/query',
    query: {
      connection_id: String(currentConnection.value.id),
      database: currentDatabase.value,
      table: tableName
    }
  })
}

// 当前库下表的快捷操作：表结构
function handleTableSchema(tableName: string) {
  if (!currentConnection.value || !currentDatabase.value) return
  updateUrlParams({ table: tableName })
  router.push({
    path: '/schema',
    query: {
      connection_id: String(currentConnection.value.id),
      database: currentDatabase.value,
      table: tableName
    }
  })
}

// 当前库下表的快捷操作：表数据
function handleTableRowData(tableName: string) {
  if (!currentConnection.value || !currentDatabase.value) return
  updateUrlParams({ table: tableName })
  router.push({
    path: '/rowdata',
    query: {
      connection_id: String(currentConnection.value.id),
      database: currentDatabase.value,
      table: tableName
    }
  })
}

// 去建表（跳转到管理 -> 表管理，并带上当前库）
function handleGoToCreateTable() {
  if (!currentConnection.value || !currentDatabase.value) return
  router.push({
    path: '/database',
    query: {
      connection_id: String(currentConnection.value.id),
      database: currentDatabase.value,
      admin_tab: 'table'
    }
  })
}

// 新建数据库（跳转到管理 -> 数据库管理）
function handleGoToCreateDatabase() {
  if (!currentConnection.value) return
  router.push({
    path: '/database',
    query: {
      connection_id: String(currentConnection.value.id),
      admin_tab: 'database'
    }
  })
}
</script>

<style scoped>
.database-list-tab {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: var(--spacing-lg);
}

.main-column {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
  min-width: 0;
}

/* 数据库网格 */
.databases-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: var(--spacing-md);
  align-content: start;
  min-height: 300px;
}

/* 数据库卡片 */
.database-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background-color: var(--color-background);
  border: 2px solid var(--color-border);
  border-radius: var(--border-radius-large);
  cursor: pointer;
  transition: all 0.2s;
}

.database-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
}

.database-card.selected {
  border-color: var(--color-primary);
  background-color: var(--color-nav-active-bg);
}

.card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 56px;
  height: 56px;
  background-color: var(--color-background-secondary);
  border-radius: var(--border-radius-medium);
  color: var(--color-primary);
  flex-shrink: 0;
}

.database-card.selected .card-icon {
  background-color: var(--color-primary);
  color: white;
}

.card-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.db-name {
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.db-status {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: var(--color-primary);
  margin-top: 4px;
}

.card-actions {
  flex-shrink: 0;
}

/* 空数据库列表 */
.empty-databases {
  grid-column: 1 / -1;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
}

/* 当前库下的表 */
.current-db-tables {
  background-color: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-large);
  padding: var(--spacing-md);
  box-shadow: var(--shadow-sm);
}

.tables-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--color-border);
}

.tables-section-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  min-width: 0;
}

.tables-section-title {
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

.new-db-entry {
  flex-shrink: 0;
}

.tables-block {
  min-height: 120px;
}

.tables-table {
  width: 100%;
}

.tables-empty {
  padding: var(--spacing-lg);
  text-align: center;
}

/* 连接信息卡片 */
.connection-info-card {
  background-color: var(--color-background);
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-sm);
  padding: var(--spacing-lg);
  height: fit-content;
  position: sticky;
  top: var(--spacing-lg);
}

.info-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding-bottom: var(--spacing-md);
  border-bottom: 1px solid var(--color-border);
  margin-bottom: var(--spacing-md);
}

.info-title {
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: var(--spacing-xs) 0;
}

.info-label {
  font-size: 13px;
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.info-value {
  font-size: 13px;
  color: var(--color-text-primary);
  text-align: right;
  word-break: break-all;
}

.info-value.code {
  font-family: var(--font-family-code);
  font-size: 12px;
}

.info-value.highlight {
  color: var(--color-primary);
  font-weight: 600;
}

/* 响应式 */
@media (max-width: 1024px) {
  .database-list-tab {
    grid-template-columns: 1fr;
  }
  
  .connection-info-card {
    position: static;
  }
}

@media (max-width: 768px) {
  .databases-grid {
    grid-template-columns: 1fr;
  }
}
</style>
