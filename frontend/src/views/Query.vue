<template>
  <div class="query-page">
    <!-- 页面标题 -->
    <PageHeader 
      :title="t('query.title')" 
      :description="t('query.description')"
    >
      <template #actions>
        <el-tooltip :content="t('query.selectConnectionHint')" :disabled="!!connectionStore.currentConnection">
          <el-button
            :disabled="!connectionStore.currentConnection"
            @click="openSaveFavoriteDialog"
          >
            <el-icon><Star /></el-icon>
            {{ t('query.favorite') }}
          </el-button>
        </el-tooltip>
        <el-button-group>
          <el-button 
            type="primary" 
            @click="handleExecute" 
            :loading="executing"
            :disabled="!canExecute || (executeMode === 'selected' && !hasSelection)"
          >
            <el-icon><CaretRight /></el-icon>
            {{ executeButtonText }}
          </el-button>
          <el-dropdown @command="handleExecuteCommand" trigger="click">
            <el-button type="primary" :disabled="!canExecute || executing">
              <el-icon><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item 
                  command="selected" 
                  :disabled="!hasSelection"
                  :class="{ 'is-active': executeMode === 'selected' }"
                >
                  <el-icon><Select /></el-icon>
                  {{ t('query.executeSelected') }}
                </el-dropdown-item>
                <el-dropdown-item 
                  command="all"
                  :class="{ 'is-active': executeMode === 'all' }"
                >
                  <el-icon><Document /></el-icon>
                  {{ t('query.executeAll') }}
                </el-dropdown-item>
                <el-dropdown-item 
                  command="explain"
                  :disabled="!canExecute"
                  divided
                >
                  <el-icon><View /></el-icon>
                  {{ t('query.viewExecPlan') }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </el-button-group>
      </template>
    </PageHeader>

    <!-- 未选择连接时的提示 -->
    <div v-if="!connectionStore.currentConnection" class="empty-state">
      <el-empty :description="t('query.selectConnectionHint')">
        <template #image>
          <el-icon :size="64" class="empty-icon"><Connection /></el-icon>
        </template>
      </el-empty>
    </div>

    <!-- 主内容区域 -->
    <template v-else>
      <!-- SQL 编辑器区域 -->
      <el-card class="editor-card">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><EditPen /></el-icon>
              {{ t('query.sqlEditor') }}
            </span>
            <div class="editor-actions">
              <el-button size="small" @click="handleFormatSQL" :title="t('common.format')">
                <el-icon><DocumentCopy /></el-icon>
                {{ t('common.format') }}
              </el-button>
              <el-button-group v-if="inTransaction" style="margin-left: 8px;">
                <el-button size="small" type="success" @click="handleCommitTransaction" :title="t('query.commit')">
                  <el-icon><Check /></el-icon>
                  {{ t('query.commit') }}
                </el-button>
                <el-button size="small" type="danger" @click="handleRollbackTransaction" :title="t('query.rollback')">
                  <el-icon><Close /></el-icon>
                  {{ t('query.rollback') }}
                </el-button>
              </el-button-group>
              <div class="editor-info" style="margin-left: 12px;">
                <el-tag size="small" type="info">
                  {{ connectionStore.currentConnection?.name }}
                </el-tag>
                <el-tag v-if="connectionStore.currentDatabase" size="small" type="success">
                  {{ connectionStore.currentDatabase }}
                </el-tag>
                <el-tag v-if="inTransaction" size="small" type="warning">
                  {{ t('query.inTransaction') }}
                </el-tag>
                <span v-if="hasSelection" class="selection-hint">
                  <el-icon><Select /></el-icon>
                  {{ t('query.selectedText') }}
                </span>
              </div>
            </div>
          </div>
        </template>
        
        <div ref="editorContainer" class="editor-container"></div>
      </el-card>

      <!-- 查询结果区域 -->
      <el-card class="result-card" v-if="hasResult || errorMessage">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><List /></el-icon>
              {{ t('query.queryResults') }}
              <el-tag v-if="result" size="small" type="success" class="result-tag">
                {{ t('query.nRecords', { n: formatNumber(result.rows.length) }) }}
              </el-tag>
              <el-tag v-if="result" size="small" type="info" class="result-tag">
                {{ t('query.timeCost', { n: result.execution_time_ms }) }}
              </el-tag>
            </span>
            <div class="header-actions" v-if="hasResult">
              <el-button size="small" @click="handleClearResult">
                <el-icon><Delete /></el-icon>
                {{ t('query.clear') }}
              </el-button>
              <el-dropdown @command="doExport" trigger="click">
                <el-button size="small" :disabled="!result || result.rows.length === 0">
                  <el-icon><Download /></el-icon>
                  {{ t('common.export') }}
                  <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="csv">{{ t('query.exportCSV') }}</el-dropdown-item>
                    <el-dropdown-item command="excel">{{ t('query.exportExcel') }}</el-dropdown-item>
                    <el-dropdown-item command="json">{{ t('query.exportJSON') }}</el-dropdown-item>
                    <el-dropdown-item command="sql">{{ t('query.exportSQL') }}</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </div>
        </template>

        <!-- 错误信息 -->
        <el-alert
          v-if="errorMessage"
          :title="errorMessage"
          type="error"
          show-icon
          :closable="true"
          @close="errorMessage = ''"
          class="error-alert"
        />

        <!-- 结果表格 -->
        <template v-if="hasResult">
          <!-- 无数据情况（如 UPDATE/DELETE 语句） -->
          <div v-if="result!.columns.length === 0" class="no-data-result">
            <el-icon :size="48" class="success-icon"><CircleCheck /></el-icon>
            <p class="success-text">{{ t('query.execSuccess') }}</p>
            <p class="affected-rows">
              {{ t('query.affectedRows') }}: <strong>{{ result!.rows_affected }}</strong>
            </p>
          </div>

          <!-- 有数据的表格 -->
          <el-table 
            v-else
            :data="tableData" 
            stripe
            border
            class="result-table"
            :max-height="tableMaxHeight"
          >
            <!-- 序号列 -->
            <el-table-column type="index" label="#" width="60" fixed="left" />

            <!-- 动态数据列 -->
            <el-table-column 
              v-for="col in result!.columns" 
              :key="col"
              :prop="col"
              :label="col"
              :min-width="getColumnWidth(col)"
              show-overflow-tooltip
            >
              <template #default="{ row }">
                <span :class="getCellClass(row[col])">
                  {{ formatCellValue(row[col]) }}
                </span>
              </template>
            </el-table-column>

            <!-- 空状态 -->
            <template #empty>
              <div class="table-empty">
                <el-icon :size="48"><DocumentRemove /></el-icon>
                <p>{{ t('query.emptyResult') }}</p>
              </div>
            </template>
          </el-table>

          <!-- 结果统计 -->
          <div class="result-footer" v-if="result!.columns.length > 0">
            <span class="result-stats">
              {{ t('query.totalNRecords', { n: formatNumber(result!.rows.length) }) }}
              <template v-if="result!.rows_affected > 0">
                | {{ t('query.affectedRows') }}: {{ result!.rows_affected }}
              </template>
            </span>
            <span class="execution-time">
              {{ t('query.executionTime', { n: result!.execution_time_ms }) }}
            </span>
          </div>
        </template>
      </el-card>
    </template>

    <!-- 执行计划对话框 -->
    <el-dialog
      v-model="showExplainDialog"
      :title="t('query.executionPlan')"
      width="900px"
      destroy-on-close
    >
      <div v-if="explainResult" class="explain-result">
        <el-table :data="explainResult.plan" stripe border>
          <el-table-column
            v-for="(value, key) in explainResult.plan[0] || {}"
            :key="key"
            :prop="key"
            :label="key"
            :min-width="120"
            show-overflow-tooltip
          />
        </el-table>
        <div class="explain-footer">
          <span>{{ t('query.executionTime', { n: explainResult.execution_time_ms }) }}</span>
        </div>
      </div>
    </el-dialog>

    <!-- 保存为收藏对话框 -->
    <el-dialog
      v-model="saveFavoriteDialogVisible"
      :title="t('query.saveAsFavorite')"
      width="560px"
      destroy-on-close
      @close="saveFavoriteForm = { name: '', description: '' }; saveFavoriteSql = ''"
    >
      <el-form :model="saveFavoriteForm" label-width="80px">
        <el-form-item :label="t('query.favoriteName')" required>
          <el-input v-model="saveFavoriteForm.name" :placeholder="t('query.favoriteNamePlaceholder')" maxlength="200" show-word-limit />
        </el-form-item>
        <el-form-item :label="t('query.favoriteDesc')">
          <el-input v-model="saveFavoriteForm.description" type="textarea" rows="2" :placeholder="t('common.optional')" />
        </el-form-item>
        <el-form-item label="SQL">
          <el-input v-model="saveFavoriteSql" type="textarea" :rows="6" readonly :placeholder="t('query.currentEditorContent')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="saveFavoriteDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" :loading="saveFavoriteLoading" @click="submitSaveFavorite">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, onActivated, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import {
  CaretRight,
  ArrowDown,
  Select,
  Document,
  Connection,
  EditPen,
  List,
  Delete,
  Download,
  CircleCheck,
  DocumentRemove,
  Star,
  DocumentCopy,
  View,
  Check,
  Close
} from '@element-plus/icons-vue'
import * as monaco from 'monaco-editor'
import PageHeader from '@/components/layout/PageHeader.vue'
import { useConnectionStore } from '@/store/connection'
import { useTheme } from '@/composables/useTheme'
import { 
  executeQuery, 
  explainQuery,
  beginTransaction,
  commitTransaction,
  rollbackTransaction,
  type QueryExecuteResponse,
  type ExplainResponse
} from '@/api/query'
import { createFavorite } from '@/api/favorite'
import { formatSQL } from '@/utils/sqlFormatter'
import { registerSqlCompletionProvider } from '@/utils/monacoSqlCompletion'

