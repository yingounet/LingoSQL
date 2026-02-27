<template>
  <el-dialog
    v-model="visible"
    title="数据库详情"
    width="600px"
  >
    <div class="database-detail" v-if="database">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="数据库名称">
          {{ database.name }}
        </el-descriptions-item>
        <el-descriptions-item label="字符集" v-if="dbType === 'mysql' && database.charset">
          {{ database.charset }}
        </el-descriptions-item>
        <el-descriptions-item label="编码" v-if="dbType === 'postgresql' && database.encoding">
          {{ database.encoding }}
        </el-descriptions-item>
        <el-descriptions-item label="排序规则" v-if="dbType === 'mysql' && database.collation">
          {{ database.collation }}
        </el-descriptions-item>
        <el-descriptions-item label="大小" v-if="database.size">
          {{ formatSize(database.size) }}
        </el-descriptions-item>
        <el-descriptions-item label="表数量" v-if="database.table_count !== undefined">
          {{ database.table_count }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间" v-if="database.created_at" :span="2">
          {{ database.created_at }}
        </el-descriptions-item>
      </el-descriptions>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { DatabaseInfo } from '@/types/databaseAdmin'

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
