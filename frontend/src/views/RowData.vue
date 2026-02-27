<template>
  <div class="rowdata-page">
    <!-- 页面标题 -->
    <PageHeader 
      :title="pageTitle" 
      :description="pageDescription"
    >
      <template #actions>
        <el-button @click="handleRefresh" :loading="loadingData">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <el-dropdown @command="doExport" trigger="click">
          <el-button type="primary" :disabled="!currentTableName || rows.length === 0">
            <el-icon><Download /></el-icon>
            导出
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="csv">导出为 CSV</el-dropdown-item>
              <el-dropdown-item command="excel">导出为 Excel</el-dropdown-item>
              <el-dropdown-item command="json">导出为 JSON</el-dropdown-item>
              <el-dropdown-item command="sql">导出为 SQL</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button @click="handleImport" :disabled="!currentTableName">
          <el-icon><Upload /></el-icon>
          导入
        </el-button>
      </template>
    </PageHeader>

    <!-- 未选择表时的提示 -->
    <div v-if="!currentTableName" class="empty-state">
      <el-empty description="请从左侧选择一个表查看数据">
        <template #image>
          <el-icon :size="64" class="empty-icon"><Grid /></el-icon>
        </template>
      </el-empty>
    </div>

    <!-- 表数据内容 -->
    <template v-else>
      <!-- 筛选表单区域 -->
      <el-card class="filter-card">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><Filter /></el-icon>
              筛选条件
              <el-tag v-if="filterConditions.length > 0" size="small" type="primary" class="condition-count-tag">
                {{ filterConditions.length }} 个条件
              </el-tag>
            </span>
            <el-button link type="primary" @click="toggleFilterExpand">
              {{ filterExpanded ? '收起' : '展开' }}
              <el-icon>
                <ArrowUp v-if="filterExpanded" />
                <ArrowDown v-else />
              </el-icon>
            </el-button>
          </div>
        </template>

        <el-collapse-transition>
          <div v-show="filterExpanded" class="filter-content">
            <!-- 筛选条件列表 -->
            <div class="filter-conditions">
              <TransitionGroup name="filter-list">
                <div 
                  v-for="(condition, index) in filterConditions" 
                  :key="condition.id"
                  class="filter-row"
                >
                  <!-- 字段选择 -->
                  <el-select 
                    v-model="condition.field"
                    placeholder="选择字段"
                    filterable
                    class="field-select"
                    @change="handleFieldChange(condition)"
                  >
                    <el-option-group label="索引字段" v-if="indexFields.length > 0">
                      <el-option 
                        v-for="field in indexFields" 
                        :key="field" 
                        :label="field" 
                        :value="field"
                      >
                        <span class="field-option">
                          <el-icon class="index-icon"><Connection /></el-icon>
                          {{ field }}
                        </span>
                      </el-option>
                    </el-option-group>
                    <el-option-group label="其他字段">
                      <el-option 
                        v-for="col in nonIndexColumns" 
                        :key="col.name" 
                        :label="col.name" 
                        :value="col.name"
                      >
                        <span class="field-option">
                          <el-icon v-if="col.isPrimary" class="key-icon"><Key /></el-icon>
                          {{ col.name }}
                          <el-tag size="small" type="info" class="type-tag">{{ col.type }}</el-tag>
                        </span>
                      </el-option>
                    </el-option-group>
                  </el-select>

                  <!-- 运算符选择 -->
                  <el-select 
                    v-model="condition.operator"
                    placeholder="运算符"
                    class="operator-select"
                  >
                    <el-option 
                      v-for="op in operatorOptions" 
                      :key="op.value" 
                      :label="op.label" 
                      :value="op.value"
                    />
                  </el-select>

                  <!-- 值输入 -->
                  <el-input 
                    v-if="!isNullOperator(condition.operator)"
                    v-model="condition.value"
                    :placeholder="getValuePlaceholder(condition)"
                    class="value-input"
                    clearable
                    @keyup.enter="handleSearch"
                  />
                  <div v-else class="null-placeholder">
                    <span>无需输入值</span>
                  </div>

                  <!-- 删除按钮 -->
                  <el-button 
                    type="danger" 
                    :icon="Delete" 
                    circle 
                    size="small"
                    @click="removeCondition(index)"
                  />
                </div>
              </TransitionGroup>

              <!-- 空状态提示 -->
              <div v-if="filterConditions.length === 0" class="empty-conditions">
                <el-icon><InfoFilled /></el-icon>
                <span>暂无筛选条件，点击下方按钮添加</span>
              </div>
            </div>

            <!-- 操作按钮栏 -->
            <div class="filter-actions">
              <div class="left-actions">
                <el-button @click="addCondition" :icon="Plus">
                  添加条件
                </el-button>
                <el-dropdown @command="handleQuickAdd" v-if="indexFields.length > 0">
                  <el-button>
                    快速添加
                    <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item 
                        v-for="field in indexFields" 
                        :key="field" 
                        :command="field"
                      >
                        <el-icon><Connection /></el-icon>
                        {{ field }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
              <div class="right-actions">
                <el-button @click="handleReset" :disabled="filterConditions.length === 0">
                  <el-icon><RefreshLeft /></el-icon>
                  重置
                </el-button>
                <el-button type="primary" @click="handleSearch" :loading="loadingData">
                  <el-icon><Search /></el-icon>
                  查询
                </el-button>
              </div>
            </div>
          </div>
        </el-collapse-transition>
      </el-card>

      <!-- 数据表格区域 -->
      <el-card class="data-card">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><List /></el-icon>
              数据记录
              <el-tag size="small" type="success" class="total-tag">
                共 {{ formatNumber(pagination.total) }} 条
              </el-tag>
            </span>
            <div class="header-actions">
              <el-button-group>
                <el-button size="small" @click="handleBatchUpdate" :disabled="filterConditions.length === 0">
                  <el-icon><Edit /></el-icon>
                  批量更新
                </el-button>
                <el-button size="small" type="danger" @click="handleBatchDelete" :disabled="filterConditions.length === 0">
                  <el-icon><Delete /></el-icon>
                  批量删除
                </el-button>
              </el-button-group>
              <el-button-group style="margin-left: 8px;">
                <el-button size="small" @click="handleCompareData">
                  <el-icon><DocumentCopy /></el-icon>
                  数据对比
                </el-button>
                <el-button size="small" @click="handleFindReplace">
                  <el-icon><Search /></el-icon>
                  查找替换
                </el-button>
              </el-button-group>
              <el-tooltip content="列设置" placement="top">
                <el-button :icon="Setting" circle size="small" @click="showColumnSettings = true" />
              </el-tooltip>
            </div>
          </div>
        </template>

        <el-table 
          :data="rows" 
          v-loading="loadingData"
          stripe
          border
          class="data-table"
          :max-height="tableMaxHeight"
          row-key="id"
          @sort-change="handleSortChange"
        >
          <!-- 序号列 -->
          <el-table-column type="index" label="#" width="60" fixed="left" :index="getRowIndex" />

          <!-- 动态数据列 -->
          <el-table-column 
            v-for="col in visibleColumns" 
            :key="col.name"
            :prop="col.name"
            :label="col.name"
            :min-width="getColumnWidth(col)"
            :sortable="col.isIndex ? 'custom' : false"
            show-overflow-tooltip
          >
            <template #header>
              <span class="column-header">
                <el-icon v-if="col.isPrimary" class="key-icon"><Key /></el-icon>
                <el-icon v-else-if="col.isIndex" class="index-icon"><Connection /></el-icon>
                {{ col.name }}
              </span>
            </template>
            <template #default="{ row, $index }">
              <!-- 行内编辑模式 -->
              <div 
                v-if="isInlineEditing(row, col.name)"
                class="inline-edit-wrapper"
              >
                <el-input
                  ref="inlineInputRef"
                  v-model="editingValue"
                  size="small"
                  :autofocus="true"
                  @blur="handleInlineEditConfirm(row, col)"
                  @keyup.enter="handleInlineEditConfirm(row, col)"
                  @keyup.escape="cancelEdit"
                />
              </div>
              <!-- 正常显示模式 -->
              <div 
                v-else
                class="cell-content"
                :class="{ editable: !col.isPrimary }"
                @dblclick="handleCellDblClick(row, col, $index)"
              >
                <span :class="getCellClass(col, row[col.name])">
                  {{ formatCellValue(col, row[col.name]) }}
                </span>
                <el-icon v-if="!col.isPrimary" class="edit-hint"><Edit /></el-icon>
              </div>
            </template>
          </el-table-column>

          <!-- 空状态 -->
          <template #empty>
            <div class="table-empty">
              <el-icon :size="48"><DocumentRemove /></el-icon>
              <p>暂无数据</p>
            </div>
          </template>
        </el-table>

        <!-- 分页组件 -->
        <div class="pagination-wrapper">
          <el-pagination
            v-model:current-page="pagination.page"
            v-model:page-size="pagination.pageSize"
            :page-sizes="[50, 100, 200, 500]"
            :total="pagination.total"
            :disabled="loadingData"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="handlePageSizeChange"
            @current-change="handlePageChange"
          />
        </div>
      </el-card>
    </template>

    <!-- 列设置抽屉 -->
    <el-drawer
      v-model="showColumnSettings"
      title="列显示设置"
      direction="rtl"
      size="300px"
    >
      <div class="column-settings">
        <el-checkbox
          v-model="selectAllColumns"
          :indeterminate="isIndeterminate"
          @change="handleSelectAllColumns"
        >
          全选
        </el-checkbox>
        <el-divider />
        <el-checkbox-group v-model="selectedColumns" class="column-list">
          <el-checkbox 
            v-for="col in columns" 
            :key="col.name" 
            :label="col.name"
            :value="col.name"
          >
            <span class="column-checkbox-label">
              <el-icon v-if="col.isPrimary" class="key-icon"><Key /></el-icon>
              <el-icon v-else-if="col.isIndex" class="index-icon"><Connection /></el-icon>
              {{ col.name }}
              <el-tag size="small" type="info">{{ col.type }}</el-tag>
            </span>
          </el-checkbox>
        </el-checkbox-group>
      </div>
    </el-drawer>

    <!-- 导入对话框 -->
    <DataImportDialog
      v-model="importDialogVisible"
      :connection-id="connectionStore.currentConnection?.id || 0"
      :database="connectionStore.currentDatabase || ''"
      :table="currentTableName || ''"
      :db-type="connectionStore.currentConnection?.db_type || 'mysql'"
      @success="handleImportSuccess"
    />

    <!-- 批量更新对话框 -->
    <el-dialog
      v-model="batchUpdateDialogVisible"
      title="批量更新"
      width="600px"
      destroy-on-close
    >
      <el-form :model="batchUpdateForm" label-width="100px">
        <el-form-item
          v-for="col in columns.filter(c => !c.isPrimary)"
          :key="col.name"
          :label="col.name"
        >
          <el-input
            v-model="batchUpdateForm[col.name]"
            :placeholder="`输入 ${col.name} 的新值（留空则不更新此字段）`"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchUpdateDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="doBatchUpdate">确定</el-button>
      </template>
    </el-dialog>

    <!-- 数据对比对话框 -->
    <el-dialog
      v-model="compareDialogVisible"
      title="数据对比"
      width="700px"
      destroy-on-close
    >
      <el-form :model="compareForm" label-width="120px">
        <el-form-item label="对比数据库">
          <el-input v-model="compareForm.database2" placeholder="留空则使用当前数据库" />
        </el-form-item>
        <el-form-item label="对比表">
          <el-select v-model="compareForm.table2" placeholder="选择要对比的表" filterable>
            <el-option
              v-for="table in connectionStore.tables"
              :key="table.name"
              :label="table.name"
              :value="table.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="主键字段">
          <el-select
            v-model="compareForm.keyColumns"
            multiple
            placeholder="选择用于对比的主键字段"
          >
            <el-option
              v-for="col in columns.filter(c => c.isPrimary || c.isIndex)"
              :key="col.name"
              :label="col.name"
              :value="col.name"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="compareDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="doCompareData">开始对比</el-button>
      </template>
    </el-dialog>

    <!-- 查找替换对话框 -->
    <el-dialog
      v-model="findReplaceDialogVisible"
      title="查找替换"
      width="600px"
      destroy-on-close
    >
      <el-form :model="findReplaceForm" label-width="100px">
        <el-form-item label="字段">
          <el-select v-model="findReplaceForm.column" placeholder="选择要查找替换的字段" filterable>
            <el-option
              v-for="col in columns.filter(c => !c.isPrimary)"
              :key="col.name"
              :label="col.name"
              :value="col.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="查找值">
          <el-input v-model="findReplaceForm.findValue" placeholder="输入要查找的值" />
        </el-form-item>
        <el-form-item label="替换为">
          <el-input v-model="findReplaceForm.replaceValue" placeholder="输入替换后的值" />
        </el-form-item>
        <el-alert
          type="warning"
          :closable="false"
          show-icon
        >
          此操作将更新所有符合当前筛选条件的记录中匹配的字段值
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="findReplaceDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="doFindReplace">确定</el-button>
      </template>
    </el-dialog>

    <!-- 长文本编辑对话框 -->
    <el-dialog
      v-model="showTextEditDialog"
      :title="`编辑字段: ${editingColumn?.name || ''}`"
      width="600px"
      :close-on-click-modal="false"
      @closed="handleDialogClosed"
    >
      <div class="text-edit-dialog">
        <div class="field-info">
          <el-tag type="info" size="small">{{ editingColumn?.type }}</el-tag>
          <span class="char-count">{{ editingValue.length }} 字符</span>
        </div>
        <el-input
          ref="textareaRef"
          v-model="editingValue"
          type="textarea"
          :rows="12"
          :placeholder="`输入 ${editingColumn?.name || ''} 的值`"
          resize="vertical"
        />
        <div class="edit-actions">
          <el-button size="small" @click="setEditingValueNull">
            设为 NULL
          </el-button>
          <el-button size="small" @click="setEditingValueEmpty">
            设为空字符串
          </el-button>
        </div>
      </div>
      <template #footer>
        <el-button @click="cancelEdit">取消</el-button>
        <el-button type="primary" @click="handleTextEditConfirm">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onActivated, reactive, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Grid, 
  List, 
  Refresh, 
  Download, 
  Upload,
  Filter, 
  Search, 
  RefreshLeft,
  Key,
  Connection,
  Setting,
  DocumentRemove,
  ArrowUp,
  ArrowDown,
  InfoFilled,
  Plus,
  Delete,
  Edit,
  DocumentCopy
} from '@element-plus/icons-vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import DataImportDialog from '@/components/import/DataImportDialog.vue'
import { useConnectionStore } from '@/store/connection'
import { useUrlState } from '@/composables/useUrlState'
import { getTableIndexes, getTableColumns, type ColumnInfo } from '@/api/schema'
import { 
  getTableData, 
  updateTableRow,
  batchUpdateData,
  batchDeleteData,
  compareData,
  findReplaceData,
  extractIndexFields,
  type ColumnDefFrontend as ColumnDef,
  type FilterCondition as ApiFilterCondition,
  type RowDataResult,
  type BatchUpdateRequest,
  type BatchDeleteRequest,
  type CompareDataRequest,
  type FindReplaceRequest
} from '@/api/rowdata'

