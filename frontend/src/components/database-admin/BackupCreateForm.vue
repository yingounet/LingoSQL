<template>
  <el-card class="backup-create-form" shadow="never">
    <template #header>
      <div class="card-header">
        <span>新建备份</span>
        <el-button link type="primary" @click="expanded = !expanded">
          {{ expanded ? '收起' : '展开' }}
          <el-icon><ArrowDown v-if="!expanded" /><ArrowUp v-else /></el-icon>
        </el-button>
      </div>
    </template>
    <el-collapse-transition>
      <el-form v-show="expanded" :model="form" label-width="120px" class="backup-form">
        <el-form-item label="数据库" required>
          <el-select
            v-model="form.database"
            placeholder="选择数据库"
            filterable
            style="width: 280px"
          >
            <el-option
              v-for="db in databaseOptions"
              :key="db.name"
              :label="db.name"
              :value="db.name"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="备份范围">
          <el-radio-group v-model="form.scope">
            <el-radio value="full">全库</el-radio>
            <el-radio value="tables">选表</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.scope === 'tables'" label="选择表">
          <div class="table-checkbox-group">
            <el-checkbox-group v-model="form.selectedTables">
              <el-checkbox
                v-for="t in tables"
                :key="t.name"
                :label="t.name"
              >
                {{ t.name }}
              </el-checkbox>
            </el-checkbox-group>
          </div>
          <div v-if="props.tables.length === 0 && form.database" class="no-tables-hint">
            该数据库暂无表
          </div>
        </el-form-item>
        <el-form-item label="仅结构">
          <el-switch v-model="form.schemaOnly" />
          <span class="form-hint">不备份数据，仅导出 CREATE TABLE 等结构</span>
        </el-form-item>
        <el-form-item label="压缩">
          <el-switch v-model="form.compress" />
          <span class="form-hint">使用 gzip 压缩备份文件</span>
        </el-form-item>
        <el-form-item label="单文件最大(MB)">
          <el-input-number v-model="form.maxFileSizeMb" :min="0" :max="2048" />
          <span class="form-hint">0 表示不拆分</span>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="submitting" :disabled="!form.database" @click="handleSubmit">
            开始备份
          </el-button>
        </el-form-item>
      </el-form>
    </el-collapse-transition>
  </el-card>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ArrowDown, ArrowUp } from '@element-plus/icons-vue'
import type { DatabaseInfo } from '@/types/databaseAdmin'
import type { TableInfo } from '@/api/table'
import type { CreateBackupRequest } from '@/api/backup'

const props = defineProps<{
  connectionId: number
  databaseOptions: DatabaseInfo[]
  selectedDatabase: string | null
  tables: TableInfo[]
  submitting?: boolean
}>()

const emit = defineEmits<{
  (e: 'submit', req: CreateBackupRequest): void
}>()

const expanded = ref(true)

const form = ref({
  database: '',
  scope: 'full' as 'full' | 'tables',
  selectedTables: [] as string[],
  schemaOnly: false,
  compress: true,
  maxFileSizeMb: 0
})

watch(
  () => props.selectedDatabase,
  (db) => {
    if (db) form.value.database = db
  },
  { immediate: true }
)

watch(
  () => props.databaseOptions,
  (opts) => {
    if (opts.length > 0 && !form.value.database && props.selectedDatabase) {
      form.value.database = props.selectedDatabase
    } else if (opts.length > 0 && !form.value.database) {
      form.value.database = opts[0].name
    }
  },
  { immediate: true }
)

function handleSubmit() {
  if (!form.value.database || !props.connectionId) return
  emit('submit', {
    connection_id: props.connectionId,
    database: form.value.database,
    tables: form.value.scope === 'tables' && form.value.selectedTables.length > 0 ? form.value.selectedTables : undefined,
    schema_only: form.value.schemaOnly,
    compress: form.value.compress,
    max_file_size_mb: form.value.maxFileSizeMb || 0
  })
}
</script>

<style scoped>
.backup-create-form {
  margin-bottom: var(--spacing-md);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.backup-form {
  padding-top: var(--spacing-sm);
}

.table-checkbox-group {
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
  max-height: 160px;
  overflow-y: auto;
  padding: var(--spacing-sm);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-medium);
}

.form-hint {
  margin-left: var(--spacing-sm);
  font-size: 12px;
  color: var(--color-text-tertiary);
}

.no-tables-hint {
  font-size: 13px;
  color: var(--color-text-tertiary);
}
</style>
