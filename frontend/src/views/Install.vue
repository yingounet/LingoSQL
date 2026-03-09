<template>
  <div class="install-container">
    <el-card class="install-card">
      <template #header>
        <h2>{{ t('install.title') }}</h2>
        <p class="install-subtitle">{{ t('install.subtitle') }}</p>
      </template>

      <el-steps :active="currentStep" finish-status="success" align-center>
        <el-step :title="t('install.stepAdmin')" />
        <el-step :title="t('install.stepConfig')" />
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
          <el-form-item :label="t('auth.username')" prop="admin.username">
            <el-input
              v-model="form.admin.username"
              :placeholder="t('auth.enterUsernameMin3')"
            />
          </el-form-item>
          <el-form-item :label="t('auth.email')" prop="admin.email">
            <el-input
              v-model="form.admin.email"
              :placeholder="t('auth.enterEmail')"
            />
          </el-form-item>
          <el-form-item :label="t('auth.password')" prop="admin.password">
            <el-input
              v-model="form.admin.password"
              type="password"
              :placeholder="t('auth.passwordRequirement')"
              show-password
            />
          </el-form-item>
          <el-form-item :label="t('auth.confirmPassword')" prop="admin.confirmPassword">
            <el-input
              v-model="form.admin.confirmPassword"
              type="password"
              :placeholder="t('auth.enterConfirmPassword')"
              show-password
            />
          </el-form-item>
        </div>

        <!-- Step 2: 系统配置 -->
        <div v-show="currentStep === 1" class="step-content">
          <el-form-item :label="t('install.siteName')" prop="settings.site_name">
            <el-input
              v-model="form.settings.site_name"
              placeholder="LingoSQL"
            />
          </el-form-item>
          <el-form-item :label="t('install.allowRegistration')" prop="settings.allow_registration">
            <el-switch v-model="form.settings.allow_registration" />
            <span class="form-tip">{{ t('install.allowRegistrationTip') }}</span>
          </el-form-item>
          <el-form-item :label="t('install.enableRateLimit')" prop="settings.rate_limit_enabled">
            <el-switch v-model="form.settings.rate_limit_enabled" />
            <span class="form-tip">{{ t('install.enableRateLimitTip') }}</span>
          </el-form-item>
          <template v-if="form.settings.rate_limit_enabled">
            <el-form-item :label="t('install.defaultRateLimit')" prop="settings.rate_limit_default_rpm">
              <el-input-number
                v-model="form.settings.rate_limit_default_rpm"
                :min="10"
                :max="1000"
              />
            </el-form-item>
            <el-form-item :label="t('install.pollingRateLimit')" prop="settings.rate_limit_polling_rpm">
              <el-input-number
                v-model="form.settings.rate_limit_polling_rpm"
                :min="5"
                :max="200"
              />
            </el-form-item>
          </template>
          <el-form-item :label="t('install.corsOrigins')" prop="settings.cors_allowed_origins">
            <el-input
              v-model="corsOriginsText"
              type="textarea"
              :rows="3"
              :placeholder="t('install.corsOriginsPlaceholder')"
            />
          </el-form-item>
        </div>

        <div class="form-actions">
          <el-button v-if="currentStep > 0" @click="currentStep--">
            {{ t('install.prevStep') }}
          </el-button>
          <el-button
            v-if="currentStep < 1"
            type="primary"
            :loading="loading"
            @click="handleNextStep"
          >
            {{ t('install.nextStep') }}
          </el-button>
          <el-button
            v-else
            type="primary"
            :loading="loading"
            @click="handleSubmit"
          >
            {{ t('install.finishInstall') }}
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import { setupInstall } from '@/api/install'
import type { FormInstance, FormRules } from 'element-plus'

const { t } = useI18n()
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
    callback(new Error(t('auth.passwordMismatch')))
  } else {
    callback()
  }
}

const stepRules = computed<FormRules>(() => ({
  'admin.username': [
    { required: true, message: t('auth.enterUsername'), trigger: 'blur' },
    { min: 3, max: 50, message: t('auth.usernameLength'), trigger: 'blur' },
  ],
  'admin.email': [
    { required: true, message: t('auth.enterEmail'), trigger: 'blur' },
    { type: 'email', message: t('auth.invalidEmail'), trigger: 'blur' },
  ],
  'admin.password': [
    { required: true, message: t('auth.enterPassword'), trigger: 'blur' },
    { min: 8, message: t('auth.passwordMinLength'), trigger: 'blur' },
    {
      pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$/,
      message: t('auth.passwordPattern'),
      trigger: 'blur',
    },
  ],
  'admin.confirmPassword': [
    { required: true, message: t('auth.enterConfirmPasswordRequired'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
  'settings.site_name': [
    { required: true, message: t('install.enterSiteName'), trigger: 'blur' },
  ],
  'settings.rate_limit_default_rpm': [
    { required: true, message: t('install.enterDefaultRateLimit'), trigger: 'blur' },
    { type: 'number', min: 10, max: 1000, message: t('install.rateRange10_1000'), trigger: 'blur' },
  ],
  'settings.rate_limit_polling_rpm': [
    { required: true, message: t('install.enterPollingRateLimit'), trigger: 'blur' },
    { type: 'number', min: 5, max: 200, message: t('install.rateRange5_200'), trigger: 'blur' },
  ],
}))

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
      ElMessage.success(t('install.installSuccess'))
      router.push('/')
    } catch (err: unknown) {
      const e = err as { response?: { data?: { message?: string } }; message?: string }
      ElMessage.error(e.response?.data?.message || e.message || t('install.installFailed'))
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