// ==================== 类型定义 ====================

/** 筛选条件 */
interface FilterCondition {
  id: string
  field: string
  operator: string
  value: string
}

/** 运算符选项 */
interface OperatorOption {
  value: string
  label: string
}

// ==================== 常量 ====================

/** 运算符选项列表 */
const operatorOptions: OperatorOption[] = [
  { value: '=', label: '等于 (=)' },
  { value: '!=', label: '不等于 (!=)' },
  { value: '<', label: '小于 (<)' },
  { value: '<=', label: '小于等于 (<=)' },
  { value: '>', label: '大于 (>)' },
  { value: '>=', label: '大于等于 (>=)' },
  { value: 'LIKE', label: '包含 (LIKE)' },
  { value: 'NOT LIKE', label: '不包含 (NOT LIKE)' },
  { value: 'IN', label: '在列表中 (IN)' },
  { value: 'NOT IN', label: '不在列表中 (NOT IN)' },
  { value: 'IS NULL', label: '为空 (IS NULL)' },
  { value: 'IS NOT NULL', label: '不为空 (IS NOT NULL)' },
]

// ==================== 组合式函数 ====================

const route = useRoute()
const connectionStore = useConnectionStore()
const { restoreFromUrl, getUrlParams } = useUrlState()

// ==================== 状态 ====================

