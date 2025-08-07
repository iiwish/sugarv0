import { onMounted, onBeforeUnmount } from 'vue'
import { usePluginManager } from './usePluginManager'
import { useWorkbookManager } from './useWorkbookManager'
import type { UniverContainerReadyEvent } from '@/core/types/events'
import { functionLmdiZhCN } from '@/plugins/custom-formulas/formulas/financial'

/**
 * 应用级组合式函数
 * 负责整个应用的启动、运行和关闭
 */
export function useApp() {
  const pluginManager = usePluginManager()
  const workbookManager = useWorkbookManager(pluginManager)

  /**
   * 运行应用
   */
  async function run() {
    try {
      console.log('启动应用...')
      // 初始化插件管理器
      await pluginManager.initialize()
      // 注册核心插件
      await pluginManager.registerCorePlugins()
      // 触发容器就绪事件，这将启动Univer的渲染
      const eventBus = pluginManager.getEventBus()
      if (eventBus) {
        const containerReadyEvent: UniverContainerReadyEvent = {
          type: 'univer:container-ready',
          payload: {
            containerId: 'univer-sheet-container',
            locales: {
              zhCN: functionLmdiZhCN
            }
          },
          timestamp: Date.now()
        }
        eventBus.emit('univer:container-ready', containerReadyEvent)
      } else {
        throw new Error('事件总线未初始化，无法启动Univer')
      }

      // 加载工作簿
      await workbookManager.loadInitialWorkbook()
      console.log('应用运行成功')
    } catch (error) {
      console.error('应用启动失败:', error)
    }
  }

  /**
   * 关闭应用
   */
  async function shutdown() {
    try {
      console.log('关闭应用...')
      // 卸载工作簿
      await workbookManager.unloadWorkbook()
      // 停用所有插件
      await pluginManager.deactivateAllPlugins()
      console.log('应用关闭成功')
    } catch (error) {
      console.error('应用关闭失败:', error)
    }
  }

  // 使用生命周期钩子自动管理应用的运行和关闭
  onMounted(run)
  onBeforeUnmount(shutdown)

  return {
    run,
    shutdown,
    pluginManager,
    workbookManager,
  }
}