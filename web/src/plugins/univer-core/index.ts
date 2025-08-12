import type { App } from 'vue'
import type { IWorkbookData } from '@univerjs/core'
import { BasePlugin } from '@/core/plugin/BasePlugin'
import type { IUniverInstanceService } from '@univerjs/core';
import type { PluginContext, PluginMetadata } from '@/core/types/plugin'
import type {
  WorkbookCreatedEvent,
  WorkbookDisposedEvent,
  UniverInitializedEvent,
  UniverDisposedEvent,
  UniverContainerReadyEvent,
  FormulaRegisteredEvent
} from '@/core/types/events'
import { UniverSheetsCorePreset } from '@univerjs/preset-sheets-core'
import { createUniver, LocaleType, mergeLocales } from '@univerjs/presets'

// 引入样式
import '@univerjs/preset-sheets-core/lib/index.css'

/**
 * Univer核心插件
 * 负责初始化和管理Univer表格实例
 */
export class UniverCorePlugin extends BasePlugin {
  private univerAPI: any = null
  private workbook: any = null
  private container: string = ''
  private _renderEngine: any = null; // 新增：保存渲染引擎

  constructor() {
    const metadata: PluginMetadata = {
      id: 'univer-core',
      name: 'Univer表格核心',
      version: '1.0.0',
      description: 'Univer表格核心功能插件',
      author: 'Sugar Team',
      dependencies: [],
      config: {
        enabled: true,
        autoStart: true,
        priority: 1000, // 最高优先级
      }
    }
    super(metadata)
  }

  async onInstall(context: PluginContext): Promise<void> {
    // 监听容器准备事件
    context.eventBus.on('univer:container-ready', this.handleContainerReady.bind(this))
    
    // 监听工作簿数据变化
    context.eventBus.on('workbook:data-changed', this.handleWorkbookDataChanged.bind(this))
    
    // 监听获取当前工作簿请求
    context.eventBus.on('workbook:get-current', this.handleGetCurrentWorkbook.bind(this))
  }

  async onActivate(context: PluginContext): Promise<void> {
    context.logger.info('Univer核心插件已激活')
  }

  async onDeactivate(context: PluginContext): Promise<void> {
    // 只清理工作簿，保留Univer实例以便快速恢复
    await this.disposeWorkbook()
    context.logger.info('Univer核心插件已停用')
  }

  async onUninstall(context: PluginContext): Promise<void> {
    await this.dispose()
    this.context?.eventBus.off('univer:container-ready', this.handleContainerReady.bind(this))
    this.context?.eventBus.off('workbook:data-changed', this.handleWorkbookDataChanged.bind(this))
    this.context?.eventBus.off('workbook:get-current', this.handleGetCurrentWorkbook.bind(this))
  }

  /**
   * 初始化Univer实例
   */
  async initializeUniver(containerId: string, locales: Record<string, any> = {}): Promise<void> {
    if (this.univerAPI) {
      this.context?.logger.warn('Univer实例已存在，跳过初始化')
      return
    }

    this.container = containerId

    try {
      // 动态导入本地化文件
      const sheetsCoreZhCN = await import('@univerjs/preset-sheets-core/lib/locales/zh-CN')
      
      // 合并本地化配置
      const mergedLocales = {
        [LocaleType.ZH_CN]: mergeLocales(
          sheetsCoreZhCN.default || sheetsCoreZhCN,
          ...Object.values(locales)
        ),
      }

      // 创建Univer实例
      const { univerAPI, renderEngine } = createUniver({
        locale: LocaleType.ZH_CN,
        locales: mergedLocales,
        presets: [
          UniverSheetsCorePreset({
            container: containerId,
          }),
        ],
      })
      this._renderEngine = renderEngine; // 保存渲染引擎

      this.univerAPI = univerAPI
      
      // 触发初始化完成事件
      const event: UniverInitializedEvent = {
        type: 'univer:initialized',
        payload: { univerAPI },
        timestamp: Date.now(),
      }
      this.context?.eventBus.emit('univer:initialized', event)
      
      this.context?.logger.info('Univer实例初始化成功')
    } catch (error) {
      this.context?.logger.error('Univer实例初始化失败:', error)
      throw error
    }
  }

