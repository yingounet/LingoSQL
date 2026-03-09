<template>
  <div class="schema-page">
    <!-- 页面标题 -->
    <PageHeader 
      :title="t('schema.title')" 
      :description="pageDescription"
    />

    <!-- 未选择表时的提示 -->
    <div v-if="!currentTableName" class="empty-state">
      <el-empty :description="t('schema.selectTableHint')">
        <template #image>
          <el-icon :size="64" class="empty-icon"><Grid /></el-icon>
        </template>
      </el-empty>
    </div>

    <!-- 表结构内容 -->
    <template v-else>
      <!-- 字段结构表格 -->
      <el-card class="schema-card">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><List /></el-icon>
              {{ t('schema.columnStructure') }}
              <el-tag size="small" type="info" class="db-type-tag">{{ dbTypeLabel }}</el-tag>
            </span>
            <el-button type="primary" size="small" @click="handleAddColumn">
              <el-icon><Plus /></el-icon>
              {{ t('schema.addColumn') }}
            </el-button>
          </div>
        </template>
        
        <el-table 
          :data="columns" 
          v-loading="loadingColumns"
          stripe
          border
          class="schema-table"
        >
          <!-- 字段名 -->
          <el-table-column prop="name" :label="t('schema.columnName')" min-width="120" fixed="left">
            <template #default="{ row }">
              <span class="column-name">
                <el-icon v-if="row.is_primary" class="key-icon"><Key /></el-icon>
                {{ row.name }}
              </span>
            </template>
          </el-table-column>

          <!-- 类型 -->
          <el-table-column prop="type" :label="t('common.type')" width="150">
            <template #default="{ row }">
              <el-tag size="small" type="info">
                {{ formatColumnType(row) }}
              </el-tag>
            </template>
          </el-table-column>

          <!-- 可空 -->
          <el-table-column prop="nullable" :label="t('schema.nullable')" width="80" align="center">
            <template #default="{ row }">
              <el-tag :type="row.nullable ? 'success' : 'danger'" size="small">
                {{ row.nullable ? 'YES' : 'NO' }}
              </el-tag>
            </template>
          </el-table-column>

          <!-- 默认值 -->
          <el-table-column prop="default_value" :label="t('schema.defaultValue')" width="140">
            <template #default="{ row }">
              <span class="default-value">{{ row.default_value ?? 'NULL' }}</span>
            </template>
          </el-table-column>

          <!-- 主键 -->
          <el-table-column prop="is_primary" :label="t('schema.primaryKey')" width="70" align="center">
            <template #default="{ row }">
              <el-icon v-if="row.is_primary" class="check-icon"><Check /></el-icon>
              <span v-else class="dash">-</span>
            </template>
          </el-table-column>

          <!-- MySQL/MariaDB: 无符号 -->
          <el-table-column 
            v-if="dbType !== 'postgresql'" 
            prop="unsigned" 
            :label="t('schema.unsigned')" 
            width="80" 
            align="center"
          >
            <template #default="{ row }">
              <el-icon v-if="row.unsigned" class="check-icon"><Check /></el-icon>
              <span v-else class="dash">-</span>
            </template>
          </el-table-column>

          <!-- MySQL/MariaDB: 自增 -->
          <el-table-column 
            v-if="dbType !== 'postgresql'" 
            prop="auto_increment" 
            :label="t('schema.autoIncrement')" 
            width="70" 
            align="center"
          >
            <template #default="{ row }">
              <el-icon v-if="row.auto_increment" class="check-icon"><Check /></el-icon>
              <span v-else class="dash">-</span>
            </template>
          </el-table-column>

          <!-- PostgreSQL: 标识列 -->
          <el-table-column 
            v-if="dbType === 'postgresql'" 
            prop="identity" 
            :label="t('schema.identity')" 
            width="100" 
            align="center"
          >
            <template #default="{ row }">
              <el-tag v-if="row.identity" size="small" type="warning">
                {{ row.identity }}
              </el-tag>
              <span v-else class="dash">-</span>
            </template>
          </el-table-column>

          <!-- PostgreSQL: 数组 -->
          <el-table-column 
            v-if="dbType === 'postgresql'" 
            prop="is_array" 
            :label="t('schema.isArray')" 
            width="70" 
            align="center"
          >
            <template #default="{ row }">
              <el-icon v-if="row.is_array" class="check-icon"><Check /></el-icon>
              <span v-else class="dash">-</span>
            </template>
          </el-table-column>

          <!-- 注释 -->
          <el-table-column prop="comment" :label="t('sidebar.comment')" min-width="150">
            <template #default="{ row }">
              <span class="comment-text">{{ row.comment || '-' }}</span>
            </template>
          </el-table-column>

          <!-- 操作 -->
          <el-table-column :label="t('common.actions')" width="120" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="handleEditColumn(row)">
                {{ t('common.edit') }}
              </el-button>
              <el-button type="danger" link size="small" @click="handleDeleteColumn(row)">
                {{ t('common.delete') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 索引信息表格 -->
      <el-card class="schema-card">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><Connection /></el-icon>
              {{ t('schema.indexInfo') }}
            </span>
            <el-button type="primary" size="small" @click="handleAddIndex">
              <el-icon><Plus /></el-icon>
              {{ t('schema.addIndex') }}
            </el-button>
          </div>
        </template>
        
        <el-table 
          :data="indexes" 
          v-loading="loadingIndexes"
          stripe
          border
          class="schema-table"
        >
          <el-table-column prop="name" :label="t('schema.indexName')" min-width="150">
            <template #default="{ row }">
              <span class="index-name">
                <el-icon v-if="row.type === 'PRIMARY'" class="key-icon"><Key /></el-icon>
                {{ row.name }}
              </span>
            </template>
          </el-table-column>

          <el-table-column prop="type" :label="t('common.type')" width="120">
            <template #default="{ row }">
              <el-tag 
                size="small" 
                :type="getIndexTypeTag(row.type)"
              >
                {{ getIndexTypeLabel(row.type) }}
              </el-tag>
            </template>
          </el-table-column>

          <el-table-column prop="method" :label="t('common.method')" width="100">
            <template #default="{ row }">
              <span class="index-method">{{ row.method || '-' }}</span>
            </template>
          </el-table-column>

          <el-table-column prop="columns" :label="t('common.columns')" min-width="200">
            <template #default="{ row }">
              <div class="index-columns">
                <el-tag 
                  v-for="col in row.columns" 
                  :key="col" 
                  size="small" 
                  type="info"
                  class="column-tag"
                >
                  {{ col }}
                </el-tag>
              </div>
            </template>
          </el-table-column>

          <!-- PostgreSQL: WHERE 条件 -->
          <el-table-column 
            v-if="dbType === 'postgresql'" 
            prop="where_clause" 
            label="WHERE" 
            width="150"
          >
            <template #default="{ row }">
              <span class="where-clause" :title="row.where_clause">
                {{ row.where_clause || '-' }}
              </span>
            </template>
          </el-table-column>

          <el-table-column prop="cardinality" :label="t('schema.cardinality')" width="100" align="right">
            <template #default="{ row }">
              <span class="cardinality">{{ row.cardinality ?? '-' }}</span>
            </template>
          </el-table-column>

          <el-table-column prop="comment" :label="t('sidebar.comment')" min-width="150">
            <template #default="{ row }">
              <span class="comment-text">{{ row.comment || '-' }}</span>
            </template>
          </el-table-column>

          <el-table-column :label="t('common.actions')" width="120" fixed="right">
            <template #default="{ row }">
              <el-button 
                type="primary" 
                link 
                size="small" 
                @click="handleEditIndex(row)"
                :disabled="row.type === 'PRIMARY'"
              >
                {{ t('common.edit') }}
              </el-button>
              <el-button 
                type="danger" 
                link 
                size="small" 
                @click="handleDeleteIndex(row)"
                :disabled="row.type === 'PRIMARY'"
              >
                {{ t('common.delete') }}
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 建表语句 -->
      <el-card class="schema-card">
        <template #header>
          <div class="card-header">
            <span class="card-title">
              <el-icon><Document /></el-icon>
              {{ t('schema.createTableDDL') }}
            </span>
            <el-button-group>
              <el-button type="primary" size="small" @click="copyDDL">
                <el-icon><CopyDocument /></el-icon>
                {{ t('common.copy') }}
              </el-button>
              <el-button size="small" @click="handleRenameTable" v-if="currentTableName">
                <el-icon><Edit /></el-icon>
                {{ t('schema.renameTable') }}
              </el-button>
            </el-button-group>
          </div>
        </template>
        <pre class="ddl-content">{{ createTableDDL }}</pre>
      </el-card>
    </template>

    <!-- 字段编辑对话框 -->
    <el-dialog
      v-model="columnDialogVisible"
      :title="editingColumn ? t('schema.editColumn') : t('schema.addColumnTitle')"
      width="700px"
      destroy-on-close
    >
      <el-form 
        ref="columnFormRef"
        :model="columnForm" 
        :rules="columnRules"
        label-width="100px"
      >
        <!-- 字段名 -->
        <el-form-item :label="t('schema.columnName')" prop="name">
          <el-input v-model="columnForm.name" :placeholder="t('schema.enterColumnName')" />
        </el-form-item>

        <!-- 类型 -->
        <el-form-item :label="t('common.type')" prop="type">
          <div class="type-row">
            <el-select 
              v-model="columnForm.type" 
              :placeholder="t('schema.selectColumnType')" 
              style="width: 200px;"
              filterable
              @change="handleTypeChange"
            >
              <el-option-group 
                v-for="group in schemaConfig.columnTypeGroups" 
                :key="group.label" 
                :label="group.label"
              >
                <el-option 
                  v-for="type in group.types" 
                  :key="type" 
                  :label="type" 
                  :value="type" 
                />
              </el-option-group>
            </el-select>
            
            <!-- 长度 -->
            <el-input-number 
              v-if="showLengthInput"
              v-model="columnForm.length" 
              :min="1" 
              :max="65535"
              :placeholder="t('schema.length')"
              style="width: 120px; margin-left: 10px;"
            />
            
            <!-- 精度和小数位 -->
            <template v-if="showPrecisionInput">
              <el-input-number 
                v-model="columnForm.precision" 
                :min="1" 
                :max="65"
                :placeholder="t('schema.precision')"
                style="width: 100px; margin-left: 10px;"
              />
              <span style="margin: 0 5px;">,</span>
              <el-input-number 
                v-model="columnForm.scale" 
                :min="0" 
                :max="30"
                :placeholder="t('schema.scale')"
                style="width: 100px;"
              />
            </template>
          </div>
        </el-form-item>

        <!-- 可空 -->
        <el-form-item :label="t('schema.nullable')">
          <el-switch v-model="columnForm.nullable" />
        </el-form-item>

        <!-- 默认值 -->
        <el-form-item :label="t('schema.defaultValue')">
          <el-input v-model="columnForm.default_value" :placeholder="t('schema.enterDefaultValue')" />
        </el-form-item>

        <!-- 主键 -->
        <el-form-item :label="t('schema.primaryKey')">
          <el-switch v-model="columnForm.is_primary" />
        </el-form-item>

        <!-- MySQL/MariaDB 特有属性 -->
        <template v-if="dbType !== 'postgresql'">
          <!-- 无符号 -->
          <el-form-item :label="t('schema.unsigned')" v-if="isNumericType(columnForm.type)">
            <el-switch v-model="columnForm.unsigned" />
          </el-form-item>

          <!-- 填充零 -->
          <el-form-item :label="t('schema.zerofill')" v-if="isNumericType(columnForm.type)">
            <el-switch v-model="columnForm.zerofill" />
          </el-form-item>

          <!-- 自增 -->
          <el-form-item :label="t('schema.autoIncrement')" v-if="isIntegerType(columnForm.type)">
            <el-switch v-model="columnForm.auto_increment" />
          </el-form-item>

          <!-- 字符集 -->
          <el-form-item :label="t('sidebar.charset')" v-if="isStringType(columnForm.type)">
            <el-select v-model="columnForm.charset" :placeholder="t('schema.selectCharset')" clearable style="width: 200px;">
              <el-option 
                v-for="cs in schemaConfig.charsets" 
                :key="cs.value" 
                :label="cs.label" 
                :value="cs.value" 
              />
            </el-select>
          </el-form-item>

          <!-- 排序规则 -->
          <el-form-item :label="t('schema.collation')" v-if="isStringType(columnForm.type)">
            <el-select v-model="columnForm.collation" :placeholder="t('schema.selectCollation')" clearable style="width: 200px;">
              <el-option 
                v-for="col in schemaConfig.collations" 
                :key="col.value" 
                :label="col.label" 
                :value="col.value" 
              />
            </el-select>
          </el-form-item>

          <!-- 更新时 -->
          <el-form-item :label="t('schema.onUpdate')" v-if="isDateTimeType(columnForm.type)">
            <el-select v-model="columnForm.on_update" :placeholder="t('schema.selectOnUpdate')" clearable style="width: 200px;">
              <el-option :label="t('schema.none')" value="" />
              <el-option label="CURRENT_TIMESTAMP" value="CURRENT_TIMESTAMP" />
            </el-select>
          </el-form-item>
        </template>

        <!-- PostgreSQL 特有属性 -->
        <template v-if="dbType === 'postgresql'">
          <!-- 标识列 -->
          <el-form-item :label="t('schema.identity')" v-if="isIntegerType(columnForm.type)">
            <el-select v-model="columnForm.identity" :placeholder="t('schema.selectIdentityType')" clearable style="width: 200px;">
              <el-option :label="t('schema.none')" value="" />
              <el-option label="GENERATED ALWAYS" value="ALWAYS" />
              <el-option label="GENERATED BY DEFAULT" value="BY DEFAULT" />
            </el-select>
          </el-form-item>

          <!-- 数组 -->
          <el-form-item :label="t('schema.isArray')">
            <el-switch v-model="columnForm.is_array" />
            <el-input-number 
              v-if="columnForm.is_array"
              v-model="columnForm.dimension" 
              :min="1" 
              :max="6"
              :placeholder="t('schema.dimension')"
              style="width: 100px; margin-left: 10px;"
            />
          </el-form-item>

          <!-- 排序规则 -->
          <el-form-item :label="t('schema.collation')" v-if="isStringType(columnForm.type)">
            <el-select v-model="columnForm.collation" :placeholder="t('schema.selectCollation')" clearable style="width: 200px;">
              <el-option 
                v-for="col in schemaConfig.collations" 
                :key="col.value" 
                :label="col.label" 
                :value="col.value" 
              />
            </el-select>
          </el-form-item>
        </template>

        <!-- 注释 -->
        <el-form-item :label="t('sidebar.comment')">
          <el-input v-model="columnForm.comment" type="textarea" :rows="2" :placeholder="t('schema.enterComment')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="columnDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSaveColumn">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <!-- 索引编辑对话框 -->
    <el-dialog
      v-model="indexDialogVisible"
      :title="editingIndex ? t('schema.editIndex') : t('schema.addIndexTitle')"
      width="600px"
      destroy-on-close
    >
      <el-form 
        ref="indexFormRef"
        :model="indexForm" 
        :rules="indexRules"
        label-width="100px"
      >
        <!-- 索引名 -->
        <el-form-item :label="t('schema.indexName')" prop="name">
          <el-input v-model="indexForm.name" :placeholder="t('schema.enterIndexName')" />
        </el-form-item>

        <!-- 类型 -->
        <el-form-item :label="t('common.type')" prop="type">
          <el-select v-model="indexForm.type" :placeholder="t('schema.selectIndexType')">
            <el-option 
              v-for="type in schemaConfig.indexTypes.filter(t => t.value !== 'PRIMARY')" 
              :key="type.value" 
              :label="type.label" 
              :value="type.value" 
            />
          </el-select>
        </el-form-item>

        <!-- 索引方法 -->
        <el-form-item :label="t('schema.indexMethod')" v-if="schemaConfig.indexMethods">
          <el-select v-model="indexForm.method" :placeholder="t('schema.selectIndexMethod')">
            <el-option 
              v-for="method in schemaConfig.indexMethods" 
              :key="method.value" 
              :label="method.label" 
              :value="method.value" 
            />
          </el-select>
        </el-form-item>

        <!-- 字段 -->
        <el-form-item :label="t('common.columns')" prop="columns">
          <el-select 
            v-model="indexForm.columns" 
            multiple 
            :placeholder="t('schema.selectColumns')"
            style="width: 100%;"
          >
            <el-option 
              v-for="col in columns" 
              :key="col.name" 
              :label="col.name" 
              :value="col.name" 
            />
          </el-select>
        </el-form-item>

        <!-- PostgreSQL: WHERE 条件 -->
        <el-form-item :label="t('schema.whereCondition')" v-if="dbType === 'postgresql'">
          <el-input 
            v-model="indexForm.where_clause" 
            :placeholder="t('schema.whereExample')"
          />
          <div class="form-tip">{{ t('schema.partialIndexTip') }}</div>
        </el-form-item>

        <!-- PostgreSQL: NULLS NOT DISTINCT -->
        <el-form-item label="NULLS NOT DISTINCT" v-if="dbType === 'postgresql' && indexForm.type === 'UNIQUE'">
          <el-switch v-model="indexForm.nulls_not_distinct" />
          <div class="form-tip">{{ t('schema.nullsNotDistinctTip') }}</div>
        </el-form-item>

        <!-- 注释 -->
        <el-form-item :label="t('sidebar.comment')">
          <el-input v-model="indexForm.comment" type="textarea" :rows="2" :placeholder="t('schema.enterComment')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="indexDialogVisible = false">{{ t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSaveIndex">{{ t('common.save') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onActivated, onDeactivated } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { Grid, List, Plus, Key, Check, Connection, Document, CopyDocument, Edit } from '@element-plus/icons-vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import { useConnectionStore } from '@/store/connection'
import { useUrlState } from '@/composables/useUrlState'
import { DB_TYPE_CONFIG } from '@/types/connection'
import { 
  getDbTypeConfig, 
  typeNeedsLength, 
  typeNeedsPrecision,
  isNumericType as checkNumericType,
  isStringType as checkStringType,
  isDateTimeType as checkDateTimeType,
  getDefaultLength,
  getDefaultPrecision,
} from '@/config/dbTypeConfig'
import { 
  getTableColumns, 
  getTableIndexes, 
  createEmptyColumn,
  createEmptyIndex,
  type ColumnInfo, 
  type IndexInfo 
} from '@/api/schema'
import { alterTable, createIndex, dropIndex, renameTable } from '@/api/tableAdmin'

const route = useRoute()
const router = useRouter()
const connectionStore = useConnectionStore()
const { restoreFromUrl, getUrlParams } = useUrlState()
const { t } = useI18n()

// 是否已完成初始化（包括从 URL 恢复状态）
const initialized = ref(false)

// 缓存上次加载的参数，用于避免不必要的重新加载
const lastLoadParams = ref({
  connectionId: null as number | null,
  database: null as string | null,
  table: null as string | null,
  dbType: null as string | null
})

// 当前数据库类型
const dbType = computed(() => connectionStore.currentConnection?.db_type || 'mysql')
const dbTypeLabel = computed(() => DB_TYPE_CONFIG[dbType.value]?.label || dbType.value)

// 数据库类型配置
const schemaConfig = computed(() => getDbTypeConfig(dbType.value))

// 当前表名（从 URL 获取）
const currentTableName = computed(() => {
  return (route.query.table as string) || null
})

// 页面描述
const pageDescription = computed(() => {
  if (!currentTableName.value) {
    return t('schema.selectTableFirst')
  }
  return t('schema.tableStructure', { name: currentTableName.value })
})

// 字段数据
const columns = ref<ColumnInfo[]>([])
const loadingColumns = ref(false)

// 索引数据
const indexes = ref<IndexInfo[]>([])
const loadingIndexes = ref(false)

// 字段编辑对话框
const columnDialogVisible = ref(false)
const editingColumn = ref<ColumnInfo | null>(null)
const columnFormRef = ref<FormInstance>()
const columnForm = ref<ColumnInfo>(createEmptyColumn(dbType.value))

const columnRules = computed<FormRules>(() => ({
  name: [{ required: true, message: t('schema.enterFieldName'), trigger: 'blur' }],
  type: [{ required: true, message: t('schema.selectFieldType'), trigger: 'change' }],
}))

// 索引编辑对话框
const indexDialogVisible = ref(false)
const editingIndex = ref<IndexInfo | null>(null)
const indexFormRef = ref<FormInstance>()
const indexForm = ref<IndexInfo>(createEmptyIndex(dbType.value))

const indexRules = computed<FormRules>(() => ({
  name: [{ required: true, message: t('schema.enterIndexNameReq'), trigger: 'blur' }],
  type: [{ required: true, message: t('schema.selectIndexTypeReq'), trigger: 'change' }],
  columns: [{ required: true, message: t('schema.selectColumnsReq'), trigger: 'change', type: 'array', min: 1 }],
}))

// 是否显示长度输入框
const showLengthInput = computed(() => typeNeedsLength(columnForm.value.type, dbType.value))

// 是否显示精度输入框
const showPrecisionInput = computed(() => typeNeedsPrecision(columnForm.value.type, dbType.value))

// 类型判断函数
function isNumericType(type: string): boolean {
  return checkNumericType(type)
}

function isIntegerType(type: string): boolean {
  const intTypes = ['TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'INTEGER', 'BIGINT']
  return intTypes.includes(type.toUpperCase())
}

function isStringType(type: string): boolean {
  return checkStringType(type)
}

function isDateTimeType(type: string): boolean {
  return checkDateTimeType(type)
}

// 格式化字段类型显示
function formatColumnType(row: ColumnInfo): string {
  let result = row.type
  
  if (row.length) {
    result += `(${row.length})`
  } else if (row.precision != null) {
    result += `(${row.precision}`
    if (row.scale != null) {
      result += `,${row.scale}`
    }
    result += ')'
  }
  
  if (row.unsigned) {
    result += ' UNSIGNED'
  }
  
  if (row.is_array) {
    result += '[]'
  }
  
  return result
}

// 生成建表 DDL 语句
function generateCreateTableDDL(
  cols: ColumnInfo[],
  idxs: IndexInfo[],
  tableName: string,
  db: string
): string {
  if (!tableName || cols.length === 0) {
    return ''
  }

  if (db === 'postgresql') {
    return generatePostgreSQLDDL(cols, idxs, tableName)
  }
  return generateMySQLDDL(cols, idxs, tableName)
}

function generateMySQLDDL(cols: ColumnInfo[], idxs: IndexInfo[], tableName: string): string {
  const q = (s: string) => '`' + s.replace(/`/g, '``') + '`'
  const parts: string[] = []

  for (const col of cols) {
    let def = `  ${q(col.name)} ${formatColumnType(col)}`
    def += col.nullable ? '' : ' NOT NULL'
    if (col.default_value != null && col.default_value !== '') {
      const dv = col.default_value.toUpperCase()
      if (dv === 'CURRENT_TIMESTAMP' || dv === 'NULL') {
        def += ` DEFAULT ${col.default_value}`
      } else {
        def += ` DEFAULT '${String(col.default_value).replace(/'/g, "''")}'`
      }
    }
    if (col.auto_increment) def += ' AUTO_INCREMENT'
    if (col.comment) def += ` COMMENT '${String(col.comment).replace(/'/g, "''")}'`
    parts.push(def)
  }

  // 主键和索引
  const pkCols = cols.filter(c => c.is_primary).map(c => c.name)
  if (pkCols.length > 0) {
    parts.push(`  PRIMARY KEY (${pkCols.map(c => q(c)).join(', ')})`)
  }

  for (const idx of idxs) {
    if (idx.type === 'PRIMARY') continue
    if (!idx.columns || idx.columns.length === 0) continue
    const colsStr = idx.columns.map(c => q(c)).join(', ')
    if (idx.type === 'UNIQUE') {
      parts.push(`  UNIQUE KEY ${q(idx.name)} (${colsStr})`)
    } else if (idx.type === 'FULLTEXT') {
      parts.push(`  FULLTEXT KEY ${q(idx.name)} (${colsStr})`)
    } else if (idx.type === 'SPATIAL') {
      parts.push(`  SPATIAL KEY ${q(idx.name)} (${colsStr})`)
    } else {
      parts.push(`  KEY ${q(idx.name)} (${colsStr})`)
    }
  }

  return `CREATE TABLE ${q(tableName)} (\n${parts.join(',\n')}\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
}

function generatePostgreSQLDDL(cols: ColumnInfo[], idxs: IndexInfo[], tableName: string): string {
  const q = (s: string) => '"' + s.replace(/"/g, '""') + '"'
  const parts: string[] = []

  for (const col of cols) {
    let def = `  ${q(col.name)} ${formatColumnType(col)}`
    if (col.identity === 'ALWAYS') {
      def += ' GENERATED ALWAYS AS IDENTITY'
    } else if (col.identity === 'BY DEFAULT') {
      def += ' GENERATED BY DEFAULT AS IDENTITY'
    } else {
      def += col.nullable ? '' : ' NOT NULL'
      if (col.default_value != null && col.default_value !== '') {
        const dv = col.default_value.toUpperCase()
        if (dv === 'CURRENT_TIMESTAMP' || dv === 'NULL') {
          def += ` DEFAULT ${col.default_value}`
        } else {
          def += ` DEFAULT '${String(col.default_value).replace(/'/g, "''")}'`
        }
      }
    }
    if (col.is_primary && cols.filter(c => c.is_primary).length === 1) {
      def += ' PRIMARY KEY'
    }
    parts.push(def)
  }

  const pkCols = cols.filter(c => c.is_primary).map(c => c.name)
  if (pkCols.length > 1) {
    parts.push(`  PRIMARY KEY (${pkCols.map(c => q(c)).join(', ')})`)
  }

  let sql = `CREATE TABLE ${q(tableName)} (\n${parts.join(',\n')}\n);\n`

  for (const idx of idxs) {
    if (idx.type === 'PRIMARY') continue
    if (!idx.columns || idx.columns.length === 0) continue
    const colsStr = idx.columns.map(c => q(c)).join(', ')
    const unique = idx.type === 'UNIQUE' ? 'UNIQUE ' : ''
    const method = idx.method ? ` USING ${idx.method}` : ''
    let idxSql = `CREATE ${unique}INDEX ${q(idx.name)} ON ${q(tableName)}${method} (${colsStr})`
    if (idx.where_clause) idxSql += ` WHERE ${idx.where_clause}`
    idxSql += ';\n'
    sql += idxSql
  }

  return sql.trim()
}

