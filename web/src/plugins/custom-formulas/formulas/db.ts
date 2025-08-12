import type { BaseValueObject, IFunctionInfo } from '@univerjs/preset-sheets-core'
import {
  ArrayValueObject,
  BaseFunction,
  FunctionType,
  NumberValueObject,
  StringValueObject,
} from '@univerjs/preset-sheets-core'
import { useUserStore } from '@/pinia/modules/user'

/**
 * 向下填充配置接口
 */
interface FillDownConfig {
  maxRows: number;           // 最大填充行数
  stopOnNonEmpty: boolean;   // 遇到非空单元格时是否停止
  showWarnings: boolean;     // 是否显示警告信息
}

/**
 * 默认填充配置
 */
const DEFAULT_FILL_CONFIG: FillDownConfig = {
  maxRows: 1000,
  stopOnNonEmpty: true,
  showWarnings: true
}

/**
 * SUGAR.CALC 自定义函数实现
 * 对指定的数据列进行聚合计算，返回单个结果
 */
export class SugarCalcFunction extends BaseFunction {
  override calculate(...args: BaseValueObject[]): BaseValueObject {
    // 参数验证：至少需要3个参数（模型名称、计算列、计算方式）
    if (args.length < 3) {
      return StringValueObject.create('#VALUE!')
    }

    // 提取参数
    const modelName = this.getStringValue(args[0])
    const calcColumn = this.getStringValue(args[1])
    const calcMethod = this.getStringValue(args[2])

    // 验证必填参数
    if (!modelName || !calcColumn || !calcMethod) {
      return StringValueObject.create('#VALUE!')
    }

    // 验证计算方式
    const validMethods = ['SUM', 'AVG', 'COUNT', 'MAX', 'MIN']
    if (!validMethods.includes(calcMethod.toUpperCase())) {
      return StringValueObject.create('#NAME?')
    }

    // 解析筛选条件（支持成对参数和冒号格式）
    const filters: Record<string, any> = {}
    let i = 3
    while (i < args.length) {
      const currentArg = this.getStringValue(args[i])
      
      // 检查是否是 "key:value" 格式
      if (currentArg.includes(':')) {
        const [filterKey, ...valueParts] = currentArg.split(':')
        const filterValue = valueParts.join(':') // 处理值中可能包含冒号的情况
        if (filterKey && filterValue) {
          filters[filterKey.trim()] = filterValue.trim()
        }
        i += 1
      } else {
        // 成对参数格式 "key", "value"
        if (i + 1 < args.length) {
          const filterKey = currentArg
          const filterValue = this.getValue(args[i + 1])
          if (filterKey) {
            filters[filterKey] = filterValue
          }
          i += 2
        } else {
          // 奇数个参数，跳过最后一个
          i += 1
        }
      }
    }

    // 调用后端API
    return this.executeCalcFormula(modelName, calcColumn, calcMethod.toUpperCase(), filters)
  }

  private executeCalcFormula(
    modelName: string,
    calcColumn: string,
    calcMethod: string,
    filters: Record<string, any>
  ): BaseValueObject {
    try {
      // 构建请求数据
      const requestData = {
        modelName,
        calcColumn,
        calcMethod,
        filters
      }

      // 发送同步请求到后端
      const xhr = new XMLHttpRequest()
      xhr.open('POST', '/api/sugarFormulaQuery/executeCalc', false) // 同步请求
      xhr.setRequestHeader('Content-Type', 'application/json')
      
      // 添加认证头
      const userStore = useUserStore()
      if (userStore.token) {
        xhr.setRequestHeader('x-token', userStore.token)
      }
      if ((userStore.userInfo as any).ID) {
        xhr.setRequestHeader('x-user-id', (userStore.userInfo as any).ID)
      }

      xhr.send(JSON.stringify(requestData))

      if (xhr.status === 200) {
        const response = JSON.parse(xhr.responseText)
        if (response.code === 0 && response.data) {
          const result = response.data.result
          if (typeof result === 'number') {
            return new NumberValueObject(result)
          } else if (result !== null && result !== undefined) {
            return StringValueObject.create(String(result))
          }
        } else {
          return StringValueObject.create(response.msg || '#ERROR!')
        }
      } else {
        return StringValueObject.create('#CONNECT!')
      }
    } catch (error) {
      return StringValueObject.create('#ERROR!')
    }

    return StringValueObject.create('#N/A')
  }

