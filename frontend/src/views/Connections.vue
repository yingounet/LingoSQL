<template>
  <div class="connections-page">
    <!-- 页面标题 -->
    <PageHeader 
      title="Database Connections" 
      description="Manage and connect to your saved database instances"
    >
      <template #actions>
        <el-button type="primary" @click="handleNew">
          <el-icon><Plus /></el-icon>
          New Connection
        </el-button>
      </template>
    </PageHeader>
    
    <!-- 连接列表组件 -->
    <div class="connections-content">
      <ConnectionList
        @new="handleNew"
        @edit="handleEdit"
        @connect="handleConnect"
      />
    </div>

    <!-- 连接表单对话框 -->
    <ConnectionForm
      v-model="formDialogVisible"
      :connection-id="editingConnectionId"
      @saved="handleSaved"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import PageHeader from '@/components/layout/PageHeader.vue'
import ConnectionList from '@/components/connection/ConnectionList.vue'
import ConnectionForm from '@/components/connection/ConnectionForm.vue'
import { useUrlState } from '@/composables/useUrlState'
import type { Connection } from '@/types/connection'
import { Plus } from '@element-plus/icons-vue'

const router = useRouter()
const { updateUrlParams } = useUrlState()

// 表单对话框状态
const formDialogVisible = ref(false)
const editingConnectionId = ref<number | null>(null)

// 新建连接
function handleNew() {
  editingConnectionId.value = null
  formDialogVisible.value = true
}

// 编辑连接
function handleEdit(id: number) {
  editingConnectionId.value = id
  formDialogVisible.value = true
}

// 连接到数据库
function handleConnect(connection: Connection) {
  // 连接成功后跳转到数据库页面，并在 URL 中保存连接 ID
  console.log('Connected to:', connection.name)
  router.push({
    path: '/database',
    query: { connection_id: String(connection.id) }
  })
}

// 保存完成
function handleSaved() {
  console.log('Connection saved')
}
</script>

<style scoped>
.connections-page {
  max-width: 1400px;
  margin: 0 auto;
}

.connections-content {
  background-color: var(--color-background);
  border-radius: var(--border-radius-large);
  box-shadow: var(--shadow-sm);
}
</style>
