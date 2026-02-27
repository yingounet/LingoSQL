<template>
  <el-dialog
    v-model="visible"
    title="修改密码"
    width="500px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="用户名">
        <el-input :value="user?.username" disabled />
      </el-form-item>
      
      <el-form-item 
        v-if="dbType === 'mysql'" 
        label="主机"
      >
        <el-input :value="user?.host || 'localhost'" disabled />
      </el-form-item>
      
      <el-form-item label="新密码" prop="newPassword">
        <el-input
          v-model="form.newPassword"
          type="password"
          placeholder="请输入新密码"
          show-password
        />
      </el-form-item>
      
      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          placeholder="请再次输入新密码"
          show-password
        />
      </el-form-item>
    </el-form>
    
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="loading">
        确认修改
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { changeUserPassword } from '@/api/userAdmin'
import { useConnectionStore } from '@/store/connection'
import type { DatabaseUser } from '@/types/userAdmin'

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
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于6位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

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
      ElMessage.error('请先选择连接')
      return
    }
    
    loading.value = true
    try {
      await changeUserPassword(connectionStore.currentConnection.id, {
        username: props.user.username,
        host: props.user.host,
        new_password: form.newPassword
      })
      ElMessage.success('密码修改成功')
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error.message || '修改密码失败')
    } finally {
      loading.value = false
    }
  })
}
</script>