  private getStringValue(valueObject: BaseValueObject): string {
    const value = this.getValue(valueObject)
    return typeof value === 'string' ? value : String(value || '')
  }

  private getValue(valueObject: BaseValueObject): any {
    if (valueObject.isArray()) {
      const array = valueObject as ArrayValueObject
      const arrayValue = array.getArrayValue()
      
      // 递归函数来深度展平嵌套数组并提取第一个有效值
      const extractFirstValue = (arr: any): any => {
        if (!Array.isArray(arr) || arr.length === 0) {
          return ''
        }
        
        const firstItem = arr[0]
        
        // 如果第一个元素还是数组，继续递归
        if (Array.isArray(firstItem)) {
          return extractFirstValue(firstItem)
        }
        
        // 如果是 BaseValueObject，获取其值
        if (firstItem && typeof firstItem === 'object' && 'getValue' in firstItem) {
          return firstItem.getValue()
        }
        
        // 否则直接返回值
        return firstItem !== null && firstItem !== undefined ? firstItem : ''
      }
      
      return extractFirstValue(arrayValue)
    }
    return valueObject.getValue()
  }
}

/**
 * SUGAR.GET 自定义函数实现
 * 获取一列或多列明细数据，结果会动态向下填充
 *
 * 功能特性：
 * - 支持单列和多列数据查询
 * - 自动向下填充多行结果
 * - 限制最大返回行数（1000行）防止性能问题
 * - 智能处理空值和类型转换
 * - 提供详细的错误信息和警告
 *
 * 使用建议：
 * - 对于大量数据，建议添加筛选条件减少结果集
 * - 多行数据会自动向下填充，请确保下方单元格为空
 * - 建议在独立区域使用，避免覆盖重要数据
 */
export class SugarGetFunction extends BaseFunction {
  override calculate(...args: BaseValueObject[]): BaseValueObject {
    // 参数验证：至少需要2个参数（模型名称、返回列）
    if (args.length < 2) {
      return StringValueObject.create('#VALUE!')
    }

    // 提取参数
    const modelName = this.getStringValue(args[0])
    const returnColumns = this.getStringValue(args[1])

    // 验证必填参数
    if (!modelName || !returnColumns) {
      return StringValueObject.create('#VALUE!')
    }

    // 解析返回列（可能是逗号分隔的多列）
    const columnList = returnColumns.split(',').map(col => col.trim()).filter(col => col)
    if (columnList.length === 0) {
      return StringValueObject.create('#VALUE!')
    }

    // 解析筛选条件（支持成对参数和冒号格式）
    const filters: Record<string, any> = {}
    let i = 2
    while (i < args.length) {
      const currentArg = this.getStringValue(args[i])
      
      // 检查是否是 "key:value" 格式
      if (currentArg.includes(':')) {
        const [filterKey, ...valueParts] = currentArg.split(':')
        const filterValue = valueParts.join(':') // 处理值中可能包含冒号的情况
        if (filterKey && filterValue) {
          filters[filterKey.trim()] = filterValue.trim()
        }
        i += 1
      } else {
        // 成对参数格式 "key", "value"
        if (i + 1 < args.length) {
          const filterKey = currentArg
          const filterValue = this.getValue(args[i + 1])
          if (filterKey) {
            filters[filterKey] = filterValue
          }
          i += 2
        } else {
          // 奇数个参数，跳过最后一个
          i += 1
        }
      }
    }

    // 调用后端API
    return this.executeGetFormula(modelName, columnList, filters)
  }