// ==================== Store ====================

const { t } = useI18n()
const connectionStore = useConnectionStore()
const { theme: uiTheme } = useTheme()

// ==================== 状态 ====================

// 编辑器相关
const editorContainer = ref<HTMLDivElement>()
let editor: monaco.editor.IStandaloneCodeEditor | null = null

// 执行状态
const executing = ref(false)
const executeMode = ref<'selected' | 'all'>('all')

// 保存为收藏
const saveFavoriteDialogVisible = ref(false)
const saveFavoriteForm = ref({ name: '', description: '' })
const saveFavoriteSql = ref('')
const saveFavoriteLoading = ref(false)

// 从收藏「使用」跳转时，若为 SELECT 则标记为待自动执行
const pendingAutoExecute = ref(false)

// 选中状态
const hasSelection = ref(false)
// 缓存选中的文本（用于下拉菜单执行所选，避免编辑器失焦导致选区丢失）
const cachedSelection = ref('')

// 查询结果
const result = ref<QueryExecuteResponse | null>(null)
const errorMessage = ref('')

// 执行计划
const explainResult = ref<ExplainResponse | null>(null)
const showExplainDialog = ref(false)

// 事务状态
const inTransaction = ref(false)
const transactionLoading = ref(false)

// 自动补全提供者
let completionProvider: monaco.IDisposable | null = null