// 建表语句（computed）
const createTableDDL = computed(() => {
  return generateCreateTableDDL(
    columns.value,
    indexes.value,
    currentTableName.value || '',
    dbType.value
  )
})

// 复制建表语句
async function copyDDL() {
  const text = createTableDDL.value
  if (!text) {
    ElMessage.warning(t('schema.noDDLToCopy'))
    return
  }
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(t('schema.copiedToClipboard'))
  } catch {
    ElMessage.error(t('schema.copyFailed'))
  }
}

// 重命名表
async function handleRenameTable() {
  if (!currentTableName.value || !connectionStore.currentConnection || !connectionStore.currentDatabase) {
    return
  }

  try {
    const { value: newName } = await ElMessageBox.prompt(
      t('schema.enterNewTableName'),
      t('schema.renameTable'),
      {
        confirmButtonText: t('common.confirm'),
        cancelButtonText: t('common.cancel'),
        inputValue: currentTableName.value,
        inputPattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/,
        inputErrorMessage: t('schema.invalidTableName')
      }
    )

    if (!newName || newName === currentTableName.value) {
      return
    }

    await renameTable(connectionStore.currentConnection.id, {
      database: connectionStore.currentDatabase,
      old_name: currentTableName.value,
      new_name: newName
    })

    ElMessage.success(t('schema.tableRenamed'))
    // 更新URL参数
    const route = useRoute()
    const router = useRouter()
    router.replace({
      path: route.path,
      query: {
        ...route.query,
        table: newName
      }
    })
  } catch (error: unknown) {
    if ((error as { message?: string }).message !== 'cancel') {
      const err = error as { response?: { data?: { message?: string } }; message?: string }
      ElMessage.error(err.response?.data?.message || err.message || t('schema.renameTableFailed'))
    }
  }
}

