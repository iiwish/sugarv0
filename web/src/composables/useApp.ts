import { onMounted, onBeforeUnmount, onActivated, onDeactivated, ref } from 'vue'
import { usePluginManager } from './usePluginManager'
import { useWorkbookManager } from './useWorkbookManager'
import type { UniverContainerReadyEvent } from '@/core/types/events'
import { functionLmdiZhCN } from '@/plugins/custom-formulas/formulas/financial'
import type { UniverCorePlugin } from '@/plugins/univer-core'

/**
 * 应用级组合式函数
 * 负责整个应用的启动、运行和关闭
 */
export function useApp() {
  const pluginManager = usePluginManager()
  const workbookManager = useWorkbookManager(pluginManager)
  const isInitialized = ref(false)
  const isActive = ref(false)

  /**
   * 运行应用
   */
  async function run() {
    try {
      console.log('启动应用...')
      
      // 如果已经初始化过，直接返回
      if (isInitialized.value) {
        return
      }

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
      isInitialized.value = true
      isActive.value = true
      console.log('应用运行成功')
    } catch (error) {
      console.error('应用启动失败:', error)
    }
  }

  // 以下函数不再需要，因为我们采用了更优化的容器分离/附加策略
  // async function reactivatePlugins() { ... } 注释掉
  // async function pause() { ... } 注释掉

  /**
   * 关闭应用（真正的卸载）
   */
  async function shutdown() {
    try {
      console.log('关闭应用...')
      // 卸载工作簿
      await workbookManager.unloadWorkbook()
      // 停用所有插件
      await pluginManager.deactivateAllPlugins()
      isInitialized.value = false
      isActive.value = false
      console.log('应用关闭成功')
    } catch (error) {
      console.error('应用关闭失败:', error)
    }
  }

  // 使用生命周期钩子自动管理应用的运行和关闭
  onMounted(run)
  onBeforeUnmount(shutdown)
  
  // 使用keep-alive的激活/停用钩子来管理插件状态
  onActivated(async () => {
    // 当组件被激活时，重新附加渲染容器
    if (isInitialized.value && !isActive.value) {
      console.log('应用被激活，重新附加Univer容器...')
      const univerCorePlugin = pluginManager.getPlugin('univer-core') as UniverCorePlugin
      if (univerCorePlugin) {
        univerCorePlugin.reattachContainer('univer-sheet-container');
      }
      isActive.value = true;
      console.log('Univer容器已重新附加');
    }
  })
  
  onDeactivated(async () => {
    // 当组件被停用时，分离渲染容器，但保持核心服务运行
    if (isActive.value) {
      console.log('应用被停用，分离Univer容器...')
      const univerCorePlugin = pluginManager.getPlugin('univer-core') as UniverCorePlugin
      if (univerCorePlugin) {
        univerCorePlugin.detachContainer();
      }
      isActive.value = false;
      console.log('Univer容器已分离');
    }
  })

  return {
    run,
    shutdown,
    pluginManager,
    workbookManager,
    isInitialized,
    isActive
  }
}