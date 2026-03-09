<template>
  <div class="login-container">
    <el-card class="login-card">
      <template #header>
        <h2>{{ t('auth.loginLingoSQL') }}</h2>
      </template>
      <el-form
        :model="form"
        :rules="rules"
        ref="formRef"
        label-position="top"
        class="login-form"
        @submit.prevent="handleLogin"
      >
        <el-form-item :label="t('auth.username')" prop="username">
          <el-input v-model="form.username" :placeholder="t('auth.enterUsername')" />
        </el-form-item>
        <el-form-item :label="t('auth.password')" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            :placeholder="t('auth.enterPassword')"
            @keyup.enter="handleLogin"
          />
        </el-form-item>
        <div class="form-actions">
          <el-form-item>
            <el-button type="primary" @click="handleLogin" :loading="loading" style="width: 100%">
              {{ t('auth.login') }}
            </el-button>
          </el-form-item>
          <el-form-item v-if="allowRegistration !== false" class="form-link-item">
            <el-link type="primary" @click="$router.push('/register')">{{ t('auth.noAccount') }}</el-link>
          </el-form-item>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
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
  password: '',
})

const rules: FormRules = {
  username: [{ required: true, message: () => t('auth.enterUsername'), trigger: 'blur' }],
  password: [{ required: true, message: () => t('auth.enterPassword'), trigger: 'blur' }],
}

const handleLogin = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      loading.value = true
      try {
        await authStore.loginUser(form)
        ElMessage.success(t('auth.loginSuccess'))
        router.push('/')
      } catch (error: any) {
        ElMessage.error(error.response?.data?.message || t('auth.loginFailed'))
      } finally {
        loading.value = false
      }
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background: var(--color-background-secondary);
  padding: var(--spacing-lg);
}

.login-card {
  width: 100%;
  max-width: 400px;
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-medium);
  background: var(--color-background);
}

.login-card :deep(.el-card__header) {
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.login-card h2 {
  margin: 0;
  text-align: center;
  font-size: var(--font-size-h2);
  font-weight: 600;
  color: var(--color-text-primary);
}

.login-card :deep(.el-card__body) {
  padding: var(--spacing-xl);
}

.login-form :deep(.el-form-item) {
  margin-bottom: var(--spacing-lg);
}

.login-form :deep(.el-form-item__label) {
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

.login-form :deep(.el-input__wrapper) {
  border-radius: var(--border-radius-small);
  box-shadow: 0 0 0 1px var(--color-border) inset;
  background-color: var(--color-background);
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--color-primary) inset;
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--color-primary) inset;
}

.login-form :deep(.el-button) {
  height: 36px;
  border-radius: var(--border-radius-small);
  font-weight: 500;
}

.login-form :deep(.el-button--primary) {
  background-color: var(--color-primary);
  border-color: var(--color-primary);
}

.login-form :deep(.el-button--primary:hover) {
  background-color: #0969d9;
  border-color: #0969d9;
}

.login-form :deep(.el-link) {
  font-size: var(--font-size-small);
}
</style>
