<template>
  <div class="backup-management">
    <!-- 当前操作库 -->
    <div class="current-db-bar" v-if="databaseOptions.length > 0">
      <span class="current-db-label">{{ t('dbAdmin.currentDb') }}</span>
      <el-select
        v-model="selectedDatabase"
        :placeholder="t('backup.selectDatabase')"
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
        <template #title>{{ t('dbAdmin.noAvailableDatabases') }}</template>
        {{ t('dbAdmin.selectDbFromList') }}
      </el-alert>
    </div>

    <div v-if="databaseOptions.length > 0 && !selectedDatabase" class="select-db-hint">
      <span>{{ t('dbAdmin.selectDbToBackup') }}</span>
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
        <el-tab-pane :label="t('backup.runningTasks')" name="tasks">
          <BackupTaskList
            :refresh-trigger="refreshTrigger"
            @download="handleTaskDownload"
          />
        </el-tab-pane>
        <el-tab-pane :label="t('backup.backupFiles')" name="files">
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
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()
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
      ElMessage.success(t('backup.backupTaskCreated'))
      refreshTrigger.value++
    } else if (res.backup_id && res.download_url) {
      ElMessage.success(t('backup.backupComplete'))
      refreshTrigger.value++
      window.open(res.download_url.startsWith('/') ? res.download_url : `/api${res.download_url}`, '_blank')
    } else {
      ElMessage.success(t('backup.backupSubmitted'))
      refreshTrigger.value++
    }
  } catch (e: any) {
    ElMessage.error(e?.message || t('backup.createBackupFailed'))
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
    ElMessage.warning(t('backup.noDownloadUrl'))
  }
}

function handleRestore(item: BackupFileInfo) {
  ElMessage.info(t('backup.restoreTodo'))
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
