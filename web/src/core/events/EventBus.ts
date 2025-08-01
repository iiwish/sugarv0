/**
 * 事件总线实现
 * 提供发布订阅模式的事件通信机制
 */

import type {
  IEventBus,
  EventListener,
  EventListenerConfig,
  EventListenerInfo,
  EventMiddleware,
  EventConfig
} from '../types/events'

export class EventBus implements IEventBus {
  private listeners: Map<string, Map<string, EventListenerInfo>> = new Map()
  private middleware: EventMiddleware[] = []
  private config: Required<EventConfig>
  private listenerIdCounter = 0

  constructor(config: EventConfig = {}) {
    this.config = {
      maxListeners: config.maxListeners ?? 100,
      enableHistory: config.enableHistory ?? false,
      historySize: config.historySize ?? 1000,
      enableStats: config.enableStats ?? false,
      middleware: config.middleware ?? []
    }
    
    this.middleware = [...this.config.middleware]
  }

  /**
   * 注册事件监听器
   */
  on<T = any>(
    event: string,
    listener: EventListener<T>,
    config: EventListenerConfig = {}
  ): string {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Map())
    }

    const eventListeners = this.listeners.get(event)!
    
    // 检查监听器数量限制
    if (eventListeners.size >= this.config.maxListeners) {
      throw new Error(`事件 "${event}" 的监听器数量已达到上限 ${this.config.maxListeners}`)
    }

    const listenerId = this.generateListenerId()
    const listenerInfo: EventListenerInfo<T> = {
      id: listenerId,
      listener,
      config: {
        once: config.once ?? false,
        priority: config.priority ?? 0,
        async: config.async ?? false
      },
      createdAt: Date.now()
    }

    eventListeners.set(listenerId, listenerInfo)
    
    // 按优先级排序
    this.sortListeners(event)

    return listenerId
  }

  /**
   * 移除事件监听器
   */
  off(event: string, listenerId?: string): void {
    if (!this.listeners.has(event)) {
      return
    }

    const eventListeners = this.listeners.get(event)!

    if (listenerId) {
      eventListeners.delete(listenerId)
    } else {
      eventListeners.clear()
    }

    // 如果没有监听器了，删除事件
    if (eventListeners.size === 0) {
      this.listeners.delete(event)
    }
  }

  /**
   * 触发事件
   */
  async emit<T = any>(event: string, data?: T): Promise<void> {
    if (!this.listeners.has(event)) {
      return
    }

    const eventListeners = this.listeners.get(event)!
    const listenersToExecute = Array.from(eventListeners.values())

    // 执行中间件
    await this.executeMiddleware(event, data, async () => {
      // 分离同步和异步监听器
      const syncListeners: EventListenerInfo<T>[] = []
      const asyncListeners: EventListenerInfo<T>[] = []

      for (const listenerInfo of listenersToExecute) {
        if (listenerInfo.config.async) {
          asyncListeners.push(listenerInfo)
        } else {
          syncListeners.push(listenerInfo)
        }
      }

      // 先执行同步监听器
      for (const listenerInfo of syncListeners) {
        await this.executeListener(event, listenerInfo, data)
      }

      // 并行执行异步监听器
      if (asyncListeners.length > 0) {
        await Promise.all(
          asyncListeners.map(listenerInfo => 
            this.executeListener(event, listenerInfo, data)
          )
        )
      }
    })
  }

  /**
   * 注册一次性事件监听器
   */
  once<T = any>(event: string, listener: EventListener<T>): string {
    return this.on(event, listener, { once: true })
  }

  /**
   * 清除事件监听器
   */
  clear(event?: string): void {
    if (event) {
      this.listeners.delete(event)
    } else {
      this.listeners.clear()
    }
  }

  /**
   * 获取事件的所有监听器
   */
  getListeners(event: string): EventListenerInfo[] {
    const eventListeners = this.listeners.get(event)
    return eventListeners ? Array.from(eventListeners.values()) : []
  }

  /**
   * 检查事件是否有监听器
   */
  hasListeners(event: string): boolean {
    const eventListeners = this.listeners.get(event)
    return eventListeners ? eventListeners.size > 0 : false
  }

  /**
   * 添加中间件
   */
  use(middleware: EventMiddleware): void {
    this.middleware.push(middleware)
  }

  /**
   * 移除中间件
   */
  removeMiddleware(middleware: EventMiddleware): void {
    const index = this.middleware.indexOf(middleware)
    if (index > -1) {
      this.middleware.splice(index, 1)
    }
  }

  /**
   * 获取所有事件名称
   */
  getEventNames(): string[] {
    return Array.from(this.listeners.keys())
  }

  /**
   * 获取事件统计信息
   */
  getStats(): Record<string, any> {
    const stats: Record<string, any> = {
      totalEvents: this.listeners.size,
      totalListeners: 0,
      eventDetails: {}
    }

    this.listeners.forEach((listeners, event) => {
      stats.totalListeners += listeners.size
      stats.eventDetails[event] = {
        listenerCount: listeners.size,
        listeners: Array.from(listeners.values()).map((l: EventListenerInfo) => ({
          id: l.id,
          priority: l.config.priority,
          once: l.config.once,
          async: l.config.async,
          createdAt: l.createdAt
        }))
      }
    })

    return stats
  }

  /**
   * 执行监听器
   */
  private async executeListener<T>(
    event: string,
    listenerInfo: EventListenerInfo<T>,
    data?: T
  ): Promise<void> {
    try {
      await listenerInfo.listener(data)

      // 如果是一次性监听器，执行后移除
      if (listenerInfo.config.once) {
        this.off(event, listenerInfo.id)
      }
    } catch (error) {
      console.error(`事件监听器执行失败 [${event}]:`, error)
      // 可以在这里触发错误事件
      // this.emit('error:listenerFailed', { event, listenerId: listenerInfo.id, error })
    }
  }

  /**
   * 执行中间件
   */
  private async executeMiddleware<T>(
    event: string,
    data: T,
    next: () => Promise<void>
  ): Promise<void> {
    let index = 0

    const executeNext = async (): Promise<void> => {
      if (index >= this.middleware.length) {
        await next()
        return
      }

      const middleware = this.middleware[index++]
      await middleware(event, data, executeNext)
    }

    await executeNext()
  }

  /**
   * 按优先级排序监听器
   */
  private sortListeners(event: string): void {
    const eventListeners = this.listeners.get(event)
    if (!eventListeners) return

    const sortedEntries = Array.from(eventListeners.entries())
      .sort(([, a], [, b]) => (b.config.priority ?? 0) - (a.config.priority ?? 0))

    eventListeners.clear()
    for (const [id, listener] of sortedEntries) {
      eventListeners.set(id, listener)
    }
  }

  /**
   * 生成监听器ID
   */
  private generateListenerId(): string {
    return `listener_${++this.listenerIdCounter}_${Date.now()}`
  }
}

// 创建全局事件总线实例
export const globalEventBus = new EventBus({
  maxListeners: 200,
  enableHistory: true,
  enableStats: true
})