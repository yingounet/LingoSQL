<template>
  <el-dialog
    v-model="dialogVisible"
    :title="t('importData.title')"
    width="800px"
    destroy-on-close
    @close="handleClose"
  >
    <el-steps :active="currentStep" finish-status="success" align-center>
      <el-step :title="t('importData.stepSelectFile')" />
      <el-step :title="t('importData.stepPreview')" />
      <el-step :title="t('importData.stepMapping')" />
      <el-step :title="t('importData.stepSettings')" />
    </el-steps>

    <!-- 步骤1: 选择文件 -->
    <div v-if="currentStep === 0" class="step-content">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :on-change="handleFileChange"
        :file-list="fileList"
        accept=".csv,.xlsx,.xls,.json,.sql"
        drag
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text" v-html="t('importData.dragOrClick')" />
        <template #tip>
          <div class="el-upload__tip">
            {{ t('importData.supportedFormats') }}
          </div>
        </template>
      </el-upload>
    </div>

    <!-- 步骤2: 预览数据 -->
    <div v-if="currentStep === 1" class="step-content">
      <div class="preview-info">
        <span>{{ t('importData.file') }}: {{ selectedFile?.name }}</span>
        <span>{{ t('importData.rowCount') }}: {{ previewData.data.length }}</span>
        <span>{{ t('importData.columnCount') }}: {{ previewData.headers.length }}</span>
      </div>
      <el-table
        :data="previewData.data.slice(0, 10)"
        stripe
        border
        max-height="400"
      >
        <el-table-column
          v-for="(header, index) in previewData.headers"
          :key="index"
          :prop="String(index)"
          :label="header"
          :min-width="120"
          show-overflow-tooltip
        >
          <template #default="{ row }">
            {{ row[index] ?? t('importData.emptyCell') }}
          </template>
        </el-table-column>
      </el-table>
      <div v-if="previewData.data.length > 10" class="preview-more">
        {{ t('importData.previewHint', { n: previewData.data.length }) }}
      </div>
    </div>

    <!-- 步骤3: 字段映射 -->
    <div v-if="currentStep === 2" class="step-content">
      <el-table :data="fieldMappingData" stripe border>
        <el-table-column :label="t('importData.fileColumn')" prop="fileColumn" width="200" />
        <el-table-column :label="t('importData.tableField')" prop="tableColumn" width="200">
          <template #default="{ row }">
            <el-select
              v-model="row.tableColumn"
              :placeholder="t('importData.selectTableField')"
              filterable
              clearable
            >
              <el-option
                v-for="col in tableColumns"
                :key="col.name"
                :label="col.name"
                :value="col.name"
              >
                <span>{{ col.name }}</span>
                <span style="color: #8492a6; font-size: 13px; margin-left: 8px;">
                  {{ col.type }}
                </span>
              </el-option>
            </el-select>
          </template>
        </el-table-column>
        <el-table-column :label="t('importData.dataType')" prop="dataType" width="150">
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ getColumnType(row.tableColumn) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="100">
          <template #default="{ row, $index }">
            <el-button
              link
              type="danger"
              size="small"
              @click="skipColumn($index)"
            >
              {{ t('importData.skip') }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 步骤4: 导入设置 -->
    <div v-if="currentStep === 3" class="step-content">
      <el-form :model="importSettings" label-width="120px">
        <el-form-item :label="t('importData.skipFirstRow')">
          <el-switch v-model="importSettings.skipFirstRow" />
          <div class="form-tip">{{ t('importData.skipFirstRowTip') }}</div>
        </el-form-item>
        <el-form-item :label="t('importData.duplicateHandling')">
          <el-radio-group v-model="importSettings.onDuplicate">
            <el-radio label="ignore">{{ t('importData.duplicateIgnore') }}</el-radio>
            <el-radio label="update">{{ t('importData.duplicateUpdate') }}</el-radio>
            <el-radio label="error">{{ t('importData.duplicateError') }}</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item :label="t('importData.batchSize')">
          <el-input-number
            v-model="importSettings.batchSize"
            :min="1"
            :max="1000"
          />
          <div class="form-tip">{{ t('importData.batchSizeTip') }}</div>
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
        <el-button v-if="currentStep > 0" @click="currentStep--">{{ t('common.prev') }}</el-button>
        <el-button
          v-if="currentStep < 3"
          type="primary"
          @click="handleNext"
          :disabled="!canNext"
        >
          {{ t('common.next') }}
        </el-button>
        <el-button
          v-if="currentStep === 3"
          type="primary"
          @click="handleImport"
          :loading="importing"
        >
          {{ t('importData.startImport') }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'
import { parseCSV, parseExcel, parseJSON, parseSQL } from '@/utils/importUtils'
import { importData } from '@/api/import'
import { getTableColumns, type ColumnInfo } from '@/api/schema'
import type { DbType } from '@/types/connection'

interface Props {
  modelValue: boolean
  connectionId: number
  database: string
  table: string
  dbType?: DbType
}

const props = withDefaults(defineProps<Props>(), {
  dbType: 'mysql'
})
const { t } = useI18n()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'success': []
}>()

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const currentStep = ref(0)
const uploadRef = ref()
const fileList = ref<File[]>([])
const selectedFile = ref<File | null>(null)
const previewData = ref<{ headers: string[]; data: unknown[][] }>({ headers: [], data: [] })
const tableColumns = ref<ColumnInfo[]>([])
const fieldMappingData = ref<Array<{ fileColumn: string; tableColumn: string }>>([])
const importSettings = ref({
  skipFirstRow: true,
  onDuplicate: 'ignore' as 'ignore' | 'update' | 'error',
  batchSize: 100
})
const importing = ref(false)

const canNext = computed(() => {
  if (currentStep.value === 0) return selectedFile.value !== null
  if (currentStep.value === 1) return previewData.value.data.length > 0
  if (currentStep.value === 2) {
    return fieldMappingData.value.some(m => m.tableColumn)
  }
  return true
})

// 加载表字段
async function loadTableColumns() {
  try {
    tableColumns.value = await getTableColumns(
      props.connectionId,
      props.database,
      props.table,
      props.dbType
    )
  } catch (error) {
    console.error('loadTableColumns failed:', error)
    ElMessage.error(t('importData.loadFieldsFailed'))
  }
}

// 文件选择
async function handleFileChange(file: any) {
  selectedFile.value = file.raw
  if (!selectedFile.value) return

  try {
    const fileName = selectedFile.value.name.toLowerCase()
    
    if (fileName.endsWith('.csv')) {
      previewData.value = await parseCSV(selectedFile.value)
    } else if (fileName.endsWith('.xlsx') || fileName.endsWith('.xls')) {
      previewData.value = await parseExcel(selectedFile.value)
    } else if (fileName.endsWith('.json')) {
      previewData.value = await parseJSON(selectedFile.value)
    } else if (fileName.endsWith('.sql')) {
      previewData.value = await parseSQL(selectedFile.value)
    } else {
      ElMessage.error(t('importData.unsupportedFormat'))
      return
    }

    // 初始化字段映射
    fieldMappingData.value = previewData.value.headers.map(header => ({
      fileColumn: header,
      tableColumn: ''
    }))

    // 自动匹配字段名
    await loadTableColumns()
    fieldMappingData.value.forEach(mapping => {
      const match = tableColumns.value.find(
        col => col.name.toLowerCase() === mapping.fileColumn.toLowerCase()
      )
      if (match) {
        mapping.tableColumn = match.name
      }
    })
  } catch (error: any) {
    ElMessage.error(error.message || t('importData.parseFileFailed'))
  }
}

// 获取字段类型
function getColumnType(tableColumn: string): string {
  if (!tableColumn) return '-'
  const col = tableColumns.value.find(c => c.name === tableColumn)
  return col?.type || '-'
}

// 跳过列
function skipColumn(index: number) {
  fieldMappingData.value[index].tableColumn = ''
}

// 下一步
function handleNext() {
  if (currentStep.value === 2) {
    // 检查是否有映射的字段
    const hasMapping = fieldMappingData.value.some(m => m.tableColumn)
    if (!hasMapping) {
      ElMessage.warning(t('importData.mapAtLeastOne'))
      return
    }
  }
  currentStep.value++
}

// 导入
async function handleImport() {
  if (!selectedFile.value) return

  importing.value = true
  try {
    // 构建字段映射
    const fieldMapping: Record<string, string> = {}
    fieldMappingData.value.forEach(m => {
      if (m.tableColumn) {
        fieldMapping[m.fileColumn] = m.tableColumn
      }
    })

    // 准备数据（如果跳过第一行，数据已经去掉了表头）
    let dataToImport = previewData.value.data
    if (importSettings.value.skipFirstRow && previewData.value.headers.length > 0) {
      // 数据已经解析时，第一行可能还在，需要检查
      // 这里假设解析时已经处理了表头
    }

    // 按映射转换数据
    const mappedData = dataToImport.map(row => {
      return previewData.value.headers.map(header => {
        const tableColumn = fieldMapping[header]
        if (!tableColumn) return null
        const index = previewData.value.headers.indexOf(header)
        return row[index]
      })
    })

    // 调用导入 API
    const result = await importData({
      connection_id: props.connectionId,
      database: props.database,
      table: props.table,
      data: mappedData,
      headers: Object.values(fieldMapping),
      field_mapping: fieldMapping,
      skip_first_row: importSettings.value.skipFirstRow,
      on_duplicate: importSettings.value.onDuplicate,
    })

    ElMessage.success(
      t('importData.importComplete', {
        success: result.inserted_rows,
        updated: result.updated_rows,
        errors: result.error_rows,
      })
    )

    emit('success')
    handleClose()
  } catch (error: any) {
    ElMessage.error(error.message || t('importData.importFailed'))
  } finally {
    importing.value = false
  }
}

// 关闭
function handleClose() {
  currentStep.value = 0
  selectedFile.value = null
  previewData.value = { headers: [], data: [] }
  fieldMappingData.value = []
  fileList.value = []
  dialogVisible.value = false
}

// 监听对话框打开，加载表字段
watch(() => props.modelValue, (val) => {
  if (val) {
    loadTableColumns()
  }
})
</script>

<style scoped>
.step-content {
  margin-top: var(--spacing-xl);
  min-height: 300px;
}

.preview-info {
  display: flex;
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-md);
  padding: var(--spacing-sm);
  background-color: var(--color-background-secondary);
  border-radius: var(--border-radius-medium);
  font-size: var(--font-size-small);
}

.preview-more {
  margin-top: var(--spacing-sm);
  text-align: center;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-small);
}

.field-mapping-table {
  margin-top: var(--spacing-md);
}

.form-tip {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin-top: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: var(--spacing-sm);
}
</style>
