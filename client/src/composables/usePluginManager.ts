import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { PluginManager } from '@/core/plugin/PluginManager'
import { EventBus } from '@/core/events/EventBus'
import { LifecycleState } from '@/core/types'
import type { IPlugin, PluginConfig, PluginState, PluginContext } from '@/core/types/plugin'
import type { IEventBus } from '@/core/types/events'

/**
 * 插件管理组合式函数
 * 提供插件系统的状态管理和操作方法
 */
export function usePluginManager() {
  // 插件管理器实例
  const pluginManager = ref<PluginManager | null>(null)
  
  // 事件总线实例
  const eventBus = ref<IEventBus | null>(null)
  
  // 已注册的插件列表
  const registeredPlugins = ref<Map<string, IPlugin>>(new Map())
  
  // 插件状态映射
  const pluginStates = reactive<Map<string, PluginState>>(new Map())
  
  // 插件配置
  const pluginConfigs = reactive<Map<string, PluginConfig>>(new Map())
  
  // 加载状态
  const isLoading = ref(false)
  
  // 错误信息
  const error = ref<string | null>(null)

  // 计算属性
  const activePlugins = computed(() => {
    return Array.from(registeredPlugins.value.values()).filter(plugin => 
      pluginStates.get(plugin.metadata.id)?.state === LifecycleState.STARTED
    )
  })

  const inactivePlugins = computed(() => {
    return Array.from(registeredPlugins.value.values()).filter(plugin => {
      const state = pluginStates.get(plugin.metadata.id)?.state
      return state === LifecycleState.STOPPED || state === LifecycleState.INITIALIZED
    })
  })

  const errorPlugins = computed(() => {
    return Array.from(registeredPlugins.value.values()).filter(plugin => 
      pluginStates.get(plugin.metadata.id)?.error
    )
  })

  const pluginCount = computed(() => registeredPlugins.value.size)

  /**
   * 初始化插件管理器
   */
  const initialize = async () => {
    try {
      isLoading.value = true
      error.value = null

      // 创建事件总线
      eventBus.value = new EventBus()
      
      // 创建插件上下文
      const context: PluginContext = {
        app: null, // 将在实际使用时设置
        store: null, // 将在实际使用时设置
        eventBus: eventBus.value,
        logger: console, // 简单的日志实现
        config: {}
      }
      
      // 创建插件管理器
      pluginManager.value = new PluginManager(context)
      
      // 监听插件事件
      setupEventListeners()
      
      // 加载插件配置
      await loadPluginConfigs()
      
      ElMessage.success('插件管理器初始化完成')
    } catch (err) {
      error.value = err instanceof Error ? err.message : '插件管理器初始化失败'
      ElMessage.error(error.value)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 设置事件监听器
   */
  const setupEventListeners = () => {
    if (!eventBus.value) return

    // 监听插件注册事件
    eventBus.value.on('plugin:registered', (data: { plugin: IPlugin }) => {
      registeredPlugins.value.set(data.plugin.metadata.id, data.plugin)
      pluginStates.set(data.plugin.metadata.id, data.plugin.state)
    })

    // 监听插件激活事件
    eventBus.value.on('plugin:activated', (data: { pluginId: string }) => {
      const state = pluginStates.get(data.pluginId)
      if (state) {
        state.state = LifecycleState.STARTED
      }
    })

    // 监听插件停用事件
    eventBus.value.on('plugin:deactivated', (data: { pluginId: string }) => {
      const state = pluginStates.get(data.pluginId)
      if (state) {
        state.state = LifecycleState.STOPPED
      }
    })

    // 监听插件错误事件
    eventBus.value.on('plugin:error', (data: { pluginId: string, error: Error }) => {
      const state = pluginStates.get(data.pluginId)
      if (state) {
        state.error = data.error
      }
      ElMessage.error(`插件 ${data.pluginId} 发生错误: ${data.error.message}`)
    })
  }

  /**
   * 注册插件
   */
  const registerPlugin = async (plugin: IPlugin) => {
    try {
      if (!pluginManager.value) {
        throw new Error('插件管理器未初始化')
      }

      await pluginManager.value.register(plugin)
      ElMessage.success(`插件 ${plugin.metadata.name} 注册成功`)
    } catch (err) {
      const message = err instanceof Error ? err.message : '插件注册失败'
      ElMessage.error(message)
      throw err
    }
  }

  /**
   * 激活插件
   */
  const activatePlugin = async (pluginId: string) => {
    try {
      if (!pluginManager.value) {
        throw new Error('插件管理器未初始化')
      }

      await pluginManager.value.activate(pluginId)
      
      const plugin = registeredPlugins.value.get(pluginId)
      ElMessage.success(`插件 ${plugin?.metadata.name || pluginId} 已激活`)
    } catch (err) {
      const message = err instanceof Error ? err.message : '插件激活失败'
      ElMessage.error(message)
      throw err
    }
  }

  /**
   * 停用插件
   */
  const deactivatePlugin = async (pluginId: string) => {
    try {
      if (!pluginManager.value) {
        throw new Error('插件管理器未初始化')
      }

      await pluginManager.value.deactivate(pluginId)
      
      const plugin = registeredPlugins.value.get(pluginId)
      ElMessage.success(`插件 ${plugin?.metadata.name || pluginId} 已停用`)
    } catch (err) {
      const message = err instanceof Error ? err.message : '插件停用失败'
      ElMessage.error(message)
      throw err
    }
  }

  /**
   * 卸载插件
   */
  const uninstallPlugin = async (pluginId: string) => {
    try {
      if (!pluginManager.value) {
        throw new Error('插件管理器未初始化')
      }

      const plugin = registeredPlugins.value.get(pluginId)
      await pluginManager.value.uninstall(pluginId)
      
      // 从本地状态中移除
      registeredPlugins.value.delete(pluginId)
      pluginStates.delete(pluginId)
      pluginConfigs.delete(pluginId)
      
      ElMessage.success(`插件 ${plugin?.metadata.name || pluginId} 已卸载`)
    } catch (err) {
      const message = err instanceof Error ? err.message : '插件卸载失败'
      ElMessage.error(message)
      throw err
    }
  }

  /**
   * 获取插件信息
   */
  const getPluginInfo = (pluginId: string) => {
    const plugin = registeredPlugins.value.get(pluginId)
    const state = pluginStates.get(pluginId)
    const config = pluginConfigs.get(pluginId)
    
    return {
      plugin,
      state,
      config
    }
  }

  /**
   * 更新插件配置
   */
  const updatePluginConfig = (pluginId: string, config: Partial<PluginConfig>) => {
    const existingConfig = pluginConfigs.get(pluginId) || { enabled: true }
    const newConfig = { ...existingConfig, ...config }
    
    pluginConfigs.set(pluginId, newConfig)
    
    // 更新插件实例的配置
    const plugin = registeredPlugins.value.get(pluginId)
    if (plugin) {
      plugin.updateConfig(newConfig)
    }
    
    savePluginConfigs()
    ElMessage.success('插件配置已更新')
  }

  /**
   * 切换插件启用状态
   */
  const togglePlugin = async (pluginId: string) => {
    const state = pluginStates.get(pluginId)
    
    if (state?.state === LifecycleState.STARTED) {
      await deactivatePlugin(pluginId)
    } else if (state?.state === LifecycleState.STOPPED || state?.state === LifecycleState.INITIALIZED) {
      await activatePlugin(pluginId)
    }
  }

  /**
   * 重新加载插件
   */
  const reloadPlugin = async (pluginId: string) => {
    try {
      const state = pluginStates.get(pluginId)
      
      if (state?.state === LifecycleState.STARTED) {
        await deactivatePlugin(pluginId)
        await activatePlugin(pluginId)
      }
      
      ElMessage.success('插件重新加载完成')
    } catch (err) {
      const message = err instanceof Error ? err.message : '插件重新加载失败'
      ElMessage.error(message)
      throw err
    }
  }

  /**
   * 获取所有插件列表
   */
  const getAllPlugins = () => {
    return Array.from(registeredPlugins.value.values()).map(plugin => ({
      id: plugin.metadata.id,
      name: plugin.metadata.name,
      version: plugin.metadata.version,
      description: plugin.metadata.description,
      author: plugin.metadata.author,
      state: pluginStates.get(plugin.metadata.id),
      config: pluginConfigs.get(plugin.metadata.id),
      plugin
    }))
  }

  /**
   * 搜索插件
   */
  const searchPlugins = (keyword: string) => {
    const allPlugins = getAllPlugins()
    
    return allPlugins.filter(item => 
      item.name.toLowerCase().includes(keyword.toLowerCase()) ||
      item.description?.toLowerCase().includes(keyword.toLowerCase()) ||
      item.id.toLowerCase().includes(keyword.toLowerCase())
    )
  }

  /**
   * 加载插件配置
   */
  const loadPluginConfigs = async () => {
    try {
      const saved = localStorage.getItem('plugin_configs')
      if (saved) {
        const configs = JSON.parse(saved)
        Object.entries(configs).forEach(([pluginId, config]) => {
          pluginConfigs.set(pluginId, config as PluginConfig)
        })
      }
    } catch (err) {
      console.warn('加载插件配置失败:', err)
    }
  }

  /**
   * 保存插件配置
   */
  const savePluginConfigs = () => {
    try {
      const configs = Object.fromEntries(pluginConfigs.entries())
      localStorage.setItem('plugin_configs', JSON.stringify(configs))
    } catch (err) {
      console.warn('保存插件配置失败:', err)
    }
  }

  /**
   * 清理资源
   */
  const cleanup = () => {
    if (eventBus.value) {
      eventBus.value.clear()
      eventBus.value = null
    }
    
    registeredPlugins.value.clear()
    pluginStates.clear()
    pluginConfigs.clear()
    pluginManager.value = null
  }

  return {
    // 状态
    pluginManager,
    eventBus,
    registeredPlugins,
    pluginStates,
    pluginConfigs,
    isLoading,
    error,
    
    // 计算属性
    activePlugins,
    inactivePlugins,
    errorPlugins,
    pluginCount,
    
    // 方法
    initialize,
    registerPlugin,
    activatePlugin,
    deactivatePlugin,
    uninstallPlugin,
    getPluginInfo,
    updatePluginConfig,
    togglePlugin,
    reloadPlugin,
    getAllPlugins,
    searchPlugins,
    cleanup
  }
}