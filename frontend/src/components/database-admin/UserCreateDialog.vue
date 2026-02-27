<template>
  <el-dialog
    v-model="visible"
    title="创建用户"
    width="500px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="用户名" prop="username">
        <el-input v-model="form.username" placeholder="请输入用户名" />
      </el-form-item>
      
      <!-- MySQL选项 -->
      <el-form-item 
        v-if="dbType === 'mysql'" 
        label="主机" 
        prop="host"
      >
        <el-select v-model="form.host" placeholder="请选择主机">
          <el-option label="localhost" value="localhost" />
          <el-option label="%" value="%" />
          <el-option label="127.0.0.1" value="127.0.0.1" />
        </el-select>
      </el-form-item>
      
      <el-form-item label="密码" prop="password">
        <el-input
          v-model="form.password"
          type="password"
          placeholder="请输入密码"
          show-password
        />
      </el-form-item>
      
      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          placeholder="请再次输入密码"
          show-password
        />
      </el-form-item>
      
      <!-- PostgreSQL选项 -->
      <el-form-item 
        v-if="dbType === 'postgresql'" 
        label="超级用户"
      >
        <el-switch v-model="form.is_superuser" />
      </el-form-item>
    </el-form>
    
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="loading">
        创建
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { createUser } from '@/api/userAdmin'
import { useConnectionStore } from '@/store/connection'
import type { CreateUserRequest } from '@/types/userAdmin'

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
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: '用户名只能包含字母、数字和下划线，且不能以数字开头', trigger: 'blur' }
  ],
  host: [
    { required: true, message: '请选择主机', trigger: 'change' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
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
      ElMessage.error('请先选择连接')
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
      ElMessage.success('用户创建成功')
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error.message || '创建用户失败')
    } finally {
      loading.value = false
    }
  })
}
</script>
