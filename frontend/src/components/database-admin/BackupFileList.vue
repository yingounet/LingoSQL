<template>
  <el-table
    v-loading="loading"
    :data="backupList"
    stripe
    border
    style="width: 100%"
  >
    <el-table-column prop="name" label="备份名称" min-width="200" show-overflow-tooltip />
    <el-table-column prop="database" label="数据库" width="140" />
    <el-table-column label="大小" width="120">
      <template #default="{ row }">
        {{ formatSize(row.size) }}
      </template>
    </el-table-column>
    <el-table-column prop="file_count" label="文件数" width="90" />
    <el-table-column prop="created_at" label="创建时间" width="180" />
    <el-table-column label="操作" width="180" fixed="right">
      <template #default="{ row }">
        <el-button size="small" type="primary" link @click="handleDownload(row)">
          下载
        </el-button>
        <el-button size="small" type="primary" link @click="handleRestore(row)">
          恢复
        </el-button>
        <el-button size="small" type="danger" link @click="handleDelete(row)">
          删除
        </el-button>
      </template>
    </el-table-column>
    <template #empty>
      <el-empty v-if="!loading" description="暂无备份文件，备份接口实现后将显示列表" />
    </template>
  </el-table>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { listBackups, deleteBackup, getBackupDownloadUrl, type BackupFileInfo } from '@/api/backup'

const props = defineProps<{
  connectionId?: number
  refreshTrigger?: number
}>()

const emit = defineEmits<{
  (e: 'restore', item: BackupFileInfo): void
}>()

const loading = ref(false)
const backupList = ref<BackupFileInfo[]>([])

function formatSize(bytes: number): string {
  if (!bytes) return '-'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(2)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(2)} MB`
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

async function loadBackups() {
  loading.value = true
  try {
    const res = await listBackups(props.connectionId, 1, 50)
    backupList.value = res?.list || []
  } catch {
    backupList.value = []
  } finally {
    loading.value = false
  }
}

function handleDownload(row: BackupFileInfo) {
  const url = getBackupDownloadUrl(row.id)
  window.open(url, '_blank')
}

function handleRestore(row: BackupFileInfo) {
  emit('restore', row)
}

async function handleDelete(row: BackupFileInfo) {
  try {
    await deleteBackup(row.id)
    ElMessage.success('已删除')
    await loadBackups()
  } catch (e: any) {
    ElMessage.error(e?.message || '删除失败')
  }
}

watch(
  () => [props.connectionId, props.refreshTrigger],
  () => {
    loadBackups()
  }
)

onMounted(() => {
  loadBackups()
})
</script>
