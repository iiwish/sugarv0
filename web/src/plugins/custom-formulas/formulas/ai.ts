import { useUserStore } from '@/pinia/modules/user'

/**
 * @description AI类公式
 * 提供AI相关的智能分析功能
 */

// 请求缓存和并发控制
interface CacheEntry {
  result: any
  timestamp: number
  ttl: number
}

interface PendingRequest {
  promise: Promise<any>
  timestamp: number
}

interface QueuedRequest {
  resolve: () => void
  cacheKey: string
  priority: number
  timestamp: number
}

class AIFormulaManager {
  private cache = new Map<string, CacheEntry>()
  private pendingRequests = new Map<string, PendingRequest>()
  private maxConcurrentRequests = 15 // 增加并发数量以支持更好的并发性能
  private currentRequests = 0
  private requestQueue: QueuedRequest[] = []
  private defaultCacheTTL = 5 * 60 * 1000 // 5分钟缓存
  private requestTimeout = 300000 // 5分钟超时
  private requestPriority = new Map<string, number>() // 请求优先级管理
  private queueProcessingLock = false // 防止队列处理竞态条件

  /**
   * 生成缓存键
   */
  private generateCacheKey(endpoint: string, data: any): string {
    return `${endpoint}:${JSON.stringify(data)}`
  }

  /**
   * 公开的生成缓存键方法，供外部使用
   */
  public generateCacheKeyPublic(endpoint: string, data: any): string {
    return this.generateCacheKey(endpoint, data)
  }

  /**
   * 检查缓存是否有效
   */
  private isCacheValid(entry: CacheEntry): boolean {
    return Date.now() - entry.timestamp < entry.ttl
  }

  /**
   * 清理过期缓存
   */
  private cleanExpiredCache(): void {
    const now = Date.now()
    const keysToDelete: string[] = []
    
    this.cache.forEach((entry, key) => {
      if (now - entry.timestamp >= entry.ttl) {
        keysToDelete.push(key)
      }
    })
    
    keysToDelete.forEach(key => {
      this.cache.delete(key)
    })
  }

  /**
   * 执行异步AI请求
   */
  async executeAIRequest(endpoint: string, requestData: any, cacheTTL?: number): Promise<any> {
    const cacheKey = this.generateCacheKey(endpoint, requestData)
    
    // 检查缓存
    const cachedEntry = this.cache.get(cacheKey)
    if (cachedEntry && this.isCacheValid(cachedEntry)) {
      return cachedEntry.result
    }

    // 检查是否有相同的请求正在进行
    const pendingRequest = this.pendingRequests.get(cacheKey)
    if (pendingRequest) {
      // 检查请求是否超时
      if (Date.now() - pendingRequest.timestamp < this.requestTimeout) {
        try {
          return await pendingRequest.promise
        } catch (error) {
          // 如果等待的请求失败，清理并重新尝试
          this.pendingRequests.delete(cacheKey)
          // 继续执行新请求
        }
      } else {
        // 清理超时的请求
        this.pendingRequests.delete(cacheKey)
      }
    }

    // 创建新的请求Promise，添加更好的错误处理
    const requestPromise = this.makeRequest(endpoint, requestData, cacheKey, cacheTTL || this.defaultCacheTTL)
      .catch(error => {
        // 确保即使请求失败也要清理pendingRequests
        this.pendingRequests.delete(cacheKey)
        throw error
      })
    
    // 记录正在进行的请求
    this.pendingRequests.set(cacheKey, {
      promise: requestPromise,
      timestamp: Date.now()
    })

    try {
      const result = await requestPromise
      return result
    } finally {
      // 清理完成的请求
      this.pendingRequests.delete(cacheKey)
    }
  }

