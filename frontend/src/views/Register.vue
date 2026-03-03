<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <h2>注册 LingoSQL</h2>
      </template>
      <el-alert
        v-if="allowRegistration === false"
        title="系统已关闭公开注册"
        type="info"
        description="请联系管理员获取账号，或使用已有账号登录。"
        show-icon
        class="register-alert"
      />
      <el-form
        v-else
        :model="form"
        :rules="rules"
        ref="formRef"
        @submit.prevent="handleRegister"
      >
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" placeholder="请输入用户名（至少3个字符）" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="至少 8 位，需包含大小写字母与数字"
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="请再次输入密码"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleRegister" :loading="loading" style="width: 100%">
            注册
          </el-button>
        </el-form-item>
        <el-form-item>
          <el-link type="primary" @click="$router.push('/login')">已有账号？立即登录</el-link>
        </el-form-item>
      </el-form>
      <div v-if="allowRegistration === false" class="register-back">
        <el-link type="primary" @click="$router.push('/login')">返回登录</el-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import { getInstallStatus } from '@/api/install'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const allowRegistration = ref<boolean | undefined>(undefined)

onMounted(async () => {
  try {
    const res = await getInstallStatus()
    if (res.data.installed && typeof res.data.allow_registration === 'boolean') {
      allowRegistration.value = res.data.allow_registration
    } else {
      allowRegistration.value = true
    }
  } catch {
    allowRegistration.value = true
  }
})

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
})

const validateConfirmPassword = (rule: any, value: string, callback: Function) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码长度至少 8 位', trigger: 'blur' },
    {
      pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$/,
      message: '密码需包含大小写字母与数字',
      trigger: 'blur',
    },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

const handleRegister = async () => {
  if (!formRef.value || allowRegistration.value === false) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await authStore.registerUser({
          username: form.username,
          email: form.email,
          password: form.password,
        })
        ElMessage.success('注册成功')
        router.push('/')
      } catch (error: any) {
        ElMessage.error(error.response?.data?.message || '注册失败')
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.register-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--color-background-secondary);
  padding: var(--spacing-lg);
}

.register-card {
  width: 100%;
  max-width: 400px;
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-medium);
  background: var(--color-background);
}

.register-card :deep(.el-card__header) {
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.register-card h2 {
  margin: 0;
  text-align: center;
  font-size: var(--font-size-h2);
  font-weight: 600;
  color: var(--color-text-primary);
}

.register-card :deep(.el-card__body) {
  padding: var(--spacing-xl);
}

.register-card :deep(.el-form-item) {
  margin-bottom: var(--spacing-lg);
}

.register-card :deep(.el-input__wrapper) {
  border-radius: var(--border-radius-small);
  box-shadow: 0 0 0 1px var(--color-border) inset;
  background-color: var(--color-background);
}

.register-card :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--color-primary) inset;
}

.register-card :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--color-primary) inset;
}

.register-card :deep(.el-button) {
  height: 36px;
  border-radius: var(--border-radius-small);
  font-weight: 500;
}

.register-card :deep(.el-button--primary) {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
}

.register-card :deep(.el-button--primary:hover) {
  background-color: #0969d9;
  border-color: #0969d9;
}

.register-card :deep(.el-link) {
  font-size: var(--font-size-small);
}

.register-alert {
  margin-bottom: var(--spacing-lg);
}

.register-back {
  margin-top: var(--spacing-lg);
  text-align: center;
}
</style>
