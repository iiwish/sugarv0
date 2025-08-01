import { ref, onUnmounted } from 'vue'
import { PluginManager } from '@/core/plugin/PluginManager'
import { EventBus } from '@/core/events/EventBus'
import { useWorkbookStore } from '@/core/store/modules/workbook'
import { useUIStore } from '@/core/store/modules/ui'
import type { PluginContext } from '@/core/types/plugin'
import type { IEventBus } from '@/core/types/events'

// 导入插件
import { univerCorePlugin } from '@/plugins/univer-core'
import { customFormulasPlugin } from '@/plugins/custom-formulas'

/**
 * 插件管理器组合式函数
 */
export function usePluginManager() {
  const pluginManager = ref<PluginManager | null>(null)
  const eventBus = ref<IEventBus | null>(null)
  const isInitialized = ref(false)

  /**
   * 初始化插件管理器
   */
  async function initialize(): Promise<void> {
    if (isInitialized.value) {
      return
    }

    try {
      // 创建事件总线实例
      eventBus.value = new EventBus()
      
      // 创建插件上下文
      const context: PluginContext = {
        app: null, // Vue应用实例，如果需要的话
        store: {
          workbook: useWorkbookStore(),
          ui: useUIStore()
        },
        eventBus: eventBus.value,
        logger: console, // 简单的日志器，可以后续替换为更完善的日志系统
        config: {
          // 全局配置
          debug: process.env.NODE_ENV === 'development',
          version: '1.0.0'
        }
      }

      // 创建插件管理器
      pluginManager.value = new PluginManager(context)
      
      isInitialized.value = true
      console.log('插件管理器初始化成功')
    } catch (error) {
      console.error('插件管理器初始化失败:', error)
      throw error
    }
  }

  /**
   * 注册核心插件
   */
  async function registerCorePlugins(): Promise<void> {
    if (!pluginManager.value) {
      throw new Error('插件管理器未初始化')
    }

    try {
      // 注册Univer核心插件
      await pluginManager.value.register(univerCorePlugin)
      await pluginManager.value.install('univer-core')
      await pluginManager.value.activate('univer-core')

      // 注册自定义公式插件
      await pluginManager.value.register(customFormulasPlugin)
      await pluginManager.value.install('custom-formulas')
      await pluginManager.value.activate('custom-formulas')

      console.log('核心插件注册完成')
    } catch (error) {
      console.error('核心插件注册失败:', error)
      throw error
    }
  }

  /**
   * 获取插件实例
   */
  function getPlugin(pluginId: string) {
    return pluginManager.value?.getPlugin(pluginId)
  }

  /**
   * 获取所有插件
   */
  function getAllPlugins() {
    return pluginManager.value?.getPlugins() || []
  }

  /**
   * 检查插件是否已激活
   */
  function isPluginActivated(pluginId: string): boolean {
    return pluginManager.value?.isActivated(pluginId) || false
  }

  /**
   * 激活插件
   */
  async function activatePlugin(pluginId: string): Promise<void> {
    if (!pluginManager.value) {
      throw new Error('插件管理器未初始化')
    }
    await pluginManager.value.activate(pluginId)
  }

  /**
   * 停用插件
   */
  async function deactivatePlugin(pluginId: string): Promise<void> {
    if (!pluginManager.value) {
      throw new Error('插件管理器未初始化')
    }
    await pluginManager.value.deactivate(pluginId)
  }

  /**
   * 停用所有插件
   */
  async function deactivateAllPlugins(): Promise<void> {
    if (!pluginManager.value) {
      return
    }

    const plugins = pluginManager.value.getPlugins()
    for (const plugin of plugins) {
      if (plugin.isActivated()) {
        try {
          await pluginManager.value.deactivate(plugin.metadata.id)
        } catch (error) {
          console.error(`停用插件 ${plugin.metadata.id} 失败:`, error)
        }
      }
    }
  }

  /**
   * 获取事件总线实例
   */
  function getEventBus(): IEventBus | null {
    return eventBus.value
  }

  // 组件卸载时清理资源
  onUnmounted(async () => {
    await deactivateAllPlugins()
  })

  return {
    // 状态
    isInitialized,
    
    // 方法
    initialize,
    registerCorePlugins,
    getPlugin,
    getAllPlugins,
    isPluginActivated,
    activatePlugin,
    deactivatePlugin,
    deactivateAllPlugins,
    getEventBus
  }
}