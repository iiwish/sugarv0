/**
 * 插件基类
 * 所有插件都应该继承此基类
 */

import type {
  IPlugin,
  PluginMetadata,
  PluginState,
  PluginContext,
  PluginHooks
} from '../types/plugin'
import { PluginHookType } from '../types/plugin'
import { LifecycleState } from '../types'

export abstract class BasePlugin implements IPlugin {
  public readonly metadata: PluginMetadata
  public readonly state: PluginState
  public readonly hooks?: PluginHooks

  protected context?: PluginContext

  constructor(metadata: PluginMetadata, hooks?: PluginHooks) {
    this.metadata = {
      ...metadata,
      config: {
        enabled: true,
        autoStart: true,
        priority: 0,
        ...metadata.config
      }
    }

    this.state = {
      id: metadata.id,
      state: LifecycleState.CREATED,
      config: this.metadata.config!
    }

    this.hooks = hooks
  }

  /**
   * 安装插件
   */
  async install(context: PluginContext): Promise<void> {
    try {
      this.setState(LifecycleState.INITIALIZING)
      this.context = context

      // 执行安装前钩子
      await this.executeHook(PluginHookType.BEFORE_INSTALL, context)

      // 执行具体的安装逻辑
      await this.onInstall(context)

      // 执行安装后钩子
      await this.executeHook(PluginHookType.AFTER_INSTALL, context)

      this.setState(LifecycleState.INITIALIZED)
    } catch (error) {
      this.setState(LifecycleState.CREATED, error as Error)
      throw error
    }
  }

  /**
   * 激活插件
   */
  async activate(context: PluginContext): Promise<void> {
    if (this.state.state !== LifecycleState.INITIALIZED) {
      throw new Error(`插件 ${this.metadata.id} 必须先安装才能激活`)
    }

    try {
      this.setState(LifecycleState.STARTING)

      // 执行激活前钩子
      await this.executeHook(PluginHookType.BEFORE_ACTIVATE, context)

      // 执行具体的激活逻辑
      await this.onActivate(context)

      // 执行激活后钩子
      await this.executeHook(PluginHookType.AFTER_ACTIVATE, context)

      this.setState(LifecycleState.STARTED)
      this.state.startTime = Date.now()
    } catch (error) {
      this.setState(LifecycleState.INITIALIZED, error as Error)
      throw error
    }
  }

  /**
   * 停用插件
   */
  async deactivate(context: PluginContext): Promise<void> {
    if (this.state.state !== LifecycleState.STARTED) {
      return // 已经停用或未激活
    }

    try {
      this.setState(LifecycleState.STOPPING)

      // 执行停用前钩子
      await this.executeHook(PluginHookType.BEFORE_DEACTIVATE, context)

      // 执行具体的停用逻辑
      await this.onDeactivate(context)

      // 执行停用后钩子
      await this.executeHook(PluginHookType.AFTER_DEACTIVATE, context)

      // 停用后回到已安装状态，而不是停止状态
      this.setState(LifecycleState.INITIALIZED)
      this.state.stopTime = Date.now()
    } catch (error) {
      this.setState(LifecycleState.STARTED, error as Error)
      throw error
    }
  }

  /**
   * 卸载插件
   */
  async uninstall(context: PluginContext): Promise<void> {
    // 如果插件正在运行，先停用
    if (this.state.state === LifecycleState.STARTED) {
      await this.deactivate(context)
    }

    if (this.state.state !== LifecycleState.STOPPED && 
        this.state.state !== LifecycleState.INITIALIZED) {
      return // 已经卸载或未安装
    }

    try {
      // 执行卸载前钩子
      await this.executeHook(PluginHookType.BEFORE_UNINSTALL, context)

      // 执行具体的卸载逻辑
      await this.onUninstall(context)

      // 执行卸载后钩子
      await this.executeHook(PluginHookType.AFTER_UNINSTALL, context)

      this.setState(LifecycleState.DESTROYED)
      this.context = undefined
    } catch (error) {
      this.setState(this.state.state, error as Error)
      throw error
    }
  }

  /**
   * 检查插件是否已安装
   */
  isInstalled(): boolean {
    return this.state.state !== LifecycleState.CREATED && 
           this.state.state !== LifecycleState.DESTROYED
  }

  /**
   * 检查插件是否已激活
   */
  isActivated(): boolean {
    return this.state.state === LifecycleState.STARTED
  }

  /**
   * 获取插件配置
   */
  getConfig<T = any>(key?: string): T {
    if (key) {
      return this.state.config.settings?.[key]
    }
    return this.state.config as T
  }

  /**
   * 更新插件配置
   */
  updateConfig(config: Partial<typeof this.state.config>): void {
    Object.assign(this.state.config, config)
  }

  /**
   * 抽象方法：子类需要实现的安装逻辑
   */
  protected abstract onInstall(context: PluginContext): Promise<void>

  /**
   * 抽象方法：子类需要实现的激活逻辑
   */
  protected abstract onActivate(context: PluginContext): Promise<void>

  /**
   * 抽象方法：子类需要实现的停用逻辑
   */
  protected abstract onDeactivate(context: PluginContext): Promise<void>

  /**
   * 抽象方法：子类需要实现的卸载逻辑
   */
  protected abstract onUninstall(context: PluginContext): Promise<void>

  /**
   * 设置插件状态
   */
  private setState(state: LifecycleState, error?: Error): void {
    this.state.state = state
    this.state.error = error
  }

  /**
   * 执行钩子函数
   */
  private async executeHook(hookType: PluginHookType, context: PluginContext): Promise<void> {
    const hook = this.hooks?.[hookType]
    if (hook) {
      try {
        await hook(context)
      } catch (error) {
        console.error(`插件 ${this.metadata.id} 的钩子 ${hookType} 执行失败:`, error)
        throw error
      }
    }
  }
}