import { BasePlugin } from '@/core/plugin/BasePlugin'
import type { PluginContext, PluginMetadata } from '@/core/types/plugin'
import type {
  UniverInitializedEvent,
  WorkbookCreatedEvent,
  FormulaRegisteredEvent
} from '@/core/types/events'

// 导入公式定义
import { financialFormulas } from './formulas/financial'
import { aiFormulas, aiFormulaManager } from './formulas/ai'
import { dbFormulas, databaseFormulaManager, forceRefreshDatabaseFormulas } from './formulas/db'

/**
 * 自定义公式插件
 * 统一管理所有自定义公式的注册和生命周期
 */
export class CustomFormulasPlugin extends BasePlugin {
  private univerAPI: any = null
  private registeredFormulas: Map<string, any> = new Map()
  private formulasRegistered = false

  constructor() {
    const metadata: PluginMetadata = {
      id: 'custom-formulas',
      name: '自定义公式集合',
      version: '1.0.0',
      description: '提供各种自定义公式功能，包括财务、统计、工程等领域的专业公式',
      author: 'Sugar Team',
      dependencies: [
        {
          name: 'univer-core',
          version: '1.0.0',
          optional: false
        }
      ], // 依赖Univer核心插件
      config: {
        enabled: true,
        autoStart: true,
        priority: 800, // 高优先级，但低于核心插件
        settings: {
          enabledCategories: ['financial', 'ai', 'db'],
          autoRegister: true,
        }
      }
    }
    super(metadata)
  }

  async onInstall(context: PluginContext): Promise<void> {
    // 监听Univer初始化完成事件，以获取univerAPI
    context.eventBus.on('univer:initialized', this.handleUniverInitialized.bind(this))
    // 监听工作簿创建事件，这是注册公式的正确时机
    context.eventBus.on('workbook:created', this.handleWorkbookCreated.bind(this))
  }

  async onActivate(context: PluginContext): Promise<void> {
    context.logger.info('自定义公式插件已激活')
    
    // 如果插件激活时，工作簿已创建但公式未注册，则立即注册
    // 这种情况可能在插件懒加载时发生
    if (this.univerAPI && !this.formulasRegistered) {
      this.context?.logger.info('插件激活时，检测到公式未注册，尝试补注册...')
      await this.registerAllFormulas()
    }
  }

  async onDeactivate(context: PluginContext): Promise<void> {
    // 注销所有已注册的公式
    await this.unregisterAllFormulas()
    context.logger.info('自定义公式插件已停用')
  }

  async onUninstall(context: PluginContext): Promise<void> {
    this.context?.eventBus.off('univer:initialized', this.handleUniverInitialized.bind(this))
    this.context?.eventBus.off('workbook:created', this.handleWorkbookCreated.bind(this))
    this.univerAPI = null
    this.registeredFormulas.clear()
    this.formulasRegistered = false // 重置状态
  }

  /**
   * 处理Univer初始化完成事件，仅获取API实例
   */
  private handleUniverInitialized(event: UniverInitializedEvent): void {
    this.univerAPI = event.payload.univerAPI
    this.context?.logger.info('CustomFormulasPlugin: Univer API has been set.')
  }

  /**
   * 处理工作簿创建事件，这是注册公式的正确时机
   */
  private async handleWorkbookCreated(event: WorkbookCreatedEvent): Promise<void> {
    this.context?.logger.info('接收到 workbook:created 事件，准备注册公式')
    
    if (this.formulasRegistered) {
      this.context?.logger.info('公式已经注册过，跳过本次操作。')
      return
    }

    if (!this.univerAPI) {
      this.context?.logger.warn('Univer API 尚未获取，无法注册公式。')
      return;
    }
    
    // 确保插件已激活
    if (this.isActivated()) {
      await this.registerAllFormulas()
    }
  }

  /**
   * 注册所有公式
   */
  private async registerAllFormulas(): Promise<void> {
    if (this.formulasRegistered) {
      return
    }

    if (!this.univerAPI) {
      this.context?.logger.warn('Univer API未初始化，无法注册公式')
      return
    }

    const formulaEngine = this.univerAPI.getFormula()
    if (!formulaEngine) {
      this.context?.logger.error('无法获取公式引擎')
      return
    }

    const enabledCategories = this.getConfig('enabledCategories') || ['financial', 'ai', 'db']
    
    try {
      // 注册各类公式
      if (enabledCategories.includes('financial')) {
        await this.registerFormulaCategory('financial', financialFormulas, formulaEngine)
      }
      
      if (enabledCategories.includes('ai')) {
        await this.registerFormulaCategory('ai', aiFormulas, formulaEngine)
      }
      
      if (enabledCategories.includes('db')) {
        await this.registerFormulaCategory('db', dbFormulas, formulaEngine)
      }

      this.context?.logger.info(`成功注册 ${this.registeredFormulas.size} 个自定义公式`)
      this.formulasRegistered = true // 标记为已注册
    } catch (error) {
      this.context?.logger.error('注册公式失败:', error)
      this.formulasRegistered = false // 注册失败，允许重试
      throw error
    }
  }