// 检查参数是否变化
function shouldReload(): boolean {
  const current = {
    connectionId: connectionStore.currentConnection?.id || null,
    database: connectionStore.currentDatabase || null,
    table: currentTableName.value,
    dbType: dbType.value
  }
  
  const changed = 
    current.connectionId !== lastLoadParams.value.connectionId ||
    current.database !== lastLoadParams.value.database ||
    current.table !== lastLoadParams.value.table ||
    current.dbType !== lastLoadParams.value.dbType
    
  if (changed) {
    lastLoadParams.value = { ...current }
  }
  
  return changed
}

// 加载表结构数据
async function loadSchemaData(skipCheck = false) {
  // 未初始化时不加载
  if (!initialized.value) {
    return
  }

  // 检查参数是否变化，如果没变化则不重新加载
  if (!skipCheck && !shouldReload()) {
    return
  }

  if (!currentTableName.value || !connectionStore.currentConnection || !connectionStore.currentDatabase) {
    columns.value = []
    indexes.value = []
    return
  }

  const connectionId = connectionStore.currentConnection.id
  const database = connectionStore.currentDatabase
  const table = currentTableName.value

  // 并行加载字段和索引
  loadingColumns.value = true
  loadingIndexes.value = true

  try {
    const [columnsData, indexesData] = await Promise.all([
      getTableColumns(connectionId, database, table, dbType.value),
      getTableIndexes(connectionId, database, table, dbType.value),
    ])
    columns.value = columnsData
    indexes.value = indexesData
  } catch (error) {
    console.error('加载表结构失败:', error)
    ElMessage.error(t('schema.loadFailed'))
  } finally {
    loadingColumns.value = false
    loadingIndexes.value = false
  }
}

