import { ref, reactive, computed } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import type { AppConfig } from '@/core/types'
import { LifecycleState } from '@/core/types'

/**
 * 应用管理组合式函数
 * 提供应用级别的状态管理和配置
 */
export function useApp() {
  // 应用状态
  const appState = ref<LifecycleState>(LifecycleState.INITIALIZING)
  
  // 应用配置
  const appConfig = reactive<AppConfig>({
    name: 'Sugar Analytics',
    version: '1.0.0',
    environment: 'development',
    debug: true,
    features: {
      darkMode: true,
      notifications: true,
      analytics: false,
      autoSave: true
    },
    api: {
      baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
      timeout: 30000,
      retryAttempts: 3
    },
    ui: {
      theme: 'light',
      language: 'zh-CN',
      pageSize: 20,
      animationDuration: 300
    }
  })

  // 加载状态
  const isLoading = ref(false)
  
  // 错误状态
  const error = ref<string | null>(null)
  
  // 通知列表
  const notifications = ref<Array<{
    id: string
    type: 'success' | 'warning' | 'error' | 'info'
    title: string
    message: string
    timestamp: Date
    read: boolean
  }>>([])

  // 计算属性
  const isDevelopment = computed(() => appConfig.environment === 'development')
  const isProduction = computed(() => appConfig.environment === 'production')
  const isDarkMode = computed(() => appConfig.ui.theme === 'dark')
  const unreadNotifications = computed(() => notifications.value.filter(n => !n.read))

  /**
   * 初始化应用
   */
  const initialize = async () => {
    try {
      appState.value = LifecycleState.INITIALIZING
      isLoading.value = true
      
      // 加载配置
      await loadConfig()
      
      // 初始化主题
      initializeTheme()
      
      // 加载通知
      loadNotifications()
      
      appState.value = LifecycleState.STARTED
      
      ElMessage.success('应用初始化完成')
    } catch (err) {
      appState.value = LifecycleState.STOPPED
      error.value = err instanceof Error ? err.message : '应用初始化失败'
      ElMessage.error(error.value)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 加载应用配置
   */
  const loadConfig = async () => {
    try {
      // 从本地存储加载配置
      const savedConfig = localStorage.getItem('app_config')
      if (savedConfig) {
        const parsed = JSON.parse(savedConfig)
        Object.assign(appConfig, parsed)
      }
      
      // 从环境变量覆盖配置
      if (import.meta.env.VITE_APP_NAME) {
        appConfig.name = import.meta.env.VITE_APP_NAME
      }
      
      if (import.meta.env.VITE_APP_VERSION) {
        appConfig.version = import.meta.env.VITE_APP_VERSION
      }
      
      if (import.meta.env.MODE) {
        appConfig.environment = import.meta.env.MODE as 'development' | 'production'
        appConfig.debug = import.meta.env.MODE === 'development'
      }
    } catch (err) {
      console.warn('加载应用配置失败:', err)
    }
  }

  /**
   * 保存应用配置
   */
  const saveConfig = () => {
    try {
      localStorage.setItem('app_config', JSON.stringify(appConfig))
      ElMessage.success('配置已保存')
    } catch (err) {
      console.error('保存应用配置失败:', err)
      ElMessage.error('保存配置失败')
    }
  }

  /**
   * 更新配置
   */
  const updateConfig = (updates: Partial<AppConfig>) => {
    Object.assign(appConfig, updates)
    saveConfig()
  }

  /**
   * 切换主题
   */
  const toggleTheme = () => {
    appConfig.ui.theme = appConfig.ui.theme === 'light' ? 'dark' : 'light'
    applyTheme()
    saveConfig()
  }

  /**
   * 应用主题
   */
  const applyTheme = () => {
    const html = document.documentElement
    if (appConfig.ui.theme === 'dark') {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }
  }

  /**
   * 初始化主题
   */
  const initializeTheme = () => {
    // 检查系统主题偏好
    if (!localStorage.getItem('app_config')) {
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      appConfig.ui.theme = prefersDark ? 'dark' : 'light'
    }
    
    applyTheme()
    
    // 监听系统主题变化
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
      if (!localStorage.getItem('app_config')) {
        appConfig.ui.theme = e.matches ? 'dark' : 'light'
        applyTheme()
      }
    })
  }

  /**
   * 添加通知
   */
  const addNotification = (notification: Omit<typeof notifications.value[0], 'id' | 'timestamp' | 'read'>) => {
    const newNotification = {
      ...notification,
      id: Date.now().toString(),
      timestamp: new Date(),
      read: false
    }
    
    notifications.value.unshift(newNotification)
    
    // 限制通知数量
    if (notifications.value.length > 50) {
      notifications.value = notifications.value.slice(0, 50)
    }
    
    saveNotifications()
    
    // 显示系统通知
    if (appConfig.features.notifications) {
      ElNotification({
        type: notification.type,
        title: notification.title,
        message: notification.message,
        duration: notification.type === 'error' ? 0 : 4500
      })
    }
  }

  /**
   * 标记通知为已读
   */
  const markNotificationAsRead = (id: string) => {
    const notification = notifications.value.find(n => n.id === id)
    if (notification) {
      notification.read = true
      saveNotifications()
    }
  }

  /**
   * 标记所有通知为已读
   */
  const markAllNotificationsAsRead = () => {
    notifications.value.forEach(n => n.read = true)
    saveNotifications()
  }

  /**
   * 删除通知
   */
  const removeNotification = (id: string) => {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index > -1) {
      notifications.value.splice(index, 1)
      saveNotifications()
    }
  }

  /**
   * 清空所有通知
   */
  const clearAllNotifications = () => {
    notifications.value = []
    saveNotifications()
    ElMessage.success('已清空所有通知')
  }

  /**
   * 保存通知到本地存储
   */
  const saveNotifications = () => {
    try {
      localStorage.setItem('app_notifications', JSON.stringify(notifications.value))
    } catch (err) {
      console.warn('保存通知失败:', err)
    }
  }

  /**
   * 从本地存储加载通知
   */
  const loadNotifications = () => {
    try {
      const saved = localStorage.getItem('app_notifications')
      if (saved) {
        const parsed = JSON.parse(saved)
        notifications.value = parsed.map((n: any) => ({
          ...n,
          timestamp: new Date(n.timestamp)
        }))
      }
    } catch (err) {
      console.warn('加载通知失败:', err)
    }
  }

  /**
   * 设置语言
   */
  const setLanguage = (language: string) => {
    appConfig.ui.language = language
    saveConfig()
    ElMessage.success(`语言已切换为 ${language}`)
  }

  /**
   * 重置应用配置
   */
  const resetConfig = () => {
    localStorage.removeItem('app_config')
    location.reload()
  }

  /**
   * 获取应用信息
   */
  const getAppInfo = () => {
    return {
      name: appConfig.name,
      version: appConfig.version,
      environment: appConfig.environment,
      buildTime: import.meta.env.VITE_BUILD_TIME || 'Unknown',
      userAgent: navigator.userAgent,
      platform: navigator.platform,
      language: navigator.language
    }
  }

  return {
    // 状态
    appState,
    appConfig,
    isLoading,
    error,
    notifications,
    
    // 计算属性
    isDevelopment,
    isProduction,
    isDarkMode,
    unreadNotifications,
    
    // 方法
    initialize,
    loadConfig,
    saveConfig,
    updateConfig,
    toggleTheme,
    addNotification,
    markNotificationAsRead,
    markAllNotificationsAsRead,
    removeNotification,
    clearAllNotifications,
    setLanguage,
    resetConfig,
    getAppInfo
  }
}