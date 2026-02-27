// 数据库类型
export type DbType = 'mysql' | 'postgresql' | 'mariadb'

// 连接方式
export type ConnectionType = 'direct' | 'ssh_tunnel'

// SSH 认证方式
export type SshAuthType = 'password' | 'private_key'

// SSL 模式
export type SslMode = 'disable' | 'require' | 'verify-ca' | 'verify-full'

// 数据库连接配置
export interface DbConfig {
  host: string
  port: number
  database?: string
  username: string
  password?: string // 编辑时可能不返回密码
  options?: {
    ssl_mode?: SslMode
    charset?: string
    timeout?: number
  }
}

// SSH 隧道配置
export interface SshConfig {
  host: string
  port: number
  username: string
  auth_type: SshAuthType
  password?: string
  private_key?: string
  passphrase?: string
}

// 数据库连接
export interface Connection {
  id: number
  name: string
  db_type: DbType
  connection_type: ConnectionType
  db_config: DbConfig
  ssh_config?: SshConfig
  is_default: boolean
  last_used_at: string | null
  created_at: string
  updated_at?: string
}

// 连接表单数据（用于创建/编辑）
export interface ConnectionFormData {
  name: string
  db_type: DbType
  connection_type: ConnectionType
  db_config: {
    host: string
    port: number
    database: string
    username: string
    password: string
    options: {
      ssl_mode: SslMode
      charset: string
      timeout: number
    }
  }
  ssh_config: {
    host: string
    port: number
    username: string
    auth_type: SshAuthType
    password: string
    private_key: string
    passphrase: string
  }
}

// 连接筛选条件
export interface ConnectionFilter {
  db_type: DbType | null
  search: string
}

// 分页信息
export interface Pagination {
  page: number
  pageSize: number
  total: number
}

// 连接测试结果
export interface TestResult {
  connected: boolean
  version?: string
  latency_ms?: number
  error?: string
}

// 连接列表项响应（后端返回的扁平化格式）
export interface ConnectionListItem {
  id: number
  name: string
  db_type: DbType
  connection_type: ConnectionType
  host: string
  port: number
  database?: string
  is_default: boolean
  last_used_at?: string | null
  created_at: string
  updated_at?: string
}

// 连接列表响应（后端原始格式）
export interface ConnectionListResponseRaw {
  list: ConnectionListItem[]
  total: number
  page: number
  page_size: number
}

// 连接列表响应（转换后的格式）
export interface ConnectionListResponse {
  list: Connection[]
  total: number
  page: number
  page_size: number
}

// 连接详情响应（后端返回格式）
export interface ConnectionDetailResponse {
  id: number
  name: string
  db_type: DbType
  connection_type: ConnectionType
  db_config?: {
    host: string
    port: number
    database?: string
    username: string
    options?: {
      ssl_mode?: string
      charset?: string
      timeout?: number
    }
  }
  ssh_config?: {
    host: string
    port: number
    username: string
    auth_type: SshAuthType
  }
  is_default: boolean
  last_used_at?: string | null
  created_at: string
  updated_at?: string
}

// 数据库类型配置
export const DB_TYPE_CONFIG: Record<DbType, { label: string; color: string; defaultPort: number }> = {
  mysql: { label: 'MySQL', color: '#4479A1', defaultPort: 3306 },
  postgresql: { label: 'PostgreSQL', color: '#336791', defaultPort: 5432 },
  mariadb: { label: 'MariaDB', color: '#C0765A', defaultPort: 3306 },
}

// 默认表单数据
export const getDefaultFormData = (): ConnectionFormData => ({
  name: '',
  db_type: 'mysql',
  connection_type: 'direct',
  db_config: {
    host: 'localhost',
    port: 3306,
    database: '',
    username: 'root',
    password: '',
    options: {
      ssl_mode: 'disable',
      charset: 'utf8mb4',
      timeout: 30,
    },
  },
  ssh_config: {
    host: '',
    port: 22,
    username: '',
    auth_type: 'password',
    password: '',
    private_key: '',
    passphrase: '',
  },
})
