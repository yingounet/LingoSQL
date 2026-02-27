/**
 * 数据库类型配置
 * 根据不同数据库类型定义字段类型、属性和索引类型
 */

import type { DbType } from '@/types/connection'

/**
 * 字段类型分组
 */
export interface ColumnTypeGroup {
  label: string
  types: string[]
}

/**
 * 字段属性配置
 */
export interface ColumnAttributeConfig {
  key: string                    // 属性键名
  label: string                  // 显示名称
  type: 'switch' | 'select' | 'input' | 'number'  // 控件类型
  options?: { value: string; label: string }[]    // 选项（select 用）
  showInTable?: boolean          // 是否在表格中显示
  editable?: boolean             // 是否可编辑（默认 true）
  width?: number                 // 表格列宽度
  appliesTo?: string[]           // 适用的字段类型（不设置则全部适用）
  dependsOn?: { key: string; value: any }  // 依赖条件
}

/**
 * 索引属性配置
 */
export interface IndexAttributeConfig {
  key: string
  label: string
  type: 'switch' | 'select' | 'input' | 'columns'
  options?: { value: string; label: string }[]
  required?: boolean
  showInTable?: boolean
}

/**
 * 数据库类型 Schema 配置
 */
export interface DbTypeSchemaConfig {
  // 字段类型分组
  columnTypeGroups: ColumnTypeGroup[]
  // 所有字段类型（扁平化）
  columnTypes: string[]
  // 字段属性配置
  columnAttributes: ColumnAttributeConfig[]
  // 索引类型
  indexTypes: { value: string; label: string }[]
  // 索引属性配置
  indexAttributes: IndexAttributeConfig[]
  // 索引方法
  indexMethods?: { value: string; label: string }[]
  // 字符集选项
  charsets?: { value: string; label: string }[]
  // 排序规则选项
  collations?: { value: string; label: string }[]
}

// ============ MySQL/MariaDB 配置 ============

