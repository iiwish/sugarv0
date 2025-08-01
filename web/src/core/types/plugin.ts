/**
 * 插件系统相关类型定义
 */

import type { LifecycleState } from './index'

// 插件基础信息
export interface PluginInfo {
  id: string
  name: string
  version: string
  description?: string
  author?: string
  homepage?: string
  keywords?: string[]
  license?: string
}

// 插件依赖
export interface PluginDependency {
  name: string
  version: string
  optional?: boolean
}

// 插件配置
export interface PluginConfig {
  enabled: boolean
  autoStart?: boolean
  priority?: number
  settings?: Record<string, any>
}

// 插件元数据
export interface PluginMetadata extends PluginInfo {
  dependencies?: PluginDependency[]
  peerDependencies?: PluginDependency[]
  config?: PluginConfig
  entry?: string
  assets?: string[]
}

// 插件状态
export interface PluginState {
  id: string
  state: LifecycleState
  error?: Error
  startTime?: number
  stopTime?: number
  config: PluginConfig
}

// 插件上下文
export interface PluginContext {
  app: any // 应用实例
  store: any // 状态管理
  eventBus: any // 事件总线
  logger: any // 日志器
  config: Record<string, any> // 全局配置
}

// 插件钩子类型
export enum PluginHookType {
  BEFORE_INSTALL = 'beforeInstall',
  AFTER_INSTALL = 'afterInstall',
  BEFORE_ACTIVATE = 'beforeActivate',
  AFTER_ACTIVATE = 'afterActivate',
  BEFORE_DEACTIVATE = 'beforeDeactivate',
  AFTER_DEACTIVATE = 'afterDeactivate',
  BEFORE_UNINSTALL = 'beforeUninstall',
  AFTER_UNINSTALL = 'afterUninstall'
}

// 插件钩子函数
export type PluginHook = (context: PluginContext) => Promise<void> | void

// 插件钩子映射
export type PluginHooks = Partial<Record<PluginHookType, PluginHook>>

// 插件接口
export interface IPlugin {
  readonly metadata: PluginMetadata
  readonly state: PluginState
  readonly hooks?: PluginHooks

  install(context: PluginContext): Promise<void>
  activate(context: PluginContext): Promise<void>
  deactivate(context: PluginContext): Promise<void>
  uninstall(context: PluginContext): Promise<void>
  isInstalled(): boolean
  isActivated(): boolean
  getConfig<T = any>(key?: string): T
  updateConfig(config: any): void
}

// 插件管理器接口
export interface IPluginManager {
  register(plugin: IPlugin): Promise<void>
  unregister(pluginId: string): Promise<void>
  install(pluginId: string): Promise<void>
  uninstall(pluginId: string): Promise<void>
  activate(pluginId: string): Promise<void>
  deactivate(pluginId: string): Promise<void>
  getPlugin(pluginId: string): IPlugin | undefined
  getPlugins(): IPlugin[]
  getPluginState(pluginId: string): PluginState | undefined
  isInstalled(pluginId: string): boolean
  isActivated(pluginId: string): boolean
}

// 插件事件类型
export enum PluginEventType {
  REGISTERED = 'plugin:registered',
  UNREGISTERED = 'plugin:unregistered',
  INSTALLED = 'plugin:installed',
  UNINSTALLED = 'plugin:uninstalled',
  ACTIVATED = 'plugin:activated',
  DEACTIVATED = 'plugin:deactivated',
  ERROR = 'plugin:error',
  STATE_CHANGED = 'plugin:stateChanged'
}

// 插件事件类型字符串
export type PluginEventTypeString =
  | 'plugin:registered'
  | 'plugin:unregistered'
  | 'plugin:installed'
  | 'plugin:uninstalled'
  | 'plugin:activated'
  | 'plugin:deactivated'
  | 'plugin:error'
  | 'plugin:stateChanged'

// 插件事件
export interface PluginEvent {
  type: PluginEventType
  pluginId: string
  plugin?: IPlugin
  state?: PluginState
  error?: Error
  timestamp: number
}

// 插件加载器接口
export interface IPluginLoader {
  load(source: string): Promise<IPlugin>
  unload(pluginId: string): Promise<void>
  supports(source: string): boolean
}

// 插件仓库接口
export interface IPluginRepository {
  search(query: string): Promise<PluginMetadata[]>
  getMetadata(pluginId: string): Promise<PluginMetadata>
  download(pluginId: string, version?: string): Promise<string>
  getVersions(pluginId: string): Promise<string[]>
}

// 面板插件特定接口
export interface IPanelPlugin extends IPlugin {
  readonly panelId: string
  readonly title: string
  readonly icon?: string
  readonly position: PanelPosition
  readonly resizable?: boolean
  readonly collapsible?: boolean
  
  createPanel(): any // 返回Vue组件
  destroyPanel(): void
}

// 面板位置
export enum PanelPosition {
  LEFT = 'left',
  RIGHT = 'right',
  TOP = 'top',
  BOTTOM = 'bottom',
  CENTER = 'center'
}

// 面板状态
export interface PanelState {
  id: string
  visible: boolean
  width?: number
  height?: number
  collapsed?: boolean
  order?: number
}