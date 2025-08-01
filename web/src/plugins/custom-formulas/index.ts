import { BasePlugin } from '@/core/plugin/BasePlugin'
import type { PluginContext, PluginMetadata } from '@/core/types/plugin'
import type {
  UniverInitializedEvent,
  WorkbookCreatedEvent,
  FormulaRegisteredEvent
} from '@/core/types/events'

// 导入公式定义
import { financialFormulas } from './formulas/financial'
import { aiFormulas } from './formulas/ai'
import { dbFormulas } from './formulas/db'

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
        // 注册公式实现
        formulaEngine.registerFunction(
          formula.name,
          formula.implementation,
          formula.config
        )

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
}

// 导出插件实例
export const customFormulasPlugin = new CustomFormulasPlugin()

// Vue插件安装函数
export default {
  install(app: any) {
    app.provide('customFormulasPlugin', customFormulasPlugin)
  },
}