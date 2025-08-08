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
      
      // 恢复插件状态（如果有的话）
      await restorePluginStates()
      
      isInitialized.value = true
      console.log('插件管理器初始化成功')
    } catch (error) {
      console.error('插件管理器初始化失败:', error)
      throw error
    }
  }

  /**
   * 保存插件状态到sessionStorage
   */
  function savePluginStates(): void {
    if (!pluginManager.value) return
    
    const states = {
      plugins: pluginManager.value.getPlugins().map(plugin => ({
        id: plugin.metadata.id,
        isInstalled: plugin.isInstalled(),
        isActivated: plugin.isActivated(),
        state: plugin.state.state
      })),
      timestamp: Date.now()
    }
    
    sessionStorage.setItem('sugar-plugin-states', JSON.stringify(states))
  }

  /**
   * 从sessionStorage恢复插件状态
   */
  async function restorePluginStates(): Promise<void> {
    const savedStates = sessionStorage.getItem('sugar-plugin-states')
    if (!savedStates || !pluginManager.value) return
    
    try {
      const states = JSON.parse(savedStates)
      // 检查状态是否过期（超过1小时）
      if (Date.now() - states.timestamp > 3600000) {
        sessionStorage.removeItem('sugar-plugin-states')
        return
      }
      
      console.log('恢复插件状态:', states.plugins)
    } catch (error) {
      console.error('恢复插件状态失败:', error)
      sessionStorage.removeItem('sugar-plugin-states')
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
      // 检查并注册Univer核心插件
      if (!pluginManager.value.getPlugin('univer-core')) {
        await pluginManager.value.register(univerCorePlugin)
      }
      
      // 确保插件已安装和激活
      if (!pluginManager.value.isInstalled('univer-core')) {
        await pluginManager.value.install('univer-core')
      }
      if (!pluginManager.value.isActivated('univer-core')) {
        await pluginManager.value.activate('univer-core')
      }

      // 检查并注册自定义公式插件
      if (!pluginManager.value.getPlugin('custom-formulas')) {
        await pluginManager.value.register(customFormulasPlugin)
      }
      
      // 确保插件已安装和激活
      if (!pluginManager.value.isInstalled('custom-formulas')) {
        await pluginManager.value.install('custom-formulas')
      }
      if (!pluginManager.value.isActivated('custom-formulas')) {
        await pluginManager.value.activate('custom-formulas')
      }

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
    
    // 保存插件状态
    savePluginStates()
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
    getEventBus,
    savePluginStates,
    restorePluginStates
  }
}