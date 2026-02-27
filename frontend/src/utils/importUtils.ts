/**
 * 数据导入工具函数
 */

import * as XLSX from 'xlsx'

/**
 * 解析 CSV 文件
 */
export function parseCSV(file: File): Promise<{ headers: string[]; data: unknown[][] }> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      try {
        const text = e.target?.result as string
        const lines = text.split('\n').filter(line => line.trim())
        
        if (lines.length === 0) {
          reject(new Error('CSV 文件为空'))
          return
        }

        // 解析 CSV（简单实现，支持引号包裹）
        const parseCSVLine = (line: string): string[] => {
          const result: string[] = []
          let current = ''
          let inQuotes = false
          
          for (let i = 0; i < line.length; i++) {
            const char = line[i]
            
            if (char === '"') {
              if (inQuotes && line[i + 1] === '"') {
                current += '"'
                i++ // 跳过下一个引号
              } else {
                inQuotes = !inQuotes
              }
            } else if (char === ',' && !inQuotes) {
              result.push(current.trim())
              current = ''
            } else {
              current += char
            }
          }
          result.push(current.trim())
          return result
        }

        const headers = parseCSVLine(lines[0])
        const data = lines.slice(1).map(line => parseCSVLine(line))

        resolve({ headers, data })
      } catch (error) {
        reject(error)
      }
    }

    reader.onerror = () => reject(new Error('读取文件失败'))
    reader.readAsText(file, 'UTF-8')
  })
}

/**
 * 解析 Excel 文件
 */
export function parseExcel(file: File, sheetIndex: number = 0): Promise<{ headers: string[]; data: unknown[][] }> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      try {
        const arrayBuffer = new Uint8Array(e.target?.result as ArrayBuffer)
        const workbook = XLSX.read(arrayBuffer, { type: 'array' })
        
        if (workbook.SheetNames.length === 0) {
          reject(new Error('Excel 文件没有工作表'))
          return
        }

        const sheetName = workbook.SheetNames[sheetIndex]
        const worksheet = workbook.Sheets[sheetName]
        const jsonData = XLSX.utils.sheet_to_json(worksheet, { header: 1, defval: null }) as unknown[][]

        if (jsonData.length === 0) {
          reject(new Error('工作表为空'))
          return
        }

        const headers = (jsonData[0] as unknown[]).map(h => String(h || ''))
        const data = jsonData.slice(1)

        resolve({ headers, data })
      } catch (error) {
        reject(error)
      }
    }

    reader.onerror = () => reject(new Error('读取文件失败'))
    reader.readAsArrayBuffer(file)
  })
}

/**
 * 解析 JSON 文件
 */
export function parseJSON(file: File): Promise<{ headers: string[]; data: unknown[][] }> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      try {
        const text = e.target?.result as string
        const jsonData = JSON.parse(text) as unknown[]

        if (!Array.isArray(jsonData) || jsonData.length === 0) {
          reject(new Error('JSON 文件格式错误：必须是对象数组'))
          return
        }

        // 获取所有可能的键作为表头
        const headersSet = new Set<string>()
        jsonData.forEach(item => {
          if (typeof item === 'object' && item !== null) {
            Object.keys(item).forEach(key => headersSet.add(key))
          }
        })
        const headers = Array.from(headersSet)

        // 转换为二维数组
        const data = jsonData.map(item => {
          if (typeof item !== 'object' || item === null) {
            return headers.map(() => null)
          }
          return headers.map(header => (item as Record<string, unknown>)[header] ?? null)
        })

        resolve({ headers, data })
      } catch (error) {
        reject(error)
      }
    }

    reader.onerror = () => reject(new Error('读取文件失败'))
    reader.readAsText(file, 'UTF-8')
  })
}

/**
 * 解析 SQL INSERT 语句文件
 */
export function parseSQL(file: File): Promise<{ headers: string[]; data: unknown[][] }> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = (e) => {
      try {
        const text = e.target?.result as string
        const insertRegex = /INSERT\s+INTO\s+[`"]?(\w+)[`"]?\s*\([^)]+\)\s*VALUES\s*\(([^)]+)\)/gi
        
        const matches = Array.from(text.matchAll(insertRegex))
        
        if (matches.length === 0) {
          reject(new Error('SQL 文件中没有找到 INSERT 语句'))
          return
        }

        // 从第一个 INSERT 语句提取表头和列名
        const firstMatch = matches[0]
        const columnsMatch = firstMatch[0].match(/\(([^)]+)\)/)
        if (!columnsMatch) {
          reject(new Error('无法解析 INSERT 语句'))
          return
        }

        const headers = columnsMatch[1]
          .split(',')
          .map(col => col.trim().replace(/[`"]/g, ''))

        // 解析所有 VALUES
        const data = matches.map(match => {
          const valuesMatch = match[0].match(/VALUES\s*\(([^)]+)\)/i)
          if (!valuesMatch) return headers.map(() => null)
          
          // 简单解析值（支持引号包裹的字符串）
          const valuesStr = valuesMatch[1]
          const values: unknown[] = []
          let current = ''
          let inQuotes = false
          let quoteChar = ''

          for (let i = 0; i < valuesStr.length; i++) {
            const char = valuesStr[i]
            
            if ((char === '"' || char === "'") && !inQuotes) {
              inQuotes = true
              quoteChar = char
            } else if (char === quoteChar && inQuotes) {
              if (valuesStr[i + 1] === quoteChar) {
                current += quoteChar
                i++
              } else {
                inQuotes = false
                quoteChar = ''
              }
            } else if (char === ',' && !inQuotes) {
              const trimmed = current.trim()
              if (trimmed === 'NULL') {
                values.push(null)
              } else if (!isNaN(Number(trimmed)) && trimmed !== '') {
                values.push(Number(trimmed))
              } else {
                values.push(trimmed)
              }
              current = ''
            } else {
              current += char
            }
          }

          // 处理最后一个值
          const trimmed = current.trim()
          if (trimmed === 'NULL') {
            values.push(null)
          } else if (!isNaN(Number(trimmed)) && trimmed !== '') {
            values.push(Number(trimmed))
          } else {
            values.push(trimmed)
          }

          return values
        })

        resolve({ headers, data })
      } catch (error) {
        reject(error)
      }
    }

    reader.onerror = () => reject(new Error('读取文件失败'))
    reader.readAsText(file, 'UTF-8')
  })
}
