<template>
  <el-dialog
    v-model="visible"
    :title="t('dbAdmin.deleteDb')"
    width="500px"
  >
    <div class="delete-confirm" v-if="database">
      <el-alert
        type="warning"
        :closable="false"
        show-icon
      >
        <template #title>
          <div class="alert-content">
            <p>{{ t('dbAdmin.deleteDbConfirm', { name: database.name }) }}</p>
            <p class="warning-text">{{ t('dbAdmin.cannotUndo') }}</p>
          </div>
        </template>
      </el-alert>
      
      <div class="confirm-input" style="margin-top: 20px">
        <el-input
          v-model="confirmText"
          :placeholder="t('dbAdmin.enterDbNameToConfirm')"
        />
      </div>
    </div>
    
    <template #footer>
      <el-button @click="visible = false">{{ t('common.cancel') }}</el-button>
      <el-button 
        type="danger" 
        @click="handleConfirm"
        :disabled="confirmText !== database?.name"
        :loading="loading"
      >
        {{ t('dbAdmin.confirmDelete') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { DatabaseInfo } from '@/types/databaseAdmin'

const { t } = useI18n()

const props = defineProps<{
  modelValue: boolean
  database: DatabaseInfo | null
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'confirm': []
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const confirmText = ref('')
const loading = ref(false)

watch(visible, (newVal) => {
  if (newVal) {
    confirmText.value = ''
  }
})

function handleConfirm() {
  emit('confirm')
  visible.value = false
}
</script>

<style scoped>
.delete-confirm {
  padding: var(--spacing-sm);
}

.alert-content {
  line-height: 1.6;
}

.warning-text {
  color: var(--el-color-warning);
  font-size: 14px;
  margin-top: 8px;
}

.confirm-input {
  margin-top: var(--spacing-md);
}
</style>
