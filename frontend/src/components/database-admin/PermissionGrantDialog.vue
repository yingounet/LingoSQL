<template>
  <el-dialog
    v-model="visible"
    :title="isGrant ? '授予权限' : '撤销权限'"
    width="500px"
    @close="handleClose"
  >
    <div v-if="target && user">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="用户">
          {{ getUserLabel(user) }}
        </el-descriptions-item>
        <el-descriptions-item label="对象类型">
          {{ getTargetTypeLabel(target.type) }}
        </el-descriptions-item>
        <el-descriptions-item label="对象名称">
          {{ target.name }}
        </el-descriptions-item>
      </el-descriptions>
      
      <div class="privileges-section" style="margin-top: 20px">
        <div class="section-title">权限类型</div>
        <el-checkbox-group v-model="selectedPrivileges">
          <el-checkbox 
            v-for="priv in availablePrivileges" 
            :key="priv"
            :label="priv"
          >
            {{ priv }}
          </el-checkbox>
        </el-checkbox-group>
      </div>
    </div>
    
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button 
        type="primary" 
        @click="handleSubmit"
        :disabled="selectedPrivileges.length === 0"
        :loading="loading"
      >
        {{ isGrant ? '授予' : '撤销' }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { PermissionNode, PrivilegeType } from '@/types/permissionAdmin'
import type { DatabaseUser } from '@/types/userAdmin'

const props = defineProps<{
  modelValue: boolean
  target: PermissionNode | null
  user: DatabaseUser | null
  dbType: 'mysql' | 'postgresql'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'grant': [data: { privileges: string[], isGrant: boolean }]
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const selectedPrivileges = ref<string[]>([])
const loading = ref(false)
const isGrant = ref(true)

const availablePrivileges: PrivilegeType[] = [
  'SELECT',
  'INSERT',
  'UPDATE',
  'DELETE',
  'CREATE',
  'DROP',
  'ALTER',
  'INDEX',
  'REFERENCES',
  'TRIGGER',
  'EXECUTE',
  'USAGE'
]

function getUserLabel(user: DatabaseUser): string {
  if (props.dbType === 'mysql') {
    return `${user.username}@${user.host || 'localhost'}`
  }
  return user.username
}

function getTargetTypeLabel(type: string): string {
  const labels: Record<string, string> = {
    database: '数据库',
    table: '表',
    column: '列'
  }
  return labels[type] || type
}

watch(visible, (newVal) => {
  if (newVal && props.target) {
    selectedPrivileges.value = [...(props.target.privileges || [])]
    isGrant.value = true
  }
})

function handleClose() {
  visible.value = false
  selectedPrivileges.value = []
}

function handleSubmit() {
  if (selectedPrivileges.value.length === 0) {
    ElMessage.warning('请至少选择一个权限')
    return
  }
  
  emit('grant', {
    privileges: selectedPrivileges.value,
    isGrant: isGrant.value
  })
  handleClose()
}
</script>

<style scoped>
.privileges-section {
  padding: var(--spacing-md);
  background-color: var(--color-background-secondary);
  border-radius: var(--border-radius-medium);
}

.section-title {
  font-weight: 600;
  margin-bottom: var(--spacing-sm);
  color: var(--color-text-primary);
}
</style>
