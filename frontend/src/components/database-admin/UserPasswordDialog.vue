<template>
  <el-dialog
    v-model="visible"
    :title="t('userAdmin.changePasswordTitle')"
    width="500px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item :label="t('auth.username')">
        <el-input :value="user?.username" disabled />
      </el-form-item>
      
      <el-form-item 
        v-if="dbType === 'mysql'" 
        :label="t('userAdmin.hostCol')"
      >
        <el-input :value="user?.host || 'localhost'" disabled />
      </el-form-item>
      
      <el-form-item :label="t('userAdmin.newPassword')" prop="newPassword">
        <el-input
          v-model="form.newPassword"
          type="password"
          :placeholder="t('userAdmin.enterNewPassword')"
          show-password
        />
      </el-form-item>
      
      <el-form-item :label="t('userAdmin.confirmNewPassword')" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          :placeholder="t('userAdmin.reenterNewPassword')"
          show-password
        />
      </el-form-item>
    </el-form>
    
    <template #footer>
      <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="loading">
        {{ t('userAdmin.confirmChange') }}
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { changeUserPassword } from '@/api/userAdmin'
import { useConnectionStore } from '@/store/connection'
import type { DatabaseUser } from '@/types/userAdmin'

const { t } = useI18n()

const props = defineProps<{
  modelValue: boolean
  user: DatabaseUser | null
  dbType: 'mysql' | 'postgresql'
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'success': []
}>()

const connectionStore = useConnectionStore()
const formRef = ref<FormInstance>()
const loading = ref(false)

const form = reactive({
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (_rule: any, value: string, callback: any) => {
  if (value !== form.newPassword) {
    callback(new Error(t('userAdmin.passwordMismatch')))
  } else {
    callback()
  }
}

const rules = computed<FormRules>(() => ({
  newPassword: [
    { required: true, message: t('userAdmin.newPasswordReq'), trigger: 'blur' },
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
  form.newPassword = ''
  form.confirmPassword = ''
}

async function handleSubmit() {
  if (!formRef.value || !props.user) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    if (!connectionStore.currentConnection) {
      ElMessage.error(t('dbAdmin.selectConnectionFirst'))
      return
    }
    
    loading.value = true
    try {
      await changeUserPassword(connectionStore.currentConnection.id, {
        username: props.user.username,
        host: props.user.host,
        new_password: form.newPassword
      })
      ElMessage.success(t('userAdmin.changePasswordSuccess'))
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error.message || t('userAdmin.changePasswordFailed'))
    } finally {
      loading.value = false
    }
  })
}
</script>