const mysqlColumnTypeGroups: ColumnTypeGroup[] = [
  {
    label: '数值类型',
    types: ['TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'BIGINT', 'DECIMAL', 'FLOAT', 'DOUBLE', 'BIT']
  },
  {
    label: '字符串类型',
    types: ['CHAR', 'VARCHAR', 'TINYTEXT', 'TEXT', 'MEDIUMTEXT', 'LONGTEXT']
  },
  {
    label: '日期时间类型',
    types: ['DATE', 'TIME', 'DATETIME', 'TIMESTAMP', 'YEAR']
  },
  {
    label: '二进制类型',
    types: ['BINARY', 'VARBINARY', 'TINYBLOB', 'BLOB', 'MEDIUMBLOB', 'LONGBLOB']
  },
  {
    label: '特殊类型',
    types: ['ENUM', 'SET', 'JSON']
  },
  {
    label: '空间类型',
    types: ['GEOMETRY', 'POINT', 'LINESTRING', 'POLYGON']
  }
]

const mysqlCharsets = [
  { value: 'utf8mb4', label: 'utf8mb4 (推荐)' },
  { value: 'utf8', label: 'utf8' },
  { value: 'latin1', label: 'latin1' },
  { value: 'gbk', label: 'gbk' },
  { value: 'gb2312', label: 'gb2312' },
  { value: 'ascii', label: 'ascii' },
  { value: 'binary', label: 'binary' },
]

const mysqlCollations = [
  { value: 'utf8mb4_general_ci', label: 'utf8mb4_general_ci' },
  { value: 'utf8mb4_unicode_ci', label: 'utf8mb4_unicode_ci' },
  { value: 'utf8mb4_bin', label: 'utf8mb4_bin' },
  { value: 'utf8_general_ci', label: 'utf8_general_ci' },
  { value: 'utf8_unicode_ci', label: 'utf8_unicode_ci' },
  { value: 'latin1_swedish_ci', label: 'latin1_swedish_ci' },
  { value: 'gbk_chinese_ci', label: 'gbk_chinese_ci' },
]

const mysqlColumnAttributes: ColumnAttributeConfig[] = [
  {
    key: 'unsigned',
    label: '无符号',
    type: 'switch',
    showInTable: true,
    width: 80,
    appliesTo: ['TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'BIGINT', 'DECIMAL', 'FLOAT', 'DOUBLE']
  },
  {
    key: 'zerofill',
    label: '填充零',
    type: 'switch',
    showInTable: false,
    appliesTo: ['TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'BIGINT', 'DECIMAL', 'FLOAT', 'DOUBLE']
  },
  {
    key: 'auto_increment',
    label: '自增',
    type: 'switch',
    showInTable: true,
    width: 80,
    appliesTo: ['TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'BIGINT']
  },
  {
    key: 'charset',
    label: '字符集',
    type: 'select',
    showInTable: false,
    appliesTo: ['CHAR', 'VARCHAR', 'TINYTEXT', 'TEXT', 'MEDIUMTEXT', 'LONGTEXT', 'ENUM', 'SET']
  },
  {
    key: 'collation',
    label: '排序规则',
    type: 'select',
    showInTable: false,
    appliesTo: ['CHAR', 'VARCHAR', 'TINYTEXT', 'TEXT', 'MEDIUMTEXT', 'LONGTEXT', 'ENUM', 'SET']
  },
  {
    key: 'on_update',
    label: '更新时',
    type: 'select',
    options: [
      { value: '', label: '无' },
      { value: 'CURRENT_TIMESTAMP', label: 'CURRENT_TIMESTAMP' },
    ],
    showInTable: false,
    appliesTo: ['TIMESTAMP', 'DATETIME']
  }
]

const mysqlIndexTypes = [
  { value: 'PRIMARY', label: '主键 (PRIMARY)' },
  { value: 'UNIQUE', label: '唯一索引 (UNIQUE)' },
  { value: 'INDEX', label: '普通索引 (INDEX)' },
  { value: 'FULLTEXT', label: '全文索引 (FULLTEXT)' },
  { value: 'SPATIAL', label: '空间索引 (SPATIAL)' },
]

const mysqlIndexMethods = [
  { value: 'BTREE', label: 'BTREE (默认)' },
  { value: 'HASH', label: 'HASH' },
]

const mysqlIndexAttributes: IndexAttributeConfig[] = [
  {
    key: 'method',
    label: '索引方法',
    type: 'select',
    showInTable: true,
  },
  {
    key: 'key_length',
    label: '前缀长度',
    type: 'number',
    showInTable: false,
  }
]

const mysqlConfig: DbTypeSchemaConfig = {
  columnTypeGroups: mysqlColumnTypeGroups,
  columnTypes: mysqlColumnTypeGroups.flatMap(g => g.types),
  columnAttributes: mysqlColumnAttributes,
  indexTypes: mysqlIndexTypes,
  indexAttributes: mysqlIndexAttributes,
  indexMethods: mysqlIndexMethods,
  charsets: mysqlCharsets,
  collations: mysqlCollations,
}

// ============ MariaDB 配置（基于 MySQL，有少量差异）============

const mariadbColumnTypeGroups: ColumnTypeGroup[] = [
  ...mysqlColumnTypeGroups.slice(0, -1), // 复用 MySQL 大部分类型
  {
    label: '特殊类型',
    types: ['ENUM', 'SET', 'JSON', 'UUID'] // MariaDB 10.7+ 支持 UUID
  },
  mysqlColumnTypeGroups[mysqlColumnTypeGroups.length - 1], // 空间类型
]

const mariadbConfig: DbTypeSchemaConfig = {
  ...mysqlConfig,
  columnTypeGroups: mariadbColumnTypeGroups,
  columnTypes: mariadbColumnTypeGroups.flatMap(g => g.types),
}

// ============ PostgreSQL 配置 ============

const postgresqlColumnTypeGroups: ColumnTypeGroup[] = [
  {
    label: '数值类型',
    types: ['SMALLINT', 'INTEGER', 'BIGINT', 'DECIMAL', 'NUMERIC', 'REAL', 'DOUBLE PRECISION']
  },
  {
    label: '自增类型',
    types: ['SMALLSERIAL', 'SERIAL', 'BIGSERIAL']
  },
  {
    label: '字符串类型',
    types: ['CHAR', 'VARCHAR', 'TEXT']
  },
  {
    label: '日期时间类型',
    types: ['DATE', 'TIME', 'TIMESTAMP', 'TIMESTAMPTZ', 'INTERVAL']
  },
  {
    label: '布尔类型',
    types: ['BOOLEAN']
  },
  {
    label: '二进制类型',
    types: ['BYTEA']
  },
  {
    label: 'JSON 类型',
    types: ['JSON', 'JSONB']
  },
  {
    label: '网络类型',
    types: ['INET', 'CIDR', 'MACADDR', 'MACADDR8']
  },
  {
    label: '其他类型',
    types: ['UUID', 'MONEY', 'XML', 'POINT', 'LINE', 'BOX', 'PATH', 'POLYGON', 'CIRCLE']
  }
]

const postgresqlCollations = [
  { value: 'default', label: '默认' },
  { value: 'C', label: 'C' },
  { value: 'POSIX', label: 'POSIX' },
  { value: 'en_US.UTF-8', label: 'en_US.UTF-8' },
  { value: 'zh_CN.UTF-8', label: 'zh_CN.UTF-8' },
]

const postgresqlColumnAttributes: ColumnAttributeConfig[] = [
  {
    key: 'identity',
    label: '标识列',
    type: 'select',
    options: [
      { value: '', label: '无' },
      { value: 'ALWAYS', label: 'GENERATED ALWAYS' },
      { value: 'BY DEFAULT', label: 'GENERATED BY DEFAULT' },
    ],
    showInTable: true,
    width: 100,
    appliesTo: ['SMALLINT', 'INTEGER', 'BIGINT']
  },
  {
    key: 'is_array',
    label: '数组',
    type: 'switch',
    showInTable: true,
    width: 70,
  },
  {
    key: 'collation',
    label: '排序规则',
    type: 'select',
    showInTable: false,
    appliesTo: ['CHAR', 'VARCHAR', 'TEXT']
  },
  {
    key: 'dimension',
    label: '数组维度',
    type: 'number',
    showInTable: false,
    dependsOn: { key: 'is_array', value: true }
  }
]

const postgresqlIndexTypes = [
  { value: 'PRIMARY', label: '主键 (PRIMARY KEY)' },
  { value: 'UNIQUE', label: '唯一索引 (UNIQUE)' },
  { value: 'INDEX', label: '普通索引 (INDEX)' },
]

const postgresqlIndexMethods = [
  { value: 'btree', label: 'B-tree (默认)' },
  { value: 'hash', label: 'Hash' },
  { value: 'gist', label: 'GiST' },
  { value: 'gin', label: 'GIN' },
  { value: 'brin', label: 'BRIN' },
  { value: 'spgist', label: 'SP-GiST' },
]

const postgresqlIndexAttributes: IndexAttributeConfig[] = [
  {
    key: 'method',
    label: '索引方法',
    type: 'select',
    showInTable: true,
  },
  {
    key: 'where_clause',
    label: 'WHERE 条件',
    type: 'input',
    showInTable: false,
  },
  {
    key: 'nulls_not_distinct',
    label: 'NULLS NOT DISTINCT',
    type: 'switch',
    showInTable: false,
  }
]

const postgresqlConfig: DbTypeSchemaConfig = {
  columnTypeGroups: postgresqlColumnTypeGroups,
  columnTypes: postgresqlColumnTypeGroups.flatMap(g => g.types),
  columnAttributes: postgresqlColumnAttributes,
  indexTypes: postgresqlIndexTypes,
  indexAttributes: postgresqlIndexAttributes,
  indexMethods: postgresqlIndexMethods,
  collations: postgresqlCollations,
}

// ============ 配置映射 ============

const dbTypeConfigs: Record<DbType, DbTypeSchemaConfig> = {
  mysql: mysqlConfig,
  mariadb: mariadbConfig,
  postgresql: postgresqlConfig,
}

/**
 * 获取数据库类型配置
 */
export function getDbTypeConfig(dbType: DbType): DbTypeSchemaConfig {
  return dbTypeConfigs[dbType] || mysqlConfig
}

/**
 * 判断字段类型是否需要长度参数
 */
export function typeNeedsLength(type: string, dbType: DbType): boolean {
  const needsLength = ['CHAR', 'VARCHAR', 'BINARY', 'VARBINARY', 'BIT']
  return needsLength.includes(type.toUpperCase())
}

/**
 * 判断字段类型是否需要精度和小数位
 */
export function typeNeedsPrecision(type: string, dbType: DbType): boolean {
  const needsPrecision = ['DECIMAL', 'NUMERIC', 'FLOAT', 'DOUBLE', 'DOUBLE PRECISION', 'REAL']
  return needsPrecision.includes(type.toUpperCase())
}

/**
 * 判断字段类型是否为数值类型
 */
export function isNumericType(type: string): boolean {
  const numericTypes = [
    'TINYINT', 'SMALLINT', 'MEDIUMINT', 'INT', 'INTEGER', 'BIGINT',
    'DECIMAL', 'NUMERIC', 'FLOAT', 'DOUBLE', 'DOUBLE PRECISION', 'REAL',
    'SMALLSERIAL', 'SERIAL', 'BIGSERIAL', 'BIT'
  ]
  return numericTypes.includes(type.toUpperCase())
}

/**
 * 判断字段类型是否为字符类型
 */
export function isStringType(type: string): boolean {
  const stringTypes = [
    'CHAR', 'VARCHAR', 'TINYTEXT', 'TEXT', 'MEDIUMTEXT', 'LONGTEXT',
    'ENUM', 'SET'
  ]
  return stringTypes.includes(type.toUpperCase())
}

/**
 * 判断字段类型是否为日期时间类型
 */
export function isDateTimeType(type: string): boolean {
  const dateTimeTypes = [
    'DATE', 'TIME', 'DATETIME', 'TIMESTAMP', 'TIMESTAMPTZ', 'YEAR', 'INTERVAL'
  ]
  return dateTimeTypes.includes(type.toUpperCase())
}

/**
 * 获取字段类型的默认长度
 */
export function getDefaultLength(type: string): number | null {
  const defaults: Record<string, number> = {
    'VARCHAR': 255,
    'CHAR': 1,
    'BINARY': 1,
    'VARBINARY': 255,
    'BIT': 1,
  }
  return defaults[type.toUpperCase()] || null
}

/**
 * 获取字段类型的默认精度和小数位
 */
export function getDefaultPrecision(type: string): { precision: number; scale: number } | null {
  const defaults: Record<string, { precision: number; scale: number }> = {
    'DECIMAL': { precision: 10, scale: 2 },
    'NUMERIC': { precision: 10, scale: 2 },
  }
  return defaults[type.toUpperCase()] || null
}