  /**
   * 注册特定类别的公式
   */
  private async registerFormulaCategory(
    category: string, 
    formulas: any[], 
    formulaEngine: any
  ): Promise<void> {
    for (const formula of formulas) {
      try {
        // 根据公式是否异步，调用不同的注册方法
        if (formula.config?.isAsync) {
          // 直接注册原始的异步函数
          // Univer会自动处理Promise，不需要包装器
          formulaEngine.registerAsyncFunction(
            formula.name,
            formula.implementation,
            formula.config // 传递完整的配置对象，包含参数信息
          )
        } else {
          formulaEngine.registerFunction(
            formula.name,
            formula.implementation,
            formula.config
          )
        }

        // 记录已注册的公式
        this.registeredFormulas.set(formula.name, {
          category,
          formula,
          registeredAt: Date.now()
        })

        // 触发公式注册事件
        const event: FormulaRegisteredEvent = {
          type: 'formula:registered',
          payload: { 
            name: formula.name, 
            category,
            config: formula.config 
          },
          timestamp: Date.now(),
        }
        this.context?.eventBus.emit('formula:registered', event)

        this.context?.logger.debug(`公式 ${formula.name} (${category}) 注册成功`)
      } catch (error) {
        this.context?.logger.error(`公式 ${formula.name} 注册失败:`, error)
      }
    }
  }

  /**
   * 注销所有公式
   */
  private async unregisterAllFormulas(): Promise<void> {
    if (!this.univerAPI || this.registeredFormulas.size === 0) {
      return
    }

    const formulaEngine = this.univerAPI.getFormula()
    if (!formulaEngine) {
      return
    }

    // 注销所有已注册的公式
    const formulaNames = Array.from(this.registeredFormulas.keys())
    for (const formulaName of formulaNames) {
      try {
        // 注意：Univer可能没有提供unregisterFunction方法
        // 这里只是清理我们的记录
        this.context?.logger.debug(`公式 ${formulaName} 已从记录中移除`)
      } catch (error) {
        this.context?.logger.error(`注销公式 ${formulaName} 失败:`, error)
      }
    }

    this.registeredFormulas.clear()
  }

  /**
   * 获取已注册的公式列表
   */
  getRegisteredFormulas(): Array<{name: string, category: string, registeredAt: number}> {
    return Array.from(this.registeredFormulas.entries()).map(([name, info]) => ({
      name,
      category: info.category,
      registeredAt: info.registeredAt
    }))
  }

  /**
   * 检查公式是否已注册
   */
  isFormulaRegistered(formulaName: string): boolean {
    return this.registeredFormulas.has(formulaName)
  }

  /**
   * 动态注册单个公式
   */
  async registerFormula(
    name: string,
    implementation: Function,
    config: any,
    category: string = 'custom'
  ): Promise<void> {
    if (!this.univerAPI) {
      throw new Error('Univer API未初始化')
    }

    const formulaEngine = this.univerAPI.getFormula()
    if (!formulaEngine) {
      throw new Error('无法获取公式引擎')
    }

    try {
      formulaEngine.registerFunction(name, implementation, config)
      
      this.registeredFormulas.set(name, {
        category,
        formula: { name, implementation, config },
        registeredAt: Date.now()
      })

      // 触发公式注册事件
      const event: FormulaRegisteredEvent = {
        type: 'formula:registered',
        payload: { name, category, config },
        timestamp: Date.now(),
      }
      this.context?.eventBus.emit('formula:registered', event)

      this.context?.logger.info(`动态注册公式 ${name} 成功`)
    } catch (error) {
      this.context?.logger.error(`动态注册公式 ${name} 失败:`, error)
      throw error
    }
  }

