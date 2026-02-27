/**
 * 数据导出工具函数
 */

import * as XLSX from 'xlsx'

/**
 * 导出为 CSV
 */
export function exportToCSV(
  data: unknown[][],
  headers: string[],
  filename: string = 'export.csv'
): void {
  // 添加 BOM 以支持中文
  const BOM = '\uFEFF'
  const csvContent = [
    headers.join(','),
    ...data.map(row =>
      row.map(cell => {
        if (cell === null || cell === undefined) return ''
        const str = String(cell)
        // 如果包含逗号、引号或换行符，需要用引号包裹
        if (str.includes(',') || str.includes('"') || str.includes('\n')) {
          return `"${str.replace(/"/g, '""')}"`
        }
        return str
      }).join(',')
    )
  ].join('\n')

  const blob = new Blob([BOM + csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

/**
 * 导出为 Excel
 */
export function exportToExcel(
  data: unknown[][],
  headers: string[],
  filename: string = 'export.xlsx',
  sheetName: string = 'Sheet1'
): void {
  // 创建工作簿
  const wb = XLSX.utils.book_new()
  
  // 准备数据（包含表头）
  const wsData = [headers, ...data]
  
  // 创建工作表
  const ws = XLSX.utils.aoa_to_sheet(wsData)
  
  // 设置列宽
  const colWidths = headers.map((_, colIndex) => {
    const maxLength = Math.max(
      headers[colIndex].length,
      ...data.map(row => String(row[colIndex] || '').length)
    )
    return { wch: Math.min(maxLength + 2, 50) }
  })
  ws['!cols'] = colWidths
  
  // 添加工作表到工作簿
  XLSX.utils.book_append_sheet(wb, ws, sheetName)
  
  // 导出文件
  XLSX.writeFile(wb, filename)
}

/**
 * 导出为 JSON
 */
export function exportToJSON(
  data: Record<string, unknown>[],
  filename: string = 'export.json'
): void {
  const jsonContent = JSON.stringify(data, null, 2)
  const blob = new Blob([jsonContent], { type: 'application/json;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

/**
 * 导出为 SQL INSERT 语句
 */
export function exportToSQL(
  data: unknown[][],
  headers: string[],
  tableName: string,
  filename: string = 'export.sql'
): void {
  const sqlStatements = data.map(row => {
    const values = row.map(cell => {
      if (cell === null || cell === undefined) return 'NULL'
      if (typeof cell === 'number') return String(cell)
      // 转义单引号
      return `'${String(cell).replace(/'/g, "''")}'`
    }).join(', ')
    return `INSERT INTO \`${tableName}\` (\`${headers.join('`, `')}\`) VALUES (${values});`
  }).join('\n')

  const blob = new Blob([sqlStatements], { type: 'text/plain;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}
