<template>
  <el-dialog
    v-model="visible"
    title="创建数据库"
    width="500px"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
    >
      <el-form-item label="数据库名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入数据库名称" />
      </el-form-item>
      
      <!-- MySQL选项 -->
      <template v-if="dbType === 'mysql'">
        <el-form-item label="字符集" prop="charset">
          <el-select v-model="form.charset" placeholder="请选择字符集">
            <el-option label="utf8mb4" value="utf8mb4" />
            <el-option label="utf8" value="utf8" />
            <el-option label="latin1" value="latin1" />
          </el-select>
        </el-form-item>
        <el-form-item label="排序规则" prop="collation">
          <el-select v-model="form.collation" placeholder="请选择排序规则">
            <el-option label="utf8mb4_unicode_ci" value="utf8mb4_unicode_ci" />
            <el-option label="utf8mb4_general_ci" value="utf8mb4_general_ci" />
            <el-option label="utf8_unicode_ci" value="utf8_unicode_ci" />
          </el-select>
        </el-form-item>
      </template>
      
      <!-- PostgreSQL选项 -->
      <template v-if="dbType === 'postgresql'">
        <el-form-item label="编码" prop="encoding">
          <el-select v-model="form.encoding" placeholder="请选择编码">
            <el-option label="UTF8" value="UTF8" />
            <el-option label="LATIN1" value="LATIN1" />
          </el-select>
        </el-form-item>
        <el-form-item label="LC_COLLATE" prop="lc_collate">
          <el-input v-model="form.lc_collate" placeholder="例如: en_US.UTF-8" />
        </el-form-item>
        <el-form-item label="LC_CTYPE" prop="lc_ctype">
          <el-input v-model="form.lc_ctype" placeholder="例如: en_US.UTF-8" />
        </el-form-item>
      </template>
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
import { ref, reactive, computed } from 'vue'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { createDatabase } from '@/api/databaseAdmin'
import { useConnectionStore } from '@/store/connection'
import type { CreateDatabaseRequest } from '@/types/databaseAdmin'

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

const form = reactive<CreateDatabaseRequest>({
  name: '',
  charset: 'utf8mb4',
  collation: 'utf8mb4_unicode_ci',
  encoding: 'UTF8',
  lc_collate: 'en_US.UTF-8',
  lc_ctype: 'en_US.UTF-8'
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入数据库名称', trigger: 'blur' },
    { pattern: /^[a-zA-Z_][a-zA-Z0-9_]*$/, message: '数据库名称只能包含字母、数字和下划线，且不能以数字开头', trigger: 'blur' }
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
    name: '',
    charset: 'utf8mb4',
    collation: 'utf8mb4_unicode_ci',
    encoding: 'UTF8',
    lc_collate: 'en_US.UTF-8',
    lc_ctype: 'en_US.UTF-8'
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
      await createDatabase(connectionStore.currentConnection.id, form)
      ElMessage.success('数据库创建成功')
      emit('success')
      handleClose()
    } catch (error: any) {
      ElMessage.error(error.message || '创建数据库失败')
    } finally {
      loading.value = false
    }
  })
}
</script>
