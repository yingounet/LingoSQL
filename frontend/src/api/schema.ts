/**
 * 表结构 API
 */

import request from '@/utils/request'
import type { ApiResponse } from '@/types/api'
import type { DbType } from '@/types/connection'

/**
 * 字段信息
 */
export interface ColumnInfo {
  name: string
  type: string
  length: number | null
  precision: number | null       // 精度（DECIMAL 等）
  scale: number | null           // 小数位数
  nullable: boolean
  default_value: string | null
  comment: string
  is_primary: boolean
  // MySQL/MariaDB 特有
  auto_increment: boolean
  unsigned?: boolean
  zerofill?: boolean
  charset?: string
  collation?: string
  on_update?: string
  // PostgreSQL 特有
  identity?: 'ALWAYS' | 'BY DEFAULT' | ''
  is_array?: boolean
  dimension?: number             // 数组维度
  // 通用
  extra: string
}

/**
 * 索引信息
 */
export interface IndexInfo {
  name: string
  type: 'PRIMARY' | 'UNIQUE' | 'INDEX' | 'FULLTEXT' | 'SPATIAL' | string
  columns: string[]
  comment: string
  cardinality: number | null
  // 索引方法
  method?: string                // BTREE, HASH, gin, gist, brin 等
  // PostgreSQL 特有
  where_clause?: string          // 部分索引条件
  nulls_not_distinct?: boolean
  // MySQL 特有
  key_length?: number[]          // 各列的前缀长度
}

/**
 * 获取表字段列表
 */
export async function getTableColumns(
  connectionId: number,
  database: string,
  table: string,
  dbType: DbType = 'mysql'
): Promise<ColumnInfo[]> {
  const response = await request.get<ApiResponse<ColumnInfo[]>>('/tables/columns', {
    params: {
      connection_id: connectionId,
      database: database,
      table: table
    }
  })
  return response.data
}

/**
 * 获取表索引列表
 */
export async function getTableIndexes(
  connectionId: number,
  database: string,
  table: string,
  dbType: DbType = 'mysql'
): Promise<IndexInfo[]> {
  const response = await request.get<ApiResponse<IndexInfo[]>>('/tables/indexes', {
    params: {
      connection_id: connectionId,
      database: database,
      table: table
    }
  })
  return response.data
}

/**
 * 创建空的字段对象
 */
export function createEmptyColumn(dbType: DbType): ColumnInfo {
  const base: ColumnInfo = {
    name: '',
    type: dbType === 'postgresql' ? 'VARCHAR' : 'VARCHAR',
    length: 255,
    precision: null,
    scale: null,
    nullable: true,
    default_value: null,
    comment: '',
    is_primary: false,
    auto_increment: false,
    extra: '',
  }

  if (dbType === 'postgresql') {
    return {
      ...base,
      identity: '',
      is_array: false,
      dimension: 1,
    }
  }

  return {
    ...base,
    unsigned: false,
    zerofill: false,
    charset: '',
    collation: '',
    on_update: '',
  }
}

/**
 * 创建空的索引对象
 */
export function createEmptyIndex(dbType: DbType): IndexInfo {
  const base: IndexInfo = {
    name: '',
    type: 'INDEX',
    columns: [],
    comment: '',
    cardinality: null,
    method: dbType === 'postgresql' ? 'btree' : 'BTREE',
  }

  if (dbType === 'postgresql') {
    return {
      ...base,
      where_clause: '',
      nulls_not_distinct: false,
    }
  }

  return {
    ...base,
    key_length: [],
  }
}
