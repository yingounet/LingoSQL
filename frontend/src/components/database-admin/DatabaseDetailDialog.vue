<template>
  <el-dialog
    v-model="visible"
    :title="t('dbAdmin.dbDetail')"
    width="600px"
  >
    <div class="database-detail" v-if="database">
      <el-descriptions :column="2" border>
        <el-descriptions-item :label="t('dbAdmin.dbName')">
          {{ database.name }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dbAdmin.charsetLabel')" v-if="dbType === 'mysql' && database.charset">
          {{ database.charset }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dbAdmin.encodingLabel')" v-if="dbType === 'postgresql' && database.encoding">
          {{ database.encoding }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dbAdmin.sortRule')" v-if="dbType === 'mysql' && database.collation">
          {{ database.collation }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dbAdmin.sizeLabel')" v-if="database.size">
          {{ formatSize(database.size) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dbAdmin.tableCount')" v-if="database.table_count !== undefined">
          {{ database.table_count }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('dbAdmin.createdAt')" v-if="database.created_at" :span="2">
          {{ database.created_at }}
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { DatabaseInfo } from '@/types/databaseAdmin'

const { t } = useI18n()

const props = defineProps<{
  modelValue: boolean
  database: DatabaseInfo | null
  dbType: 'mysql' | 'postgresql'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

function formatSize(size: number): string {
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`
  if (size < 1024 * 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(2)} MB`
  return `${(size / (1024 * 1024 * 1024)).toFixed(2)} GB`
}
</script>

<style scoped>
.database-detail {
  padding: var(--spacing-sm);
}
</style>
