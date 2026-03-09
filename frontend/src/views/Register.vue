<template>
  <div class="register-container">
    <el-card class="register-card">
      <template #header>
        <h2>{{ t('auth.registerLingoSQL') }}</h2>
      </template>
      <el-alert
        v-if="allowRegistration === false"
        :title="t('auth.registrationClosed')"
        type="info"
        :description="t('auth.registrationClosedDesc')"
        show-icon
        class="register-alert"
      />
      <el-form
        v-else
        :model="form"
        :rules="rules"
        ref="formRef"
        label-position="top"
        class="register-form"
        @submit.prevent="handleRegister"
      >
        <el-form-item :label="t('auth.username')" prop="username">
          <el-input v-model="form.username" :placeholder="t('auth.enterUsernameMin3')" />
        </el-form-item>
        <el-form-item :label="t('auth.email')" prop="email">
          <el-input v-model="form.email" :placeholder="t('auth.enterEmail')" />
        </el-form-item>
        <el-form-item :label="t('auth.password')" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            :placeholder="t('auth.passwordRequirement')"
          />
        </el-form-item>
        <el-form-item :label="t('auth.confirmPassword')" prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            :placeholder="t('auth.enterConfirmPassword')"
          />
        </el-form-item>
        <div class="form-actions">
          <el-form-item>
            <el-button type="primary" @click="handleRegister" :loading="loading" style="width: 100%">
              {{ t('auth.register') }}
            </el-button>
          </el-form-item>
          <el-form-item class="form-link-item">
            <el-link type="primary" @click="$router.push('/login')">{{ t('auth.hasAccount') }}</el-link>
          </el-form-item>
        </div>
      </el-form>
      <div v-if="allowRegistration === false" class="register-back">
        <el-link type="primary" @click="$router.push('/login')">{{ t('auth.backToLogin') }}</el-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'
import { getInstallStatus } from '@/api/install'
import type { FormInstance, FormRules } from 'element-plus'

const { t } = useI18n()
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
    callback(new Error(t('auth.passwordMismatch')))
  } else {
    callback()
  }
}

const rules = computed<FormRules>(() => ({
  username: [
    { required: true, message: t('auth.enterUsername'), trigger: 'blur' },
    { min: 3, max: 50, message: t('auth.usernameLength'), trigger: 'blur' },
  ],
  email: [
    { required: true, message: t('auth.enterEmail'), trigger: 'blur' },
    { type: 'email', message: t('auth.invalidEmail'), trigger: 'blur' },
  ],
  password: [
    { required: true, message: t('auth.enterPassword'), trigger: 'blur' },
    { min: 8, message: t('auth.passwordMinLength'), trigger: 'blur' },
    {
      pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d).+$/,
      message: t('auth.passwordPattern'),
      trigger: 'blur',
    },
  ],
  confirmPassword: [
    { required: true, message: t('auth.enterConfirmPasswordRequired'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}))

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
        ElMessage.success(t('auth.registerSuccess'))
        router.push('/')
      } catch (error: any) {
        ElMessage.error(error.response?.data?.message || t('auth.registerFailed'))
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

.register-form :deep(.el-form-item) {
  margin-bottom: var(--spacing-lg);
}

.register-form :deep(.el-form-item__label) {
  padding-bottom: var(--spacing-xs);
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-actions {
  margin-top: var(--spacing-xl);
  padding-top: var(--spacing-md);
}

.form-actions .form-link-item {
  margin-bottom: 0;
}

.form-actions .form-link-item :deep(.el-form-item__content) {
  justify-content: center;
}

.register-form :deep(.el-input__wrapper) {
  border-radius: var(--border-radius-small);
  box-shadow: 0 0 0 1px var(--color-border) inset;
  background-color: var(--color-background);
}

.register-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--color-primary) inset;
}

.register-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--color-primary) inset;
}

.register-form :deep(.el-button) {
  height: 36px;
  border-radius: var(--border-radius-small);
  font-weight: 500;
}

.register-form :deep(.el-button--primary) {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
}

.register-form :deep(.el-button--primary:hover) {
  background-color: #0969d9;
  border-color: #0969d9;
}

.register-form :deep(.el-link) {
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