// 是否已完成初始化
const initialized = ref(false)

// 缓存上次加载的参数，用于避免不必要的重新加载
const lastLoadParams = ref({
  connectionId: null as number | null,
  database: null as string | null,
  table: null as string | null
})

// 当前表名（从 URL 获取）
const currentTableName = computed(() => {
  return (route.query.table as string) || null
})

// 页面标题和描述
const pageTitle = computed(() => {
  if (!currentTableName.value) return '表数据'
  return `表数据: ${currentTableName.value}`
})

const pageDescription = computed(() => {
  if (!currentTableName.value) return '请先选择一个表'
  return `查看和筛选 ${currentTableName.value} 表的数据记录`
})

// 数据加载状态
const loadingData = ref(false)
const loadingIndexes = ref(false)

// 索引字段
const indexFields = ref<string[]>([])

// 列信息
const columns = ref<ColumnDef[]>([])
const tableColumns = ref<ColumnInfo[]>([])

// 非索引列（用于字段选择器）
const nonIndexColumns = computed(() => {
  return columns.value.filter(col => !indexFields.value.includes(col.name))
})

// 数据行
const rows = ref<Record<string, unknown>[]>([])

// 筛选条件
const filterConditions = ref<FilterCondition[]>([])
const filterExpanded = ref(true)

