<template>
  <header class="header-bar">
    <!-- 左侧区域 -->
    <div class="header-left">
      <button class="menu-toggle" @click="$emit('toggle-sidebar')">
        <el-icon :size="20"><Menu /></el-icon>
      </button>
      <router-link to="/" class="logo">
        <span class="logo-text">LingoSQL</span>
      </router-link>
      <div class="search-box">
        <el-icon :size="16"><Search /></el-icon>
        <input 
          type="text" 
          :placeholder="t('header.searchPlaceholder')" 
          v-model="searchKeyword"
          @keyup.enter="handleSearch"
        />
      </div>
    </div>
    
    <!-- 中间导航 -->
    <nav class="header-nav">
      <el-tooltip
        v-for="item in navItems"
        :key="item.path"
        :content="getDisabledReason(item)"
        :disabled="!isNavDisabled(item)"
        placement="bottom"
      >
        <span>
          <router-link
            v-if="!isNavDisabled(item)"
            :to="getNavRoute(item.path)"
            class="nav-item"
            :class="{ active: isActive(item.path) }"
          >
            {{ item.label }}
          </router-link>
          <span v-else class="nav-item disabled">
            {{ item.label }}
          </span>
        </span>
      </el-tooltip>
    </nav>
    
    <!-- 右侧区域 -->
    <div class="header-right">
      <el-button circle @click="toggleTheme" :title="theme === 'light' ? t('header.switchToDark') : t('header.switchToLight')">
        <el-icon><Sunny v-if="theme === 'dark'" /><Moon v-else /></el-icon>
      </el-button>
      <button class="icon-btn" :title="t('header.notifications')">
        <el-icon :size="20"><Bell /></el-icon>
      </button>
      <button class="icon-btn" :title="t('header.settings')" @click="$router.push('/settings')">
        <el-icon :size="20"><Setting /></el-icon>
      </button>
      <el-dropdown @command="handleUserCommand">
        <div class="user-info">
          <el-avatar :size="32">
            {{ userInitial }}
          </el-avatar>
          <span class="username">{{ authStore.user?.username }}</span>
          <el-icon :size="12"><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item command="profile">
              <el-icon><User /></el-icon>
              {{ t('header.profile') }}
            </el-dropdown-item>
            <el-dropdown-item command="logout" divided>
              <el-icon><SwitchButton /></el-icon>
              {{ t('header.logout') }}
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { useConnectionStore } from '@/store/connection'
import { useTheme } from '@/composables/useTheme'
import { 
  Menu, 
  Search, 
  Bell, 
  Setting, 
  ArrowDown, 
  User, 
  SwitchButton,
  Sunny,
  Moon
} from '@element-plus/icons-vue'

defineProps<{
  collapsed: boolean
}>()

defineEmits<{
  'toggle-sidebar': []
}>()

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const connectionStore = useConnectionStore()
const { theme, toggleTheme } = useTheme()

const searchKeyword = ref('')

// 导航菜单配置（使用 computed 以支持 i18n 切换）
const { t } = useI18n()
const navItems = computed(() => [
  { path: '/', label: t('nav.dashboard') },
  { path: '/database', label: t('nav.database'), requiresConnection: true },
  { path: '/schema', label: t('nav.schema'), requiresConnection: true, requiresDatabase: true },
  { path: '/rowdata', label: t('nav.data'), requiresConnection: true, requiresDatabase: true },
  { path: '/history', label: t('nav.history') },
  { path: '/favorites', label: t('nav.favorites') },
  { path: '/query', label: t('nav.query'), requiresConnection: true }
])
const hasConnection = computed(() => !!connectionStore.currentConnection)
const hasDatabase = computed(() => !!connectionStore.currentDatabase)

function isNavDisabled(item: { requiresConnection?: boolean; requiresDatabase?: boolean }) {
  if (item.requiresConnection && !hasConnection.value) return true
  if (item.requiresDatabase && !hasDatabase.value) return true
  return false
}

function getDisabledReason(item: { requiresConnection?: boolean; requiresDatabase?: boolean }) {
  if (item.requiresDatabase && !hasDatabase.value) {
    return t('header.selectDatabaseFirst')
  }
  if (item.requiresConnection && !hasConnection.value) {
    return t('header.establishConnectionFirst')
  }
  return ''
}

