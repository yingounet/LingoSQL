<template>
  <div class="dashboard-page">
    <el-alert
      v-if="showSecurityNotice"
      class="security-notice"
      type="warning"
      show-icon
      closable
      @close="handleDismissNotice"
      :title="t('dashboard.securityNoticeTitle')"
      :description="t('dashboard.securityNoticeDesc')"
    />
    <!-- 页面标题 -->
    <PageHeader 
      :title="t('dashboard.title')" 
      :description="t('dashboard.description')"
    />
    
    <!-- 快捷操作 -->
    <div class="quick-actions">
      <h2 class="section-title">
        <el-icon :size="18"><Lightning /></el-icon>
        {{ t('dashboard.quickActions') }}
      </h2>
      <div class="actions-grid">
        <div class="action-card" @click="handleNewQuery">
          <div class="action-icon">
            <el-icon :size="24"><EditPen /></el-icon>
          </div>
          <div class="action-info">
            <span class="action-title">{{ t('dashboard.newQuery') }}</span>
            <span class="action-desc">{{ t('dashboard.newQueryDesc') }}</span>
          </div>
        </div>
        
        <div class="action-card" @click="handleCreateTable">
          <div class="action-icon">
            <el-icon :size="24"><Grid /></el-icon>
          </div>
          <div class="action-info">
            <span class="action-title">{{ t('dashboard.createTable') }}</span>
            <span class="action-desc">{{ t('dashboard.createTableDesc') }}</span>
          </div>
        </div>
        
        <div class="action-card" @click="handleImportCSV">
          <div class="action-icon">
            <el-icon :size="24"><Upload /></el-icon>
          </div>
          <div class="action-info">
            <span class="action-title">{{ t('dashboard.importCSV') }}</span>
            <span class="action-desc">{{ t('dashboard.importCSVDesc') }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 收藏的 SQL -->
    <div class="recent-favorites-section">
      <div class="recent-favorites-header">
        <h2 class="section-title">
          <el-icon :size="18"><Star /></el-icon>
          {{ t('dashboard.favoritedSQL') }}
        </h2>
        <span class="section-subtitle">{{ t('dashboard.recentlyUsed') }}</span>
        <router-link to="/favorites" class="view-all-link">{{ t('dashboard.viewAll') }}</router-link>
      </div>
      <div v-loading="loadingRecentFavorites" class="recent-favorites-list">
        <template v-if="recentFavorites.length > 0">
          <div
            v-for="fav in recentFavorites"
            :key="fav.id"
            class="favorite-chip"
            @click="handleUseFavorite(fav)"
            :title="fav.name"
          >
            <el-icon :size="14" class="chip-icon"><Document /></el-icon>
            <span class="chip-name">{{ fav.name }}</span>
          </div>
        </template>
        <div v-else-if="!loadingRecentFavorites" class="recent-favorites-empty">
          <el-icon :size="28"><Star /></el-icon>
          <span>{{ t('dashboard.noFavorites') }}</span>
          <span class="empty-hint">{{ t('dashboard.noFavoritesHint') }}</span>
        </div>
      </div>
    </div>
    
    <!-- 数据库连接 -->
    <div class="connections-section">
      <div class="connections-header">
        <h2 class="section-title">
          <el-icon :size="18"><ConnectionIcon /></el-icon>
          {{ t('dashboard.dbConnections') }}
        </h2>
        <el-button type="primary" @click="handleNewConnection">
          <el-icon><Plus /></el-icon>
          {{ t('connection.newConnection') }}
        </el-button>
      </div>
      <ConnectionList
        embedded
        @new="handleNewConnection"
        @edit="handleEditConnection"
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
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'
import { useConnectionStore } from '@/store/connection'
import PageHeader from '@/components/layout/PageHeader.vue'
import ConnectionList from '@/components/connection/ConnectionList.vue'
import ConnectionForm from '@/components/connection/ConnectionForm.vue'
import { getFavorites, recordFavoriteUse } from '@/api/favorite'
import type { Connection } from '@/types/connection'
import type { Favorite } from '@/types/favorite'
import { 
  Connection as ConnectionIcon, 
  Lightning, 
  EditPen, 
  Grid, 
  Upload,
  Plus,
  Star,
  Document
} from '@element-plus/icons-vue'

const { t } = useI18n()
const router = useRouter()
const connectionStore = useConnectionStore()

const noticeKey = 'lingosql.security.notice.dismissed'
const showSecurityNotice = ref(true)

// 连接表单对话框
const formDialogVisible = ref(false)
const editingConnectionId = ref<number | null>(null)

// 快捷操作
function handleNewQuery() {
  router.push('/query')
}

function handleCreateTable() {
  // TODO: 打开表设计器
  console.log('Create table')
}

function handleImportCSV() {
  // TODO: 打开 CSV 导入
  console.log('Import CSV')
}

// 连接管理
function handleNewConnection() {
  editingConnectionId.value = null
  formDialogVisible.value = true
}

function handleEditConnection(id: number) {
  editingConnectionId.value = id
  formDialogVisible.value = true
}

function handleConnect(connection: Connection) {
  router.push({
    path: '/database',
    query: { connection_id: String(connection.id) }
  })
}

function handleSaved() {
  formDialogVisible.value = false
}

// 最近使用的收藏
const loadingRecentFavorites = ref(false)
const recentFavorites = ref<Favorite[]>([])

function isSelectStatement(sql: string): boolean {
  let s = sql.trim()
  s = s.replace(/^--[^\n]*\n?/gm, '').trim()
  s = s.replace(/\/\*[\s\S]*?\*\//g, '').trim()
  return /^select\b/i.test(s)
}

async function loadRecentFavorites() {
  loadingRecentFavorites.value = true
  try {
    const list = await getFavorites({ sort: 'last_used_at' })
    recentFavorites.value = list.slice(0, 8)
  } catch {
    recentFavorites.value = []
  } finally {
    loadingRecentFavorites.value = false
  }
}

async function handleUseFavorite(fav: Favorite) {
  const needRestore =
    !connectionStore.currentConnection ||
    connectionStore.currentConnection.id !== fav.connection_id ||
    (fav.database && connectionStore.currentDatabase !== fav.database)
  if (needRestore) {
    try {
      const ok = await connectionStore.restoreState(fav.connection_id, fav.database || undefined)
      if (!ok) {
        ElMessage.error(t('dashboard.cannotSwitchConnection'))
        return
      }
    } catch (e: unknown) {
      ElMessage.error((e as Error).message || t('dashboard.switchFailed'))
      return
    }
  }
  try {
    await recordFavoriteUse(fav.id)
  } catch {
    // 忽略记录失败，仍跳转
  }
  router.push({
    path: '/query',
    query: {
      connection_id: String(fav.connection_id),
      ...(fav.database ? { database: fav.database } : {}),
    },
    state: {
      initialSql: fav.sql_query,
      autoExecute: isSelectStatement(fav.sql_query),
    },
  })
}

function handleDismissNotice() {
  showSecurityNotice.value = false
  localStorage.setItem(noticeKey, '1')
}

// 加载数据
onMounted(() => {
  showSecurityNotice.value = localStorage.getItem(noticeKey) !== '1'
  connectionStore.fetchConnections()
  loadRecentFavorites()
})
</script>

<style scoped>
.dashboard-page {
  max-width: 1400px;
  margin: 0 auto;
}

.security-notice {
  margin-bottom: var(--spacing-lg);
}

/* 快捷操作 */
.quick-actions {
  margin-bottom: var(--spacing-xl);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 var(--spacing-md);
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-md);
}

.action-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-lg);
  background-color: var(--color-background);
  border: 1px solid var(--color-border);
  border-radius: var(--border-radius-large);
  cursor: pointer;
  transition: all 0.2s;
}

