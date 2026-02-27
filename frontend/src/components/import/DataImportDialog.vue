<template>
  <el-dialog
    v-model="dialogVisible"
    title="导入数据"
    width="800px"
    destroy-on-close
    @close="handleClose"
  >
    <el-steps :active="currentStep" finish-status="success" align-center>
      <el-step title="选择文件" />
      <el-step title="预览数据" />
      <el-step title="字段映射" />
      <el-step title="导入设置" />
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
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            支持 CSV、Excel、JSON、SQL 文件
          </div>
        </template>
      </el-upload>
    </div>

    <!-- 步骤2: 预览数据 -->
    <div v-if="currentStep === 1" class="step-content">
      <div class="preview-info">
        <span>文件: {{ selectedFile?.name }}</span>
        <span>行数: {{ previewData.data.length }}</span>
        <span>列数: {{ previewData.headers.length }}</span>
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
            {{ row[index] ?? '(空)' }}
          </template>
        </el-table-column>
      </el-table>
      <div v-if="previewData.data.length > 10" class="preview-more">
        仅显示前 10 行，共 {{ previewData.data.length }} 行
      </div>
    </div>

    <!-- 步骤3: 字段映射 -->
    <div v-if="currentStep === 2" class="step-content">
      <el-table :data="fieldMappingData" stripe border>
        <el-table-column label="文件列名" prop="fileColumn" width="200" />
        <el-table-column label="表字段名" prop="tableColumn" width="200">
          <template #default="{ row }">
            <el-select
              v-model="row.tableColumn"
              placeholder="选择表字段"
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
        <el-table-column label="数据类型" prop="dataType" width="150">
          <template #default="{ row }">
            <el-tag size="small" type="info">
              {{ getColumnType(row.tableColumn) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="{ row, $index }">
            <el-button
              link
              type="danger"
              size="small"
              @click="skipColumn($index)"
            >
              跳过
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 步骤4: 导入设置 -->
    <div v-if="currentStep === 3" class="step-content">
      <el-form :model="importSettings" label-width="120px">
        <el-form-item label="跳过第一行">
          <el-switch v-model="importSettings.skipFirstRow" />
          <div class="form-tip">如果文件第一行是表头，请开启此选项</div>
        </el-form-item>
        <el-form-item label="重复数据处理">
          <el-radio-group v-model="importSettings.onDuplicate">
            <el-radio label="ignore">忽略</el-radio>
            <el-radio label="update">更新</el-radio>
            <el-radio label="error">报错</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="批量大小">
          <el-input-number
            v-model="importSettings.batchSize"
            :min="1"
            :max="1000"
          />
          <div class="form-tip">每次导入的行数，建议 100-500</div>
        </el-form-item>
      </el-form>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button v-if="currentStep > 0" @click="currentStep--">上一步</el-button>
        <el-button
          v-if="currentStep < 3"
          type="primary"
          @click="handleNext"
          :disabled="!canNext"
        >
          下一步
        </el-button>
        <el-button
          v-if="currentStep === 3"
          type="primary"
          @click="handleImport"
          :loading="importing"
        >
          开始导入
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
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
    console.error('加载表字段失败:', error)
    ElMessage.error('加载表字段失败')
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
      ElMessage.error('不支持的文件格式')
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
    ElMessage.error(error.message || '解析文件失败')
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
      ElMessage.warning('请至少映射一个字段')
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
      `导入完成：成功 ${result.inserted_rows} 行，更新 ${result.updated_rows} 行，错误 ${result.error_rows} 行`
    )

    emit('success')
    handleClose()
  } catch (error: any) {
    ElMessage.error(error.message || '导入失败')
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
