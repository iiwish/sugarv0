/**
 * 事件系统类型定义
 */

// 基础事件接口
export interface BaseEvent {
  type: string
  timestamp: number
  payload?: any
}

// 事件监听器类型
export type EventListener<T = any> = (event: T) => void | Promise<void>

// 事件总线接口
export interface IEventBus {
  on<T = any>(eventType: string, listener: EventListener<T>): void
  off<T = any>(eventType: string, listener: EventListener<T>): void
  once<T = any>(eventType: string, listener: EventListener<T>): void
  emit<T = any>(eventType: string, event: T): void
  clear(): void
}

// Univer 相关事件
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

// 工作簿相关事件
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

export interface WorkbookDataChangedEvent extends BaseEvent {
  type: 'workbook:data-changed'
  payload: {
    workbook: any
    changes: any
  }
}

// 公式相关事件
export interface FormulaRegisteredEvent extends BaseEvent {
  type: 'formula:registered'
  payload: {
    name: string
    category?: string
    config: any
  }
}

export interface FormulaRefreshRequestedEvent extends BaseEvent {
  type: 'formulas:refresh-requested'
  payload: {
    source: string
  }
}

// 聊天相关事件
export interface ChatMessageEvent extends BaseEvent {
  type: 'chat:message'
  payload: {
    message: string
    sender: 'user' | 'assistant'
    timestamp: number
  }
}

export interface ChatResponseEvent extends BaseEvent {
  type: 'chat:response'
  payload: {
    response: string
    messageId?: string
  }
}

// 联合类型
export type AppEvent = 
  | UniverInitializedEvent
  | UniverDisposedEvent
  | UniverContainerReadyEvent
  | WorkbookCreatedEvent
  | WorkbookDisposedEvent
  | WorkbookDataChangedEvent
  | FormulaRegisteredEvent
  | FormulaRefreshRequestedEvent
  | ChatMessageEvent
  | ChatResponseEvent