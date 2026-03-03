<template>
  <div class="install-container">
    <el-card class="install-card">
      <template #header>
        <h2>LingoSQL 安装引导</h2>
        <p class="install-subtitle">欢迎！请完成以下步骤完成首次配置</p>
      </template>

      <el-steps :active="currentStep" finish-status="success" align-center>
        <el-step title="创建管理员" />
        <el-step title="系统配置" />
      </el-steps>

      <el-form
        ref="formRef"
        :model="form"
        :rules="stepRules"
        label-position="top"
        class="install-form"
      >
        <!-- Step 1: 管理员账号 -->
        <div v-show="currentStep === 0" class="step-content">
          <el-form-item label="用户名" prop="admin.username">
            <el-input
              v-model="form.admin.username"
              placeholder="请输入管理员用户名（3-50 字符）"
            />
          </el-form-item>
          <el-form-item label="邮箱" prop="admin.email">
            <el-input
              v-model="form.admin.email"
              placeholder="请输入管理员邮箱"
            />
          </el-form-item>
          <el-form-item label="密码" prop="admin.password">
            <el-input
              v-model="form.admin.password"
              type="password"
              placeholder="至少 8 位，需包含大小写字母与数字"
              show-password
            />
          </el-form-item>
          <el-form-item label="确认密码" prop="admin.confirmPassword">
            <el-input
              v-model="form.admin.confirmPassword"
              type="password"
              placeholder="请再次输入密码"
              show-password
            />
          </el-form-item>
        </div>

        <!-- Step 2: 系统配置 -->
        <div v-show="currentStep === 1" class="step-content">
          <el-form-item label="站点名称" prop="settings.site_name">
            <el-input
              v-model="form.settings.site_name"
              placeholder="LingoSQL"
            />
          </el-form-item>
          <el-form-item label="允许公开注册" prop="settings.allow_registration">
            <el-switch v-model="form.settings.allow_registration" />
            <span class="form-tip">开启后，访客可自行注册账号</span>
          </el-form-item>
          <el-form-item label="启用 API 限流" prop="settings.rate_limit_enabled">
            <el-switch v-model="form.settings.rate_limit_enabled" />
            <span class="form-tip">限制接口请求频率，提升稳定性</span>
          </el-form-item>
          <template v-if="form.settings.rate_limit_enabled">
            <el-form-item label="默认 API 限流（RPM）" prop="settings.rate_limit_default_rpm">
              <el-input-number
                v-model="form.settings.rate_limit_default_rpm"
                :min="10"
                :max="1000"
              />
            </el-form-item>
            <el-form-item label="轮询接口限流（RPM）" prop="settings.rate_limit_polling_rpm">
              <el-input-number
                v-model="form.settings.rate_limit_polling_rpm"
                :min="5"
                :max="200"
              />
            </el-form-item>
          </template>
          <el-form-item label="CORS 允许来源" prop="settings.cors_allowed_origins">
            <el-input
              v-model="corsOriginsText"
              type="textarea"
              :rows="3"
              placeholder="每行一个 URL，如 http://localhost:5173"
            />
          </el-form-item>
        </div>

        <div class="form-actions">
          <el-button v-if="currentStep > 0" @click="currentStep--">
            上一步
          </el-button>
          <el-button
            v-if="currentStep < 1"
            type="primary"
            :loading="loading"
            @click="handleNextStep"
          >
            下一步
          </el-button>
          <el-button
            v-else
            type="primary"
            :loading="loading"
            @click="handleSubmit"
          >
            完成安装
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import { setupInstall } from '@/api/install'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const currentStep = ref(0)

const form = reactive({
  admin: {
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
  },
  settings: {
    site_name: 'LingoSQL',
    allow_registration: false,
    rate_limit_enabled: true,
    rate_limit_default_rpm: 120,
    rate_limit_polling_rpm: 30,
    cors_allowed_origins: ['http://localhost:5173'] as string[],
  },
})

