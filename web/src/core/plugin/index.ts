/**
 * 插件系统入口
 */

import { PluginManager } from './PluginManager'

export { BasePlugin } from './BasePlugin'
export { PluginManager } from './PluginManager'
export * from '../types/plugin'

// 创建全局插件管理器实例
import type { PluginContext } from '../types/plugin'

let globalPluginManager: PluginManager | null = null

/**
 * 初始化全局插件管理器
 */
export function initializePluginManager(context: PluginContext): PluginManager {
  if (globalPluginManager) {
    console.warn('插件管理器已经初始化')
    return globalPluginManager
  }
  
  globalPluginManager = new PluginManager(context)
  return globalPluginManager
}

/**
 * 获取全局插件管理器
 */
export function getPluginManager(): PluginManager {
  if (!globalPluginManager) {
    throw new Error('插件管理器未初始化，请先调用 initializePluginManager')
  }
  return globalPluginManager
}

/**
 * 销毁全局插件管理器
 */
export async function destroyPluginManager(): Promise<void> {
  if (globalPluginManager) {
    // 停用所有插件
    await globalPluginManager.deactivateAll()
    globalPluginManager = null
  }
}