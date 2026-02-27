<template>
  <footer class="status-bar">
    <!-- 左侧：连接状态 -->
    <div class="status-left">
      <span class="connection-status" :class="statusClass">
        <span class="status-dot"></span>
        {{ connectionLabel }}
      </span>
      <span class="separator">|</span>
      <span class="db-info" v-if="dbInfo">{{ dbInfo }}</span>
    </div>
    
    <!-- 中间：执行信息 -->
    <div class="status-center">
      <span v-if="executionTime" class="execution-time">
        <el-icon :size="12"><Clock /></el-icon>
        {{ executionTime }}
      </span>
      <span v-if="rowCount !== null" class="row-count">
        <el-icon :size="12"><Document /></el-icon>
        {{ rowCount }} rows
      </span>
    </div>
    
    <!-- 右侧：编辑器信息 -->
    <div class="status-right">
      <span class="encoding">UTF-8</span>
      <span class="cursor-position">LN {{ line }}, COL {{ column }}</span>
      <span class="version">VER {{ appVersion }}</span>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useConnectionStore } from '@/store/connection'
import { Clock, Document } from '@element-plus/icons-vue'

const connectionStore = useConnectionStore()

// 应用版本
const appVersion = ref('1.0.0')

// 模拟数据（实际应从 store 获取）
const executionTime = ref<string | null>('0.14s')
const rowCount = ref<number | null>(42)
const line = ref(1)
const column = ref(1)

// 连接状态
const statusClass = computed(() => ({
  connected: !!connectionStore.currentConnection,
  disconnected: !connectionStore.currentConnection
}))

const connectionLabel = computed(() => {
  if (!connectionStore.currentConnection) return 'NO CONNECTION'
  return `${connectionStore.currentConnection.name.toUpperCase()} CONNECTED`
})

// 数据库信息
const dbInfo = computed(() => {
  const conn = connectionStore.currentConnection
  if (!conn) return ''
  return `${conn.db_type?.toUpperCase() || 'SQL'}`
})
</script>

<style scoped>
.status-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 32px;
  background-color: var(--color-background);
  border-top: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-md);
  font-size: 12px;
  z-index: 1000;
}

.status-left,
.status-center,
.status-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

/* 连接状态 */
.connection-status {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-weight: 500;
}

.connection-status .status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.connection-status.connected .status-dot {
  background-color: var(--color-success);
}

.connection-status.connected {
  color: var(--color-success);
}

.connection-status.disconnected .status-dot {
  background-color: var(--color-text-tertiary);
}

.connection-status.disconnected {
  color: var(--color-text-tertiary);
}

.separator {
  color: var(--color-border);
}

.db-info {
  color: var(--color-text-secondary);
}

/* 执行信息 */
.execution-time,
.row-count {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--color-text-secondary);
}

/* 右侧信息 */
.encoding,
.cursor-position,
.version {
  color: var(--color-text-tertiary);
}

.version {
  color: var(--color-primary);
  font-weight: 500;
}

/* 响应式 */
@media (max-width: 768px) {
  .status-center {
    display: none;
  }
  
  .cursor-position {
    display: none;
  }
}
</style>
