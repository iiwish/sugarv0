/**
 * 状态管理入口
 */

import { createPinia } from 'pinia'
import type { App } from 'vue'

// 导出所有store模块
export { useWorkbookStore } from './modules/workbook'
export { useUIStore } from './modules/ui'

// 创建pinia实例
export const pinia = createPinia()

/**
 * 安装状态管理
 */
export function setupStore(app: App) {
  app.use(pinia)
}

/**
 * 重置所有store状态
 */
export function resetAllStores() {
  // 这里可以调用各个store的reset方法
  // 或者重新创建pinia实例
}