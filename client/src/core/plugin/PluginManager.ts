/**
 * 插件管理器
 * 负责插件的注册、安装、激活、停用和卸载
 */

import type {
  IPlugin,
  IPluginManager,
  PluginContext,
  PluginState,
  PluginEvent
} from '../types/plugin'
import { PluginEventType } from '../types/plugin'
import { LifecycleState } from '../types'

export class PluginManager implements IPluginManager {
  private plugins: Map<string, IPlugin> = new Map()
  private context: PluginContext

  constructor(context: PluginContext) {
    this.context = context
  }

  /**
   * 注册插件
   */
  async register(plugin: IPlugin): Promise<void> {
    const pluginId = plugin.metadata.id

    if (this.plugins.has(pluginId)) {
      throw new Error(`插件 ${pluginId} 已经注册`)
    }

    this.plugins.set(pluginId, plugin)
    
    // 触发插件注册事件
    this.emitPluginEvent(PluginEventType.REGISTERED, pluginId, plugin)
    
    this.context.logger.info(`插件 ${pluginId} 注册成功`)
  }

  /**
   * 注销插件
   */
  async unregister(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未找到`)
    }

    // 如果插件已安装，先卸载
    if (plugin.isInstalled()) {
      await this.uninstall(pluginId)
    }

    this.plugins.delete(pluginId)
    
    // 触发插件注销事件
    this.emitPluginEvent(PluginEventType.UNREGISTERED, pluginId, plugin)
    
    this.context.logger.info(`插件 ${pluginId} 注销成功`)
  }

  /**
   * 安装插件
   */
  async install(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未找到`)
    }

    if (plugin.isInstalled()) {
      this.context.logger.warn(`插件 ${pluginId} 已经安装`)
      return
    }

    try {
      // 检查依赖
      await this.checkDependencies(plugin)
      
      // 安装插件
      await plugin.install(this.context)
      
      // 触发插件安装事件
      this.emitPluginEvent(PluginEventType.INSTALLED, pluginId, plugin)
      
      this.context.logger.info(`插件 ${pluginId} 安装成功`)
    } catch (error) {
      this.context.logger.error(`插件 ${pluginId} 安装失败:`, error)
      throw error
    }
  }

  /**
   * 卸载插件
   */
  async uninstall(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未找到`)
    }

    if (!plugin.isInstalled()) {
      this.context.logger.warn(`插件 ${pluginId} 未安装`)
      return
    }

    try {
      // 卸载插件
      await plugin.uninstall(this.context)
      
      // 触发插件卸载事件
      this.emitPluginEvent(PluginEventType.UNINSTALLED, pluginId, plugin)
      
      this.context.logger.info(`插件 ${pluginId} 卸载成功`)
    } catch (error) {
      this.context.logger.error(`插件 ${pluginId} 卸载失败:`, error)
      throw error
    }
  }

  /**
   * 激活插件
   */
  async activate(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未找到`)
    }

    if (!plugin.isInstalled()) {
      throw new Error(`插件 ${pluginId} 必须先安装才能激活`)
    }

    if (plugin.isActivated()) {
      this.context.logger.warn(`插件 ${pluginId} 已经激活`)
      return
    }

    try {
      // 激活插件
      await plugin.activate(this.context)
      
      // 触发插件激活事件
      this.emitPluginEvent(PluginEventType.ACTIVATED, pluginId, plugin)
      
      this.context.logger.info(`插件 ${pluginId} 激活成功`)
    } catch (error) {
      this.context.logger.error(`插件 ${pluginId} 激活失败:`, error)
      throw error
    }
  }

  /**
   * 停用插件
   */
  async deactivate(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未找到`)
    }

    if (!plugin.isActivated()) {
      this.context.logger.warn(`插件 ${pluginId} 未激活`)
      return
    }

    try {
      // 停用插件
      await plugin.deactivate(this.context)
      
      // 触发插件停用事件
      this.emitPluginEvent(PluginEventType.DEACTIVATED, pluginId, plugin)
      
      this.context.logger.info(`插件 ${pluginId} 停用成功`)
    } catch (error) {
      this.context.logger.error(`插件 ${pluginId} 停用失败:`, error)
      throw error
    }
  }

  /**
   * 获取插件实例
   */
  getPlugin(pluginId: string): IPlugin | undefined {
    return this.plugins.get(pluginId)
  }

  /**
   * 获取所有插件
   */
  getPlugins(): IPlugin[] {
    return Array.from(this.plugins.values())
  }

  /**
   * 获取插件状态
   */
  getPluginState(pluginId: string): PluginState | undefined {
    const plugin = this.plugins.get(pluginId)
    return plugin?.state
  }

  /**
   * 检查插件是否已安装
   */
  isInstalled(pluginId: string): boolean {
    const plugin = this.plugins.get(pluginId)
    return plugin?.isInstalled() || false
  }

  /**
   * 检查插件是否已激活
   */
  isActivated(pluginId: string): boolean {
    const plugin = this.plugins.get(pluginId)
    return plugin?.isActivated() || false
  }

  /**
   * 按优先级排序插件
   */
  getPluginsByPriority(): IPlugin[] {
    return this.getPlugins().sort((a, b) => {
      const priorityA = a.metadata.config?.priority || 0
      const priorityB = b.metadata.config?.priority || 0
      return priorityB - priorityA // 高优先级在前
    })
  }

  /**
   * 获取已激活的插件
   */
  getActivatedPlugins(): IPlugin[] {
    return this.getPlugins().filter(plugin => plugin.isActivated())
  }

  /**
   * 检查插件依赖
   */
  private async checkDependencies(plugin: IPlugin): Promise<void> {
    const dependencies = plugin.metadata.dependencies || []
    
    for (const dependency of dependencies) {
      const dependencyPlugin = this.plugins.get(dependency.name)
      
      if (!dependencyPlugin) {
        if (!dependency.optional) {
          throw new Error(`插件 ${plugin.metadata.id} 依赖的插件 ${dependency.name} 未找到`)
        }
        continue
      }

      if (!dependencyPlugin.isInstalled()) {
        if (!dependency.optional) {
          throw new Error(`插件 ${plugin.metadata.id} 依赖的插件 ${dependency.name} 未安装`)
        }
      }
    }
  }

  /**
   * 触发插件事件
   */
  private emitPluginEvent(
    type: PluginEventType,
    pluginId: string,
    plugin?: IPlugin,
    error?: Error
  ): void {
    const event: PluginEvent = {
      type,
      pluginId,
      plugin,
      state: plugin?.state,
      error,
      timestamp: Date.now()
    }

    this.context.eventBus.emit(type, event)
  }

  /**
   * 获取调试信息
   */
  getDebugInfo(): Record<string, any> {
    const info: Record<string, any> = {
      totalPlugins: this.plugins.size,
      installedPlugins: 0,
      activatedPlugins: 0,
      plugins: {}
    }

    this.plugins.forEach((plugin, id) => {
      if (plugin.isInstalled()) {
        info.installedPlugins++
      }
      if (plugin.isActivated()) {
        info.activatedPlugins++
      }
      
      info.plugins[id] = {
        state: plugin.state.state,
        version: plugin.metadata.version,
        priority: plugin.metadata.config?.priority || 0
      }
    })

    return info
  }
}