// 用户名首字母
const userInitial = computed(() => {
  return authStore.user?.username?.charAt(0).toUpperCase() || 'U'
})

// 判断导航是否激活
function isActive(path: string): boolean {
  // 检查是否有 activeNav meta 配置
  const activeNav = route.meta.activeNav as string | undefined
  if (activeNav) {
    return path === activeNav
  }
  
  if (path === '/') {
    return route.path === '/'
  }
  return route.path.startsWith(path)
}

// 生成导航链接的 query 参数
function getNavQuery(path: string): Record<string, string> {
  const query: Record<string, string> = {}
  const currentQuery = route.query
  
  // 对于 query、schema、rowdata、database 页面，保留连接与库表参数
  if (['/query', '/schema', '/rowdata', '/database'].includes(path)) {
    if (currentQuery.connection_id) {
      query.connection_id = String(currentQuery.connection_id)
    }
    if (currentQuery.database) {
      query.database = String(currentQuery.database)
    }
    if (path !== '/database' && currentQuery.table) {
      query.table = String(currentQuery.table)
    }
  } else {
    // 其他页面只保留 connection_id
    if (currentQuery.connection_id) {
      query.connection_id = String(currentQuery.connection_id)
    }
  }
  
  return query
}

// 生成导航路由对象
function getNavRoute(path: string) {
  const query = getNavQuery(path)
  return {
    path,
    query: Object.keys(query).length > 0 ? query : undefined
  }
}

// 搜索处理
function handleSearch() {
  if (searchKeyword.value.trim()) {
    console.log('Search:', searchKeyword.value)
    // TODO: 实现搜索功能
  }
}

// 用户菜单命令
function handleUserCommand(command: string) {
  if (command === 'logout') {
    authStore.logout()
    router.push('/login')
  } else if (command === 'profile') {
    router.push('/settings')
  }
}
</script>

<style scoped>
.header-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background-color: var(--color-background);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--spacing-md);
  z-index: 1000;
}

/* 左侧区域 */
.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.menu-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  border-radius: var(--border-radius-small);
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: all 0.2s;
}

.menu-toggle:hover {
  background-color: var(--color-background-secondary);
  color: var(--color-text-primary);
}

.logo {
  display: flex;
  align-items: center;
  text-decoration: none;
}

.logo-text {
  font-size: 18px;
  font-weight: 700;
  color: var(--color-primary);
}

.search-box {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  background-color: var(--color-background-secondary);
  border-radius: var(--border-radius-small);
  min-width: 280px;
}

.search-box input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  outline: none;
}

.search-box input::placeholder {
  color: var(--color-text-tertiary);
}

.search-box .el-icon {
  color: var(--color-text-tertiary);
}

/* 中间导航 */
.header-nav {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.nav-item {
  padding: var(--spacing-sm) var(--spacing-md);
  font-size: var(--font-size-body);
  color: var(--color-text-secondary);
  text-decoration: none;
  border-radius: var(--border-radius-small);
  transition: all 0.2s;
}

.nav-item:hover {
  color: var(--color-text-primary);
  background-color: var(--color-background-secondary);
}

.nav-item.active {
  color: var(--color-primary);
  font-weight: 600;
  background-color: var(--color-nav-active-bg);
}

.nav-item.disabled {
  color: var(--color-text-tertiary);
  cursor: not-allowed;
  background-color: transparent;
}

/* 右侧区域 */
.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.icon-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: none;
  background: transparent;
  border-radius: var(--border-radius-small);
  cursor: pointer;
  color: var(--color-text-secondary);
  transition: all 0.2s;
}

.icon-btn:hover {
  background-color: var(--color-background-secondary);
  color: var(--color-text-primary);
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: 4px 8px;
  border-radius: var(--border-radius-small);
  cursor: pointer;
  transition: all 0.2s;
}

.user-info:hover {
  background-color: var(--color-background-secondary);
}

.username {
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
  font-weight: 500;
}

/* 响应式 */
@media (max-width: 1024px) {
  .search-box {
    min-width: 200px;
  }
}

@media (max-width: 768px) {
  .search-box {
    display: none;
  }
  
  .header-nav {
    display: none;
  }
  
  .username {
    display: none;
  }
}
</style>
