/**
 * 事件总线实现
 */

import type { IEventBus, EventListener } from '../types/events'

export class EventBus implements IEventBus {
  private listeners: Map<string, Set<EventListener>> = new Map()

  /**
   * 注册事件监听器
   */
  on<T = any>(eventType: string, listener: EventListener<T>): void {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set())
    }
    this.listeners.get(eventType)!.add(listener)
  }

  /**
   * 移除事件监听器
   */
  off<T = any>(eventType: string, listener: EventListener<T>): void {
    const eventListeners = this.listeners.get(eventType)
    if (eventListeners) {
      eventListeners.delete(listener)
      if (eventListeners.size === 0) {
        this.listeners.delete(eventType)
      }
    }
  }

  /**
   * 注册一次性事件监听器
   */
  once<T = any>(eventType: string, listener: EventListener<T>): void {
    const onceListener: EventListener<T> = (event: T) => {
      this.off(eventType, onceListener)
      listener(event)
    }
    this.on(eventType, onceListener)
  }

  /**
   * 触发事件
   */
  emit<T = any>(eventType: string, event: T): void {
    const eventListeners = this.listeners.get(eventType)
    if (eventListeners) {
      // 创建监听器副本，避免在执行过程中修改原集合
      const listenersArray = Array.from(eventListeners)
      
      for (const listener of listenersArray) {
        try {
          const result = listener(event)
          // 如果监听器返回 Promise，处理可能的错误
          if (result instanceof Promise) {
            result.catch(error => {
              console.error(`事件监听器执行失败 [${eventType}]:`, error)
            })
          }
        } catch (error) {
          console.error(`事件监听器执行失败 [${eventType}]:`, error)
        }
      }
    }
  }

  /**
   * 清除所有事件监听器
   */
  clear(): void {
    this.listeners.clear()
  }

  /**
   * 获取指定事件类型的监听器数量
   */
  getListenerCount(eventType: string): number {
    return this.listeners.get(eventType)?.size || 0
  }

  /**
   * 获取所有事件类型
   */
  getEventTypes(): string[] {
    return Array.from(this.listeners.keys())
  }

  /**
   * 检查是否有指定事件类型的监听器
   */
  hasListeners(eventType: string): boolean {
    return this.getListenerCount(eventType) > 0
  }

  /**
   * 获取调试信息
   */
  getDebugInfo(): Record<string, number> {
    const info: Record<string, number> = {}
    this.listeners.forEach((listeners, eventType) => {
      info[eventType] = listeners.size
    })
    return info
  }
}