  /**
   * 实际执行HTTP请求
   */
  private async makeRequest(endpoint: string, requestData: any, cacheKey: string, cacheTTL: number): Promise<any> {
    // 优化的并发控制：使用优先级队列
    if (this.currentRequests >= this.maxConcurrentRequests) {
      await new Promise<void>((resolve) => {
        // 设置请求优先级（AI公式优先级较低，避免阻塞其他公式）
        const priority = this.requestPriority.get(cacheKey) || 1
        const queuedRequest: QueuedRequest = {
          resolve,
          cacheKey,
          priority,
          timestamp: Date.now()
        }
        
        this.requestQueue.push(queuedRequest)
        
        // 按优先级排序队列（优先级高的先执行，时间早的优先）
        this.requestQueue.sort((a, b) => {
          if (a.priority !== b.priority) {
            return b.priority - a.priority // 优先级高的先执行
          }
          return a.timestamp - b.timestamp // 相同优先级按时间排序
        })
      })
    }

    this.currentRequests++

    try {
      const userStore = useUserStore()
      
      // 创建可取消的请求
      const controller = new AbortController()
      const timeoutId = setTimeout(() => controller.abort(), this.requestTimeout)
      
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          ...(userStore.token && { 'x-token': userStore.token }),
          ...((userStore.userInfo as any)?.ID && { 'x-user-id': (userStore.userInfo as any).ID })
        },
        body: JSON.stringify(requestData),
        signal: controller.signal
      })

      clearTimeout(timeoutId)

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`)
      }

      const result = await response.json()
      
      if (result.code === 0) {
        // 缓存完整的响应结果，包含 data 和 msg
        const processedResult = {
          data: result.data,
          msg: result.msg,
          code: result.code
        }
        
        this.cache.set(cacheKey, {
          result: processedResult,
          timestamp: Date.now(),
          ttl: cacheTTL
        })
        
        // 定期清理过期缓存
        if (Math.random() < 0.1) { // 10%概率清理
          this.cleanExpiredCache()
        }
        
        return processedResult
      } else {
        throw new Error(result.msg || 'API请求失败')
      }
    } catch (error) {
      if (error.name === 'TimeoutError' || error.name === 'AbortError') {
        throw new Error('请求超时或被中止')
      } else {
        throw error
      }
    } finally {
      this.currentRequests--
      
      // 清理优先级记录
      this.requestPriority.delete(cacheKey)
      
      // 处理队列中的下一个请求
      this.processNextQueuedRequest()
    }
  }

  /**
   * 处理队列中的下一个请求
   */
  private processNextQueuedRequest(): void {
    if (this.queueProcessingLock || this.requestQueue.length === 0) {
      return
    }

    this.queueProcessingLock = true
    
    try {
      const nextRequest = this.requestQueue.shift()
      if (nextRequest) {
        // 使用 setTimeout 确保异步执行，避免阻塞当前请求
        setTimeout(() => {
          nextRequest.resolve()
          this.queueProcessingLock = false
          // 递归处理下一个请求
          this.processNextQueuedRequest()
        }, 0)
      } else {
        this.queueProcessingLock = false
      }
    } catch (error) {
      this.queueProcessingLock = false
      console.error('处理队列请求时发生错误:', error)
    }
  }

  /**
   * 设置请求优先级
   */
  setRequestPriority(cacheKey: string, priority: number): void {
    this.requestPriority.set(cacheKey, priority)
  }

  /**
   * 获取当前并发状态
   */
  getConcurrencyStatus(): { current: number; max: number; queued: number } {
    return {
      current: this.currentRequests,
      max: this.maxConcurrentRequests,
      queued: this.requestQueue.length
    }
  }

  /**
   * 清理所有缓存
   */
  clearCache(): void {
    this.cache.clear()
  }

  /**
   * 获取缓存统计信息
   */
  getCacheStats(): { size: number; pendingRequests: number; currentRequests: number; queuedRequests: number } {
    return {
      size: this.cache.size,
      pendingRequests: this.pendingRequests.size,
      currentRequests: this.currentRequests,
      queuedRequests: this.requestQueue.length
    }
  }

  /**
   * 清理超时的待处理请求
   */
  private cleanupTimeoutRequests(): void {
    const now = Date.now()
    const timeoutKeys: string[] = []
    
    this.pendingRequests.forEach((request, key) => {
      if (now - request.timestamp >= this.requestTimeout) {
        timeoutKeys.push(key)
      }
    })
    
    timeoutKeys.forEach(key => {
      this.pendingRequests.delete(key)
    })
    
    if (timeoutKeys.length > 0) {
      console.warn(`清理了 ${timeoutKeys.length} 个超时的AI请求`)
    }
  }

  /**
   * 定期清理超时请求和过期缓存
   */
  startPeriodicCleanup(intervalMs: number = 60000): () => void {
    const cleanupInterval = setInterval(() => {
      this.cleanupTimeoutRequests()
      this.cleanExpiredCache()
    }, intervalMs)

    return () => {
      clearInterval(cleanupInterval)
    }
  }
}

// 全局AI公式管理器实例
const aiFormulaManager = new AIFormulaManager()

// 启动定期清理
aiFormulaManager.startPeriodicCleanup(60000) // 每分钟清理一次

/**
 * AI.FETCH 函数的中文本地化
 */
export const functionSugarAiFetchZhCN = {
  formula: {
    functionList: {
      'AI.FETCH': {
        description: '使用AI Agent获取和分析数据，根据自然语言描述智能提取相关信息并返回分析结论。',
        abstract: 'AI智能数据分析',
        links: [
          {
            title: '教学',
            url: 'https://univer.ai',
          },
        ],
        functionParameter: {
          agentName: {
            name: 'Agent名称',
            detail: '指定要使用的AI Agent名称',
          },
          description: {
            name: '分析需求',
            detail: '用自然语言描述您的数据分析需求',
          },
          dataRange: {
            name: '数据范围',
            detail: '可选的数据范围，如果提供则优先使用该范围的数据',
          },
        },
      },
    },
  },
}

/**
 * AI.EXPLAINRANGE 函数的中文本地化
 */
export const functionSugarAiExplainRangeZhCN = {
  formula: {
    functionList: {
      'AI.EXPLAINRANGE': {
        description: '使用AI分析指定数据范围，提供智能的数据洞察和解释。',
        abstract: 'AI数据范围分析',
        links: [
          {
            title: '教学',
            url: 'https://univer.ai',
          },
        ],
        functionParameter: {
          dataSource: {
            name: '数据源',
            detail: '要分析的数据范围或数组',
          },
          description: {
            name: '分析需求',
            detail: '用自然语言描述您希望AI如何分析这些数据',
          },
        },
      },
    },
  },
}

/**
 * 处理AI公式结果
 */
function processAIResult(result: any): any {
  if (!result) {
    return '#N/A'
  }

  // 处理完整的API响应结构 { code, data: { text }, msg }
  if (result.data && typeof result.data === 'object') {
    // 优先返回 data.text
    if (result.data.text) {
      return result.data.text
    }
    
    // 如果取不到 text，则显示 msg 的内容
    if (result.msg) {
      return result.msg
    }
    
    // 如果 data 中有其他文本字段，尝试返回
    if (typeof result.data === 'string') {
      return result.data
    }
  }
  
  // 兼容旧格式：直接包含 text 字段
  if (result.text) {
    return result.text
  }
  
  // 如果有错误信息，返回错误
  if (result.error) {
    return result.error
  }
  
  // 如果有 msg 字段，返回 msg
  if (result.msg) {
    return result.msg
  }
  
  // 如果结果是字符串，直接返回
  if (typeof result === 'string') {
    return result
  }
  
  return '#N/A'
}

/**
 * AI公式定义
 */
export const aiFormulas = [
  {
    name: 'AI.FETCH',
    implementation: async (agentName: any, description: any, dataRange?: any) => {
      // 参数验证和错误值检查
      if (!agentName || !description) {
        return '#VALUE!'
      }

      // 检查是否为Excel错误值
      const isExcelError = (value: any): boolean => {
        if (typeof value === 'string') {
          return /^#(NAME\?|VALUE!|REF!|DIV\/0!|NUM!|N\/A|NULL!)$/.test(value)
        }
        return false
      }

      // 如果参数包含Excel错误值，直接返回该错误
      if (isExcelError(agentName)) {
        return agentName
      }
      if (isExcelError(description)) {
        return description
      }
      if (dataRange && isExcelError(dataRange)) {
        return dataRange
      }

      // 转换参数为字符串
      const agentNameStr = String(agentName || '')
      const descriptionStr = String(description || '')
      const dataRangeStr = dataRange ? String(dataRange) : undefined

      try {
        // 构建请求数据
        const requestData: any = {
          agentName: agentNameStr,
          description: descriptionStr,
        }

        // 如果提供了数据范围，添加到请求中
        if (dataRangeStr) {
          requestData.dataRange = dataRangeStr
        }

        // 设置较高的优先级以确保AI公式能够及时执行
        const cacheKey = aiFormulaManager.generateCacheKeyPublic('/api/sugarFormulaQuery/executeAiFetch', requestData)
        aiFormulaManager.setRequestPriority(cacheKey, 5) // 设置中等优先级

        try {
          const result = await aiFormulaManager.executeAIRequest(
            '/api/sugarFormulaQuery/executeAiFetch',
            requestData
          )
          return processAIResult(result)
        } catch (error) {
          console.error('AIFETCH: 执行异常:', error)
          if (error.message.includes('超时')) {
            return '#TIMEOUT!'
          } else if (error.message.includes('中止')) {
            return '#ABORTED!'
          } else if (error.message.includes('网络')) {
            return '#NETWORK!'
          } else {
            return '#ERROR!'
          }
        }
      } catch (error) {
        console.error('AIFETCH: 同步异常:', error)
        return '#ERROR!'
      }
    },
    config: {
      isAsync: true,
      description: {
        functionName: 'AI.FETCH',
        description: '使用AI Agent获取和分析数据，根据自然语言描述智能提取相关信息并返回分析结论。',
        abstract: 'AI智能数据分析',
        functionParameter: [
          {
            name: 'Agent名称',
            detail: '指定要使用的AI Agent名称',
            example: '"数据分析助手"',
            require: 1,
            repeat: 0,
          },
          {
            name: '分析需求',
            detail: '用自然语言描述您的数据分析需求',
            example: '"帮我分析销售数据的趋势"',
            require: 1,
            repeat: 0,
          },
          {
            name: '数据范围',
            detail: '可选的数据范围，如果提供则优先使用该范围的数据',
            example: '"A1:C10"',
            require: 0,
            repeat: 0,
          },
        ],
      },
      locales: {
        zhCN: functionSugarAiFetchZhCN,
      },
    },
    locales: functionSugarAiFetchZhCN,
  },
  {
    name: 'AI.EXPLAINRANGE',
    implementation: async (dataSource: any, description: any) => {
      // 参数验证
      if (!dataSource || !description) {
        return '#VALUE!'
      }

      // 检查是否为Excel错误值
      const isExcelError = (value: any): boolean => {
        if (typeof value === 'string') {
          return /^#(NAME\?|VALUE!|REF!|DIV\/0!|NUM!|N\/A|NULL!)$/.test(value)
        }
        return false
      }

      // 如果参数包含Excel错误值，直接返回该错误
      if (isExcelError(dataSource)) {
        return dataSource
      }
      if (isExcelError(description)) {
        return description
      }

      // 转换参数
      const descriptionStr = String(description || '')
      
      // 处理数据源 - 将其转换为二维数组
      let dataArray: any[][] = []
      
      try {
        if (Array.isArray(dataSource)) {
          // 如果已经是数组，确保是二维数组格式
          if (Array.isArray(dataSource[0])) {
            dataArray = dataSource
          } else {
            // 如果是一维数组，转换为二维数组（单列）
            dataArray = dataSource.map(item => [item])
          }
        } else {
          // 如果是单个值，转换为1x1的二维数组
          dataArray = [[dataSource]]
        }

        // 构建请求数据
        const requestData = {
          dataSource: dataArray,
          description: descriptionStr,
        }

        // 设置较高的优先级以确保AI公式能够及时执行
        const cacheKey = aiFormulaManager.generateCacheKeyPublic('/api/sugarFormulaQuery/executeAiExplainRange', requestData)
        aiFormulaManager.setRequestPriority(cacheKey, 5) // 设置中等优先级

        try {
          const result = await aiFormulaManager.executeAIRequest(
            '/api/sugarFormulaQuery/executeAiExplainRange',
            requestData
          )
          return processAIResult(result)
        } catch (error) {
          console.error('AIEXPLAINRANGE: 执行异常:', error)
          if (error.message.includes('超时')) {
            return '#TIMEOUT!'
          } else if (error.message.includes('中止')) {
            return '#ABORTED!'
          } else if (error.message.includes('网络')) {
            return '#NETWORK!'
          } else {
            return '#ERROR!'
          }
        }
      } catch (error) {
        console.error('AIEXPLAINRANGE: 同步异常:', error)
        return '#ERROR!'
      }
    },
    config: {
      isAsync: true, // 标记为异步函数
      description: {
        functionName: 'AI.EXPLAINRANGE',
        description: '使用AI分析指定数据范围，提供智能的数据洞察和解释。',
        abstract: 'AI数据范围分析',
        functionParameter: [
          {
            name: '数据源',
            detail: '要分析的数据范围或数组',
            example: 'A1:C10',
            require: 1,
            repeat: 0,
          },
          {
            name: '分析需求',
            detail: '用自然语言描述您希望AI如何分析这些数据',
            example: '"分析这些数据的趋势和异常值"',
            require: 1,
            repeat: 0,
          },
        ],
      },
      locales: {
        zhCN: functionSugarAiExplainRangeZhCN,
      },
    },
    locales: functionSugarAiExplainRangeZhCN,
  },
]

// 导出AI公式管理器，供外部使用
export { aiFormulaManager }