  private executeGetFormula(
    modelName: string,
    returnColumns: string[],
    filters: Record<string, any>
  ): BaseValueObject {
    try {
      // 构建请求数据
      const requestData = {
        modelName,
        returnColumns,
        filters
      }

      // 发送同步请求到后端
      const xhr = new XMLHttpRequest()
      xhr.open('POST', '/api/sugarFormulaQuery/executeGet', false) // 同步请求
      xhr.setRequestHeader('Content-Type', 'application/json')
      
      // 添加认证头
      const userStore = useUserStore()
      if (userStore.token) {
        xhr.setRequestHeader('x-token', userStore.token)
      }
      if ((userStore.userInfo as any).ID) {
        xhr.setRequestHeader('x-user-id', (userStore.userInfo as any).ID)
      }

      xhr.send(JSON.stringify(requestData))

      if (xhr.status === 200) {
        const response = JSON.parse(xhr.responseText)
        if (response.code === 0 && response.data && response.data.results) {
          const results = response.data.results
          
          // 如果没有结果，返回空值
          if (!results || results.length === 0) {
            return StringValueObject.create('')
          }
          
          // 限制最大返回行数，避免性能问题
          const MAX_ROWS = 1000
          const limitedResults = results.slice(0, MAX_ROWS)
          
          if (results.length > MAX_ROWS) {
            console.warn(`SUGAR.GET: 数据行数 (${results.length}) 超过限制 (${MAX_ROWS})，已截取前 ${MAX_ROWS} 行`)
          }
          
          // 如果只有一列，返回一维数组
          if (returnColumns.length === 1) {
            const columnName = returnColumns[0]
            const values = limitedResults.map((row: any) => {
              const val = row[columnName]
              return val !== null && val !== undefined ? val : ''
            })
            
            // 如果只有一行，直接返回该值
            if (values.length === 1) {
              const singleValue = values[0]
              return typeof singleValue === 'number'
                ? new NumberValueObject(singleValue)
                : StringValueObject.create(String(singleValue))
            }
            
            // 多行数据，返回数组以支持向下填充
            return ArrayValueObject.create({
              calculateValueList: values.map((val: any) =>
                typeof val === 'number' ? [new NumberValueObject(val)] : [StringValueObject.create(String(val))]
              ),
              rowCount: values.length,
              columnCount: 1,
              unitId: '',
              sheetId: '',
              row: 0,
              column: 0,
            })
          } else {
            // 多列返回二维数组
            const rows = limitedResults.map((row: any) =>
              returnColumns.map(col => {
                const val = row[col]
                const processedVal = val !== null && val !== undefined ? val : ''
                return typeof processedVal === 'number'
                  ? new NumberValueObject(processedVal)
                  : StringValueObject.create(String(processedVal))
              })
            )
            
            return ArrayValueObject.create({
              calculateValueList: rows,
              rowCount: rows.length,
              columnCount: returnColumns.length,
              unitId: '',
              sheetId: '',
              row: 0,
              column: 0,
            })
          }
        } else {
          const errorMsg = response.msg || response.error || '#ERROR!'
          console.error('SUGAR.GET: 服务器返回错误:', errorMsg)
          return StringValueObject.create(errorMsg)
        }
      } else {
        console.error('SUGAR.GET: HTTP请求失败:', xhr.status, xhr.statusText)
        return StringValueObject.create('#CONNECT!')
      }
    } catch (error) {
      console.error('SUGAR.GET: 执行异常:', error)
      return StringValueObject.create('#ERROR!')
    }

    return StringValueObject.create('#N/A')
  }

  private getStringValue(valueObject: BaseValueObject): string {
    const value = this.getValue(valueObject)
    return typeof value === 'string' ? value : String(value || '')
  }

  private getValue(valueObject: BaseValueObject): any {
    if (valueObject.isArray()) {
      const array = valueObject as ArrayValueObject
      const arrayValue = array.getArrayValue()
      
      // 递归函数来深度展平嵌套数组并提取第一个有效值
      const extractFirstValue = (arr: any): any => {
        if (!Array.isArray(arr) || arr.length === 0) {
          return ''
        }
        
        const firstItem = arr[0]
        
        // 如果第一个元素还是数组，继续递归
        if (Array.isArray(firstItem)) {
          return extractFirstValue(firstItem)
        }
        
        // 如果是 BaseValueObject，获取其值
        if (firstItem && typeof firstItem === 'object' && 'getValue' in firstItem) {
          return firstItem.getValue()
        }
        
        // 否则直接返回值
        return firstItem !== null && firstItem !== undefined ? firstItem : ''
      }
      
      return extractFirstValue(arrayValue)
    }
    return valueObject.getValue()
  }
}

/**
 * SUGAR.CALC 函数的中文本地化
 */
export const functionSugarCalcZhCN = {
  formula: {
    functionList: {
      'SUGAR.CALC': {
        description: '对指定语义模型中的数据列进行聚合计算，返回单个结果。',
        abstract: '语义模型聚合计算',
        links: [
          {
            title: '教学',
            url: 'https://univer.ai',
          },
        ],
        functionParameter: {
          modelName: {
            name: '模型名称',
            detail: '语义模型的友好名称',
          },
          calcColumn: {
            name: '计算列',
            detail: '需要进行聚合计算的列的友好名称',
          },
          calcMethod: {
            name: '计算方式',
            detail: '支持 SUM, AVG, COUNT, MAX, MIN',
          },
          filters: {
            name: '筛选条件',
            detail: '可选的筛选条件，格式为：筛选列1, 筛选值1, 筛选列2, 筛选值2...',
          },
        },
      },
    },
  },
}

