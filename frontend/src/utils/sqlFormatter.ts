/**
 * SQL 格式化工具
 */

import { format } from 'sql-formatter'

/**
 * 格式化 SQL 语句
 */
export function formatSQL(sql: string, dbType: 'mysql' | 'postgresql' = 'mysql'): string {
  try {
    const language = dbType === 'postgresql' ? 'postgresql' : 'mysql'
    return format(sql, {
      language,
      tabWidth: 2,
      keywordCase: 'upper',
      indentStyle: 'standard',
      linesBetweenQueries: 2,
    })
  } catch (error) {
    console.error('SQL格式化失败:', error)
    return sql // 格式化失败时返回原SQL
  }
}

/**
 * 压缩 SQL 语句（移除多余空格和换行）
 */
export function minifySQL(sql: string): string {
  return sql
    .replace(/\s+/g, ' ')
    .replace(/\s*([,;()])\s*/g, '$1')
    .trim()
}
