<template>
  <div class="settings-page">
    <PageHeader
      title="个人设置"
      description="管理您的账户信息与登录密码"
    />

    <div class="settings-content">
      <!-- 个人信息维护 -->
      <el-card class="content-card">
        <template #header>
          <span>个人信息</span>
        </template>
        <el-form
          ref="profileFormRef"
          :model="profileForm"
          :rules="profileRules"
          label-width="100px"
          @submit.prevent="handleSaveProfile"
        >
          <el-form-item label="用户名" prop="username">
            <el-input v-model="profileForm.username" placeholder="请输入用户名（至少3个字符）" />
          </el-form-item>
          <el-form-item label="邮箱" prop="email">
            <el-input v-model="profileForm.email" placeholder="请输入邮箱" />
          </el-form-item>
          <el-form-item label="注册时间">
            <el-input :model-value="createdAtDisplay" disabled />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="profileLoading" @click="handleSaveProfile">
              保存
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 登录密码修改 -->
      <el-card class="content-card">
        <template #header>
          <span>修改密码</span>
        </template>
        <el-form
          ref="passwordFormRef"
          :model="passwordForm"
          :rules="passwordRules"
          label-width="100px"
          @submit.prevent="handleChangePassword"
        >
          <el-form-item label="当前密码" prop="current_password">
            <el-input
              v-model="passwordForm.current_password"
              type="password"
              placeholder="请输入当前密码"
              show-password
            />
          </el-form-item>
          <el-form-item label="新密码" prop="new_password">
            <el-input
              v-model="passwordForm.new_password"
              type="password"
              placeholder="请输入新密码（至少6个字符）"
              show-password
            />
          </el-form-item>
          <el-form-item label="确认新密码" prop="confirm_password">
            <el-input
              v-model="passwordForm.confirm_password"
              type="password"
              placeholder="请再次输入新密码"
              show-password
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="passwordLoading" @click="handleChangePassword">
              修改密码
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import PageHeader from '@/components/layout/PageHeader.vue'
import { useAuthStore } from '@/store/auth'
import { updateProfile, changePassword } from '@/api/auth'

const authStore = useAuthStore()

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
    callback(new Error('两次输入的新密码不一致'))
  } else {
    callback()
  }
}

const profileRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 50, message: '用户名长度在 3 到 50 个字符', trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
}

const passwordRules: FormRules = {
  current_password: [
    { required: true, message: '请输入当前密码', trigger: 'blur' },
  ],
  new_password: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度至少 6 个字符', trigger: 'blur' },
  ],
  confirm_password: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
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
      ElMessage.success('个人信息已更新')
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
      ElMessage.success('密码已修改')
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