// ==================== 计算属性 ====================

// 是否可以执行
const canExecute = computed(() => {
  return connectionStore.currentConnection && connectionStore.currentDatabase
})

// 执行按钮文案
const executeButtonText = computed(() => {
  return executeMode.value === 'selected' ? t('query.executeSelected') : t('query.executeAll')
})

// 是否有结果
const hasResult = computed(() => {
  return result.value !== null
})

// 表格数据（将二维数组转换为对象数组）
const tableData = computed(() => {
  if (!result.value || result.value.columns.length === 0) {
    return []
  }
  
  return result.value.rows.map(row => {
    const obj: Record<string, unknown> = {}
    result.value!.columns.forEach((col, index) => {
      obj[col] = row[index]
    })
    return obj
  })
})

// 表格最大高度
const tableMaxHeight = computed(() => {
  return Math.max(300, window.innerHeight - 550)
})

// ==================== 工具函数 ====================

// 格式化数字
function formatNumber(num: number): string {
  return num.toLocaleString()
}

// 获取列宽度
function getColumnWidth(col: string): number {
  if (col.length > 20) {
    return 200
  }
  if (col.length > 10) {
    return 150
  }
  return 120
}

// 获取单元格样式类
function getCellClass(value: unknown): string {
  const classes: string[] = ['cell-value']
  
  if (typeof value === 'number') {
    classes.push('numeric')
  }
  if (value === null || value === undefined) {
    classes.push('null-value')
  }
  
  return classes.join(' ')
}

