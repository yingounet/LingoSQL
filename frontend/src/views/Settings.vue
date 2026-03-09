<template>
  <div class="settings-page">
    <PageHeader
      :title="t('settings.title')"
      :description="t('settings.description')"
    />

    <div class="settings-content">
      <!-- 语言设置 -->
      <el-card class="content-card">
        <template #header>
          <span>{{ t('settings.language') }}</span>
        </template>
        <el-select
          :model-value="locale"
          :placeholder="t('settings.language')"
          style="width: 240px"
          @change="handleLocaleChange"
        >
          <el-option
            v-for="item in supportedLocales"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
      </el-card>

      <!-- 个人信息维护 -->
      <el-card class="content-card">
        <template #header>
          <span>{{ t('settings.profile') }}</span>
        </template>
        <el-form
          ref="profileFormRef"
          :model="profileForm"
          :rules="profileRules"
          label-width="100px"
          @submit.prevent="handleSaveProfile"
        >
          <el-form-item :label="t('settings.username')" prop="username">
            <el-input v-model="profileForm.username" :placeholder="t('settings.enterUsername')" />
          </el-form-item>
          <el-form-item :label="t('settings.email')" prop="email">
            <el-input v-model="profileForm.email" :placeholder="t('settings.enterEmail')" />
          </el-form-item>
          <el-form-item :label="t('settings.registrationTime')">
            <el-input :model-value="createdAtDisplay" disabled />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="profileLoading" @click="handleSaveProfile">
              {{ t('common.save') }}
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 登录密码修改 -->
      <el-card class="content-card">
        <template #header>
          <span>{{ t('settings.changePassword') }}</span>
        </template>
        <el-form
          ref="passwordFormRef"
          :model="passwordForm"
          :rules="passwordRules"
          label-width="100px"
          @submit.prevent="handleChangePassword"
        >
          <el-form-item :label="t('settings.currentPassword')" prop="current_password">
            <el-input
              v-model="passwordForm.current_password"
              type="password"
              :placeholder="t('settings.enterCurrentPassword')"
              show-password
            />
          </el-form-item>
          <el-form-item :label="t('settings.newPassword')" prop="new_password">
            <el-input
              v-model="passwordForm.new_password"
              type="password"
              :placeholder="t('settings.enterNewPassword')"
              show-password
            />
          </el-form-item>
          <el-form-item :label="t('settings.confirmPassword')" prop="confirm_password">
            <el-input
              v-model="passwordForm.confirm_password"
              type="password"
              :placeholder="t('settings.enterConfirmPassword')"
              show-password
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">
              {{ t('settings.changePassword') }}
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import PageHeader from '@/components/layout/PageHeader.vue'
import { useAuthStore } from '@/store/auth'
import { useLocale } from '@/composables/useLocale'
import { updateProfile, changePassword } from '@/api/auth'

const { t } = useI18n()
const { locale, changeLocale, supportedLocales } = useLocale()
const authStore = useAuthStore()

async function handleLocaleChange(newLocale: string) {
  await changeLocale(newLocale as import('@/i18n').AppLocale)
  ElMessage.success(t('common.success'))
}

const profileFormRef = ref<FormInstance>()
const passwordFormRef = ref<FormInstance>()
const profileLoading = ref(false)
const passwordLoading = ref(false)

const profileForm = reactive({
  username: '',
  email: '',
})

const passwordForm = reactive({
  current_password: '',
  new_password: '',
  confirm_password: '',
})

const createdAtDisplay = computed(() => {
  const raw = authStore.user?.created_at
  if (!raw) return '—'
  try {
    const d = new Date(raw)
    return d.toLocaleString('zh-CN')
  } catch {
    return raw
  }
})

function validateConfirmPassword(_rule: unknown, value: string, callback: (e?: Error) => void) {
  if (value !== passwordForm.new_password) {
    callback(new Error(t('settings.passwordMismatch')))
  } else {
    callback()
  }
}

const profileRules: FormRules = {
  username: [
    { required: true, message: t('auth.enterUsername'), trigger: 'blur' },
    { min: 3, max: 50, message: t('settings.usernameLength'), trigger: 'blur' },
  ],
  email: [
    { required: true, message: t('settings.enterEmail'), trigger: 'blur' },
    { type: 'email', message: t('settings.invalidEmail'), trigger: 'blur' },
  ],
}

const passwordRules: FormRules = {
  current_password: [
    { required: true, message: t('settings.enterCurrentPassword'), trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: t('settings.enterNewPassword'), trigger: 'blur' },
    { min: 6, message: t('settings.enterNewPassword'), trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: t('settings.enterConfirmPassword'), trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' },
  ],
}

watch(
  () => authStore.user,
  (u) => {
    if (u) {
      profileForm.username = u.username
      profileForm.email = u.email
    }
  },
  { immediate: true }
)

async function handleSaveProfile() {
  if (!profileFormRef.value) return
  await profileFormRef.value.validate(async (valid) => {
    if (!valid) return
    profileLoading.value = true
    try {
      await updateProfile({
        username: profileForm.username,
        email: profileForm.email,
      })
      ElMessage.success(t('settings.profileUpdated'))
      await authStore.fetchUser()
    } catch (error: unknown) {
      const msg = error instanceof Error ? error.message : '更新失败'
      ElMessage.error(msg)
    } finally {
      profileLoading.value = false
    }
  })
}

async function handleChangePassword() {
  if (!passwordFormRef.value) return
  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return
    passwordLoading.value = true
    try {
      await changePassword({
        current_password: passwordForm.current_password,
        new_password: passwordForm.new_password,
      })
      ElMessage.success(t('settings.passwordChanged'))
      passwordForm.current_password = ''
      passwordForm.new_password = ''
      passwordForm.confirm_password = ''
      passwordFormRef.value.resetFields()
    } catch (error: unknown) {
      const msg = error instanceof Error ? error.message : '修改密码失败'
      ElMessage.error(msg)
    } finally {
      passwordLoading.value = false
    }
  })
}
</script>

<style scoped>
.settings-page {
  max-width: 1400px;
  margin: 0 auto;
}

.settings-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.content-card {
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-light);
  background: var(--color-background);
}

.content-card :deep(.el-card__header) {
  border-bottom: 1px solid var(--color-border);
  padding: var(--spacing-lg);
  font-weight: 600;
  color: var(--color-text-primary);
}

.content-card :deep(.el-card__body) {
  padding: var(--spacing-xl);
}
</style>
