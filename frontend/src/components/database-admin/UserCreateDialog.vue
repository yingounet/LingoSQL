<template>
  <el-dialog
    v-model="visible"
    :title="t('userAdmin.createUserTitle')"
    width="500px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item :label="t('auth.username')" prop="username">
        <el-input v-model="form.username" :placeholder="t('userAdmin.enterUsername')" />
      </el-form-item>
      
      <!-- MySQL选项 -->
      <el-form-item 
        v-if="dbType === 'mysql'" 
        :label="t('userAdmin.hostCol')" 
        prop="host"
      >
        <el-select v-model="form.host" :placeholder="t('userAdmin.selectHost')">
          <el-option label="localhost" value="localhost" />
          <el-option label="%" value="%" />
          <el-option label="127.0.0.1" value="127.0.0.1" />
        </el-select>
      </el-form-item>
      
      <el-form-item :label="t('auth.password')" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          :placeholder="t('userAdmin.enterPasswordPlaceholder')"
          show-password
        />
      </el-form-item>
      
      <el-form-item :label="t('auth.confirmPassword')" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          :placeholder="t('userAdmin.reenterPassword')"
          show-password
        />
      </el-form-item>
      
      <!-- PostgreSQL选项 -->
      <el-form-item 
        v-if="dbType === 'postgresql'" 
        :label="t('userAdmin.superUser')"
      >
        <el-switch v-model="form.is_superuser" />
      </el-form-item>
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
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { createUser } from '@/api/userAdmin'
import { useConnectionStore } from '@/store/connection'
import type { CreateUserRequest } from '@/types/userAdmin'

const { t } = useI18n()

const props = defineProps<{
  modelValue: boolean
  dbType: 'mysql' | 'postgresql'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'success': []
}>()

const connectionStore = useConnectionStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive<CreateUserRequest & { confirmPassword: string }>({
  username: '',
  host: 'localhost',
  password: '',
  confirmPassword: '',
  is_superuser: false
})

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== form.password) {
    callback(new Error(t('userAdmin.passwordMismatch')))
  } else {
    callback()
  }
}

const rules = computed<FormRules>(() => ({
  username: [
    { required: true, message: t('userAdmin.usernameReq'), trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: t('userAdmin.usernamePattern'), trigger: 'blur' }
  ],
  host: [
    { required: true, message: t('userAdmin.selectHostReq'), trigger: 'change' }
  ],
  password: [
    { required: true, message: t('userAdmin.passwordReq'), trigger: 'blur' },
    { min: 6, message: t('userAdmin.passwordMinLength'), trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: t('userAdmin.confirmPasswordReq'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
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
    username: '',
    host: 'localhost',
    password: '',
    confirmPassword: '',
    is_superuser: false
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
      const request: CreateUserRequest = {
        username: form.username,
        password: form.password
      }
      
      if (props.dbType === 'mysql') {
        request.host = form.host
      } else {
        request.is_superuser = form.is_superuser
      }
      
      await createUser(connectionStore.currentConnection.id, request)
      ElMessage.success(t('userAdmin.createUserSuccess'))
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error.message || t('userAdmin.createUserFailed'))
    } finally {
      loading.value = false
    }
  })
}
</script>