// 格式化单元格值
function formatCellValue(value: unknown): string {
  if (value === null || value === undefined) {
    return 'NULL'
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

// ==================== 编辑器操作 ====================

// 初始化编辑器
function initEditor() {
  if (!editorContainer.value) return

  // 配置 Monaco Editor 使用 CDN 加载 workers
  self.MonacoEnvironment = {
    getWorker: function () {
      return new Worker(
        URL.createObjectURL(
          new Blob(
            ['self.onmessage = function() {}'],
            { type: 'text/javascript' }
          )
        )
      )
    }
  }

  editor = monaco.editor.create(editorContainer.value, {
    value: t('query.editorDefault'),
    language: 'sql',
    theme: uiTheme.value === 'dark' ? 'vs-dark' : 'vs',
    minimap: { enabled: false },
    lineNumbers: 'on',
    automaticLayout: true,
    fontSize: 14,
    fontFamily: "Monaco, 'Courier New', Consolas, monospace",
    scrollBeyondLastLine: false,
    wordWrap: 'on',
    tabSize: 2,
    insertSpaces: true,
    renderWhitespace: 'selection',
    scrollbar: {
      verticalScrollbarSize: 10,
      horizontalScrollbarSize: 10,
    },
    suggestOnTriggerCharacters: true,
    quickSuggestions: true,
    padding: {
      top: 10,
      bottom: 10,
    },
  })

  // 监听选中变化
  editor.onDidChangeCursorSelection(() => {
    updateSelectionState()
  })

  // 监听内容变化
  editor.onDidChangeModelContent(() => {
    updateSelectionState()
  })

  // 添加快捷键
  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
    handleExecute()
  })

  // 添加 F5 快捷键
  editor.addCommand(monaco.KeyCode.F5, () => {
    handleExecute()
  })

  // 添加格式化快捷键 Ctrl+Shift+F
  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyMod.Shift | monaco.KeyCode.KeyF, () => {
    handleFormatSQL()
  })

  // 注册自动补全提供者
  updateCompletionProvider()
}

function applyEditorTheme() {
  if (!editor) return
  monaco.editor.setTheme(uiTheme.value === 'dark' ? 'vs-dark' : 'vs')
}

// 更新选中状态
function updateSelectionState() {
  if (!editor) {
    hasSelection.value = false
    cachedSelection.value = ''
    executeMode.value = 'all'
    return
  }
  
  const selection = editor.getSelection()
  if (!selection) {
    hasSelection.value = false
    cachedSelection.value = ''
    executeMode.value = 'all'
    return
  }
  
  const selectedText = editor.getModel()?.getValueInRange(selection)
  const trimmedText = selectedText?.trim() || ''
  const hadSelection = hasSelection.value
  hasSelection.value = trimmedText.length > 0
  
  // 缓存选中的文本（当有选中内容时更新缓存）
  if (hasSelection.value) {
    cachedSelection.value = trimmedText
    // 自动切换到"执行所选"模式
    executeMode.value = 'selected'
  } else if (hadSelection) {
    // 从有选中变为无选中时，切换回"执行所有"模式
    executeMode.value = 'all'
  }
}

// 获取要执行的 SQL
function getExecuteSQL(): string {
  if (!editor) return ''
  
  if (executeMode.value === 'selected') {
    // 优先使用缓存的选中文本（避免编辑器失焦导致选区丢失）
    if (cachedSelection.value) {
      return cachedSelection.value
    }
    // 如果缓存为空，尝试实时获取
    const selection = editor.getSelection()
    if (selection) {
      const selectedText = editor.getModel()?.getValueInRange(selection)
      return selectedText?.trim() || ''
    }
  }
  
  return editor.getValue().trim()
}

// 从路由 state 预填 SQL（从收藏页「使用」跳转时）；若带 autoExecute 则稍后自动执行
function applyInitialSqlFromState() {
  const state = window.history.state as { initialSql?: string; autoExecute?: boolean } | null
  if (!state?.initialSql || typeof state.initialSql !== 'string') return
  const model = editor?.getModel()
  if (model) {
    model.setValue(state.initialSql)
    if (state.autoExecute) {
      pendingAutoExecute.value = true
    }
    window.history.replaceState({ ...state, initialSql: undefined, autoExecute: undefined }, '')
  }
}

