<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑连接' : '新建连接'"
    width="680px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-position="top"
      class="connection-form"
    >
      <!-- 连接类型选择 -->
      <el-form-item label="连接方式">
        <el-radio-group v-model="formData.connection_type" class="connection-type-radio">
          <el-radio-button value="direct">
            <el-icon><Connection /></el-icon>
            直连
          </el-radio-button>
          <el-radio-button value="ssh_tunnel">
            <el-icon><Lock /></el-icon>
            SSH 隧道
          </el-radio-button>
        </el-radio-group>
      </el-form-item>

      <!-- 数据库类型选择 -->
      <el-form-item label="数据库类型" prop="db_type">
        <div class="db-type-cards">
          <div
            v-for="(config, type) in DB_TYPE_CONFIG"
            :key="type"
            class="db-type-card"
            :class="{ selected: formData.db_type === type }"
            @click="handleDbTypeChange(type as DbType)"
          >
            <div class="db-type-icon" :style="{ backgroundColor: config.color + '20' }">
              <span :style="{ color: config.color }">{{ config.label[0] }}</span>
            </div>
            <span class="db-type-label">{{ config.label }}</span>
          </div>
        </div>
      </el-form-item>

      <!-- 连接名称 -->
      <el-form-item label="连接名称" prop="name">
        <el-input
          v-model="formData.name"
          placeholder="例如：Production MySQL"
          maxlength="100"
        />
      </el-form-item>

      <!-- 数据库配置区域 -->
      <div class="form-section">
        <h4 class="section-title">数据库配置</h4>
        
        <el-row :gutter="16">
          <el-col :span="16">
            <el-form-item label="主机地址" prop="db_config.host">
              <el-input
                v-model="formData.db_config.host"
                placeholder="localhost 或 IP 地址"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="端口" prop="db_config.port">
              <el-input-number
                v-model="formData.db_config.port"
                :min="1"
                :max="65535"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="用户名" prop="db_config.username">
              <el-input
                v-model="formData.db_config.username"
                placeholder="数据库用户名"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="密码" prop="db_config.password">
              <el-input
                v-model="formData.db_config.password"
                type="password"
                placeholder="数据库密码"
                show-password
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="数据库名（可选）">
          <el-input
            v-model="formData.db_config.database"
            placeholder="默认连接的数据库"
          />
        </el-form-item>

        <!-- SSL 选项 -->
        <el-form-item label="SSL 模式">
          <el-select v-model="formData.db_config.options.ssl_mode" style="width: 100%">
            <el-option value="disable" label="禁用 (disable)" />
            <el-option value="require" label="要求 (require)" />
            <el-option value="verify-ca" label="验证 CA (verify-ca)" />
            <el-option value="verify-full" label="完全验证 (verify-full)" />
          </el-select>
        </el-form-item>
      </div>

      <!-- SSH 配置区域（条件显示） -->
      <div class="form-section" v-if="formData.connection_type === 'ssh_tunnel'">
        <h4 class="section-title">
          <el-icon><Lock /></el-icon>
          SSH 隧道配置
        </h4>

        <el-row :gutter="16">
          <el-col :span="16">
            <el-form-item label="SSH 主机" prop="ssh_config.host">
              <el-input
                v-model="formData.ssh_config.host"
                placeholder="SSH 服务器地址"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="SSH 端口" prop="ssh_config.port">
              <el-input-number
                v-model="formData.ssh_config.port"
                :min="1"
                :max="65535"
                controls-position="right"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="SSH 用户名" prop="ssh_config.username">
          <el-input
            v-model="formData.ssh_config.username"
            placeholder="SSH 登录用户名"
          />
        </el-form-item>

        <el-form-item label="认证方式">
          <el-radio-group v-model="formData.ssh_config.auth_type">
            <el-radio value="password">密码认证</el-radio>
            <el-radio value="private_key">私钥认证</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 密码认证 -->
        <el-form-item
          v-if="formData.ssh_config.auth_type === 'password'"
          label="SSH 密码"
          prop="ssh_config.password"
        >
          <el-input
            v-model="formData.ssh_config.password"
            type="password"
            placeholder="SSH 登录密码"
            show-password
          />
        </el-form-item>

        <!-- 私钥认证 -->
        <template v-if="formData.ssh_config.auth_type === 'private_key'">
          <el-form-item label="私钥" prop="ssh_config.private_key">
            <div class="private-key-input">
              <el-input
                v-model="privateKeyFileName"
                placeholder="选择私钥文件..."
                readonly
              >
                <template #append>
                  <el-button @click="triggerFileInput">
                    <el-icon><Upload /></el-icon>
                    选择文件
                  </el-button>
                </template>
              </el-input>
              <input
                ref="fileInputRef"
                type="file"
                accept=".pem,.key,.ppk"
                style="display: none"
                @change="handleFileSelect"
              />
            </div>
          </el-form-item>

          <el-form-item label="私钥密码（可选）">
            <el-input
              v-model="formData.ssh_config.passphrase"
              type="password"
              placeholder="如果私钥有密码保护"
              show-password
            />
          </el-form-item>
        </template>
      </div>

      <!-- 测试结果 -->
      <div v-if="testResult" class="test-result" :class="testResult.connected ? 'success' : 'error'">
        <el-icon v-if="testResult.connected"><CircleCheck /></el-icon>
        <el-icon v-else><CircleClose /></el-icon>
        <div class="test-result-content">
          <span class="test-result-title">
            {{ testResult.connected ? '连接成功' : '连接失败' }}
          </span>
          <span v-if="testResult.connected" class="test-result-detail">
            {{ testResult.version }} · 延迟 {{ testResult.latency_ms }}ms
          </span>
          <span v-else class="test-result-detail">
            {{ testResult.error }}
          </span>
        </div>
      </div>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button
          @click="handleTestConnection"
          :loading="testing"
          :disabled="!canTest"
        >
          {{ testing ? '测试中...' : '测试连接' }}
        </el-button>
        <div class="footer-right">
          <el-button @click="handleClose">取消</el-button>
          <el-button
            type="primary"
            @click="handleSave"
            :loading="saving"
          >
            {{ saving ? '保存中...' : '保存' }}
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Connection, Lock, Upload, CircleCheck, CircleClose } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useConnectionStore } from '@/store/connection'
import {
  DB_TYPE_CONFIG,
  getDefaultFormData,
  type DbType,
  type ConnectionFormData,
  type TestResult,
} from '@/types/connection'

