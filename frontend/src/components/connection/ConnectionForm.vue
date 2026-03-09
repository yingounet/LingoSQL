<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? t('connection.editConnection') : t('connection.newConnection')"
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
      <el-form-item :label="t('connection.connMethod')">
        <el-radio-group v-model="formData.connection_type" class="connection-type-radio">
          <el-radio-button value="direct">
            <el-icon><Connection /></el-icon>
            {{ t('connection.direct') }}
          </el-radio-button>
          <el-radio-button value="ssh_tunnel">
            <el-icon><Lock /></el-icon>
            {{ t('connection.sshTunnel') }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>

      <!-- 数据库类型选择 -->
      <el-form-item :label="t('connection.dbType')" prop="db_type">
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
      <el-form-item :label="t('connection.connName')" prop="name">
        <el-input
          v-model="formData.name"
          :placeholder="t('connection.connNamePlaceholder')"
          maxlength="100"
        />
      </el-form-item>

      <!-- 数据库配置区域 -->
      <div class="form-section">
        <h4 class="section-title">{{ t('connection.dbConfig') }}</h4>
        
        <el-row :gutter="16">
          <el-col :span="16">
            <el-form-item :label="t('connection.hostAddress')" prop="db_config.host">
              <el-input
                v-model="formData.db_config.host"
                :placeholder="t('connection.hostPlaceholder')"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('connection.port')" prop="db_config.port">
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
            <el-form-item :label="t('connection.username')" prop="db_config.username">
              <el-input
                v-model="formData.db_config.username"
                :placeholder="t('connection.dbUsernamePlaceholder')"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item :label="t('connection.password')" prop="db_config.password">
              <el-input
                v-model="formData.db_config.password"
                type="password"
                :placeholder="t('connection.dbPasswordPlaceholder')"
                show-password
              />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item :label="t('connection.dbNameOptional')">
          <el-input
            v-model="formData.db_config.database"
            :placeholder="t('connection.defaultDb')"
          />
        </el-form-item>

        <!-- SSL 选项 -->
        <el-form-item :label="t('connection.sslMode')">
          <el-select v-model="formData.db_config.options.ssl_mode" style="width: 100%">
            <el-option value="disable" :label="t('connection.sslDisable')" />
            <el-option value="require" :label="t('connection.sslRequire')" />
            <el-option value="verify-ca" :label="t('connection.sslVerifyCa')" />
            <el-option value="verify-full" :label="t('connection.sslVerifyFull')" />
          </el-select>
        </el-form-item>
      </div>

      <!-- SSH 配置区域（条件显示） -->
      <div class="form-section" v-if="formData.connection_type === 'ssh_tunnel'">
        <h4 class="section-title">
          <el-icon><Lock /></el-icon>
          {{ t('connection.sshConfig') }}
        </h4>

        <el-row :gutter="16">
          <el-col :span="16">
            <el-form-item :label="t('connection.sshHost')" prop="ssh_config.host">
              <el-input
                v-model="formData.ssh_config.host"
                :placeholder="t('connection.sshHostPlaceholder')"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item :label="t('connection.sshPort')" prop="ssh_config.port">
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

        <el-form-item :label="t('connection.sshUsername')" prop="ssh_config.username">
          <el-input
            v-model="formData.ssh_config.username"
            :placeholder="t('connection.sshUsernamePlaceholder')"
          />
        </el-form-item>

        <el-form-item :label="t('connection.authMethod')">
          <el-radio-group v-model="formData.ssh_config.auth_type">
            <el-radio value="password">{{ t('connection.passwordAuth') }}</el-radio>
            <el-radio value="private_key">{{ t('connection.privateKeyAuth') }}</el-radio>
          </el-radio-group>
        </el-form-item>

        <!-- 密码认证 -->
        <el-form-item
          v-if="formData.ssh_config.auth_type === 'password'"
          :label="t('connection.sshPassword')"
          prop="ssh_config.password"
        >
          <el-input
            v-model="formData.ssh_config.password"
            type="password"
            :placeholder="t('connection.sshPasswordPlaceholder')"
            show-password
          />
        </el-form-item>

        <!-- 私钥认证 -->
        <template v-if="formData.ssh_config.auth_type === 'private_key'">
          <el-form-item :label="t('connection.privateKey')" prop="ssh_config.private_key">
            <div class="private-key-input">
              <el-input
                v-model="privateKeyFileName"
                :placeholder="t('connection.selectKeyFile')"
                readonly
              >
                <template #append>
                  <el-button @click="triggerFileInput">
                    <el-icon><Upload /></el-icon>
                    {{ t('connection.selectFile') }}
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

          <el-form-item :label="t('connection.privateKeyPassphrase')">
            <el-input
              v-model="formData.ssh_config.passphrase"
              type="password"
              :placeholder="t('connection.keyPassphrasePlaceholder')"
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
            {{ testResult.connected ? t('connection.testSuccess') : t('connection.testFailed') }}
          </span>
          <span v-if="testResult.connected" class="test-result-detail">
            {{ testResult.version }} · {{ t('connection.latency') }} {{ testResult.latency_ms }}ms
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
          {{ testing ? t('connection.testing') : t('connection.testConnection') }}
        </el-button>
        <div class="footer-right">
          <el-button @click="handleClose">{{ t('common.cancel') }}</el-button>
          <el-button
            type="primary"
            @click="handleSave"
            :loading="saving"
          >
            {{ saving ? t('connection.saving') : t('common.save') }}
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
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

const { t } = useI18n()

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

// 表单验证规则（computed 以响应语言切换）
const formRules = computed<FormRules>(() => ({
  name: [
    { required: true, message: t('connection.enterConnName'), trigger: 'blur' },
    { max: 100, message: t('connection.connNameTooLong'), trigger: 'blur' },
  ],
  db_type: [
    { required: true, message: t('connection.selectDbType'), trigger: 'change' },
  ],
  'db_config.host': [
    { required: true, message: t('connection.enterHost'), trigger: 'blur' },
  ],
  'db_config.port': [
    { required: true, message: t('connection.enterPort'), trigger: 'blur' },
  ],
  'db_config.username': [
    { required: true, message: t('connection.enterUsername'), trigger: 'blur' },
  ],
  'db_config.password': [
    { required: true, message: t('connection.enterPassword'), trigger: 'blur' },
  ],
  'ssh_config.host': [
    {
      validator: (_rule, value, callback) => {
        if (formData.value.connection_type === 'ssh_tunnel' && !value) {
          callback(new Error(t('connection.enterSshHost')))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  'ssh_config.username': [
    {
      validator: (_rule, value, callback) => {
        if (formData.value.connection_type === 'ssh_tunnel' && !value) {
          callback(new Error(t('connection.enterSshUsername')))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  'ssh_config.password': [
    {
      validator: (_rule, value, callback) => {
        if (
          formData.value.connection_type === 'ssh_tunnel' &&
          formData.value.ssh_config.auth_type === 'password' &&
          !value
        ) {
          callback(new Error(t('connection.enterSshPassword')))
        } else {
          callback()
        }
      },
      trigger: 'blur',
    },
  ],
  'ssh_config.private_key': [
    {
      validator: (_rule, value, callback) => {
        if (
          formData.value.connection_type === 'ssh_tunnel' &&
          formData.value.ssh_config.auth_type === 'private_key' &&
          !value
        ) {
          callback(new Error(t('connection.selectKeyFileReq')))
        } else {
          callback()
        }
      },
      trigger: 'change',
    },
  ],
}))

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
      error: error.message || t('connection.testFailed'),
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
      ElMessage.success(t('connection.connUpdated'))
    } else {
      await store.createConnection(formData.value)
      ElMessage.success(t('connection.connCreated'))
    }
    emit('saved')
    handleClose()
  } catch (error: any) {
    ElMessage.error(error.message || t('connection.saveFailed'))
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
        password: '',
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
