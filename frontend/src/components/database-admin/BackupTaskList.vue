<template>
  <el-table
    v-loading="loading"
    :data="backupTasks"
    stripe
    border
    style="width: 100%"
  >
    <el-table-column prop="id" label="任务 ID" width="90" />
    <el-table-column label="数据库" min-width="120">
      <template #default="{ row }">
        {{ getTaskDatabase(row) }}
      </template>
    </el-table-column>
    <el-table-column label="备份范围" width="100">
      <template #default="{ row }">
        {{ getTaskScope(row) }}
      </template>
    </el-table-column>
    <el-table-column prop="status" label="状态" width="100">
      <template #default="{ row }">
        <el-tag :type="statusTagType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="进度" width="100">
      <template #default="{ row }">
        {{ row.progress }}%
      </template>
    </el-table-column>
    <el-table-column prop="created_at" label="创建时间" width="180" />
    <el-table-column label="操作" width="140" fixed="right">
      <template #default="{ row }">
        <el-button v-if="row.status === 'running' || row.status === 'pending'" size="small" @click="handleCancel(row)">
          取消
        </el-button>
        <el-button v-if="row.status === 'success'" size="small" type="primary" link @click="handleDownload(row)">
          下载
        </el-button>
      </template>
    </el-table-column>
    <template #empty>
      <el-empty v-if="!loading" description="暂无备份任务" />
    </template>
  </el-table>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getTasks, getTask, cancelTask, type TaskInfo } from '@/api/backup'

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
      return `选表(${p.tables.length})`
    }
    return '全库'
  } catch {
    return '全库'
  }
}

function statusText(status: string): string {
  const map: Record<string, string> = {
    pending: '等待中',
    running: '进行中',
    success: '成功',
    failed: '失败',
    canceled: '已取消'
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
    ElMessage.success('已取消')
    await loadTasks()
  } catch (e: any) {
    ElMessage.error(e?.message || '取消失败')
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
