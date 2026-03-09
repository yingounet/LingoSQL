<template>
  <div class="database-management">
    <!-- 操作栏 -->
    <div class="action-bar">
      <el-button type="primary" @click="handleCreateDatabase">
        <el-icon><Plus /></el-icon>
        {{ t('dbAdmin.createDb') }}
      </el-button>
    </div>
    
    <!-- 数据库列表表格 -->
    <el-table
      :data="databaseList"
      v-loading="loading"
      stripe
      border
      style="width: 100%"
    >
      <el-table-column prop="name" :label="t('dbAdmin.dbName')" sortable />
      <el-table-column prop="charset" :label="t('dbAdmin.charsetLabel')" v-if="dbType === 'mysql'" />
      <el-table-column prop="encoding" :label="t('dbAdmin.encodingLabel')" v-if="dbType === 'postgresql'" />
      <el-table-column prop="size" :label="t('dbAdmin.sizeLabel')" :formatter="formatSize" />
      <el-table-column prop="table_count" :label="t('dbAdmin.tableCount')" />
      <el-table-column prop="created_at" :label="t('dbAdmin.createdAt')" />
      <el-table-column :label="t('common.actions')" width="250" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleViewDetail(row)">
            {{ t('dbAdmin.detail') }}
          </el-button>
          <el-button 
            v-if="dbType === 'postgresql'"
            size="small" 
            @click="handleRename(row)"
          >
            {{ t('dbAdmin.rename') }}
          </el-button>
          <el-button 
            size="small" 
            type="danger" 
            @click="handleDelete(row)"
          >
            {{ t('common.delete') }}
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 统计信息 -->
    <div class="table-footer">
      <span>{{ t('dbAdmin.totalNDbs', { n: databaseList.length }) }}</span>
    </div>
    
    <!-- 对话框组件 -->
    <DatabaseCreateDialog
      v-model="showCreateDialog"
      :db-type="dbType"
      @success="handleCreateSuccess"
    />
    <DatabaseDetailDialog
      v-model="showDetailDialog"
      :database="selectedDatabase"
      :db-type="dbType"
    />
    <DatabaseDeleteConfirm
      v-model="showDeleteDialog"
      :database="selectedDatabase"
      @confirm="handleDeleteConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useConnectionStore } from '@/store/connection'
import { getDatabaseList, deleteDatabase } from '@/api/databaseAdmin'
import type { DatabaseInfo } from '@/types/databaseAdmin'
import DatabaseCreateDialog from './DatabaseCreateDialog.vue'
import DatabaseDetailDialog from './DatabaseDetailDialog.vue'
import DatabaseDeleteConfirm from './DatabaseDeleteConfirm.vue'

const { t } = useI18n()
const connectionStore = useConnectionStore()

const currentConnection = computed(() => connectionStore.currentConnection)
const dbType = computed(() => currentConnection.value?.db_type || 'mysql')

const databaseList = ref<DatabaseInfo[]>([])
const loading = ref(false)
const showCreateDialog = ref(false)
const showDetailDialog = ref(false)
const showDeleteDialog = ref(false)
const selectedDatabase = ref<DatabaseInfo | null>(null)

// 加载数据库列表
async function loadDatabaseList() {
  if (!currentConnection.value) return
  
  loading.value = true
  try {
    databaseList.value = await getDatabaseList(currentConnection.value.id)
  } catch (error: any) {
    ElMessage.error(error.message || t('dbAdmin.loadDbListFailed'))
  } finally {
    loading.value = false
  }
}

// 格式化文件大小
function formatSize(row: DatabaseInfo) {
  if (!row.size) return '-'
  const size = row.size
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(2)} MB`
  return `${(size / (1024 * 1024 * 1024)).toFixed(2)} GB`
}

// 创建数据库
function handleCreateDatabase() {
  showCreateDialog.value = true
}

// 创建成功
async function handleCreateSuccess() {
  await loadDatabaseList()
  ElMessage.success(t('dbAdmin.dbCreated'))
}

// 查看详情
function handleViewDetail(row: DatabaseInfo) {
  selectedDatabase.value = row
  showDetailDialog.value = true
}

// 重命名
function handleRename(row: DatabaseInfo) {
  // TODO: 实现重命名功能
  ElMessage.info(t('dbAdmin.renameTodo'))
}

// 删除
function handleDelete(row: DatabaseInfo) {
  selectedDatabase.value = row
  showDeleteDialog.value = true
}

// 确认删除
async function handleDeleteConfirm() {
  if (!selectedDatabase.value || !currentConnection.value) return
  
  try {
    await deleteDatabase(currentConnection.value.id, selectedDatabase.value.name)
    ElMessage.success(t('dbAdmin.dbDeleted'))
    await loadDatabaseList()
  } catch (error: any) {
    ElMessage.error(error.message || t('dbAdmin.deleteDbFailed'))
  } finally {
    showDeleteDialog.value = false
    selectedDatabase.value = null
  }
}

onMounted(() => {
  loadDatabaseList()
})
</script>

<style scoped>
.database-management {
  padding: var(--spacing-md);
}

.action-bar {
  margin-bottom: var(--spacing-md);
}

.table-footer {
  margin-top: var(--spacing-md);
  padding: var(--spacing-sm);
  text-align: right;
  color: var(--color-text-secondary);
  font-size: 14px;
}
</style>