// 更新自动补全提供者
function updateCompletionProvider() {
  // 移除旧的提供者
  if (completionProvider) {
    completionProvider.dispose()
    completionProvider = null
  }

  // 注册新的提供者
  if (connectionStore.currentConnection && connectionStore.currentDatabase) {
    completionProvider = registerSqlCompletionProvider(
      connectionStore.currentConnection,
      connectionStore.currentDatabase
    )
  }
}

// 格式化 SQL
function handleFormatSQL() {
  if (!editor) return
  
  const model = editor.getModel()
  if (!model) return

  const sql = model.getValue()
  const dbType = connectionStore.currentConnection?.db_type || 'mysql'
  const formatted = formatSQL(sql, dbType)
  
  model.setValue(formatted)
  ElMessage.success(t('query.sqlFormatted'))
}

// 查看执行计划
async function handleExplainSQL() {
  if (!editor || !connectionStore.currentConnection || !connectionStore.currentDatabase) return

  const sql = getExecuteSQL()
  if (!sql) {
    ElMessage.warning(t('query.enterSql'))
    return
  }

  try {
    explainResult.value = await explainQuery({
      connection_id: connectionStore.currentConnection.id,
      database: connectionStore.currentDatabase,
      sql
    })
    showExplainDialog.value = true
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    ElMessage.error(err.response?.data?.message || err.message || t('query.getExecPlanFailed'))
  }
}

// 开始事务
async function handleBeginTransaction() {
  if (!connectionStore.currentConnection) return

  transactionLoading.value = true
  try {
    await beginTransaction(
      connectionStore.currentConnection.id,
      connectionStore.currentDatabase || undefined
    )
    inTransaction.value = true
    ElMessage.success(t('query.txStarted'))
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    ElMessage.error(err.response?.data?.message || err.message || t('query.txStartFailed'))
  } finally {
    transactionLoading.value = false
  }
}

// 提交事务
async function handleCommitTransaction() {
  if (!connectionStore.currentConnection) return

  transactionLoading.value = true
  try {
    await commitTransaction(
      connectionStore.currentConnection.id,
      connectionStore.currentDatabase || undefined
    )
    inTransaction.value = false
    ElMessage.success(t('query.txCommitted'))
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    ElMessage.error(err.response?.data?.message || err.message || t('query.txCommitFailed'))
  } finally {
    transactionLoading.value = false
  }
}

// 回滚事务
async function handleRollbackTransaction() {
  if (!connectionStore.currentConnection) return

  transactionLoading.value = true
  try {
    await rollbackTransaction(
      connectionStore.currentConnection.id,
      connectionStore.currentDatabase || undefined
    )
    inTransaction.value = false
    ElMessage.success(t('query.txRolledBack'))
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    ElMessage.error(err.response?.data?.message || err.message || t('query.txRollbackFailed'))
  } finally {
    transactionLoading.value = false
  }
}

// 销毁编辑器
function destroyEditor() {
  if (completionProvider) {
    completionProvider.dispose()
    completionProvider = null
  }
  if (editor) {
    editor.dispose()
    editor = null
  }
}

// ==================== 执行操作 ====================

// 执行查询
async function handleExecute() {
  const sql = getExecuteSQL()
  
  if (!sql) {
    ElMessage.warning(t('query.enterSql'))
    return
  }
  
  if (!connectionStore.currentConnection) {
    ElMessage.warning(t('query.selectDbConnection'))
    return
  }
  
  if (!connectionStore.currentDatabase) {
    ElMessage.warning(t('query.selectDatabase'))
    return
  }
  
  executing.value = true
  errorMessage.value = ''
  
  try {
    result.value = await executeQuery({
      connection_id: connectionStore.currentConnection.id,
      database: connectionStore.currentDatabase,
      sql
    })
    
    ElMessage.success(t('query.queryExecSuccess', { n: result.value.execution_time_ms }))
  } catch (error: unknown) {
    console.error('查询执行失败:', error)
    const err = error as { 
      response?: { 
        data?: { 
          message?: string
          data?: { error?: string; query_id?: number }
        } 
      }
      message?: string 
    }
    // 优先使用详细错误信息，其次使用响应消息，最后使用通用错误
    const detailError = err.response?.data?.data?.error
    const responseMessage = err.response?.data?.message
    errorMessage.value = detailError || responseMessage || err.message || t('query.queryExecFailed')
    result.value = null
    ElMessage.error(errorMessage.value)
  } finally {
    executing.value = false
  }
}

