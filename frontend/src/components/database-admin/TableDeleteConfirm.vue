<template>
  <el-dialog
    v-model="visible"
    title="删除表"
    width="500px"
  >
    <div class="delete-confirm" v-if="tableName">
      <el-alert type="warning" :closable="false" show-icon>
        <template #title>
          <div class="alert-content">
            <p>您确定要删除表 <strong>{{ tableName }}</strong>（数据库：{{ database }}）吗？</p>
            <p class="warning-text">此操作不可恢复，请谨慎操作！</p>
          </div>
        </template>
      </el-alert>
      <div class="confirm-input" style="margin-top: 20px">
        <el-input
          v-model="confirmText"
          placeholder="请输入表名以确认删除"
        />
      </div>
    </div>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button
        type="danger"
        :disabled="confirmText !== tableName"
        :loading="loading"
        @click="handleConfirm"
      >
        确认删除
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

const props = defineProps<{
  modelValue: boolean
  database: string
  tableName: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'confirm': []
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const confirmText = ref('')
const loading = ref(false)

watch(visible, (v) => {
  if (v) confirmText.value = ''
})

function handleConfirm() {
  emit('confirm')
  visible.value = false
}
</script>

<style scoped>
.delete-confirm {
  padding: var(--spacing-sm);
}

.alert-content {
  line-height: 1.6;
}

.warning-text {
  color: var(--el-color-warning);
  font-size: 14px;
  margin-top: 8px;
}

.confirm-input {
  margin-top: var(--spacing-md);
}
</style>