/**
 * SUGAR.GET 函数的中文本地化
 */
export const functionSugarGetZhCN = {
  formula: {
    functionList: {
      'SUGAR.GET': {
        description: '从指定语义模型中获取一列或多列明细数据，结果会动态向下填充。',
        abstract: '语义模型数据查询',
        links: [
          {
            title: '教学',
            url: 'https://univer.ai',
          },
        ],
        functionParameter: {
          modelName: {
            name: '模型名称',
            detail: '语义模型的友好名称',
          },
          returnColumns: {
            name: '返回列',
            detail: '需要返回的列名，多列用逗号分隔',
          },
          filters: {
            name: '筛选条件',
            detail: '可选的筛选条件，格式为：筛选列1, 筛选值1, 筛选列2, 筛选值2...',
          },
        },
      },
    },
  },
}

/**
 * 公式刷新缓存管理
 */
class FormulaRefreshCache {
  private static instance: FormulaRefreshCache
  private cache = new Map<string, { timestamp: number; result: any }>()
  private readonly CACHE_DURATION = 30000 // 30秒缓存

  static getInstance(): FormulaRefreshCache {
    if (!FormulaRefreshCache.instance) {
      FormulaRefreshCache.instance = new FormulaRefreshCache()
    }
    return FormulaRefreshCache.instance
  }

  getCacheKey(formulaName: string, params: any[]): string {
    return `${formulaName}:${JSON.stringify(params)}`
  }

  get(key: string): any | null {
    const cached = this.cache.get(key)
    if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
      return cached.result
    }
    this.cache.delete(key)
    return null
  }

  set(key: string, result: any): void {
    this.cache.set(key, { timestamp: Date.now(), result })
  }

  clear(): void {
    this.cache.clear()
  }

  clearExpired(): void {
    const now = Date.now()
    const keysToDelete: string[] = []
    
    this.cache.forEach((value, key) => {
      if (now - value.timestamp >= this.CACHE_DURATION) {
        keysToDelete.push(key)
      }
    })
    
    keysToDelete.forEach(key => this.cache.delete(key))
  }
}

/**
 * 强制刷新数据库公式
 */
export function forceRefreshDatabaseFormulas(): void {
  const cache = FormulaRefreshCache.getInstance()
  cache.clear()
  console.log('数据库公式缓存已清空，下次调用将重新获取数据')
}

/**
 * 数据库类公式定义
 */