// 分页
const pagination = reactive({
  page: 1,
  pageSize: 100,
  total: 0
})

// 排序
const sortState = reactive({
  prop: '',
  order: '' as '' | 'ascending' | 'descending'
})

// 列显示设置
const showColumnSettings = ref(false)
const selectedColumns = ref<string[]>([])

// 单元格编辑状态
const editingCell = ref<{ rowIndex: number; columnName: string } | null>(null)
const editingRow = ref<Record<string, unknown> | null>(null)
const editingColumn = ref<ColumnDef | null>(null)
const editingValue = ref<string>('')
const editingOriginalValue = ref<unknown>(null)
const showTextEditDialog = ref(false)
const inlineInputRef = ref()
const textareaRef = ref()

// 长文本阈值（超过此长度使用对话框编辑）
const LONG_TEXT_THRESHOLD = 50

// ==================== 计算属性 ====================

// 计算可见列
const visibleColumns = computed(() => {
  if (selectedColumns.value.length === 0) {
    return columns.value
  }
  return columns.value.filter(col => selectedColumns.value.includes(col.name))
})

// 全选列状态
const selectAllColumns = computed({
  get: () => selectedColumns.value.length === columns.value.length,
  set: (val) => {
    selectedColumns.value = val ? columns.value.map(c => c.name) : []
  }
})

const isIndeterminate = computed(() => {
  return selectedColumns.value.length > 0 && selectedColumns.value.length < columns.value.length
})

// 表格最大高度
const tableMaxHeight = computed(() => {
  return window.innerHeight - 400
})

// ==================== 工具函数 ====================

// 生成唯一 ID
function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2)
}

// 格式化数字
function formatNumber(num: number): string {
  return num.toLocaleString()
}

// 获取行索引
function getRowIndex(index: number): number {
  return (pagination.page - 1) * pagination.pageSize + index + 1
}

// 获取列宽度
function getColumnWidth(col: ColumnDef): number {
  const type = col.type.toLowerCase()
  if (type.includes('text') || type.includes('json')) {
    return 200
  }
  if (type.includes('datetime') || type.includes('timestamp')) {
    return 180
  }
  if (type.includes('date')) {
    return 120
  }
  if (col.name.length > 15) {
    return 180
  }
  return 140
}

// 获取单元格样式类
function getCellClass(col: ColumnDef, value: unknown): string {
  const classes: string[] = ['cell-value']
  const type = col.type.toLowerCase()
  
  if (type.includes('int') || type.includes('decimal') || type.includes('float') || type.includes('double') || type.includes('numeric')) {
    classes.push('numeric')
  }
  if (value === null || value === undefined || value === '') {
    classes.push('null-value')
  }
  if (col.isPrimary) {
    classes.push('primary-value')
  }
  
  return classes.join(' ')
}