.action-card:hover {
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
}

.action-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background-color: var(--color-background-secondary);
  border-radius: var(--border-radius-medium);
  color: var(--color-primary);
}

.action-info {
  display: flex;
  flex-direction: column;
}

.action-title {
  font-size: var(--font-size-body);
  font-weight: 600;
  color: var(--color-text-primary);
}

.action-desc {
  font-size: 12px;
  color: var(--color-text-tertiary);
}

/* 最近使用 */
.recent-favorites-section {
  margin-bottom: var(--spacing-xl);
  background-color: var(--color-background);
  border-radius: var(--border-radius-large);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-border);
}

.recent-favorites-header {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  margin-bottom: var(--spacing-md);
  flex-wrap: wrap;
}

.recent-favorites-header .section-title {
  margin: 0;
}

.section-subtitle {
  font-size: 12px;
  color: var(--color-text-tertiary);
}

.view-all-link {
  margin-left: auto;
  font-size: 13px;
  color: var(--color-primary);
  text-decoration: none;
}

.view-all-link:hover {
  text-decoration: underline;
}

.recent-favorites-list {
  min-height: 60px;
  display: flex;
  flex-wrap: wrap;
  gap: var(--spacing-sm);
}

.favorite-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background-color: var(--color-background-secondary);
  border: 1px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  max-width: 200px;
}

.favorite-chip:hover {
  background-color: var(--color-nav-active-bg);
  border-color: var(--color-primary);
}

.chip-icon {
  color: var(--color-text-tertiary);
  flex-shrink: 0;
}

.favorite-chip:hover .chip-icon {
  color: var(--color-primary);
}

.chip-name {
  font-size: 13px;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.favorite-chip:hover .chip-name {
  color: var(--color-primary);
}

.recent-favorites-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-lg);
  color: var(--color-text-tertiary);
  font-size: 13px;
  gap: 4px;
  width: 100%;
}

.recent-favorites-empty .el-icon {
  color: var(--color-text-tertiary);
  opacity: 0.5;
}

.empty-hint {
  font-size: 12px;
  color: var(--color-text-tertiary);
  opacity: 0.7;
}

/* 数据库连接区域 */
.connections-section {
  background-color: var(--color-background);
  border-radius: var(--border-radius-large);
  padding: var(--spacing-lg);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--color-border);
  border-left: 3px solid var(--color-primary);
}

.connections-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--spacing-md);
}

.connections-header .section-title {
  margin: 0;
}

/* 响应式 */
@media (max-width: 768px) {
  .actions-grid {
    grid-template-columns: 1fr;
  }
}
</style>
