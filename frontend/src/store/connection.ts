import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  Connection,
  ConnectionFormData,
  ConnectionFilter,
  Pagination,
  TestResult,
  DbType,
} from '@/types/connection'
import * as connectionApi from '@/api/connection'
import * as databaseApi from '@/api/database'
import * as tableApi from '@/api/table'
import type { TableInfo } from '@/api/table'

export const useConnectionStore = defineStore('connection', () => {
  // 状态
  const connections = ref<Connection[]>([])
  const currentConnection = ref<Connection | null>(null)
  const databases = ref<string[]>([])
  const currentDatabase = ref<string | null>(null)
  const loadingDatabases = ref(false)
  const tables = ref<TableInfo[]>([])
  const loadingTables = ref(false)
  const loading = ref(false)
  const filter = ref<ConnectionFilter>({
    db_type: null,
    search: '',
  })
  const pagination = ref<Pagination>({
    page: 1,
    pageSize: 10,
    total: 0,
  })

  // 计算属性
  const filteredConnections = computed(() => {
    let result = connections.value

    if (filter.value.db_type) {
      result = result.filter(c => c.db_type === filter.value.db_type)
    }

    if (filter.value.search) {
      const keyword = filter.value.search.toLowerCase()
      result = result.filter(c =>
        c.name.toLowerCase().includes(keyword) ||
        c.db_config.host.toLowerCase().includes(keyword)
      )
    }

    return result
  })

  const totalPages = computed(() => 
    Math.ceil(pagination.value.total / pagination.value.pageSize)
  )

  // Actions
  async function fetchConnections(params?: {
    page?: number
    pageSize?: number
    dbType?: DbType | null
    search?: string
  }) {
    loading.value = true
    try {
      const response = await connectionApi.getConnections({
        page: params?.page || pagination.value.page,
        page_size: params?.pageSize || pagination.value.pageSize,
        db_type: params?.dbType !== undefined ? params.dbType : filter.value.db_type,
        search: params?.search !== undefined ? params.search : filter.value.search,
      })

      connections.value = response.list
      pagination.value.total = response.total
      pagination.value.page = response.page
      pagination.value.pageSize = response.page_size

      // 同步筛选条件
      if (params?.dbType !== undefined) {
        filter.value.db_type = params.dbType
      }
      if (params?.search !== undefined) {
        filter.value.search = params.search
      }
    } finally {
      loading.value = false
    }
  }

  async function getConnection(id: number): Promise<Connection | null> {
    return await connectionApi.getConnection(id)
  }

  async function createConnection(data: ConnectionFormData): Promise<Connection> {
    const connection = await connectionApi.createConnection(data)
    await fetchConnections() // 刷新列表
    return connection
  }

  async function updateConnection(id: number, data: ConnectionFormData): Promise<Connection> {
    const connection = await connectionApi.updateConnection(id, data)
    await fetchConnections() // 刷新列表
    return connection
  }

  async function deleteConnection(id: number): Promise<void> {
    await connectionApi.deleteConnection(id)
    await fetchConnections() // 刷新列表
  }

  async function testConnection(id: number): Promise<TestResult> {
    return await connectionApi.testConnection(id)
  }

  async function testConnectionConfig(data: {
    db_type: DbType
    connection_type: 'direct' | 'ssh_tunnel'
    db_config: ConnectionFormData['db_config']
    ssh_config?: ConnectionFormData['ssh_config']
  }): Promise<TestResult> {
    return await connectionApi.testConnectionConfig(data)
  }

  async function setDefaultConnection(id: number): Promise<void> {
    await connectionApi.setDefaultConnection(id)
    await fetchConnections() // 刷新列表
  }

  async function connectTo(id: number): Promise<void> {
    // 更新最后使用时间
    await connectionApi.updateLastUsed(id)

    const connection = connections.value.find(c => c.id === id)
    if (!connection) {
      throw new Error('连接不存在')
    }

    // 先设置当前连接，便于后续数据库/表加载
    currentConnection.value = connection
    connection.last_used_at = new Date().toISOString()

    try {
      // 获取数据库列表，失败则回滚连接状态
      await fetchDatabases(id, true)
    } catch (error) {
      currentConnection.value = null
      currentDatabase.value = null
      databases.value = []
      tables.value = []
      throw error
    }
  }

  // 获取数据库列表
  async function fetchDatabases(connectionId: number, throwOnError = false): Promise<void> {
    loadingDatabases.value = true
    try {
      databases.value = await databaseApi.getDatabases(connectionId)
      // 如果有数据库列表，默认选择第一个并加载表
      if (databases.value.length > 0 && !currentDatabase.value) {
        await setCurrentDatabase(databases.value[0])
      }
    } catch (error) {
      console.error('获取数据库列表失败:', error)
      databases.value = []
      if (throwOnError) {
        throw error
      }
    } finally {
      loadingDatabases.value = false
    }
  }

  // 设置当前数据库
  async function setCurrentDatabase(database: string | null): Promise<void> {
    currentDatabase.value = database
    // 切换数据库时自动获取表列表
    if (database && currentConnection.value) {
      await fetchTables()
    } else {
      tables.value = []
    }
  }

  // 获取表列表
  async function fetchTables(): Promise<void> {
    if (!currentConnection.value || !currentDatabase.value) {
      tables.value = []
      return
    }

    loadingTables.value = true
    try {
      const list = await tableApi.getTables(currentConnection.value.id, currentDatabase.value)
      tables.value = Array.isArray(list) ? list : []
    } catch (error) {
      console.error('获取表列表失败，请检查网络或数据库权限:', error)
      tables.value = []
    } finally {
      loadingTables.value = false
    }
  }

  // 清除数据库选择（断开连接时调用）
  function clearDatabaseSelection(): void {
    databases.value = []
    currentDatabase.value = null
    tables.value = []
  }

  // 断开连接
  function disconnect(): void {
    currentConnection.value = null
    clearDatabaseSelection()
  }

  // 根据 ID 恢复连接状态（用于 URL 状态恢复）
  async function restoreState(connectionId: number, database?: string | null): Promise<boolean> {
    try {
      // 1. 获取连接详情
      const connection = await connectionApi.getConnection(connectionId)
      if (!connection) {
        console.error('连接不存在:', connectionId)
        return false
      }

      // 2. 设置当前连接
      currentConnection.value = connection

      // 3. 获取数据库列表
      await fetchDatabases(connectionId)

      // 4. 如果有 database 参数且在数据库列表中，设置 currentDatabase
      if (database && databases.value.includes(database)) {
        await setCurrentDatabase(database)
      } else if (databases.value.length > 0 && !currentDatabase.value) {
        // 如果没有指定数据库但有数据库列表，默认选择第一个
        currentDatabase.value = databases.value[0]
      }

      return true
    } catch (error) {
      console.error('恢复连接状态失败:', error)
      return false
    }
  }

  function setFilter(newFilter: Partial<ConnectionFilter>) {
    filter.value = { ...filter.value, ...newFilter }
    pagination.value.page = 1 // 重置到第一页
    fetchConnections()
  }

  function setPage(page: number) {
    pagination.value.page = page
    fetchConnections()
  }

  function resetFilter() {
    filter.value = {
      db_type: null,
      search: '',
    }
    pagination.value.page = 1
    fetchConnections()
  }

  return {
    // 状态
    connections,
    currentConnection,
    databases,
    currentDatabase,
    tables,
    loading,
    loadingDatabases,
    loadingTables,
    filter,
    pagination,
    // 计算属性
    filteredConnections,
    totalPages,
    // Actions
    fetchConnections,
    getConnection,
    createConnection,
    updateConnection,
    deleteConnection,
    testConnection,
    testConnectionConfig,
    setDefaultConnection,
    connectTo,
    fetchDatabases,
    setCurrentDatabase,
    fetchTables,
    clearDatabaseSelection,
    disconnect,
    restoreState,
    setFilter,
    setPage,
    resetFilter,
  }
})
