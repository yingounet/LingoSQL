<template>
  <el-dialog
    v-model="visible"
    :title="isGrant ? t('userAdmin.grantPermission') : t('userAdmin.revokePermission')"
    width="500px"
    @close="handleClose"
  >
    <div v-if="target && user">
      <el-descriptions :column="1" border>
        <el-descriptions-item :label="t('userAdmin.user')">
          {{ getUserLabel(user) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('userAdmin.objectType')">
          {{ getTargetTypeLabel(target.type) }}
        </el-descriptions-item>
        <el-descriptions-item :label="t('userAdmin.objectName')">
          {{ target.name }}
        </el-descriptions-item>
      </el-descriptions>
      
      <div class="privileges-section" style="margin-top: 20px">
        <div class="section-title">{{ t('userAdmin.permissionType') }}</div>
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
      <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
      <el-button 
        type="primary" 
        @click="handleSubmit"
        :disabled="selectedPrivileges.length === 0"
        :loading="loading"
      >
        {{ isGrant ? t('userAdmin.grant') : t('userAdmin.revoke') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { PermissionNode, PrivilegeType } from '@/types/permissionAdmin'
import type { DatabaseUser } from '@/types/userAdmin'

const { t } = useI18n()

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
    database: t('userAdmin.objDatabase'),
    table: t('userAdmin.objTable'),
    column: t('userAdmin.objColumn')
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
    ElMessage.warning(t('userAdmin.selectAtLeastOne'))
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
