<template>
  <div class="favorites-page">
    <PageHeader
      title="收藏"
      description="收藏的 SQL 语句，可按连接与数据库筛选"
    />

    <!-- 无连接时提示 -->
    <div v-if="!connectionStore.currentConnection" class="filter-hint">
      <el-alert type="info" :closable="false" show-icon>
        当前未选择连接，显示全部收藏；默认按最近使用时间排序。
      </el-alert>
    </div>

    <el-card class="content-card">
      <!-- 筛选与排序 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-radio-group v-model="scope" size="default" @change="loadList">
            <el-radio-button value="current_db" :disabled="!hasCurrentDb">
              当前数据库
            </el-radio-button>
            <el-radio-button value="current_connection" :disabled="!connectionStore.currentConnection">
              当前连接
            </el-radio-button>
            <el-radio-button value="all">
              全部
            </el-radio-button>
          </el-radio-group>
          <el-radio-group v-model="sort" size="default" class="sort-group" @change="loadList">
            <el-radio-button value="created_at">收藏时间</el-radio-button>
            <el-radio-button value="last_used_at">最近使用</el-radio-button>
          </el-radio-group>
        </div>
      </div>

      <div v-loading="loading" class="table-wrap">
        <el-table
          :data="list"
          stripe
          style="width: 100%"
          empty-text="暂无收藏"
        >
          <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
          <el-table-column prop="description" label="描述" min-width="120" show-overflow-tooltip />
          <el-table-column prop="connection_name" label="连接" width="120" show-overflow-tooltip />
          <el-table-column prop="database" label="数据库" width="100" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.database || '-' }}
            </template>
          </el-table-column>
          <el-table-column label="SQL" min-width="220">
            <template #default="{ row }">
              <code class="sql-preview">{{ sqlPreview(row.sql_query) }}</code>
            </template>
          </el-table-column>
          <el-table-column label="收藏时间" width="160">
            <template #default="{ row }">
              {{ formatDateTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="最近使用" width="160">
            <template #default="{ row }">
              {{ row.last_used_at ? formatDateTime(row.last_used_at) : '-' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="handleUse(row)">
                使用
              </el-button>
              <el-button link type="primary" size="small" @click="handleEdit(row)">
                编辑
              </el-button>
              <el-button link type="danger" size="small" @click="handleDelete(row)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      title="编辑收藏"
      width="560px"
      destroy-on-close
      @close="editForm = null"
    >
      <el-form v-if="editForm" :model="editForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="editForm.name" placeholder="收藏名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" rows="2" placeholder="可选" />
        </el-form-item>
        <el-form-item label="SQL">
          <el-input v-model="editForm.sql_query" type="textarea" :rows="6" placeholder="SQL 语句" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitEdit">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import PageHeader from '@/components/layout/PageHeader.vue'
import { useConnectionStore } from '@/store/connection'
import { getFavorites, recordFavoriteUse, updateFavorite, deleteFavorite } from '@/api/favorite'
import type { Favorite } from '@/types/favorite'

const router = useRouter()
const connectionStore = useConnectionStore()

const scope = ref<'current_db' | 'current_connection' | 'all'>('all')
const sort = ref<'created_at' | 'last_used_at'>('last_used_at')
const loading = ref(false)
const list = ref<Favorite[]>([])
const editDialogVisible = ref(false)
const saving = ref(false)
const editForm = ref<{ id: number; name: string; description: string; sql_query: string } | null>(null)

const hasCurrentDb = computed(() => !!(
  connectionStore.currentConnection &&
  connectionStore.currentDatabase
))

// 无连接时默认「最近使用」排序
watch(
  () => connectionStore.currentConnection,
  () => {
    if (!connectionStore.currentConnection && sort.value !== 'last_used_at') {
      sort.value = 'last_used_at'
      loadList()
    }
  },
  { immediate: false }
)

function getListParams() {
  const params: { connection_id?: number; database?: string; sort: 'created_at' | 'last_used_at' } = {
    sort: sort.value,
  }
  if (scope.value === 'current_db' && connectionStore.currentConnection && connectionStore.currentDatabase) {
    params.connection_id = connectionStore.currentConnection.id
    params.database = connectionStore.currentDatabase
  } else if (scope.value === 'current_connection' && connectionStore.currentConnection) {
    params.connection_id = connectionStore.currentConnection.id
  }
  return params
}

async function loadList() {
  loading.value = true
  try {
    const params = getListParams()
    list.value = await getFavorites({
      connection_id: params.connection_id ?? null,
      database: params.database ?? null,
      sort: params.sort,
    })
  } catch (e: any) {
    ElMessage.error(e.message || '加载收藏列表失败')
  } finally {
    loading.value = false
  }
}

function sqlPreview(sql: string): string {
  const t = sql.replace(/\s+/g, ' ').trim()
  return t.length > 80 ? t.slice(0, 80) + '...' : t
}

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

/** 判断是否为 SELECT 查询（用于从收藏使用时是否自动执行） */
function isSelectStatement(sql: string): boolean {
  let s = sql.trim()
  s = s.replace(/^--[^\n]*\n?/gm, '').trim()
  s = s.replace(/\/\*[\s\S]*?\*\//g, '').trim()
  return /^select\b/i.test(s)
}

async function handleUse(fav: Favorite) {
  const needRestore =
    !connectionStore.currentConnection ||
    connectionStore.currentConnection.id !== fav.connection_id ||
    (fav.database && connectionStore.currentDatabase !== fav.database)
  if (needRestore) {
    try {
      const ok = await connectionStore.restoreState(fav.connection_id, fav.database || undefined)
      if (!ok) {
        ElMessage.error('无法切换连接或数据库')
        return
      }
    } catch (e: any) {
      ElMessage.error(e.message || '切换连接失败')
      return
    }
  }
  try {
    await recordFavoriteUse(fav.id)
  } catch {
    // 忽略记录失败，仍跳转
  }
  router.push({
    path: '/query',
    query: {
      connection_id: String(fav.connection_id),
      ...(fav.database ? { database: fav.database } : {}),
    },
    state: {
      initialSql: fav.sql_query,
      autoExecute: isSelectStatement(fav.sql_query),
    },
  })
}

function handleEdit(row: Favorite) {
  editForm.value = {
    id: row.id,
    name: row.name,
    description: row.description || '',
    sql_query: row.sql_query,
  }
  editDialogVisible.value = true
}

async function submitEdit() {
  if (!editForm.value) return
  saving.value = true
  try {
    await updateFavorite(editForm.value.id, {
      name: editForm.value.name,
      description: editForm.value.description,
      sql_query: editForm.value.sql_query,
    })
    ElMessage.success('已更新')
    editDialogVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e.message || '更新失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: Favorite) {
  try {
    await ElMessageBox.confirm(`确定删除收藏「${row.name}」？`, '删除收藏', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deleteFavorite(row.id)
    ElMessage.success('已删除')
    loadList()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

onMounted(() => {
  if (connectionStore.currentConnection && connectionStore.currentDatabase) {
    scope.value = 'current_db'
  }
  if (connectionStore.currentConnection) {
    sort.value = 'created_at'
  }
  loadList()
})
</script>

<style scoped>
.favorites-page {
  max-width: 1400px;
  margin: 0 auto;
  width: 100%;
}

.filter-hint {
  margin-bottom: var(--spacing-md);
}

.content-card {
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
  background: var(--color-background);
  margin-top: var(--spacing-lg);
}

.content-card :deep(.el-card__body) {
  padding: var(--spacing-xl);
}

.toolbar {
  margin-bottom: var(--spacing-lg);
}

.toolbar-left {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--spacing-lg);
}

.sort-group {
  margin-left: var(--spacing-md);
}

.table-wrap {
  min-height: 200px;
}

.sql-preview {
  font-family: Monaco, Consolas, 'Courier New', monospace;
  font-size: 12px;
  color: var(--color-text-secondary);
  background: var(--color-background-secondary);
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
