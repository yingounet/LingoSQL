/**
 * Monaco Editor SQL 自动补全提供者
 */

import * as monaco from 'monaco-editor'
import type { Connection } from '@/types/connection'
import { getTables } from '@/api/table'
import { getTableColumns } from '@/api/schema'

/**
 * SQL 关键字列表
 */
const SQL_KEYWORDS = [
  'SELECT', 'FROM', 'WHERE', 'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'DROP', 'ALTER',
  'TABLE', 'DATABASE', 'INDEX', 'VIEW', 'PROCEDURE', 'FUNCTION', 'TRIGGER',
  'JOIN', 'INNER', 'LEFT', 'RIGHT', 'FULL', 'OUTER', 'ON', 'AS', 'AND', 'OR', 'NOT',
  'IN', 'EXISTS', 'LIKE', 'BETWEEN', 'IS', 'NULL', 'ORDER', 'BY', 'GROUP', 'HAVING',
  'LIMIT', 'OFFSET', 'UNION', 'ALL', 'DISTINCT', 'COUNT', 'SUM', 'AVG', 'MAX', 'MIN',
  'CASE', 'WHEN', 'THEN', 'ELSE', 'END', 'IF', 'ELSEIF', 'BEGIN', 'COMMIT', 'ROLLBACK',
  'TRANSACTION', 'GRANT', 'REVOKE', 'PRIMARY', 'KEY', 'FOREIGN', 'REFERENCES',
  'CONSTRAINT', 'UNIQUE', 'CHECK', 'DEFAULT', 'AUTO_INCREMENT', 'IDENTITY',
  'VARCHAR', 'CHAR', 'TEXT', 'INT', 'INTEGER', 'BIGINT', 'SMALLINT', 'TINYINT',
  'DECIMAL', 'NUMERIC', 'FLOAT', 'DOUBLE', 'REAL', 'BOOLEAN', 'DATE', 'TIME',
  'DATETIME', 'TIMESTAMP', 'YEAR', 'BLOB', 'JSON', 'ARRAY'
]

/**
 * 创建 SQL 自动补全提供者
 */
export function createSqlCompletionProvider(
  connection: Connection | null,
  database: string | null
): monaco.languages.CompletionItemProvider {
  return {
    provideCompletionItems: async (model, position) => {
      const word = model.getWordUntilPosition(position)
      const range = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: word.startColumn,
        endColumn: word.endColumn,
      }

      const suggestions: monaco.languages.CompletionItem[] = []

      // 添加 SQL 关键字
      SQL_KEYWORDS.forEach(keyword => {
        suggestions.push({
          label: keyword,
          kind: monaco.languages.CompletionItemKind.Keyword,
          insertText: keyword,
          range,
          detail: 'SQL 关键字',
        })
      })

      // 如果已连接数据库，添加表和字段建议
      if (connection && database) {
        try {
          // 获取表列表
          const tables = await getTables(connection.id, database)
          
          tables.forEach(table => {
            // 添加表名建议
            suggestions.push({
              label: table.name,
              kind: monaco.languages.CompletionItemKind.Class,
              insertText: table.name,
              range,
              detail: `表: ${table.name}`,
              documentation: table.engine ? `引擎: ${table.engine}` : undefined,
            })

            // 添加表名.字段名建议
            suggestions.push({
              label: `${table.name}.*`,
              kind: monaco.languages.CompletionItemKind.Field,
              insertText: `${table.name}.*`,
              range,
              detail: `表 ${table.name} 的所有字段`,
            })
          })

          // 获取当前行的文本，尝试解析表名
          const lineText = model.getLineContent(position.lineNumber)
          const beforeCursor = lineText.substring(0, position.column - 1)
          
          // 检查是否在 FROM 或 JOIN 之后
          const fromMatch = beforeCursor.match(/\bFROM\s+(\w+)\s*$/i)
          const joinMatch = beforeCursor.match(/\bJOIN\s+(\w+)\s*$/i)
          const tableAliasMatch = beforeCursor.match(/\bFROM\s+\w+\s+(\w+)\s*$/i)
          
          let tableName: string | null = null
          if (fromMatch) tableName = fromMatch[1]
          else if (joinMatch) tableName = joinMatch[1]
          else if (tableAliasMatch) tableName = tableAliasMatch[1]

          // 如果找到表名，添加字段建议
          if (tableName) {
            const table = tables.find(t => t.name === tableName || t.name.toLowerCase() === tableName.toLowerCase())
            if (table) {
              try {
                const columns = await getTableColumns(connection.id, database, table.name, connection.db_type)
                columns.forEach(column => {
                  suggestions.push({
                    label: column.name,
                    kind: monaco.languages.CompletionItemKind.Field,
                    insertText: column.name,
                    range,
                    detail: `字段: ${column.type}`,
                    documentation: column.comment || undefined,
                  })
                })
              } catch (error) {
                console.error('获取字段列表失败:', error)
              }
            }
          }
        } catch (error) {
          console.error('获取表列表失败:', error)
        }
      }

      return { suggestions }
    },
    triggerCharacters: ['.', ' '],
  }
}

/**
 * 注册 SQL 语言的自定义补全提供者
 */
export function registerSqlCompletionProvider(
  connection: Connection | null,
  database: string | null
): monaco.IDisposable {
  return monaco.languages.registerCompletionItemProvider(
    'sql',
    createSqlCompletionProvider(connection, database)
  )
}