  /**
   * 刷新所有数据库公式
   * 通过强制重新计算来更新公式结果
   */
  async refreshDatabaseFormulas(): Promise<{ success: number; failed: number; errors: string[] }> {
    if (!this.univerAPI) {
      throw new Error('Univer API未初始化')
    }

    const result = {
      success: 0,
      failed: 0,
      errors: [] as string[]
    }

    try {
      this.context?.logger.info('开始刷新数据库公式...')

      // 步骤1: 清空数据库公式缓存
      try {
        forceRefreshDatabaseFormulas()
        result.success++
        this.context?.logger.info('数据库公式缓存已清空')
      } catch (error) {
        result.failed++
        const errorMsg = `清空缓存失败: ${error}`
        result.errors.push(errorMsg)
        this.context?.logger.warn(errorMsg)
      }

      // 步骤2: 尝试通过公式引擎进行全局重新计算
      const formulaEngine = this.univerAPI.getFormula()
      if (formulaEngine) {
        try {
          // 尝试多种重新计算方法
          if (formulaEngine.calculate) {
            formulaEngine.calculate()
            result.success++
            this.context?.logger.info('通过公式引擎重新计算成功')
          } else if (formulaEngine.recalculate) {
            formulaEngine.recalculate()
            result.success++
            this.context?.logger.info('通过公式引擎重新计算成功')
          } else if (formulaEngine.refresh) {
            formulaEngine.refresh()
            result.success++
            this.context?.logger.info('通过公式引擎刷新成功')
          } else {
            this.context?.logger.warn('公式引擎没有可用的重新计算方法')
          }
        } catch (error) {
          result.failed++
          const errorMsg = `公式引擎重新计算失败: ${error}`
          result.errors.push(errorMsg)
          this.context?.logger.warn(errorMsg)
        }
      }

      // 步骤3: 尝试通过事件总线触发全局重新计算
      try {
        this.context?.eventBus.emit('formulas:refresh-requested', {
          type: 'formulas:refresh-requested',
          payload: { source: 'custom-formulas-plugin' },
          timestamp: Date.now()
        })
        result.success++
        this.context?.logger.info('通过事件总线触发重新计算')
      } catch (error) {
        result.failed++
        const errorMsg = `事件总线触发失败: ${error}`
        result.errors.push(errorMsg)
        this.context?.logger.warn(errorMsg)
      }

      // 步骤4: 尝试通过univerAPI的其他重新计算方法
      try {
        if (this.univerAPI.recalculate) {
          this.univerAPI.recalculate()
          result.success++
          this.context?.logger.info('通过univerAPI重新计算成功')
        } else if (this.univerAPI.calculate) {
          this.univerAPI.calculate()
          result.success++
          this.context?.logger.info('通过univerAPI计算成功')
        }
      } catch (error) {
        result.failed++
        const errorMsg = `univerAPI重新计算失败: ${error}`
        result.errors.push(errorMsg)
        this.context?.logger.warn(errorMsg)
      }

      this.context?.logger.info(`数据库公式刷新完成: 成功 ${result.success}, 失败 ${result.failed}`)
      return result

    } catch (error) {
      result.failed++
      const errorMsg = `刷新数据库公式时发生错误: ${error}`
      result.errors.push(errorMsg)
      this.context?.logger.error(errorMsg)
      return result
    }
  }

  /**
   * 获取数据库公式的统计信息
   */
  getDatabaseFormulaStats(): { total: number; byType: Record<string, number> } {
    const dbFormulas = Array.from(this.registeredFormulas.entries())
      .filter(([name, info]) => info.category === 'db')
    
    const stats = {
      total: dbFormulas.length,
      byType: {} as Record<string, number>
    }

    dbFormulas.forEach(([name]) => {
      stats.byType[name] = (stats.byType[name] || 0) + 1
    })

    return stats
  }

  /**
   * 获取所有公式管理器的并发状态
   */
  getConcurrencyStatus(): {
    ai: { current: number; max: number; queued: number; cacheSize: number }
    database: { current: number; max: number; queued: number; cacheSize: number }
  } {
    return {
      ai: {
        ...aiFormulaManager.getConcurrencyStatus(),
        cacheSize: aiFormulaManager.getCacheStats().size
      },
      database: {
        ...databaseFormulaManager.getConcurrencyStatus(),
        cacheSize: databaseFormulaManager.getCacheStats().size
      }
    }
  }

  /**
   * 清理所有公式缓存
   */
  clearAllFormulaCache(): void {
    aiFormulaManager.clearCache()
    databaseFormulaManager.clearCache()
    forceRefreshDatabaseFormulas()
    this.context?.logger.info('所有公式缓存已清理')
  }

  /**
   * 监控并发计算状态
   */
  startConcurrencyMonitoring(intervalMs: number = 5000): () => void {
    const monitoringInterval = setInterval(() => {
      const status = this.getConcurrencyStatus()
      
      // 记录并发状态
      this.context?.logger.debug('公式并发状态:', {
        ai: `${status.ai.current}/${status.ai.max} (队列: ${status.ai.queued}, 缓存: ${status.ai.cacheSize})`,
        database: `${status.database.current}/${status.database.max} (队列: ${status.database.queued}, 缓存: ${status.database.cacheSize})`
      })

      // 检查是否有阻塞情况
      if (status.ai.queued > 10 || status.database.queued > 10) {
        this.context?.logger.warn('检测到公式计算队列积压:', status)
      }
    }, intervalMs)

    // 返回停止监控的函数
    return () => {
      clearInterval(monitoringInterval)
      this.context?.logger.info('公式并发监控已停止')
    }
  }
}

// 导出插件实例
export const customFormulasPlugin = new CustomFormulasPlugin()

// Vue插件安装函数
export default {
  install(app: any) {
    app.provide('customFormulasPlugin', customFormulasPlugin)
  },
}