// 处理下拉菜单命令
function handleExecuteCommand(command: 'selected' | 'all' | 'explain') {
  if (command === 'explain') {
    handleExplainSQL()
  } else {
    executeMode.value = command
  }
}

// 清除结果
function handleClearResult() {
  result.value = null
  errorMessage.value = ''
}

// 导出结果
const exportDialogVisible = ref(false)
const exportFormat = ref<'csv' | 'excel' | 'json' | 'sql'>('csv')

function handleExport() {
  if (!result.value || result.value.rows.length === 0) {
    ElMessage.warning(t('query.noDataToExport'))
    return
  }
  exportDialogVisible.value = true
}

async function doExport(format: 'csv' | 'excel' | 'json' | 'sql') {
  if (!result.value) return

  const { exportToCSV, exportToExcel, exportToJSON, exportToSQL } = await import('@/utils/exportUtils')
  
  // 转换为对象数组
  const objectData = result.value.rows.map(row => {
    const obj: Record<string, unknown> = {}
    result.value!.columns.forEach((col, index) => {
      obj[col] = row[index]
    })
    return obj
  })

  // 转换为二维数组
  const arrayData = result.value.rows

  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5)

  try {
    switch (format) {
      case 'csv':
        exportToCSV(arrayData, result.value.columns, `query_result_${timestamp}.csv`)
        break
      case 'excel':
        exportToExcel(arrayData, result.value.columns, `query_result_${timestamp}.xlsx`)
        break
      case 'json':
        exportToJSON(objectData, `query_result_${timestamp}.json`)
        break
      case 'sql':
        // 需要表名，这里使用占位符
        const tableName = connectionStore.currentDatabase ? 'table_name' : 'table_name'
        exportToSQL(arrayData, result.value.columns, tableName, `query_result_${timestamp}.sql`)
        break
    }
    ElMessage.success(t('query.exportSuccess'))
    exportDialogVisible.value = false
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error(t('query.exportFailed'))
  }
}

// ==================== 保存为收藏 ====================

function openSaveFavoriteDialog() {
  if (!connectionStore.currentConnection) return
  const sql = editor?.getModel()?.getValue() ?? ''
  if (!sql.trim()) {
    ElMessage.warning(t('query.enterSqlToFavorite'))
    return
  }
  saveFavoriteSql.value = sql
  saveFavoriteForm.value = { name: '', description: '' }
  saveFavoriteDialogVisible.value = true
}

async function submitSaveFavorite() {
  const name = saveFavoriteForm.value.name.trim()
  if (!name) {
    ElMessage.warning(t('query.enterFavoriteName'))
    return
  }
  if (!saveFavoriteSql.value.trim()) {
    ElMessage.warning(t('query.enterSqlToFavorite'))
    return
  }
  if (!connectionStore.currentConnection) return
  saveFavoriteLoading.value = true
  try {
    await createFavorite({
      connection_id: connectionStore.currentConnection.id,
      database: connectionStore.currentDatabase || undefined,
      name,
      sql_query: saveFavoriteSql.value,
      description: saveFavoriteForm.value.description?.trim() || undefined,
    })
    ElMessage.success(t('query.favorited'))
    saveFavoriteDialogVisible.value = false
  } catch (e: unknown) {
    ElMessage.error((e as Error)?.message ?? t('query.favoriteFailed'))
  } finally {
    saveFavoriteLoading.value = false
  }
}

// ==================== 生命周期 ====================

onMounted(() => {
  nextTick(() => {
    initEditor()
    nextTick(() => {
      applyInitialSqlFromState()
    })
  })
})

onBeforeUnmount(() => {
  destroyEditor()
})

// 组件激活时（keep-alive 缓存后切换回来时）
// watch 会自动处理连接和数据库变化，这里不需要额外逻辑
onActivated(() => {
  applyInitialSqlFromState()
})

watch(uiTheme, () => {
  applyEditorTheme()
})

