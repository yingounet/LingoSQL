<template>
  <div class="backup-management">
    <!-- 当前操作库 -->
    <div class="current-db-bar" v-if="databaseOptions.length > 0">
      <span class="current-db-label">当前操作库：</span>
      <el-select
        v-model="selectedDatabase"
        placeholder="选择数据库"
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
    </div>
    <div v-else class="no-db-tip">
      <el-alert type="info" :closable="false" show-icon>
        <template #title>暂无可用数据库</template>
        请先在「数据库列表」Tab 中选择数据库，或前往「数据库管理」创建数据库。
      </el-alert>
    </div>

    <div v-if="databaseOptions.length > 0 && !selectedDatabase" class="select-db-hint">
      <span>请在上方选择要备份的数据库</span>
    </div>

    <template v-if="databaseOptions.length > 0 && selectedDatabase">
      <BackupCreateForm
        :connection-id="currentConnection!.id"
        :database-options="databaseOptions"
        :selected-database="selectedDatabase"
        :tables="tables"
        :submitting="backupSubmitting"
        @submit="handleCreateBackup"
      />

      <el-tabs v-model="activeListTab" class="backup-tabs">
        <el-tab-pane label="进行中任务" name="tasks">
          <BackupTaskList
            :refresh-trigger="refreshTrigger"
            @download="handleTaskDownload"
          />
        </el-tab-pane>
        <el-tab-pane label="备份文件列表" name="files">
          <BackupFileList
            :connection-id="currentConnection?.id"
            :refresh-trigger="refreshTrigger"
            @restore="handleRestore"
          />
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useConnectionStore } from '@/store/connection'
import { getDatabaseList } from '@/api/databaseAdmin'
import { getTables } from '@/api/table'
import { createBackup, type CreateBackupRequest, type TaskInfo, type BackupFileInfo } from '@/api/backup'
import type { DatabaseInfo } from '@/types/databaseAdmin'
import type { TableInfo } from '@/api/table'
import BackupCreateForm from './BackupCreateForm.vue'
import BackupTaskList from './BackupTaskList.vue'
import BackupFileList from './BackupFileList.vue'

const connectionStore = useConnectionStore()
const currentConnection = computed(() => connectionStore.currentConnection)

const databaseOptions = ref<DatabaseInfo[]>([])
const selectedDatabase = ref<string | null>(null)
const tables = ref<TableInfo[]>([])
const loadingTables = ref(false)
const activeListTab = ref('tasks')
const refreshTrigger = ref(0)
const backupSubmitting = ref(false)

async function loadDatabaseOptions() {
  if (!currentConnection.value) return
  try {
    databaseOptions.value = await getDatabaseList(currentConnection.value.id)
  } catch {
    databaseOptions.value = []
  }
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
  } catch {
    tables.value = []
  } finally {
    loadingTables.value = false
  }
}

async function onDatabaseChange(db: string | null) {
  await connectionStore.setCurrentDatabase(db ?? null)
  await loadTables()
}

async function handleCreateBackup(req: CreateBackupRequest) {
  if (!currentConnection.value) return
  backupSubmitting.value = true
  try {
    const res = await createBackup(req)
    if (res.task_id) {
      ElMessage.success('备份任务已创建，请在下方查看进度')
      refreshTrigger.value++
    } else if (res.backup_id && res.download_url) {
      ElMessage.success('备份完成')
      refreshTrigger.value++
      window.open(res.download_url.startsWith('/') ? res.download_url : `/api${res.download_url}`, '_blank')
    } else {
      ElMessage.success('备份任务已提交')
      refreshTrigger.value++
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '创建备份失败')
  } finally {
    backupSubmitting.value = false
  }
}

function handleTaskDownload(task: TaskInfo) {
  if (task.status !== 'success' || !task.result) return
  try {
    const result = JSON.parse(task.result)
    const url = result?.download_url || result?.download_URL
    if (url) {
      window.open(url.startsWith('/') ? url : `/api${url}`, '_blank')
    }
  } catch {
    ElMessage.warning('暂无下载地址')
  }
}

function handleRestore(item: BackupFileInfo) {
  ElMessage.info('恢复功能需在后端实现后启用')
}

watch(currentConnection, async () => {
  selectedDatabase.value = null
  tables.value = []
  await loadDatabaseOptions()
  if (selectedDatabase.value) await loadTables()
})

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
  if (selectedDatabase.value) await loadTables()
})
</script>

<style scoped>
.backup-management {
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

.backup-tabs {
  margin-top: var(--spacing-md);
}
</style>
