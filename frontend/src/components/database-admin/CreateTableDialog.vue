<template>
  <el-dialog
    v-model="visible"
    title="创建表"
    width="560px"
    :close-on-click-modal="false"
    @closed="onClosed"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="表名" prop="table_name">
        <el-input v-model="form.table_name" placeholder="请输入表名" />
      </el-form-item>
      <el-form-item label="建表 DDL" prop="create_ddl">
        <el-input
          v-model="form.create_ddl"
          type="textarea"
          :rows="6"
          :placeholder="dbType === 'postgresql' ? '可选。留空将创建仅含 id SERIAL PRIMARY KEY 的默认表；填写时请用 PostgreSQL 语法（双引号标识符、SERIAL 等）' : '可选。留空将创建仅含 id 主键的默认表；也可填写完整 CREATE TABLE 语句'"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        创建
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed as vueComputed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createTable } from '@/api/tableAdmin'
import { useConnectionStore } from '@/store/connection'

const connectionStore = useConnectionStore()
const dbType = vueComputed(() => connectionStore.currentConnection?.db_type ?? 'mysql')

const props = defineProps<{
  modelValue: boolean
  database: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'success': []
}>()

const visible = vueComputed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const formRef = ref<FormInstance>()
const loading = ref(false)
const form = ref({
  table_name: '',
  create_ddl: ''
})

const rules: FormRules = {
  table_name: [
    { required: true, message: '请输入表名', trigger: 'blur' },
    { min: 1, max: 64, message: '表名长度 1-64', trigger: 'blur' }
  ]
}

watch(visible, (v) => {
  if (v) {
    form.value = { table_name: '', create_ddl: '' }
    formRef.value?.clearValidate()
  }
})

function onClosed() {
  formRef.value?.resetFields()
}

async function handleSubmit() {
  if (!formRef.value || !props.database) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    const connectionId = connectionStore.currentConnection?.id
    if (!connectionId) {
      ElMessage.error('请先选择连接')
      return
    }
    loading.value = true
    try {
      await createTable(connectionId, {
        database: props.database,
        table_name: form.value.table_name.trim(),
        create_ddl: form.value.create_ddl?.trim() || undefined
      })
      visible.value = false
      emit('success')
    } catch (error: any) {
      ElMessage.error(error.message || '创建表失败')
    } finally {
      loading.value = false
    }
  })
}
</script>