// Props
const props = defineProps<{
  modelValue: boolean
  connectionId?: number | null
}>()

// Emits
const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'saved'): void
}>()

// Store
const store = useConnectionStore()

// 本地状态
const formRef = ref<FormInstance>()
const formData = ref<ConnectionFormData>(getDefaultFormData())
const isEdit = ref(false)
const testing = ref(false)
const saving = ref(false)
const testResult = ref<TestResult | null>(null)
const privateKeyFileName = ref('')
const fileInputRef = ref<HTMLInputElement>()

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const canTest = computed(() => {
  return (
    formData.value.db_config.host &&
    formData.value.db_config.username &&
    (formData.value.connection_type === 'direct' ||
      formData.value.ssh_config.host)
  )
})

// 表单验证规则
const formRules: FormRules = {
  name: [
    { required: true, message: '请输入连接名称', trigger: 'blur' },
    { max: 100, message: '连接名称不能超过100个字符', trigger: 'blur' },
  ],
  db_type: [
    { required: true, message: '请选择数据库类型', trigger: 'change' },
  ],
  'db_config.host': [
    { required: true, message: '请输入主机地址', trigger: 'blur' },
  ],
  'db_config.port': [
    { required: true, message: '请输入端口号', trigger: 'blur' },
  ],
  'db_config.username': [
    { required: true, message: '请输入用户名', trigger: 'blur' },
  ],
  'db_config.password': [
    { required: true, message: '请输入密码', trigger: 'blur' },
  ],
  'ssh_config.host': [
    {
      validator: (rule, value, callback) => {
        if (formData.value.connection_type === 'ssh_tunnel' && !value) {
          callback(new Error('请输入 SSH 主机地址'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  'ssh_config.username': [
    {
      validator: (rule, value, callback) => {
        if (formData.value.connection_type === 'ssh_tunnel' && !value) {
          callback(new Error('请输入 SSH 用户名'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  'ssh_config.password': [
    {
      validator: (rule, value, callback) => {
        if (
          formData.value.connection_type === 'ssh_tunnel' &&
          formData.value.ssh_config.auth_type === 'password' &&
          !value
        ) {
          callback(new Error('请输入 SSH 密码'))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  'ssh_config.private_key': [
    {
      validator: (rule, value, callback) => {
        if (
          formData.value.connection_type === 'ssh_tunnel' &&
          formData.value.ssh_config.auth_type === 'private_key' &&
          !value
        ) {
          callback(new Error('请选择私钥文件'))
        } else {
          callback()
        }
      },
      trigger: 'change',
    },
  ],
}

// 数据库类型变更
function handleDbTypeChange(type: DbType) {
  formData.value.db_type = type
  formData.value.db_config.port = DB_TYPE_CONFIG[type].defaultPort
  testResult.value = null
}

// 触发文件选择
function triggerFileInput() {
  fileInputRef.value?.click()
}

// 处理文件选择
function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    privateKeyFileName.value = file.name
    const reader = new FileReader()
    reader.onload = (e) => {
      formData.value.ssh_config.private_key = e.target?.result as string
    }
    reader.readAsText(file)
  }
}

// 测试连接
async function handleTestConnection() {
  testing.value = true
  testResult.value = null

  try {
    const result = await store.testConnectionConfig({
      db_type: formData.value.db_type,
      connection_type: formData.value.connection_type,
      db_config: formData.value.db_config,
      ssh_config: formData.value.connection_type === 'ssh_tunnel'
        ? formData.value.ssh_config
        : undefined,
    })
    testResult.value = result
  } catch (error: any) {
    testResult.value = {
      connected: false,
      error: error.message || '测试失败',
    }
  } finally {
    testing.value = false
  }
}

// 保存
async function handleSave() {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  saving.value = true

  try {
    if (isEdit.value && props.connectionId) {
      await store.updateConnection(props.connectionId, formData.value)
      ElMessage.success('连接已更新')
    } else {
      await store.createConnection(formData.value)
      ElMessage.success('连接已创建')
    }
    emit('saved')
    handleClose()
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
  } finally {
    saving.value = false
  }
}

// 关闭
function handleClose() {
  visible.value = false
  resetForm()
}

// 重置表单
function resetForm() {
  formData.value = getDefaultFormData()
  isEdit.value = false
  testResult.value = null
  privateKeyFileName.value = ''
  formRef.value?.resetFields()
}

// 加载连接数据（编辑模式）
async function loadConnection(id: number) {
  const connection = await store.getConnection(id)
  if (connection) {
    isEdit.value = true
    formData.value = {
      name: connection.name,
      db_type: connection.db_type,
      connection_type: connection.connection_type,
      db_config: {
        host: connection.db_config.host,
        port: connection.db_config.port,
        database: connection.db_config.database || '',
        username: connection.db_config.username,
        password: '', // 密码不回显
        options: {
          ssl_mode: connection.db_config.options?.ssl_mode || 'disable',
          charset: connection.db_config.options?.charset || 'utf8mb4',
          timeout: connection.db_config.options?.timeout || 30,
        },
      },
      ssh_config: {
        host: connection.ssh_config?.host || '',
        port: connection.ssh_config?.port || 22,
        username: connection.ssh_config?.username || '',
        auth_type: connection.ssh_config?.auth_type || 'password',
        password: '',
        private_key: '',
        passphrase: '',
      },
    }
  }
}

// 监听 connectionId 变化
watch(
  () => props.connectionId,
  (newId) => {
    if (newId && props.modelValue) {
      loadConnection(newId)
    }
  }
)

// 监听对话框打开
watch(
  () => props.modelValue,
  (newVal) => {
    if (newVal) {
      if (props.connectionId) {
        loadConnection(props.connectionId)
      } else {
        resetForm()
      }
    }
  }
)
</script>

<style scoped>
.connection-form {
  max-height: 60vh;
  overflow-y: auto;
  padding-right: var(--spacing-sm);
}

/* 连接类型选择 */
.connection-type-radio {
  width: 100%;
}

.connection-type-radio :deep(.el-radio-button__inner) {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-lg);
}

/* 数据库类型卡片 */
.db-type-cards {
  display: flex;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.db-type-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  border: 2px solid var(--color-border);
  border-radius: var(--border-radius-medium);
  cursor: pointer;
  transition: all 0.2s;
  min-width: 100px;
}

.db-type-card:hover {
  border-color: var(--color-primary);
  background-color: var(--color-background-secondary);
}

.db-type-card.selected {
  border-color: var(--color-primary);
  background-color: var(--color-nav-active-bg);
}

.db-type-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--border-radius-medium);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 700;
}

.db-type-label {
  font-size: var(--font-size-small);
  font-weight: 500;
  color: var(--color-text-primary);
}

/* 表单区域 */
.form-section {
  margin-top: var(--spacing-lg);
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  margin: 0 0 var(--spacing-md) 0;
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

/* 私钥输入 */
.private-key-input {
  width: 100%;
}

/* 测试结果 */
.test-result {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  padding: var(--spacing-md);
  border-radius: var(--border-radius-medium);
  margin-top: var(--spacing-lg);
}

.test-result.success {
  background-color: #f0f9eb;
  color: var(--color-success);
}

.test-result.error {
  background-color: #fef0f0;
  color: var(--color-danger);
}

.test-result .el-icon {
  font-size: 20px;
  margin-top: 2px;
}

.test-result-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.test-result-title {
  font-weight: 600;
}

.test-result-detail {
  font-size: var(--font-size-small);
  opacity: 0.8;
}

/* 对话框底部 */
.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.footer-right {
  display: flex;
  gap: var(--spacing-sm);
}

/* 响应式 */
@media (max-width: 640px) {
  .db-type-cards {
    justify-content: center;
  }

  .db-type-card {
    min-width: 80px;
    padding: var(--spacing-sm);
  }
}
</style>
