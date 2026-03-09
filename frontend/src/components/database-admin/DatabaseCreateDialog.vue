<template>
  <el-dialog
    v-model="visible"
    :title="t('dbAdmin.createDb')"
    width="500px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item :label="t('dbAdmin.dbName')" prop="name">
        <el-input v-model="form.name" :placeholder="t('dbAdmin.enterDbName')" />
      </el-form-item>
      
      <!-- MySQL选项 -->
      <template v-if="dbType === 'mysql'">
        <el-form-item :label="t('dbAdmin.charsetLabel')" prop="charset">
          <el-select v-model="form.charset" :placeholder="t('dbAdmin.selectCharset')">
            <el-option label="utf8mb4" value="utf8mb4" />
            <el-option label="utf8" value="utf8" />
            <el-option label="latin1" value="latin1" />
          </el-select>
        </el-form-item>
        <el-form-item :label="t('dbAdmin.sortRule')" prop="collation">
          <el-select v-model="form.collation" :placeholder="t('dbAdmin.selectSortRule')">
            <el-option label="utf8mb4_unicode_ci" value="utf8mb4_unicode_ci" />
            <el-option label="utf8mb4_general_ci" value="utf8mb4_general_ci" />
            <el-option label="utf8_unicode_ci" value="utf8_unicode_ci" />
          </el-select>
        </el-form-item>
      </template>
      
      <!-- PostgreSQL选项 -->
      <template v-if="dbType === 'postgresql'">
        <el-form-item :label="t('dbAdmin.encodingLabel')" prop="encoding">
          <el-select v-model="form.encoding" :placeholder="t('dbAdmin.selectEncoding')">
            <el-option label="UTF8" value="UTF8" />
            <el-option label="LATIN1" value="LATIN1" />
          </el-select>
        </el-form-item>
        <el-form-item label="LC_COLLATE" prop="lc_collate">
          <el-input v-model="form.lc_collate" placeholder="例如: en_US.UTF-8" />
        </el-form-item>
        <el-form-item label="LC_CTYPE" prop="lc_ctype">
          <el-input v-model="form.lc_ctype" placeholder="例如: en_US.UTF-8" />
        </el-form-item>
      </template>
    </el-form>
    
    <template #footer>
      <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="loading">
        {{ t('common.create') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { createDatabase } from '@/api/databaseAdmin'
import { useConnectionStore } from '@/store/connection'
import type { CreateDatabaseRequest } from '@/types/databaseAdmin'

const props = defineProps<{
  modelValue: boolean
  dbType: 'mysql' | 'postgresql'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'success': []
}>()

const { t } = useI18n()
const connectionStore = useConnectionStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive<CreateDatabaseRequest>({
  name: '',
  charset: 'utf8mb4',
  collation: 'utf8mb4_unicode_ci',
  encoding: 'UTF8',
  lc_collate: 'en_US.UTF-8',
  lc_ctype: 'en_US.UTF-8'
})

const rules = computed<FormRules>(() => ({
  name: [
    { required: true, message: t('dbAdmin.enterDbNameReq'), trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: t('dbAdmin.dbNamePattern'), trigger: 'blur' }
  ]
}))

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

function handleClose() {
  visible.value = false
  formRef.value?.resetFields()
  Object.assign(form, {
    name: '',
    charset: 'utf8mb4',
    collation: 'utf8mb4_unicode_ci',
    encoding: 'UTF8',
    lc_collate: 'en_US.UTF-8',
    lc_ctype: 'en_US.UTF-8'
  })
}

async function handleSubmit() {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (!connectionStore.currentConnection) {
      ElMessage.error(t('dbAdmin.selectConnectionFirst'))
      return
    }
    
    loading.value = true
    try {
      await createDatabase(connectionStore.currentConnection.id, form)
      ElMessage.success(t('dbAdmin.dbCreateSuccess'))
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error.message || t('dbAdmin.createDbFailed'))
    } finally {
      loading.value = false
    }
  })
}
</script>