// 初始化：从 URL 恢复状态
async function initFromUrl(forceReload = false) {
  const { connectionId } = getUrlParams()
  
  // 如果 URL 有 connection_id 但 store 中没有当前连接，需要恢复状态
  if (connectionId && !connectionStore.currentConnection) {
    try {
      await restoreFromUrl()
    } catch (error) {
      console.error('恢复连接状态失败:', error)
    }
  }
  
  // 标记已初始化
  initialized.value = true
  
  // 检查参数是否变化
  const currentParams = {
    connectionId: connectionStore.currentConnection?.id || null,
    database: connectionStore.currentDatabase || null,
    table: currentTableName.value,
    dbType: dbType.value
  }
  
  // 如果参数没变化且不是强制加载，则不重新加载
  if (!forceReload && 
      currentParams.connectionId === lastLoadParams.value.connectionId &&
      currentParams.database === lastLoadParams.value.database &&
      currentParams.table === lastLoadParams.value.table &&
      currentParams.dbType === lastLoadParams.value.dbType) {
    return
  }
  
  // 更新参数缓存
  lastLoadParams.value = { ...currentParams }
  
  // 加载数据（跳过检查，因为已经在 initFromUrl 中检查过了）
  await loadSchemaData(true)
}

// 防抖定时器，切换表时减少并发请求
let schemaLoadDebounceTimer: ReturnType<typeof setTimeout> | null = null

