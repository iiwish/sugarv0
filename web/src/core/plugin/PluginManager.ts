/**
 * 插件管理器
 * 负责插件的注册、安装、激活、停用和卸载
 */

import type {
  IPlugin,
  IPluginManager,
  PluginContext,
  PluginState,
  PluginEvent,
  PluginEventTypeString
} from '../types/plugin'
import { PluginEventType } from '../types/plugin'
import { LifecycleState } from '../types'
import { EventBus } from '../events/EventBus'

export class PluginManager implements IPluginManager {
  private plugins: Map<string, IPlugin> = new Map()
  private context: PluginContext
  private eventBus: EventBus

  constructor(context: PluginContext) {
    this.context = context
    this.eventBus = new EventBus()
  }

  /**
   * 注册插件
   */
  async register(plugin: IPlugin): Promise<void> {
    const pluginId = plugin.metadata.id

    if (this.plugins.has(pluginId)) {
      throw new Error(`插件 ${pluginId} 已经注册`)
    }

    // 检查依赖
    await this.checkDependencies(plugin)

    this.plugins.set(pluginId, plugin)

    // 触发注册事件
    await this.emitPluginEvent(PluginEventType.REGISTERED, pluginId, plugin)

    console.log(`插件 ${pluginId} 注册成功`)
  }

  /**
   * 取消注册插件
   */
  async unregister(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      return
    }

    // 如果插件已安装，先卸载
    if (plugin.isInstalled()) {
      await this.uninstall(pluginId)
    }

    this.plugins.delete(pluginId)

    // 触发取消注册事件
    await this.emitPluginEvent(PluginEventType.UNREGISTERED, pluginId, plugin)

    console.log(`插件 ${pluginId} 取消注册成功`)
  }

  /**
   * 安装插件
   */
  async install(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未注册`)
    }

    if (plugin.isInstalled()) {
      console.warn(`插件 ${pluginId} 已经安装`)
      return
    }

    try {
      await plugin.install(this.context)

      // 触发安装事件
      await this.emitPluginEvent(PluginEventType.INSTALLED, pluginId, plugin)

      console.log(`插件 ${pluginId} 安装成功`)

      // 如果配置为自动启动，则激活插件
      if (plugin.getConfig('autoStart')) {
        await this.activate(pluginId)
      }
    } catch (error) {
      console.error(`插件 ${pluginId} 安装失败:`, error)
      throw error
    }
  }

  /**
   * 卸载插件
   */
  async uninstall(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未注册`)
    }

    if (!plugin.isInstalled()) {
      console.warn(`插件 ${pluginId} 未安装`)
      return
    }

    try {
      await plugin.uninstall(this.context)

      // 触发卸载事件
      await this.emitPluginEvent(PluginEventType.UNINSTALLED, pluginId, plugin)

      console.log(`插件 ${pluginId} 卸载成功`)
    } catch (error) {
      console.error(`插件 ${pluginId} 卸载失败:`, error)
      throw error
    }
  }

  /**
   * 激活插件
   */
  async activate(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未注册`)
    }

    if (plugin.isActivated()) {
      console.warn(`插件 ${pluginId} 已经激活`)
      return
    }

    // 如果插件未安装，先安装
    if (!plugin.isInstalled()) {
      console.log(`插件 ${pluginId} 未安装，正在自动安装...`)
      await this.install(pluginId)
    }

    try {
      await plugin.activate(this.context)

      // 触发激活事件
      await this.emitPluginEvent(PluginEventType.ACTIVATED, pluginId, plugin)

      console.log(`插件 ${pluginId} 激活成功`)
    } catch (error) {
      console.error(`插件 ${pluginId} 激活失败:`, error)
      throw error
    }
  }

  /**
   * 停用插件
   */
  async deactivate(pluginId: string): Promise<void> {
    const plugin = this.plugins.get(pluginId)
    if (!plugin) {
      throw new Error(`插件 ${pluginId} 未注册`)
    }

    if (!plugin.isActivated()) {
      console.warn(`插件 ${pluginId} 未激活`)
      return
    }

    try {
      await plugin.deactivate(this.context)

      // 触发停用事件
      await this.emitPluginEvent(PluginEventType.DEACTIVATED, pluginId, plugin)

      console.log(`插件 ${pluginId} 停用成功`)
    } catch (error) {
      console.error(`插件 ${pluginId} 停用失败:`, error)
      throw error
    }
  }

  /**
   * 获取插件
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
    return plugin?.isInstalled() ?? false
  }

  /**
   * 检查插件是否已激活
   */
  isActivated(pluginId: string): boolean {
    const plugin = this.plugins.get(pluginId)
    return plugin?.isActivated() ?? false
  }

  /**
   * 批量安装插件
   */
  async installAll(pluginIds?: string[]): Promise<void> {
    const targetPlugins = pluginIds || Array.from(this.plugins.keys())
    
    for (const pluginId of targetPlugins) {
      try {
        await this.install(pluginId)
      } catch (error) {
        console.error(`批量安装插件 ${pluginId} 失败:`, error)
      }
    }
  }

  /**
   * 批量激活插件
   */
  async activateAll(pluginIds?: string[]): Promise<void> {
    const targetPlugins = pluginIds || Array.from(this.plugins.keys())
    
    for (const pluginId of targetPlugins) {
      try {
        await this.activate(pluginId)
      } catch (error) {
        console.error(`批量激活插件 ${pluginId} 失败:`, error)
      }
    }
  }

  /**
   * 批量停用插件
   */
  async deactivateAll(pluginIds?: string[]): Promise<void> {
    const targetPlugins = pluginIds || Array.from(this.plugins.keys())
    
    for (const pluginId of targetPlugins) {
      try {
        await this.deactivate(pluginId)
      } catch (error) {
        console.error(`批量停用插件 ${pluginId} 失败:`, error)
      }
    }
  }

  /**
   * 获取插件统计信息
   */
  getStats(): Record<string, any> {
    const stats = {
      total: this.plugins.size,
      installed: 0,
      activated: 0,
      byState: {} as Record<string, number>
    }

    this.plugins.forEach((plugin) => {
      const state = plugin.state.state
      stats.byState[state] = (stats.byState[state] || 0) + 1

      if (plugin.isInstalled()) {
        stats.installed++
      }
      if (plugin.isActivated()) {
        stats.activated++
      }
    })

    return stats
  }

  /**
   * 监听插件事件
   */
  onPluginEvent(listener: (event: PluginEvent) => void): void {
    this.eventBus.on('plugin:*', listener)
  }

  /**
   * 检查插件依赖
   */
  private async checkDependencies(plugin: IPlugin): Promise<void> {
    const dependencies = plugin.metadata.dependencies || []
    
    for (const dep of dependencies) {
      const depPlugin = this.plugins.get(dep.name)
      if (!depPlugin) {
        throw new Error(`插件 ${plugin.metadata.id} 依赖的插件 ${dep.name} 未注册`)
      }

      // 这里可以添加版本检查逻辑
      // if (!this.isVersionCompatible(depPlugin.metadata.version, dep.version)) {
      //   throw new Error(`插件版本不兼容`)
      // }
    }
  }

  /**
   * 触发插件事件
   */
  private async emitPluginEvent(
    type: PluginEventType | PluginEventTypeString,
    pluginId: string,
    plugin?: IPlugin,
    error?: Error
  ): Promise<void> {
    const event: PluginEvent = {
      type: type as PluginEventType,
      pluginId,
      plugin,
      state: plugin?.state,
      error,
      timestamp: Date.now()
    }

    await this.eventBus.emit(type, event)
    await this.eventBus.emit('plugin:*', event)
  }
}