export const dbFormulas = [
  {
    name: 'SUGAR.CALC',
    implementation: (modelName: any, calcColumn: any, calcMethod: any, ...filters: any[]) => {
      // 参数验证：至少需要3个参数（模型名称、计算列、计算方式）
      if (!modelName || !calcColumn || !calcMethod) {
        return '#VALUE!'
      }

      // 转换参数为字符串
      const modelNameStr = String(modelName || '')
      const calcColumnStr = String(calcColumn || '')
      const calcMethodStr = String(calcMethod || '').toUpperCase()

      // 验证计算方式
      const validMethods = ['SUM', 'AVG', 'COUNT', 'MAX', 'MIN']
      if (!validMethods.includes(calcMethodStr)) {
        return '#NAME?'
      }

      // 解析筛选条件（支持成对参数和冒号格式）
      const filterObj: Record<string, any> = {}
      let i = 0
      while (i < filters.length) {
        const currentArg = String(filters[i] || '')
        
        // 检查是否是 "key:value" 格式
        if (currentArg.includes(':')) {
          const [filterKey, ...valueParts] = currentArg.split(':')
          const filterValue = valueParts.join(':') // 处理值中可能包含冒号的情况
          if (filterKey && filterValue) {
            filterObj[filterKey.trim()] = filterValue.trim()
          }
          i += 1
        } else {
          // 成对参数格式 "key", "value"
          if (i + 1 < filters.length) {
            const filterKey = currentArg
            let filterValue = filters[i + 1]
            
            // 处理嵌套数组的情况（单元格引用）
            if (Array.isArray(filterValue)) {
              // 递归展平嵌套数组并提取第一个有效值
              const extractFirstValue = (arr: any): any => {
                if (!Array.isArray(arr) || arr.length === 0) {
                  return ''
                }
                const firstItem = arr[0]
                if (Array.isArray(firstItem)) {
                  return extractFirstValue(firstItem)
                }
                return firstItem !== null && firstItem !== undefined ? firstItem : ''
              }
              filterValue = extractFirstValue(filterValue)
            }
            
            if (filterKey) {
              filterObj[filterKey] = filterValue
            }
            i += 2
          } else {
            // 奇数个参数，跳过最后一个
            i += 1
          }
        }
      }

      try {
        // 构建请求数据
        const requestData = {
          modelName: modelNameStr,
          calcColumn: calcColumnStr,
          calcMethod: calcMethodStr,
          filters: filterObj
        }

        // 发送同步请求到后端
        const xhr = new XMLHttpRequest()
        xhr.open('POST', '/api/sugarFormulaQuery/executeCalc', false) // 同步请求
        xhr.setRequestHeader('Content-Type', 'application/json')
        
        // 添加认证头
        const userStore = useUserStore()
        if (userStore.token) {
          xhr.setRequestHeader('x-token', userStore.token)
        }
        if ((userStore.userInfo as any).ID) {
          xhr.setRequestHeader('x-user-id', (userStore.userInfo as any).ID)
        }

        xhr.send(JSON.stringify(requestData))

        if (xhr.status === 200) {
          const response = JSON.parse(xhr.responseText)
          if (response.code === 0 && response.data) {
            const result = response.data.result
            if (typeof result === 'number') {
              return result
            } else if (result !== null && result !== undefined) {
              return String(result)
            }
          } else {
            return response.msg || '#ERROR!'
          }
        } else {
          return '#CONNECT!'
        }
      } catch (error) {
        return '#ERROR!'
      }

      return '#N/A'
    },
    config: {
      description: {
        functionName: 'SUGAR.CALC',
        description: 'formula.functionList.SUGAR.CALC.description',
        abstract: 'formula.functionList.SUGAR.CALC.abstract',
        functionParameter: [
          {
            name: 'formula.functionList.SUGAR.CALC.functionParameter.modelName.name',
            detail: 'formula.functionList.SUGAR.CALC.functionParameter.modelName.detail',
            example: '"业务指标查询"',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.SUGAR.CALC.functionParameter.calcColumn.name',
            detail: 'formula.functionList.SUGAR.CALC.functionParameter.calcColumn.detail',
            example: '"指标金额"',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.SUGAR.CALC.functionParameter.calcMethod.name',
            detail: 'formula.functionList.SUGAR.CALC.functionParameter.calcMethod.detail',
            example: '"SUM"',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.SUGAR.CALC.functionParameter.filters.name',
            detail: 'formula.functionList.SUGAR.CALC.functionParameter.filters.detail',
            example: '"战区名称", "华东战区"',
            require: 0,
            repeat: 1,
          },
        ],
      },
      locales: {
        zhCN: functionSugarCalcZhCN,
      },
    },
    locales: functionSugarCalcZhCN,
  },
  {
    name: 'SUGAR.GET',
    implementation: (modelName: any, returnColumns: any, ...filters: any[]) => {
      // 参数验证：至少需要2个参数（模型名称、返回列）
      if (!modelName || !returnColumns) {
        return '#VALUE!'
      }

      // 转换参数为字符串
      const modelNameStr = String(modelName || '')
      const returnColumnsStr = String(returnColumns || '')

      // 解析返回列（可能是逗号分隔的多列）
      const columnList = returnColumnsStr.split(',').map(col => col.trim()).filter(col => col)
      if (columnList.length === 0) {
        return '#VALUE!'
      }

      // 解析筛选条件（支持成对参数和冒号格式）
      const filterObj: Record<string, any> = {}
      let i = 0
      while (i < filters.length) {
        const currentArg = String(filters[i] || '')
        
        // 检查是否是 "key:value" 格式
        if (currentArg.includes(':')) {
          const [filterKey, ...valueParts] = currentArg.split(':')
          const filterValue = valueParts.join(':') // 处理值中可能包含冒号的情况
          if (filterKey && filterValue) {
            filterObj[filterKey.trim()] = filterValue.trim()
          }
          i += 1
        } else {
          // 成对参数格式 "key", "value"
          if (i + 1 < filters.length) {
            const filterKey = currentArg
            let filterValue = filters[i + 1]
            
            // 处理嵌套数组的情况（单元格引用）
            if (Array.isArray(filterValue)) {
              // 递归展平嵌套数组并提取第一个有效值
              const extractFirstValue = (arr: any): any => {
                if (!Array.isArray(arr) || arr.length === 0) {
                  return ''
                }
                const firstItem = arr[0]
                if (Array.isArray(firstItem)) {
                  return extractFirstValue(firstItem)
                }
                return firstItem !== null && firstItem !== undefined ? firstItem : ''
              }
              filterValue = extractFirstValue(filterValue)
            }
            
            if (filterKey) {
              filterObj[filterKey] = filterValue
            }
            i += 2
          } else {
            // 奇数个参数，跳过最后一个
            i += 1
          }
        }
      }

      try {
        // 构建请求数据
        const requestData = {
          modelName: modelNameStr,
          returnColumns: columnList,
          filters: filterObj
        }

        // 发送同步请求到后端
        const xhr = new XMLHttpRequest()
        xhr.open('POST', '/api/sugarFormulaQuery/executeGet', false) // 同步请求
        xhr.setRequestHeader('Content-Type', 'application/json')
        
        // 添加认证头
        const userStore = useUserStore()
        if (userStore.token) {
          xhr.setRequestHeader('x-token', userStore.token)
        }
        if ((userStore.userInfo as any).ID) {
          xhr.setRequestHeader('x-user-id', (userStore.userInfo as any).ID)
        }

        xhr.send(JSON.stringify(requestData))

        if (xhr.status === 200) {
          const response = JSON.parse(xhr.responseText)
          if (response.code === 0 && response.data && response.data.results) {
            const results = response.data.results
            
            // 限制最大返回行数，避免性能问题
            const MAX_ROWS = 1000
            const limitedResults = results.slice(0, MAX_ROWS)
            
            if (results.length > MAX_ROWS) {
              console.warn(`SUGAR.GET: 数据行数 (${results.length}) 超过限制 (${MAX_ROWS})，已截取前 ${MAX_ROWS} 行`)
            }
            
            // 如果只有一列且只有一行，返回单个值
            if (columnList.length === 1 && limitedResults.length === 1) {
              const columnName = columnList[0]
              const value = limitedResults[0][columnName]
              return value !== null && value !== undefined ? value : ''
            }
            
            // 如果只有一列但多行，返回数组以支持向下填充
            if (columnList.length === 1 && limitedResults.length > 1) {
              const columnName = columnList[0]
              const values = limitedResults.map((row: any) => {
                const val = row[columnName]
                return val !== null && val !== undefined ? val : ''
              })
              
              // 返回纵向数组格式：每个值作为一行
              return values.map(val => [val])
            }
            
            // 多列情况，返回二维数组
            if (columnList.length > 1) {
              const rows = limitedResults.map((row: any) =>
                columnList.map(col => {
                  const val = row[col]
                  return val !== null && val !== undefined ? val : ''
                })
              )
              return rows
            }
            
            // 兜底情况：单列无数据
            return ''
          } else {
            return response.msg || '#ERROR!'
          }
        } else {
          return '#CONNECT!'
        }
      } catch (error) {
        return '#ERROR!'
      }

      return '#N/A'
    },
    config: {
      description: {
        functionName: 'SUGAR.GET',
        description: 'formula.functionList.SUGAR.GET.description',
        abstract: 'formula.functionList.SUGAR.GET.abstract',
        functionParameter: [
          {
            name: 'formula.functionList.SUGAR.GET.functionParameter.modelName.name',
            detail: 'formula.functionList.SUGAR.GET.functionParameter.modelName.detail',
            example: '"业务指标查询"',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.SUGAR.GET.functionParameter.returnColumns.name',
            detail: 'formula.functionList.SUGAR.GET.functionParameter.returnColumns.detail',
            example: '"城市名称"',
            require: 1,
            repeat: 0,
          },
          {
            name: 'formula.functionList.SUGAR.GET.functionParameter.filters.name',
            detail: 'formula.functionList.SUGAR.GET.functionParameter.filters.detail',
            example: '"战区名称", "华东战区"',
            require: 0,
            repeat: 1,
          },
        ],
      },
      locales: {
        zhCN: functionSugarGetZhCN,
      },
    },
    locales: functionSugarGetZhCN,
  },
];
