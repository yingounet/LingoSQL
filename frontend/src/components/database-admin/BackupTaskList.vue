<template>
  <el-table
    v-loading="loading"
    :data="backupTasks"
    stripe
    border
    style="width: 100%"
  >
    <el-table-column prop="id" :label="t('backup.taskId')" width="90" />
    <el-table-column :label="t('backup.backupDatabase')" min-width="120">
      <template #default="{ row }">
        {{ getTaskDatabase(row) }}
      </template>
    </el-table-column>
    <el-table-column :label="t('backup.scope')" width="100">
      <template #default="{ row }">
        {{ getTaskScope(row) }}
      </template>
    </el-table-column>
    <el-table-column prop="status" :label="t('historyList.status')" width="100">
      <template #default="{ row }">
        <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column :label="t('backup.progress')" width="100">
      <template #default="{ row }">
        {{ row.progress }}%
      </template>
    </el-table-column>
    <el-table-column prop="created_at" :label="t('dbAdmin.createdAt')" width="180" />
    <el-table-column :label="t('common.actions')" width="140" fixed="right">
      <template #default="{ row }">
        <el-button v-if="row.status === 'running' || row.status === 'pending'" size="small" @click="handleCancel(row)">
          {{ t('common.cancel') }}
        </el-button>
        <el-button v-if="row.status === 'success'" size="small" type="primary" link @click="handleDownload(row)">
          {{ t('backup.download') }}
        </el-button>
      </template>
    </el-table-column>
    <template #empty>
      <el-empty v-if="!loading" :description="t('backup.noBackupTasks')" />
    </template>
  </el-table>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { getTasks, getTask, cancelTask, type TaskInfo } from '@/api/backup'

const { t } = useI18n()

const props = defineProps<{
  refreshTrigger?: number
}>()

const emit = defineEmits<{
  (e: 'download', task: TaskInfo): void
}>()

const loading = ref(false)
const backupTasks = ref<TaskInfo[]>([])
let pollTimer: ReturnType<typeof setInterval> | null = null

function getTaskDatabase(row: TaskInfo): string {
  try {
    const p = JSON.parse(row.payload || '{}')
    return p.database || '-'
  } catch {
    return '-'
  }
}

function getTaskScope(row: TaskInfo): string {
  try {
    const p = JSON.parse(row.payload || '{}')
    if (p.tables && Array.isArray(p.tables) && p.tables.length > 0) {
      return t('backup.selectedTables', { n: p.tables.length })
    }
    return t('backup.fullDbLabel')
  } catch {
    return t('backup.fullDbLabel')
  }
}

function statusText(status: string): string {
  const map: Record<string, string> = {
    pending: t('backup.statusPending'),
    running: t('backup.statusRunning'),
    success: t('backup.statusSuccess'),
    failed: t('backup.statusFailed'),
    canceled: t('backup.statusCancelled')
  }
  return map[status] || status
}

function statusTagType(status: string): string {
  const map: Record<string, string> = {
    pending: 'info',
    running: 'warning',
    success: 'success',
    failed: 'danger',
    canceled: 'info'
  }
  return map[status] || ''
}

async function loadTasks() {
  loading.value = true
  try {
    const res = await getTasks(1, 50)
    backupTasks.value = (res.list || []).filter((t) => t.type === 'BACKUP_DATABASE')
    startPollingIfNeeded()
  } catch (e: any) {
    backupTasks.value = []
  } finally {
    loading.value = false
  }
}

function startPollingIfNeeded() {
  const hasRunning = backupTasks.value.some((t) => t.status === 'running' || t.status === 'pending')
  if (hasRunning && !pollTimer) {
    pollTimer = setInterval(() => {
      loadTasks()
    }, 4000)
  } else if (!hasRunning && pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

async function handleCancel(row: TaskInfo) {
  if (row.status !== 'running' && row.status !== 'pending') return
  try {
    await cancelTask(row.id)
    ElMessage.success(t('backup.cancelled'))
    await loadTasks()
  } catch (e: any) {
    ElMessage.error(e?.message || t('backup.cancelFailed'))
  }
}

function handleDownload(row: TaskInfo) {
  emit('download', row)
}

watch(
  () => props.refreshTrigger,
  () => {
    loadTasks()
  }
)

onMounted(() => {
  loadTasks()
})

onUnmounted(() => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
})
</script>