// 合并多个 watch，防抖加载，避免切换表时触发大量请求
watch(
  [
    () => connectionStore.currentConnection?.id,
    () => connectionStore.currentDatabase,
    currentTableName,
    dbType
  ],
  () => {
    if (!initialized.value) return
    if (schemaLoadDebounceTimer) clearTimeout(schemaLoadDebounceTimer)
    schemaLoadDebounceTimer = setTimeout(() => {
      schemaLoadDebounceTimer = null
      loadSchemaData()
    }, 150)
  },
  { deep: false }
)

// 组件挂载时初始化
onMounted(() => {
  initFromUrl(true)
})

// 组件激活时检查是否需要重新加载（keep-alive 缓存后切换回来时）
onActivated(() => {
  if (initialized.value) {
    initFromUrl(false)
  }
})

// 离开页面时清理防抖定时器
onDeactivated(() => {
  if (schemaLoadDebounceTimer) {
    clearTimeout(schemaLoadDebounceTimer)
    schemaLoadDebounceTimer = null
  }
})

// 类型变更处理
function handleTypeChange(type: string) {
  // 设置默认长度
  const defaultLen = getDefaultLength(type)
  if (defaultLen !== null) {
    columnForm.value.length = defaultLen
  } else {
    columnForm.value.length = null
  }
  
  // 设置默认精度
  const defaultPrec = getDefaultPrecision(type)
  if (defaultPrec) {
    columnForm.value.precision = defaultPrec.precision
    columnForm.value.scale = defaultPrec.scale
  } else {
    columnForm.value.precision = null
    columnForm.value.scale = null
  }
}