  /**
   * 创建工作簿
   */
  async createWorkbook(data: IWorkbookData): Promise<any> {
    if (!this.univerAPI) {
      throw new Error('Univer实例未初始化')
    }

    try {
      // 如果工作簿已存在，先销毁它
      if (this.workbook) {
        await this.disposeWorkbook()
      }

      // 检查是否已存在相同ID的工作簿，如果存在则先销毁
      const existingWorkbook = this.univerAPI.getUniverSheet(data.id)
      if (existingWorkbook) {
        this.univerAPI.disposeUnit(data.id)
        this.context?.logger.info(`已销毁存在的工作簿: ${data.id}`)
      }

      this.workbook = this.univerAPI.createWorkbook(data)
      
      // 触发工作簿创建事件
      const event: WorkbookCreatedEvent = {
        type: 'workbook:created',
        payload: {
          workbook: this.workbook,
          data,
        },
        timestamp: Date.now(),
      }
      
      this.context?.eventBus.emit('workbook:created', event)
      
      this.context?.logger.info('工作簿创建成功')
      return this.workbook
    } catch (error) {
      this.context?.logger.error('工作簿创建失败:', error)
      throw error
    }
  }

  /**
   * 获取Univer API实例
   */
  getUniverAPI(): any {
    return this.univerAPI
  }

  /**
   * 获取当前工作簿
   */
  getCurrentWorkbook(): any {
    return this.workbook
  }

  /**
   * 获取公式引擎
   */
  getFormulaEngine(): any {
    return this.univerAPI?.getFormula()
  }

  /**
   * 注册自定义公式
   */
  registerFormula(name: string, implementation: Function, config: any): void {
    const formulaEngine = this.getFormulaEngine()
    if (!formulaEngine) {
      throw new Error('公式引擎未初始化')
    }

    try {
      formulaEngine.registerFunction(name, implementation, config)
      this.context?.logger.info(`自定义公式 ${name} 注册成功`)
      
      // 触发公式注册事件
      const event: FormulaRegisteredEvent = {
        type: 'formula:registered',
        payload: { name, config },
        timestamp: Date.now(),
      }
      this.context?.eventBus.emit('formula:registered', event)
    } catch (error) {
      this.context?.logger.error(`自定义公式 ${name} 注册失败:`, error)
      throw error
    }
  }

  /**
   * 销毁Univer实例
   */
  private async dispose(): Promise<void> {
    await this.disposeWorkbook()

    if (this.univerAPI) {
      this.univerAPI.dispose()
      this.univerAPI = null
      
      // 触发Univer销毁事件
      const event: UniverDisposedEvent = {
        type: 'univer:disposed',
        payload: {},
        timestamp: Date.now(),
      }
      this.context?.eventBus.emit('univer:disposed', event)
    }
  }

  /**
   * 只销毁工作簿，保留Univer实例
   */
  private async disposeWorkbook(): Promise<void> {
    if (this.workbook) {
      // 触发工作簿销毁事件
      const event: WorkbookDisposedEvent = {
        type: 'workbook:disposed',
        payload: { workbook: this.workbook },
        timestamp: Date.now(),
      }
      
      this.context?.eventBus.emit('workbook:disposed', event)
      this.workbook = null
    }
  }

  /**
   * 处理容器准备事件
   */
  private async handleContainerReady(event: UniverContainerReadyEvent): Promise<void> {
    const { containerId, locales } = event.payload
    await this.initializeUniver(containerId, locales)
  }

  /**
   * 处理工作簿数据变化事件
   */
  private handleWorkbookDataChanged(event: any): void {
    // 可以在这里处理工作簿数据变化的逻辑
    this.context?.logger.debug('工作簿数据已变化:', event.payload)
  }

  /**
   * 处理获取当前工作簿请求
   */
  private handleGetCurrentWorkbook(event: any): void {
    this.context?.logger.debug('收到获取当前工作簿请求:', event.payload)
    
    // 响应当前工作簿
    const responseEvent = {
      type: 'workbook:current-response',
      payload: {
        requestId: event.payload?.requestId,
        workbook: this.workbook
      },
      timestamp: Date.now()
    }
    
    this.context?.eventBus.emit('workbook:current-response', responseEvent)
  }

  /**
   * 分离渲染容器
   */
  public detachContainer(): void {
    if (this._renderEngine) {
      this._renderEngine.setContainer(null);
      this.context?.logger.info('Univer 渲染容器已分离');
    }
  }

  /**
   * 重新附加渲染容器
   */
  public reattachContainer(containerId: string): void {
    if (this._renderEngine) {
      const container = document.getElementById(containerId);
      if (container) {
        this._renderEngine.setContainer(container);
        this.context?.logger.info('Univer 渲染容器已重新附加');
      } else {
        this.context?.logger.error(`无法找到容器 #${containerId}`);
      }
    }
  }
}

// 导出插件实例
export const univerCorePlugin = new UniverCorePlugin()

// Vue插件安装函数
export default {
  install(app: App) {
    // 可以在这里注册全局组件或提供依赖注入
    app.provide('univerCorePlugin', univerCorePlugin)
  },
}