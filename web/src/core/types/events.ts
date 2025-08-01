/**
 * 事件系统相关类型定义
 */

// 事件监听器函数
export type EventListener<T = any> = (data: T) => void | Promise<void>

// 事件监听器配置
export interface EventListenerConfig {
  once?: boolean // 是否只执行一次
  priority?: number // 优先级，数字越大优先级越高
  async?: boolean // 是否异步执行
}

// 事件监听器信息
export interface EventListenerInfo<T = any> {
  id: string
  listener: EventListener<T>
  config: EventListenerConfig
  createdAt: number
}

// 事件总线接口
export interface IEventBus {
  on<T = any>(event: string, listener: EventListener<T>, config?: EventListenerConfig): string
  off(event: string, listenerId?: string): void
  emit<T = any>(event: string, data?: T): Promise<void>
  once<T = any>(event: string, listener: EventListener<T>): string
  clear(event?: string): void
  getListeners(event: string): EventListenerInfo[]
  hasListeners(event: string): boolean
}

// 系统事件类型
export enum SystemEventType {
  // 应用生命周期
  APP_INIT = 'app:init',
  APP_READY = 'app:ready',
  APP_DESTROY = 'app:destroy',
  
  // 工作簿事件
  WORKBOOK_CREATED = 'workbook:created',
  WORKBOOK_OPENED = 'workbook:opened',
  WORKBOOK_SAVED = 'workbook:saved',
  WORKBOOK_CLOSED = 'workbook:closed',
  WORKBOOK_CHANGED = 'workbook:changed',
  
  // 工作表事件
  SHEET_CREATED = 'sheet:created',
  SHEET_DELETED = 'sheet:deleted',
  SHEET_RENAMED = 'sheet:renamed',
  SHEET_ACTIVATED = 'sheet:activated',
  
  // 单元格事件
  CELL_SELECTED = 'cell:selected',
  CELL_EDITED = 'cell:edited',
  CELL_VALUE_CHANGED = 'cell:valueChanged',
  CELL_FORMULA_CHANGED = 'cell:formulaChanged',
  
  // 公式事件
  FORMULA_CALCULATED = 'formula:calculated',
  FORMULA_ERROR = 'formula:error',
  
  // UI事件
  UI_THEME_CHANGED = 'ui:themeChanged',
  UI_LANGUAGE_CHANGED = 'ui:languageChanged',
  UI_PANEL_TOGGLED = 'ui:panelToggled',
  
  // 插件事件
  PLUGIN_LOADED = 'plugin:loaded',
  PLUGIN_UNLOADED = 'plugin:unloaded',
  PLUGIN_ACTIVATED = 'plugin:activated',
  PLUGIN_DEACTIVATED = 'plugin:deactivated',
  
  // 错误事件
  ERROR_OCCURRED = 'error:occurred',
  WARNING_OCCURRED = 'warning:occurred'
}

// 事件数据基类
export interface BaseEventData {
  timestamp: number
  source?: string
  userId?: string
}

// 工作簿事件数据
export interface WorkbookEventData extends BaseEventData {
  workbookId: string
  workbookName?: string
  data?: any
}

// 工作表事件数据
export interface SheetEventData extends BaseEventData {
  workbookId: string
  sheetId: string
  sheetName?: string
  data?: any
}

// 单元格事件数据
export interface CellEventData extends BaseEventData {
  workbookId: string
  sheetId: string
  row: number
  col: number
  oldValue?: any
  newValue?: any
  formula?: string
}

// 公式事件数据
export interface FormulaEventData extends BaseEventData {
  workbookId: string
  sheetId: string
  cellId: string
  formula: string
  result?: any
  error?: any
  calculationTime?: number
}

// UI事件数据
export interface UIEventData extends BaseEventData {
  component?: string
  action?: string
  data?: any
}

// 插件事件数据
export interface PluginEventData extends BaseEventData {
  pluginId: string
  pluginName?: string
  version?: string
  data?: any
}

// 错误事件数据
export interface ErrorEventData extends BaseEventData {
  error: Error
  context?: any
  stack?: string
  level: 'error' | 'warning' | 'info'
}

// 事件过滤器
export interface EventFilter {
  source?: string
  userId?: string
  timeRange?: {
    start: number
    end: number
  }
  data?: Record<string, any>
}

// 事件历史记录
export interface EventHistory {
  id: string
  event: string
  data: any
  timestamp: number
  source?: string
  userId?: string
}

// 事件统计
export interface EventStats {
  totalEvents: number
  eventsByType: Record<string, number>
  eventsBySource: Record<string, number>
  averageProcessingTime: number
  errorRate: number
}

// 事件中间件
export type EventMiddleware = (
  event: string,
  data: any,
  next: () => Promise<void>
) => Promise<void>

// 事件配置
export interface EventConfig {
  maxListeners?: number // 单个事件最大监听器数量
  enableHistory?: boolean // 是否启用事件历史
  historySize?: number // 历史记录最大数量
  enableStats?: boolean // 是否启用统计
  middleware?: EventMiddleware[] // 中间件
}

// 具体事件类型定义
export interface BaseEvent {
  type: string
  payload: any
  timestamp: number
}

// 工作簿事件
export interface WorkbookCreatedEvent extends BaseEvent {
  type: 'workbook:created'
  payload: {
    workbook: any
    data: any
  }
}

export interface WorkbookDisposedEvent extends BaseEvent {
  type: 'workbook:disposed'
  payload: {
    workbook: any
  }
}

// 公式事件
export interface FormulaRegisteredEvent extends BaseEvent {
  type: 'formula:registered'
  payload: {
    name: string
    config: any
    category?: string
  }
}

// Univer事件
export interface UniverInitializedEvent extends BaseEvent {
  type: 'univer:initialized'
  payload: {
    univerAPI: any
  }
}

export interface UniverDisposedEvent extends BaseEvent {
  type: 'univer:disposed'
  payload: {}
}

export interface UniverContainerReadyEvent extends BaseEvent {
  type: 'univer:container-ready'
  payload: {
    containerId: string
    locales?: Record<string, any>
  }
}