// 监听连接和数据库变化，清除结果并更新自动补全
watch(
  [
    () => connectionStore.currentConnection?.id,
    () => connectionStore.currentDatabase
  ],
  () => {
    handleClearResult()
    updateCompletionProvider()
    // 切换连接或数据库时，结束事务
    if (inTransaction.value) {
      inTransaction.value = false
    }
  },
  { deep: false }
)

// 从收藏进入且为 SELECT 时，连接/数据库就绪后自动执行
watch(
  [
    () => connectionStore.currentConnection,
    () => connectionStore.currentDatabase,
    pendingAutoExecute
  ],
  () => {
    if (
      pendingAutoExecute.value &&
      connectionStore.currentConnection &&
      connectionStore.currentDatabase
    ) {
      pendingAutoExecute.value = false
      nextTick(() => handleExecute())
    }
  },
  { deep: false }
)
</script>

<style scoped>
.query-page {
  max-width: 1600px;
  margin: 0 auto;
  width: 100%;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  background-color: var(--color-background);
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
}

.empty-icon {
  color: var(--color-text-tertiary);
}

/* ==================== 下拉菜单选中项 ==================== */
:deep(.el-dropdown-menu__item.is-active) {
  color: var(--color-primary);
  background-color: var(--color-primary-soft);
}

/* ==================== 卡片通用样式 ==================== */
.editor-card,
.result-card {
  margin-bottom: var(--spacing-md);
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
}

.editor-card :deep(.el-card__header),
.result-card :deep(.el-card__header) {
  padding: var(--spacing-sm) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.editor-card :deep(.el-card__body) {
  padding: 0;
}

.result-card :deep(.el-card__body) {
  padding: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

.result-tag {
  margin-left: var(--spacing-xs);
}

.header-actions {
  display: flex;
  gap: var(--spacing-xs);
}

/* ==================== 编辑器信息 ==================== */
.editor-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.editor-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.selection-hint {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: var(--font-size-small);
  color: var(--color-primary);
}

/* ==================== 编辑器容器 ==================== */
.editor-container {
  height: 300px;
  border-radius: 0 0 var(--border-radius-large) var(--border-radius-large);
  overflow: hidden;
}

/* ==================== 错误提示 ==================== */
.error-alert {
  margin: var(--spacing-md);
}

/* ==================== 无数据结果 ==================== */
.no-data-result {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xxl);
  text-align: center;
}

.success-icon {
  color: var(--color-success);
}

.success-text {
  margin: var(--spacing-md) 0 var(--spacing-sm);
  font-size: var(--font-size-h3);
  font-weight: 600;
  color: var(--color-text-primary);
}

.affected-rows {
  font-size: var(--font-size-body);
  color: var(--color-text-secondary);
}

.affected-rows strong {
  color: var(--color-primary);
}

/* ==================== 结果表格 ==================== */
.result-table {
  width: 100%;
}

.result-table :deep(.el-table__header th) {
  background-color: var(--color-background-secondary);
  font-weight: 600;
}

.cell-value {
  font-family: var(--font-family-code);
  font-size: 13px;
}

.cell-value.numeric {
  text-align: right;
  display: block;
}

.cell-value.null-value {
  color: var(--color-text-tertiary);
  font-style: italic;
}

.table-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-xl);
  color: var(--color-text-tertiary);
}

.table-empty p {
  margin-top: var(--spacing-sm);
}

/* ==================== 结果底部 ==================== */
.result-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--spacing-sm) var(--spacing-lg);
  border-top: 1px solid var(--color-border);
  background-color: var(--color-background-secondary);
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

.result-stats {
  font-weight: 500;
}

.execution-time {
  color: var(--color-text-tertiary);
}

/* ==================== 执行计划对话框 ==================== */
.explain-result {
  padding: var(--spacing-md) 0;
}

.explain-footer {
  margin-top: var(--spacing-md);
  padding-top: var(--spacing-md);
  border-top: 1px solid var(--color-border);
  text-align: right;
  color: var(--color-text-secondary);
  font-size: var(--font-size-small);
}

/* ==================== 响应式 ==================== */
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .editor-actions {
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
  }
  
  .editor-info {
    flex-wrap: wrap;
    margin-left: 0 !important;
    margin-top: var(--spacing-sm);
  }
  
  .result-footer {
    flex-direction: column;
    gap: var(--spacing-xs);
    text-align: center;
  }
}
</style>
