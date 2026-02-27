/**
 * URL 状态管理 composable
 * 负责 URL query 参数与连接状态的双向同步
 */

import { useRoute, useRouter } from 'vue-router'
import { useConnectionStore } from '@/store/connection'

export interface UrlStateParams {
  connectionId: number | null
  database: string | null
  table: string | null
}

export function useUrlState() {
  const route = useRoute()
  const router = useRouter()
  const connectionStore = useConnectionStore()

  /**
   * 从 URL 读取状态参数
   */
  function getUrlParams(): UrlStateParams {
    return {
      connectionId: route.query.connection_id ? Number(route.query.connection_id) : null,
      database: (route.query.database as string) || null,
      table: (route.query.table as string) || null,
    }
  }

  /**
   * 更新 URL 参数
   * 使用 router.replace 避免产生额外的历史记录
   */
  function updateUrlParams(params: Partial<UrlStateParams>) {
    const query: Record<string, string | undefined> = { ...route.query as Record<string, string | undefined> }

    // 更新 connectionId
    if (params.connectionId !== undefined) {
      if (params.connectionId !== null) {
        query.connection_id = String(params.connectionId)
      } else {
        delete query.connection_id
      }
      // 切换连接时清空数据库和表
      delete query.database
      delete query.table
    }

    // 更新 database
    if (params.database !== undefined) {
      if (params.database !== null) {
        query.database = params.database
      } else {
        delete query.database
      }
      // 切换数据库时清空表
      if (params.table === undefined) {
        delete query.table
      }
    }

    // 更新 table
    if (params.table !== undefined) {
      if (params.table !== null) {
        query.table = params.table
      } else {
        delete query.table
      }
    }

    router.replace({ query })
  }

  /**
   * 根据 URL 参数恢复状态
   * 返回恢复的表名（如果有）
   */
  async function restoreFromUrl(): Promise<string | null> {
    const { connectionId, database, table } = getUrlParams()

    // 没有连接 ID，不需要恢复
    if (!connectionId) {
      return null
    }

    // 检查是否已经是当前连接
    if (connectionStore.currentConnection?.id === connectionId) {
      // 已经是当前连接，只需检查数据库和表
      if (database && connectionStore.currentDatabase !== database) {
        await connectionStore.setCurrentDatabase(database)
      }
      return table
    }

    try {
      // 恢复连接状态
      const success = await connectionStore.restoreState(connectionId, database)
      if (success) {
        return table
      }
    } catch (error) {
      console.error('恢复连接状态失败:', error)
      // 恢复失败，清除 URL 参数
      updateUrlParams({ connectionId: null })
    }

    return null
  }

  /**
   * 清除所有 URL 状态参数
   */
  function clearUrlParams() {
    const query = { ...route.query as Record<string, string | undefined> }
    delete query.connection_id
    delete query.database
    delete query.table
    router.replace({ query })
  }

  return {
    getUrlParams,
    updateUrlParams,
    restoreFromUrl,
    clearUrlParams,
  }
}
