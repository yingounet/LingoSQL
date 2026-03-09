<template>
  <div class="history-list">
    <!-- 加载状态 -->
    <div v-loading="loading" class="loading-container">
      <!-- 历史记录表格 -->
      <el-table
        :data="data"
        stripe
        style="width: 100%"
        :empty-text="t('historyList.noHistory')"
      >
        <!-- SQL 语句 -->
        <el-table-column prop="sql_query" :label="t('historyList.sqlStatement')" min-width="300">
          <template #default="{ row }">
            <div class="sql-cell">
              <code class="sql-code">{{ row.sql_query }}</code>
              <el-button
                link
                type="primary"
                size="small"
                @click="handleCopySQL(row.sql_query)"
                class="copy-btn"
              >
                <el-icon><CopyDocument /></el-icon>
                {{ t('historyList.copySql') }}
              </el-button>
            </div>
          </template>
        </el-table-column>

        <!-- 操作类型（仅系统执行显示） -->
        <el-table-column
          v-if="type === 'system'"
          prop="operation_type"
          :label="t('historyList.operationType')"
          width="150"
        >
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ formatOperationType(row.operation_type) }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- 执行时间 -->
        <el-table-column prop="execution_time_ms" :label="t('historyList.executionTime')" width="120">
          <template #default="{ row }">
            <span>{{ row.execution_time_ms }} ms</span>
          </template>
        </el-table-column>

        <!-- 受影响行数 -->
        <el-table-column prop="rows_affected" :label="t('historyList.affectedRows')" width="120">
          <template #default="{ row }">
            <span>{{ row.rows_affected }}</span>
          </template>
        </el-table-column>

        <!-- 状态 -->
        <el-table-column prop="success" :label="t('historyList.status')" width="100">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'" size="small">
              {{ row.success ? t('historyList.statusSuccess') : t('historyList.statusFailed') }}
            </el-tag>
          </template>
        </el-table-column>

        <!-- 错误信息 -->
        <el-table-column
          v-if="hasError"
          prop="error_message"
          :label="t('historyList.errorMessage')"
          min-width="200"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            <span v-if="row.error_message" class="error-text">
              {{ row.error_message }}
            </span>
            <span v-else>-</span>
          </template>
        </el-table-column>

        <!-- 创建时间 -->
        <el-table-column prop="created_at" :label="t('historyList.createdAt')" width="180">
          <template #default="{ row }">
            <span>{{ formatDateTime(row.created_at) }}</span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container" v-if="total > 0">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { CopyDocument } from '@element-plus/icons-vue'
import type { QueryHistory } from '@/types/history'

interface Props {
  connectionId: number
  type: 'user' | 'system'
  loading: boolean
  data: QueryHistory[]
  total: number
  page: number
  pageSize: number
}

const props = defineProps<Props>()
const { t } = useI18n()

const emit = defineEmits<{
  pageChange: [page: number]
}>()

const currentPage = ref(props.page)
const pageSize = ref(props.pageSize)

// 是否有错误记录
const hasError = computed(() => {
  return props.data.some(item => !item.success)
})

const operationTypeMap = computed<Record<string, string>>(() => ({
  GET_DATABASES: t('historyList.opGetDatabases'),
  GET_TABLES: t('historyList.opGetTables'),
  GET_TABLE_ROWS: t('historyList.opGetTableData'),
  GET_TABLE_COLUMNS: t('historyList.opGetTableSchema'),
  GET_TABLE_INDEXES: t('historyList.opGetTableIndexes'),
  UPDATE_TABLE_ROW: t('historyList.opUpdateTableData'),
  USE_DATABASE: t('historyList.opSwitchDatabase'),
}))

function formatOperationType(type?: string): string {
  if (!type) return '-'
  return operationTypeMap.value[type] || type
}

// 格式化日期时间
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

// 复制 SQL
function handleCopySQL(sql: string) {
  navigator.clipboard.writeText(sql).then(() => {
    ElMessage.success(t('historyList.copiedToClipboard'))
  }).catch(() => {
    ElMessage.error(t('historyList.copyFailed'))
  })
}

// 分页大小变化
function handleSizeChange(size: number) {
  pageSize.value = size
  currentPage.value = 1
  emit('pageChange', 1)
}

// 页码变化
function handlePageChange(page: number) {
  currentPage.value = page
  emit('pageChange', page)
}
</script>

<style scoped>
.history-list {
  width: 100%;
}

.loading-container {
  min-height: 400px;
}

.sql-cell {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.sql-code {
  flex: 1;
  font-family: 'Monaco', 'Consolas', 'Courier New', monospace;
  font-size: 12px;
  color: var(--color-text-primary);
  background: var(--color-background-secondary);
  padding: 4px 8px;
  border-radius: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.copy-btn {
  flex-shrink: 0;
}

.error-text {
  color: var(--color-error);
  font-size: 12px;
}

.pagination-container {
  margin-top: var(--spacing-lg);
  display: flex;
  justify-content: flex-end;
}
</style>