const corsOriginsText = computed({
  get: () => form.settings.cors_allowed_origins.join('\n'),
  set: (val: string) => {
    form.settings.cors_allowed_origins = val
      .split(/[\n,]/)
      .map((s) => s.trim())
      .filter(Boolean)
  },
})

const validateConfirmPassword = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value !== form.admin.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const stepRules: FormRules = {
  'admin.username': [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' },
  ],
  'admin.email': [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  'admin.password': [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码长度至少 8 位', trigger: 'blur' },
    {
      pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$/,
      message: '密码需包含大小写字母与数字',
      trigger: 'blur',
    },
  ],
  'admin.confirmPassword': [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
  'settings.site_name': [
    { required: true, message: '请输入站点名称', trigger: 'blur' },
  ],
  'settings.rate_limit_default_rpm': [
    { required: true, message: '请输入默认限流值', trigger: 'blur' },
    { type: 'number', min: 10, max: 1000, message: '范围 10-1000', trigger: 'blur' },
  ],
  'settings.rate_limit_polling_rpm': [
    { required: true, message: '请输入轮询限流值', trigger: 'blur' },
    { type: 'number', min: 5, max: 200, message: '范围 5-200', trigger: 'blur' },
  ],
}

const step0Fields = ['admin.username', 'admin.email', 'admin.password', 'admin.confirmPassword']
const getStep1Fields = () => {
  const base = ['settings.site_name']
  if (form.settings.rate_limit_enabled) {
    base.push('settings.rate_limit_default_rpm', 'settings.rate_limit_polling_rpm')
  }
  return base
}

const handleNextStep = async () => {
  if (!formRef.value) return
  await formRef.value.validateField(step0Fields, (valid) => {
    if (valid) {
      currentStep.value++
    }
  })
}

const handleSubmit = async () => {
  if (!formRef.value) return

  const fieldsToValidate = [...step0Fields, ...getStep1Fields()]
  await formRef.value.validateField(fieldsToValidate, async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const res = await setupInstall({
        admin: {
          username: form.admin.username,
          email: form.admin.email,
          password: form.admin.password,
        },
        settings: {
          site_name: form.settings.site_name || 'LingoSQL',
          allow_registration: form.settings.allow_registration,
          rate_limit_enabled: form.settings.rate_limit_enabled,
          rate_limit_default_rpm: form.settings.rate_limit_default_rpm,
          rate_limit_polling_rpm: form.settings.rate_limit_polling_rpm,
          cors_allowed_origins:
            form.settings.cors_allowed_origins.length > 0
              ? form.settings.cors_allowed_origins
              : ['http://localhost:5173'],
        },
      })
      authStore.setAuth(res.data.token, res.data.user)
      ElMessage.success('安装成功')
      router.push('/')
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } }; message?: string }
      ElMessage.error(e.response?.data?.message || e.message || '安装失败')
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.install-container {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  min-height: 100vh;
  background: var(--color-background-secondary);
  padding: var(--spacing-xl);
}

.install-card {
  width: 100%;
  max-width: 480px;
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-medium);
  background: var(--color-background);
}

.install-card :deep(.el-card__header) {
  padding: var(--spacing-xl);
  border-bottom: 1px solid var(--color-border);
}

.install-card h2 {
  margin: 0 0 var(--spacing-sm);
  text-align: center;
  font-size: var(--font-size-h2);
  font-weight: 600;
  color: var(--color-text-primary);
}

.install-subtitle {
  margin: 0;
  text-align: center;
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

.install-form {
  padding: var(--spacing-xl) 0;
}

.step-content {
  margin: var(--spacing-xl) 0;
}

.form-tip {
  margin-left: var(--spacing-sm);
  font-size: var(--font-size-small);
  color: var(--color-text-secondary);
}

.form-actions {
  display: flex;
  justify-content: center;
  gap: var(--spacing-md);
  margin-top: var(--spacing-xl);
}

.install-card :deep(.el-input__wrapper),
.install-card :deep(.el-textarea__inner) {
  border-radius: var(--border-radius-small);
}

.install-card :deep(.el-button) {
  border-radius: var(--border-radius-small);
}
</style>