// 格式化单元格值
function formatCellValue(col: ColumnDef, value: unknown): string {
  if (value === null || value === undefined) {
    return 'NULL'
  }
  if (value === '') {
    return '(空)'
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

// 判断是否为空值运算符
function isNullOperator(operator: string): boolean {
  return operator === 'IS NULL' || operator === 'IS NOT NULL'
}

// 获取值输入框的 placeholder
function getValuePlaceholder(condition: FilterCondition): string {
  if (condition.operator === 'LIKE' || condition.operator === 'NOT LIKE') {
    return '输入关键词（自动添加 %）'
  }
  if (condition.operator === 'IN' || condition.operator === 'NOT IN') {
    return '多个值用逗号分隔'
  }
  return `输入 ${condition.field || '值'}`
}

// 判断是否为长文本类型
function isLongTextType(col: ColumnDef): boolean {
  const type = col.type.toLowerCase()
  return type.includes('text') || type.includes('json') || type.includes('blob')
}

// 判断值是否为长字符串
function isLongValue(value: unknown): boolean {
  if (value === null || value === undefined) return false
  const strValue = typeof value === 'object' ? JSON.stringify(value) : String(value)
  return strValue.length > LONG_TEXT_THRESHOLD
}

// 判断是否需要使用对话框编辑
function shouldUseDialogEdit(col: ColumnDef, value: unknown): boolean {
  return isLongTextType(col) || isLongValue(value)
}

// 判断当前单元格是否处于行内编辑状态
function isInlineEditing(row: Record<string, unknown>, columnName: string): boolean {
  if (!editingCell.value || showTextEditDialog.value) return false
  return editingRow.value === row && editingCell.value.columnName === columnName
}

// ==================== 单元格编辑操作 ====================

// 双击单元格开始编辑
function handleCellDblClick(row: Record<string, unknown>, col: ColumnDef, rowIndex: number) {
  // 主键不允许编辑
  if (col.isPrimary) {
    ElMessage.warning('主键字段不允许编辑')
    return
  }

  const value = row[col.name]
  
  // 保存原始值
  editingOriginalValue.value = value
  editingRow.value = row
  editingColumn.value = col
  editingCell.value = { rowIndex, columnName: col.name }
  
  // 格式化编辑值
  if (value === null || value === undefined) {
    editingValue.value = ''
  } else if (typeof value === 'object') {
    editingValue.value = JSON.stringify(value, null, 2)
  } else {
    editingValue.value = String(value)
  }

  // 判断使用行内编辑还是对话框编辑
  if (shouldUseDialogEdit(col, value)) {
    showTextEditDialog.value = true
    // 聚焦 textarea
    nextTick(() => {
      textareaRef.value?.focus()
    })
  } else {
    // 行内编辑，聚焦输入框
    nextTick(() => {
      inlineInputRef.value?.focus()
    })
  }
}

// 取消编辑
function cancelEdit() {
  editingCell.value = null
  editingRow.value = null
  editingColumn.value = null
  editingValue.value = ''
  editingOriginalValue.value = null
  showTextEditDialog.value = false
}

// 行内编辑确认
function handleInlineEditConfirm(row: Record<string, unknown>, col: ColumnDef) {
  applyEdit(row, col)
}

// 对话框编辑确认
function handleTextEditConfirm() {
  if (editingRow.value && editingColumn.value) {
    applyEdit(editingRow.value, editingColumn.value)
  }
  showTextEditDialog.value = false
}

// 对话框关闭处理
function handleDialogClosed() {
  // 如果没有通过确认按钮关闭，恢复原值
  if (editingRow.value && editingColumn.value) {
    editingRow.value[editingColumn.value.name] = editingOriginalValue.value
  }
  cancelEdit()
}

// 应用编辑
async function applyEdit(row: Record<string, unknown>, col: ColumnDef) {
  const newValue = editingValue.value
  const originalValue = editingOriginalValue.value
  
  // 检查值是否有变化
  const originalStr = originalValue === null || originalValue === undefined 
    ? '' 
    : typeof originalValue === 'object' 
      ? JSON.stringify(originalValue)
      : String(originalValue)
  
  if (newValue !== originalStr) {
    // 构建主键
    const primaryKey: Record<string, unknown> = {}
    columns.value.forEach(c => {
      if (c.isPrimary) {
        primaryKey[c.name] = row[c.name]
      }
    })
    
    // 检查是否有主键
    if (Object.keys(primaryKey).length === 0) {
      ElMessage.error('无法更新：表没有主键')
      cancelEdit()
      return
    }
    
    // 构建更新数据
    const updateData: Record<string, unknown> = {
      [col.name]: newValue === '' ? null : newValue
    }
    
    try {
      // 调用后端 API 更新数据
      if (!connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
        ElMessage.error('连接信息不完整')
        cancelEdit()
        return
      }
      
      const affectedRows = await updateTableRow(
        connectionStore.currentConnection.id,
        connectionStore.currentDatabase,
        currentTableName.value,
        primaryKey,
        updateData
      )
      
      if (affectedRows > 0) {
        // 更新本地行数据
        row[col.name] = newValue === '' ? null : newValue
        ElMessage.success(`字段 "${col.name}" 已更新`)
      } else {
        ElMessage.warning('未找到要更新的记录')
      }
    } catch (error) {
      console.error('更新数据失败:', error)
      ElMessage.error('更新数据失败: ' + (error instanceof Error ? error.message : String(error)))
    }
  }
  
  cancelEdit()
}

// 设置编辑值为 NULL
function setEditingValueNull() {
  editingValue.value = ''
}

// 设置编辑值为空字符串
function setEditingValueEmpty() {
  editingValue.value = ''
}

// ==================== 筛选条件操作 ====================

// 切换筛选展开状态
function toggleFilterExpand() {
  filterExpanded.value = !filterExpanded.value
}

// 添加筛选条件
function addCondition(field?: string) {
  filterConditions.value.push({
    id: generateId(),
    field: field || '',
    operator: '=',
    value: ''
  })
}

// 移除筛选条件
function removeCondition(index: number) {
  filterConditions.value.splice(index, 1)
}

// 字段变更处理
function handleFieldChange(condition: FilterCondition) {
  // 重置值
  condition.value = ''
}

// 快速添加索引字段
function handleQuickAdd(field: string) {
  // 检查是否已添加
  const exists = filterConditions.value.some(c => c.field === field)
  if (!exists) {
    addCondition(field)
  } else {
    ElMessage.warning(`字段 "${field}" 已在筛选条件中`)
  }
}

// 全选列处理
function handleSelectAllColumns(val: boolean) {
  selectedColumns.value = val ? columns.value.map(c => c.name) : []
}

// ==================== 数据加载 ====================

// 检查参数是否变化
function shouldReload(): boolean {
  const current = {
    connectionId: connectionStore.currentConnection?.id || null,
    database: connectionStore.currentDatabase || null,
    table: currentTableName.value
  }
  
  const changed = 
    current.connectionId !== lastLoadParams.value.connectionId ||
    current.database !== lastLoadParams.value.database ||
    current.table !== lastLoadParams.value.table
    
  if (changed) {
    lastLoadParams.value = { ...current }
  }
  
  return changed
}

// 加载索引字段
async function loadIndexFields() {
  if (!currentTableName.value || !connectionStore.currentConnection || !connectionStore.currentDatabase) {
    indexFields.value = []
    return
  }

  loadingIndexes.value = true
  try {
    const indexes = await getTableIndexes(
      connectionStore.currentConnection.id,
      connectionStore.currentDatabase,
      currentTableName.value
    )
    indexFields.value = extractIndexFields(indexes)
  } catch (error) {
    console.error('加载索引字段失败:', error)
    indexFields.value = []
  } finally {
    loadingIndexes.value = false
  }
}

// 加载列信息
async function loadColumnInfo() {
  if (!currentTableName.value || !connectionStore.currentConnection || !connectionStore.currentDatabase) {
    tableColumns.value = []
    return
  }

  try {
    tableColumns.value = await getTableColumns(
      connectionStore.currentConnection.id,
      connectionStore.currentDatabase,
      currentTableName.value
    )
  } catch (error) {
    console.error('加载列信息失败:', error)
    tableColumns.value = []
  }
}

// 构建筛选条件数组（用于后端 API）
function buildFilterArray(): ApiFilterCondition[] {
  const filters: ApiFilterCondition[] = []
  
  filterConditions.value.forEach(condition => {
    if (!condition.field) return
    
    // 空值运算符不需要值
    if (isNullOperator(condition.operator)) {
      filters.push({
        field: condition.field,
        operator: condition.operator,
        value: ''
      })
      return
    }
    
    // 其他运算符需要值
    if (condition.value === '' || condition.value === null || condition.value === undefined) {
      return
    }
    
    filters.push({
      field: condition.field,
      operator: condition.operator,
      value: condition.value
    })
  })
  
  return filters
}

// 加载表数据
async function loadTableData() {
  if (!initialized.value) return
  
  if (!currentTableName.value || !connectionStore.currentConnection || !connectionStore.currentDatabase) {
    rows.value = []
    columns.value = []
    pagination.total = 0
    return
  }

  loadingData.value = true
  try {
    const filters = buildFilterArray()

    const result: RowDataResult = await getTableData(
      connectionStore.currentConnection.id,
      connectionStore.currentDatabase,
      currentTableName.value,
      filters,
      { page: pagination.page, pageSize: pagination.pageSize }
    )

    columns.value = result.columns
    rows.value = result.rows
    pagination.total = result.total

    // 初始化选中的列
    if (selectedColumns.value.length === 0) {
      selectedColumns.value = result.columns.map(c => c.name)
    }
  } catch (error) {
    console.error('加载表数据失败:', error)
    ElMessage.error('加载表数据失败')
    rows.value = []
    columns.value = []
    pagination.total = 0
  } finally {
    loadingData.value = false
  }
}

// 加载所有数据（索引字段、列信息、表数据）
async function loadAllData(skipCheck = false) {
  // 保存旧的表名，用于检测表名是否变化
  const oldTable = lastLoadParams.value.table
  
  // 检查参数是否变化，如果没变化则不重新加载
  if (!skipCheck && !shouldReload()) {
    return
  }
  
  // 如果表名变化，重置筛选和分页
  if (oldTable !== currentTableName.value) {
    filterConditions.value = []
    pagination.page = 1
    selectedColumns.value = []
  }
  
  await loadIndexFields()
  await loadColumnInfo()
  await loadTableData()
}

// 初始化：从 URL 恢复状态
async function initFromUrl(forceReload = false) {
  const { connectionId } = getUrlParams()
  
  if (connectionId && !connectionStore.currentConnection) {
    try {
      await restoreFromUrl()
    } catch (error) {
      console.error('恢复连接状态失败:', error)
    }
  }
  
  initialized.value = true
  
  // 检查参数是否变化
  const currentParams = {
    connectionId: connectionStore.currentConnection?.id || null,
    database: connectionStore.currentDatabase || null,
    table: currentTableName.value
  }
  
  // 如果参数没变化且不是强制加载，则不重新加载
  if (!forceReload && 
      currentParams.connectionId === lastLoadParams.value.connectionId &&
      currentParams.database === lastLoadParams.value.database &&
      currentParams.table === lastLoadParams.value.table) {
    return
  }
  
  // 更新参数缓存
  lastLoadParams.value = { ...currentParams }
  
  // 加载索引字段和数据（跳过检查，因为已经在 initFromUrl 中检查过了）
  await loadAllData(true)
}

// ==================== 事件处理 ====================

// 搜索
function handleSearch() {
  pagination.page = 1
  loadTableData()
}

// 重置筛选
function handleReset() {
  filterConditions.value = []
  pagination.page = 1
  loadTableData()
}

// 刷新
async function handleRefresh() {
  await loadIndexFields()
  await loadColumnInfo()
  await loadTableData()
}

// 导出
const exportDialogVisible = ref(false)

async function handleExport() {
  if (!currentTableName.value || rows.value.length === 0) {
    ElMessage.warning('没有可导出的数据')
    return
  }
  exportDialogVisible.value = true
}

async function doExport(format: 'csv' | 'excel' | 'json' | 'sql') {
  if (rows.value.length === 0) return

  const { exportToCSV, exportToExcel, exportToJSON, exportToSQL } = await import('@/utils/exportUtils')
  
  // 转换为二维数组
  const arrayData = rows.value.map(row => 
    columns.value.map(col => row[col.name] ?? null)
  )
  const headers = columns.value.map(col => col.name)

  // 转换为对象数组
  const objectData = rows.value

  const timestamp = new Date().toISOString().replace(/[:.]/g, '-').slice(0, -5)
  const tableName = currentTableName.value || 'table'

  try {
    switch (format) {
      case 'csv':
        exportToCSV(arrayData, headers, `${tableName}_${timestamp}.csv`)
        break
      case 'excel':
        exportToExcel(arrayData, headers, `${tableName}_${timestamp}.xlsx`)
        break
      case 'json':
        exportToJSON(objectData, `${tableName}_${timestamp}.json`)
        break
      case 'sql':
        exportToSQL(arrayData, headers, tableName, `${tableName}_${timestamp}.sql`)
        break
    }
    ElMessage.success('导出成功')
    exportDialogVisible.value = false
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  }
}

// 导入
const importDialogVisible = ref(false)

function handleImport() {
  if (!currentTableName.value) {
    ElMessage.warning('请先选择表')
    return
  }
  importDialogVisible.value = true
}

function handleImportSuccess() {
  // 重新加载数据
  loadTableData()
}

// 排序变化
function handleSortChange({ prop, order }: { prop: string; order: string | null }) {
  sortState.prop = prop || ''
  sortState.order = (order || '') as '' | 'ascending' | 'descending'
  // Mock 数据不支持排序，仅记录状态
  console.log('Sort changed:', sortState)
}

// 分页大小变化
function handlePageSizeChange() {
  pagination.page = 1
  loadTableData()
}

// 页码变化
function handlePageChange() {
  loadTableData()
}

// 批量更新
const batchUpdateDialogVisible = ref(false)
const batchUpdateForm = ref<Record<string, unknown>>({})

async function handleBatchUpdate() {
  if (filterConditions.value.length === 0) {
    ElMessage.warning('请先设置筛选条件')
    return
  }
  batchUpdateDialogVisible.value = true
}

async function doBatchUpdate() {
  if (!connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
    return
  }

  try {
    const filters = buildFilterArray()
    const result = await batchUpdateData({
      connection_id: connectionStore.currentConnection.id,
      database: connectionStore.currentDatabase,
      table: currentTableName.value,
      filters,
      update_data: batchUpdateForm.value
    } as BatchUpdateRequest)

    ElMessage.success(`批量更新成功，影响 ${result.affected_rows} 行`)
    batchUpdateDialogVisible.value = false
    batchUpdateForm.value = {}
    await loadTableData()
  } catch (error: any) {
    ElMessage.error(error.message || '批量更新失败')
  }
}

// 批量删除
async function handleBatchDelete() {
  if (filterConditions.value.length === 0) {
    ElMessage.warning('请先设置筛选条件')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除符合筛选条件的 ${pagination.total} 条记录吗？此操作不可撤销。`,
      '批量删除确认',
      { type: 'warning' }
    )

    if (!connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
      return
    }

    const filters = buildFilterArray()
    const result = await batchDeleteData({
      connection_id: connectionStore.currentConnection.id,
      database: connectionStore.currentDatabase,
      table: currentTableName.value,
      filters
    } as BatchDeleteRequest)

    ElMessage.success(`批量删除成功，删除 ${result.affected_rows} 行`)
    await loadTableData()
  } catch (error: any) {
    if (error.message !== 'cancel') {
      ElMessage.error(error.message || '批量删除失败')
    }
  }
}

// 数据对比
const compareDialogVisible = ref(false)
const compareForm = ref({
  database2: '',
  table2: '',
  keyColumns: [] as string[]
})

async function handleCompareData() {
  if (!connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
    return
  }
  compareForm.value.database2 = connectionStore.currentDatabase
  compareForm.value.table2 = ''
  compareForm.value.keyColumns = columns.value.filter(c => c.isPrimary).map(c => c.name)
  compareDialogVisible.value = true
}

async function doCompareData() {
  if (!connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
    return
  }

  if (!compareForm.value.table2) {
    ElMessage.warning('请选择要对比的表')
    return
  }

  if (compareForm.value.keyColumns.length === 0) {
    ElMessage.warning('请选择至少一个主键字段')
    return
  }

  try {
    const result = await compareData({
      connection_id: connectionStore.currentConnection.id,
      database1: connectionStore.currentDatabase,
      table1: currentTableName.value,
      database2: compareForm.value.database2 || undefined,
      table2: compareForm.value.table2,
      key_columns: compareForm.value.keyColumns
    } as CompareDataRequest)

    ElMessage.success(
      `对比完成：仅表1有 ${result.only_in_table1.length} 条，仅表2有 ${result.only_in_table2.length} 条，不同 ${result.different.length} 条，相同 ${result.same_count} 条`
    )
    // TODO: 显示对比结果详情
    compareDialogVisible.value = false
  } catch (error: any) {
    ElMessage.error(error.message || '数据对比失败')
  }
}

// 查找替换
const findReplaceDialogVisible = ref(false)
const findReplaceForm = ref({
  column: '',
  findValue: '',
  replaceValue: ''
})

function handleFindReplace() {
  if (!currentTableName.value) {
    ElMessage.warning('请先选择表')
    return
  }
  findReplaceForm.value.column = ''
  findReplaceForm.value.findValue = ''
  findReplaceForm.value.replaceValue = ''
  findReplaceDialogVisible.value = true
}

async function doFindReplace() {
  if (!connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
    return
  }

  if (!findReplaceForm.value.column) {
    ElMessage.warning('请选择要查找替换的字段')
    return
  }

  if (!findReplaceForm.value.findValue) {
    ElMessage.warning('请输入查找值')
    return
  }

  try {
    const result = await findReplaceData({
      connection_id: connectionStore.currentConnection.id,
      database: connectionStore.currentDatabase,
      table: currentTableName.value,
      column: findReplaceForm.value.column,
      find_value: findReplaceForm.value.findValue,
      replace_value: findReplaceForm.value.replaceValue,
      filters: buildFilterArray()
    } as FindReplaceRequest)

    ElMessage.success(`查找替换完成：匹配 ${result.matched_rows} 行，更新 ${result.affected_rows} 行`)
    findReplaceDialogVisible.value = false
    await loadTableData()
  } catch (error: any) {
    ElMessage.error(error.message || '查找替换失败')
  }
}

// ==================== 监听器 ====================

// 合并多个 watch，只在参数真正变化时才加载数据
watch(
  [
    () => connectionStore.currentConnection?.id,
    () => connectionStore.currentDatabase,
    currentTableName
  ],
  () => {
    if (initialized.value) {
      loadAllData()
    }
  },
  { deep: false }
)

// ==================== 生命周期 ====================

// 组件挂载时初始化
onMounted(() => {
  initFromUrl(true)
})

// 组件激活时检查是否需要重新加载（keep-alive 缓存后切换回来时）
onActivated(() => {
  if (initialized.value) {
    // 检查参数是否变化，如果变化则重新加载
    initFromUrl(false)
  }
})
</script>

<style scoped>
.rowdata-page {
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

/* ==================== 筛选卡片 ==================== */
.filter-card {
  margin-bottom: var(--spacing-md);
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
}

.filter-card :deep(.el-card__header) {
  padding: var(--spacing-sm) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.filter-card :deep(.el-card__body) {
  padding: 0;
}

.filter-content {
  padding: var(--spacing-md) var(--spacing-lg);
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

.condition-count-tag,
.total-tag {
  margin-left: var(--spacing-xs);
}

/* ==================== 筛选条件 ==================== */
.filter-conditions {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  margin-bottom: var(--spacing-md);
}

.filter-row {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  background-color: var(--color-background-secondary);
  border-radius: var(--border-radius-medium);
  transition: all 0.2s;
}

.filter-row:hover {
  background-color: var(--color-background-hover);
}

.field-select {
  width: 180px;
  flex-shrink: 0;
}

.operator-select {
  width: 160px;
  flex-shrink: 0;
}

.value-input {
  flex: 1;
  min-width: 200px;
}

.null-placeholder {
  flex: 1;
  min-width: 200px;
  padding: 0 var(--spacing-sm);
  color: var(--color-text-tertiary);
  font-style: italic;
  font-size: 13px;
}

.field-option {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.field-option .type-tag {
  margin-left: auto;
  font-size: 10px;
}

.empty-conditions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-lg);
  color: var(--color-text-tertiary);
  font-size: 13px;
  background-color: var(--color-background-secondary);
  border-radius: var(--border-radius-medium);
  border: 1px dashed var(--color-border);
}

/* ==================== 筛选操作按钮 ==================== */
.filter-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: var(--spacing-sm);
  border-top: 1px solid var(--color-border);
}

.left-actions,
.right-actions {
  display: flex;
  gap: var(--spacing-sm);
}

/* ==================== 列表动画 ==================== */
.filter-list-enter-active,
.filter-list-leave-active {
  transition: all 0.3s ease;
}

.filter-list-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.filter-list-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

.filter-list-move {
  transition: transform 0.3s ease;
}

/* ==================== 数据卡片 ==================== */
.data-card {
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
}

.data-card :deep(.el-card__header) {
  padding: var(--spacing-sm) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.data-card :deep(.el-card__body) {
  padding: 0;
}

.header-actions {
  display: flex;
  gap: var(--spacing-xs);
}

/* ==================== 数据表格 ==================== */
.data-table {
  width: 100%;
}

.data-table :deep(.el-table__header th) {
  background-color: var(--color-background-secondary);
  font-weight: 600;
}

.column-header {
  display: flex;
  align-items: center;
  gap: 4px;
}

.key-icon {
  color: var(--color-warning);
}

.index-icon {
  color: var(--color-primary);
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

.cell-value.primary-value {
  font-weight: 600;
  color: var(--color-primary);
}

/* ==================== 单元格编辑 ==================== */
.cell-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--spacing-xs);
  cursor: default;
  min-height: 24px;
}

.cell-content.editable {
  cursor: pointer;
  border-radius: 2px;
  transition: background-color 0.2s;
}

.cell-content.editable:hover {
  background-color: var(--color-primary-light);
}

.cell-content .edit-hint {
  opacity: 0;
  color: var(--color-text-tertiary);
  font-size: 12px;
  flex-shrink: 0;
  transition: opacity 0.2s;
}

.cell-content.editable:hover .edit-hint {
  opacity: 1;
}

.inline-edit-wrapper {
  min-width: 100px;
}

.inline-edit-wrapper :deep(.el-input__inner) {
  font-family: var(--font-family-code);
  font-size: 13px;
}

/* ==================== 长文本编辑对话框 ==================== */
.text-edit-dialog {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.text-edit-dialog .field-info {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.text-edit-dialog .char-count {
  font-size: 12px;
  color: var(--color-text-tertiary);
}

.text-edit-dialog :deep(.el-textarea__inner) {
  font-family: var(--font-family-code);
  font-size: 13px;
  line-height: 1.6;
}

.text-edit-dialog .edit-actions {
  display: flex;
  gap: var(--spacing-sm);
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

/* ==================== 分页 ==================== */
.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: var(--spacing-md) var(--spacing-lg);
  border-top: 1px solid var(--color-border);
  background-color: var(--color-background);
}

/* ==================== 列设置 ==================== */
.column-settings {
  padding: var(--spacing-md);
}

.column-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.column-list :deep(.el-checkbox) {
  width: 100%;
  margin-right: 0;
}

.column-checkbox-label {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.column-checkbox-label .el-tag {
  margin-left: auto;
}

/* ==================== 响应式 ==================== */
@media (max-width: 1024px) {
  .filter-row {
    flex-wrap: wrap;
  }
  
  .field-select,
  .operator-select {
    width: calc(50% - var(--spacing-sm) / 2);
  }
  
  .value-input,
  .null-placeholder {
    width: 100%;
    min-width: unset;
  }
  
  .filter-actions {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
  
  .left-actions,
  .right-actions {
    width: 100%;
    justify-content: center;
  }
}

@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .field-select,
  .operator-select {
    width: 100%;
  }
}
</style>
