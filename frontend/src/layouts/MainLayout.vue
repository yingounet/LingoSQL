<template>
  <div class="main-layout" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
    <!-- 顶部导航栏 -->
    <HeaderBar 
      :collapsed="sidebarCollapsed"
      @toggle-sidebar="toggleSidebar"
    />
    
    <!-- 主体区域 -->
    <div class="layout-body">
      <!-- 左侧边栏 -->
      <Sidebar 
        ref="sidebarRef"
        :collapsed="sidebarCollapsed"
        @collapse="toggleSidebar"
      />
      
      <!-- 主内容区 -->
      <main class="main-content">
        <router-view v-slot="{ Component, route }">
          <keep-alive>
            <component 
              v-if="route.meta.keepAlive" 
              :is="Component" 
              :key="route.path" 
            />
          </keep-alive>
          <component 
            v-if="!route.meta.keepAlive" 
            :is="Component" 
            :key="route.path" 
          />
        </router-view>
      </main>
    </div>
    
    <!-- 底部状态栏 -->
    <StatusBar />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import HeaderBar from '@/components/layout/HeaderBar.vue'
import Sidebar from '@/components/layout/Sidebar.vue'
import StatusBar from '@/components/layout/StatusBar.vue'
import { useUrlState } from '@/composables/useUrlState'

// 侧边栏折叠状态
const sidebarCollapsed = ref(false)

// Sidebar 组件引用
const sidebarRef = ref<InstanceType<typeof Sidebar> | null>(null)

// URL 状态管理
const { restoreFromUrl } = useUrlState()

// 切换侧边栏
function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

// 响应式处理
function handleResize() {
  if (window.innerWidth < 768) {
    sidebarCollapsed.value = true
  }
}

// 从 URL 恢复状态
async function restoreStateFromUrl() {
  try {
    const tableName = await restoreFromUrl()
    // 如果有表名需要恢复，等待 DOM 更新后设置选中的表
    if (tableName && sidebarRef.value) {
      await nextTick()
      // 等待表列表加载完成
      setTimeout(() => {
        sidebarRef.value?.setSelectedTable(tableName)
      }, 100)
    }
  } catch (error) {
    console.error('恢复 URL 状态失败:', error)
  }
}

onMounted(() => {
  handleResize()
  window.addEventListener('resize', handleResize)
  // 恢复 URL 状态
  restoreStateFromUrl()
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
/* 布局尺寸变量 */
.main-layout {
  --header-height: 56px;
  --sidebar-width: 240px;
  --sidebar-collapsed-width: 64px;
  --status-bar-height: 32px;
  
  display: flex;
  flex-direction: column;
  height: 100vh;
  background-color: var(--color-background-secondary);
}

.layout-body {
  display: flex;
  flex: 1;
  overflow: hidden;
  margin-top: var(--header-height);
}

.main-content {
  flex: 1;
  overflow-y: auto;
  margin-left: var(--sidebar-width);
  margin-bottom: var(--status-bar-height);
  padding: var(--spacing-lg);
  transition: margin-left 0.3s ease;
}

.sidebar-collapsed .main-content {
  margin-left: var(--sidebar-collapsed-width);
}

/* 响应式 */
@media (max-width: 768px) {
  .main-content {
    margin-left: 0;
    padding: var(--spacing-md);
  }
  
  .sidebar-collapsed .main-content {
    margin-left: 0;
  }
}
</style>
