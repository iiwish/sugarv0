/**
 * 事件系统入口
 */

import { globalEventBus } from './EventBus'

export { EventBus, globalEventBus } from './EventBus'
export * from '../types/events'

// 便捷的事件发射器
export const emit = globalEventBus.emit.bind(globalEventBus)
export const on = globalEventBus.on.bind(globalEventBus)
export const off = globalEventBus.off.bind(globalEventBus)
export const once = globalEventBus.once.bind(globalEventBus)