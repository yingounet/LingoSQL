<template>
  <div class="favorites-page">
    <PageHeader
      :title="t('favorites.title')"
      :description="t('favorites.description')"
    />

    <!-- 无连接时提示 -->
    <div v-if="!connectionStore.currentConnection" class="filter-hint">
      <el-alert type="info" :closable="false" show-icon>
        {{ t('favorites.noConnectionHint') }}
      </el-alert>
    </div>

    <el-card class="content-card">
      <!-- 筛选与排序 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-radio-group v-model="scope" size="default" @change="loadList">
            <el-radio-button value="current_db" :disabled="!hasCurrentDb">
              {{ t('favorites.currentDatabase') }}
            </el-radio-button>
            <el-radio-button value="current_connection" :disabled="!connectionStore.currentConnection">
              {{ t('favorites.currentConnection') }}
            </el-radio-button>
            <el-radio-button value="all">
              {{ t('favorites.all') }}
            </el-radio-button>
          </el-radio-group>
          <el-radio-group v-model="sort" size="default" class="sort-group" @change="loadList">
            <el-radio-button value="created_at">{{ t('favorites.favoriteTime') }}</el-radio-button>
            <el-radio-button value="last_used_at">{{ t('favorites.lastUsed') }}</el-radio-button>
          </el-radio-group>
        </div>
      </div>

      <div v-loading="loading" class="table-wrap">
        <el-table
          :data="list"
          stripe
          style="width: 100%"
          :empty-text="t('favorites.noFavorites')"
        >
          <el-table-column prop="name" :label="t('favorites.favoriteName')" min-width="140" show-overflow-tooltip />
          <el-table-column prop="description" :label="t('common.description')" min-width="120" show-overflow-tooltip />
          <el-table-column prop="connection_name" :label="t('favorites.connection')" width="120" show-overflow-tooltip />
          <el-table-column prop="database" :label="t('favorites.database')" width="100" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.database || '-' }}
            </template>
          </el-table-column>
          <el-table-column :label="t('favorites.sql')" min-width="220">
            <template #default="{ row }">
              <code class="sql-preview">{{ sqlPreview(row.sql_query) }}</code>
            </template>
          </el-table-column>
          <el-table-column :label="t('favorites.favoriteTime')" width="160">
            <template #default="{ row }">
              {{ formatDateTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column :label="t('favorites.lastUsed')" width="160">
            <template #default="{ row }">
              {{ row.last_used_at ? formatDateTime(row.last_used_at) : '-' }}
            </template>
          </el-table-column>
          <el-table-column :label="t('common.actions')" width="200" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="handleUse(row)">
                {{ t('common.use') }}
              </el-button>
              <el-button link type="primary" size="small" @click="handleEdit(row)">
                {{ t('common.edit') }}
              </el-button>
              <el-button link type="danger" size="small" @click="handleDelete(row)">
                {{ t('common.delete') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="editDialogVisible"
      :title="t('favorites.editFavorite')"
      width="560px"
      destroy-on-close
      @close="editForm = null"
    >
      <el-form v-if="editForm" :model="editForm" label-width="80px">
        <el-form-item :label="t('favorites.favoriteName')">
          <el-input v-model="editForm.name" :placeholder="t('favorites.favoriteNamePlaceholder')" />
        </el-form-item>
        <el-form-item :label="t('common.description')">
          <el-input v-model="editForm.description" type="textarea" rows="2" :placeholder="t('common.optional')" />
        </el-form-item>
        <el-form-item :label="t('favorites.sql')">
          <el-input v-model="editForm.sql_query" type="textarea" :rows="6" :placeholder="t('favorites.sqlStatement')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saving" @click="submitEdit">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import PageHeader from '@/components/layout/PageHeader.vue'
import { useConnectionStore } from '@/store/connection'
import { getFavorites, recordFavoriteUse, updateFavorite, deleteFavorite } from '@/api/favorite'
import type { Favorite } from '@/types/favorite'

const { t } = useI18n()
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
    ElMessage.error(e.message || t('favorites.loadFailed'))
  } finally {
    loading.value = false
  }
}

function sqlPreview(sql: string): string {
  const trimmed = sql.replace(/\s+/g, ' ').trim()
  return trimmed.length > 80 ? trimmed.slice(0, 80) + '...' : trimmed
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
        ElMessage.error(t('favorites.cannotSwitchConnection'))
        return
      }
    } catch (e: any) {
      ElMessage.error(e.message || t('favorites.switchFailed'))
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
    ElMessage.success(t('favorites.updated'))
    editDialogVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e.message || t('favorites.updateFailed'))
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: Favorite) {
  try {
    await ElMessageBox.confirm(t('favorites.deleteConfirm', { name: row.name }), t('favorites.deleteFavoriteTitle'), {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await deleteFavorite(row.id)
    ElMessage.success(t('favorites.deleted'))
    loadList()
  } catch (e: any) {
    ElMessage.error(e.message || t('favorites.deleteFailed'))
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