// 获取索引类型标签颜色
function getIndexTypeTag(type: string): 'primary' | 'success' | 'warning' | 'info' {
  switch (type) {
    case 'PRIMARY':
      return 'primary'
    case 'UNIQUE':
      return 'success'
    case 'FULLTEXT':
    case 'SPATIAL':
      return 'warning'
    default:
      return 'info'
  }
}

// 获取索引类型标签文本
function getIndexTypeLabel(type: string): string {
  const item = schemaConfig.value.indexTypes.find(t => t.value === type)
  return item?.label || type
}

// 新增字段
function handleAddColumn() {
  editingColumn.value = null
  columnForm.value = createEmptyColumn(dbType.value)
  columnDialogVisible.value = true
}

// 编辑字段
function handleEditColumn(row: ColumnInfo) {
  editingColumn.value = row
  columnForm.value = { ...row }
  columnDialogVisible.value = true
}

// 删除字段
async function handleDeleteColumn(row: ColumnInfo) {
  if (!connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
    ElMessage.error(t('schema.connectionIncomplete'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('schema.deleteColumnConfirm', { name: row.name }),
      t('schema.deleteConfirm'),
      { type: 'warning' }
    )

    await alterTable(connectionStore.currentConnection.id, {
      database: connectionStore.currentDatabase,
      table: currentTableName.value,
      operations: [{
        type: 'drop_column',
        old_column_name: row.name
      }]
    })

    // 重新加载表结构
    await loadSchemaData(true)
    ElMessage.success(t('schema.columnDeleted'))
  } catch (error: unknown) {
    if ((error as { message?: string }).message !== 'cancel') {
      const err = error as { response?: { data?: { message?: string } }; message?: string }
      ElMessage.error(err.response?.data?.message || err.message || t('schema.deleteColumnFailed'))
    }
  }
}

// 保存字段
async function handleSaveColumn() {
  if (!columnFormRef.value || !connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
    return
  }
  
  try {
    await columnFormRef.value.validate()
    
    if (editingColumn.value) {
      // 编辑模式：修改字段
      const isRename = editingColumn.value.name !== columnForm.value.name
      
      await alterTable(connectionStore.currentConnection.id, {
        database: connectionStore.currentDatabase,
        table: currentTableName.value,
        operations: [{
          type: isRename ? 'rename_column' : 'modify_column',
          old_column_name: editingColumn.value.name,
          new_column_name: isRename ? columnForm.value.name : undefined,
          column: columnForm.value
        }]
      })
      ElMessage.success(t('schema.columnUpdated'))
    } else {
      // 新增模式：添加字段
      await alterTable(connectionStore.currentConnection.id, {
        database: connectionStore.currentDatabase,
        table: currentTableName.value,
        operations: [{
          type: 'add_column',
          column: columnForm.value
        }]
      })
      ElMessage.success(t('schema.columnAdded'))
    }
    
    // 重新加载表结构
    await loadSchemaData(true)
    columnDialogVisible.value = false
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    ElMessage.error(err.response?.data?.message || err.message || t('schema.saveColumnFailed'))
  }
}

// 新增索引
function handleAddIndex() {
  editingIndex.value = null
  indexForm.value = createEmptyIndex(dbType.value)
  indexDialogVisible.value = true
}

// 编辑索引
function handleEditIndex(row: IndexInfo) {
  editingIndex.value = row
  indexForm.value = { ...row }
  indexDialogVisible.value = true
}

// 删除索引
async function handleDeleteIndex(row: IndexInfo) {
  if (!connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
    ElMessage.error(t('schema.connectionIncomplete'))
    return
  }

  try {
    await ElMessageBox.confirm(
      t('schema.deleteIndexConfirm', { name: row.name }),
      t('schema.deleteConfirm'),
      { type: 'warning' }
    )

    await dropIndex(connectionStore.currentConnection.id, {
      database: connectionStore.currentDatabase,
      table: currentTableName.value,
      index_name: row.name
    })

    // 重新加载表结构
    await loadSchemaData(true)
    ElMessage.success(t('schema.indexDeleted'))
  } catch (error: unknown) {
    if ((error as { message?: string }).message !== 'cancel') {
      const err = error as { response?: { data?: { message?: string } }; message?: string }
      ElMessage.error(err.response?.data?.message || err.message || t('schema.deleteIndexFailed'))
    }
  }
}

// 保存索引
async function handleSaveIndex() {
  if (!indexFormRef.value || !connectionStore.currentConnection || !connectionStore.currentDatabase || !currentTableName.value) {
    return
  }
  
  try {
    await indexFormRef.value.validate()
    
    if (editingIndex.value) {
      // 编辑模式：先删除旧索引，再创建新索引
      await dropIndex(connectionStore.currentConnection.id, {
        database: connectionStore.currentDatabase,
        table: currentTableName.value,
        index_name: editingIndex.value.name
      })
    }
    
    // 创建索引
    await createIndex(connectionStore.currentConnection.id, {
      database: connectionStore.currentDatabase,
      table: currentTableName.value,
      index: indexForm.value
    })
    
    // 重新加载表结构
    await loadSchemaData(true)
    ElMessage.success(editingIndex.value ? t('schema.indexUpdated') : t('schema.indexAdded'))
    indexDialogVisible.value = false
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    ElMessage.error(err.response?.data?.message || err.message || t('schema.saveIndexFailed'))
  }
}
</script>

<style scoped>
.schema-page {
  max-width: 1400px;
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

.schema-card {
  margin-bottom: var(--spacing-lg);
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
}

.schema-card :deep(.el-card__header) {
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.schema-card :deep(.el-card__body) {
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

.db-type-tag {
  margin-left: var(--spacing-sm);
}

.schema-table {
  width: 100%;
}

.schema-table :deep(.el-table__header th) {
  background-color: var(--color-background-secondary);
  font-weight: 600;
}

.column-name,
.index-name {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-family: var(--font-family-code);
  font-weight: 500;
}

.key-icon {
  color: var(--color-warning);
}

.check-icon {
  color: var(--color-success);
}

.dash {
  color: var(--color-text-tertiary);
}

.default-value {
  font-family: var(--font-family-code);
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

.comment-text {
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

.index-columns {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.column-tag {
  font-family: var(--font-family-code);
}

.index-method {
  font-family: var(--font-family-code);
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

.where-clause {
  font-family: var(--font-family-code);
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cardinality {
  font-family: var(--font-family-code);
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

.type-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 5px;
}

.form-tip {
  font-size: 12px;
  color: var(--color-text-tertiary);
  margin-top: 4px;
}

.ddl-content {
  font-family: var(--font-family-code);
  font-size: var(--font-size-small);
  padding: var(--spacing-md) var(--spacing-lg);
  margin: 0;
  background-color: var(--color-background-secondary);
  border-radius: var(--border-radius-small);
  overflow-x: auto;
  white-space: pre;
  color: var(--color-text-primary);
}

.schema-card .ddl-content {
  border-radius: 0;
}

/* 响应式 */
@media (max-width: 768px) {
  .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .type-row {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
