<template>
  <div class="table-management">
    <!-- 当前操作库提示 -->
    <div class="current-db-bar" v-if="databaseOptions.length > 0">
      <span class="current-db-label">{{ t('dbAdmin.currentDb') }}</span>
      <el-select
        v-model="selectedDatabase"
        :placeholder="t('sidebar.selectDatabase')"
        clearable
        filterable
        class="current-db-select"
        @change="onDatabaseChange"
      >
        <el-option
          v-for="db in databaseOptions"
          :key="db.name"
          :label="db.name"
          :value="db.name"
        />
      </el-select>
      <el-button
        type="primary"
        :disabled="!selectedDatabase"
        @click="handleCreateTable"
      >
        <el-icon><Plus /></el-icon>
        {{ t('dbAdmin.createTable') }}
      </el-button>
    </div>
    <div v-else class="no-db-tip">
      <el-alert
        type="info"
        :closable="false"
        show-icon
      >
        <template #title>{{ t('dbAdmin.noAvailableDatabases') }}</template>
        {{ t('dbAdmin.selectDbFromList') }}
      </el-alert>
    </div>

    <div v-if="databaseOptions.length > 0 && !selectedDatabase" class="select-db-hint">
      <span>{{ t('dbAdmin.selectDbAbove') }}</span>
    </div>

    <el-table
      v-if="databaseOptions.length > 0 && selectedDatabase"
      v-loading="loadingTables"
      :data="tables"
      stripe
      border
      style="width: 100%"
    >
      <el-table-column prop="name" :label="t('dbAdmin.tableName')" />
      <el-table-column prop="engine" :label="t('sidebar.engine')" v-if="dbType === 'mysql'" />
      <el-table-column prop="rows" :label="t('sidebar.rows')" />
      <el-table-column :label="t('common.actions')" width="120" fixed="right">
        <template #default="{ row }">
          <el-button
            size="small"
            type="danger"
            @click="handleDeleteTable(row)"
          >
            {{ t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
      <template #empty>
        <div v-if="!loadingTables" class="tables-empty-state">
          <el-empty :description="t('dbAdmin.noTablesCreateHint')">
            <el-button type="primary" @click="handleCreateTable">{{ t('dbAdmin.createTable') }}</el-button>
          </el-empty>
        </div>
      </template>
    </el-table>

    <div class="table-footer" v-if="databaseOptions.length > 0 && selectedDatabase">
      <span>{{ t('dbAdmin.totalNTables', { n: tables.length }) }}</span>
    </div>

    <CreateTableDialog
      v-model="showCreateDialog"
      :database="selectedDatabase || ''"
      @success="handleCreateSuccess"
    />
    <TableDeleteConfirm
      v-model="showDeleteDialog"
      :database="selectedDatabase || ''"
      :table-name="tableToDelete"
      @confirm="handleDeleteConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useConnectionStore } from '@/store/connection'
import { getDatabaseList } from '@/api/databaseAdmin'
import { getTables } from '@/api/table'
import { dropTable } from '@/api/tableAdmin'
import type { DatabaseInfo } from '@/types/databaseAdmin'
import type { TableInfo } from '@/api/table'
import CreateTableDialog from './CreateTableDialog.vue'
import TableDeleteConfirm from './TableDeleteConfirm.vue'

const { t } = useI18n()
const connectionStore = useConnectionStore()
const currentConnection = computed(() => connectionStore.currentConnection)
const dbType = computed(() => currentConnection.value?.db_type || 'mysql')

const databaseOptions = ref<DatabaseInfo[]>([])
const selectedDatabase = ref<string | null>(null)
const tables = ref<TableInfo[]>([])
const loadingTables = ref(false)
const showCreateDialog = ref(false)
const showDeleteDialog = ref(false)
const tableToDelete = ref('')

async function loadDatabaseOptions() {
  if (!currentConnection.value) return
  try {
    databaseOptions.value = await getDatabaseList(currentConnection.value.id)
  } catch (_) {
    databaseOptions.value = []
  }
  // 默认使用 store 的当前库，与列表 Tab / 侧边栏一致
  const storeDb = connectionStore.currentDatabase
  if (storeDb && databaseOptions.value.some((d) => d.name === storeDb)) {
    selectedDatabase.value = storeDb
  } else if (databaseOptions.value.length > 0 && !selectedDatabase.value) {
    selectedDatabase.value = databaseOptions.value[0].name
    await connectionStore.setCurrentDatabase(selectedDatabase.value)
  }
}

async function loadTables() {
  if (!currentConnection.value || !selectedDatabase.value) {
    tables.value = []
    return
  }
  loadingTables.value = true
  try {
    const result = await getTables(currentConnection.value.id, selectedDatabase.value)
    tables.value = Array.isArray(result) ? result : []
  } catch (error: any) {
    ElMessage.error(error.message || t('dbAdmin.loadTablesFailed'))
    tables.value = []
  } finally {
    loadingTables.value = false
  }
}

async function onDatabaseChange(db: string | null) {
  await connectionStore.setCurrentDatabase(db ?? null)
  await loadTables()
}

function handleCreateTable() {
  showCreateDialog.value = true
}

async function handleCreateSuccess() {
  await loadTables()
  ElMessage.success(t('dbAdmin.tableCreated'))
}

function handleDeleteTable(row: TableInfo) {
  tableToDelete.value = row.name
  showDeleteDialog.value = true
}

async function handleDeleteConfirm() {
  if (!currentConnection.value || !selectedDatabase.value || !tableToDelete.value) return
  try {
    await dropTable(currentConnection.value.id, selectedDatabase.value, tableToDelete.value)
    ElMessage.success(t('dbAdmin.tableDeleted'))
    await loadTables()
  } catch (error: any) {
    ElMessage.error(error.message || t('dbAdmin.deleteTableFailed'))
  } finally {
    showDeleteDialog.value = false
    tableToDelete.value = ''
  }
}

// 连接变化：清空后重新加载选项并恢复与 store 一致的当前库
watch(currentConnection, async () => {
  selectedDatabase.value = null
  tables.value = []
  await loadDatabaseOptions()
  if (selectedDatabase.value) {
    await loadTables()
  }
})

// 与列表 Tab / 侧边栏同步：当 store 的当前库在外部被修改时更新本页选中项
watch(
  () => connectionStore.currentDatabase,
  (storeDb) => {
    if (storeDb && storeDb !== selectedDatabase.value) {
      selectedDatabase.value = storeDb
      loadTables()
    }
  }
)

onMounted(async () => {
  await loadDatabaseOptions()
  if (selectedDatabase.value) {
    await loadTables()
  }
})
</script>

<style scoped>
.table-management {
  padding: var(--spacing-md);
}

.current-db-bar {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
  flex-wrap: wrap;
}

.current-db-label {
  font-size: 14px;
  color: var(--color-text-secondary);
}

.current-db-select {
  width: 280px;
}

.no-db-tip {
  margin-bottom: var(--spacing-md);
}

.select-db-hint {
  padding: var(--spacing-md);
  color: var(--color-text-tertiary);
  font-size: 13px;
  margin-bottom: var(--spacing-sm);
}

.table-footer {
  margin-top: var(--spacing-md);
  padding: var(--spacing-sm);
  text-align: right;
  color: var(--color-text-secondary);
  font-size: 14px;
}

.tables-empty-state {
  padding: var(--spacing-lg);
}
</style>
