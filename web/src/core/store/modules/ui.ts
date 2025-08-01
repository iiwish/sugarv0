/**
 * UI状态管理
 */

import { defineStore } from 'pinia'
import type { PanelState } from '../../types/plugin'
import { PanelPosition } from '../../types/plugin'

interface UIStoreState {
  // 主题设置
  theme: {
    mode: 'light' | 'dark' | 'auto'
    primaryColor: string
    fontSize: number
  }
  // 语言设置
  locale: string
  // 面板状态
  panels: Record<string, PanelState>
  // 侧边栏状态
  sidebar: {
    collapsed: boolean
    width: number
  }
  // 工具栏状态
  toolbar: {
    visible: boolean
    items: string[]
  }
  // 状态栏
  statusBar: {
    visible: boolean
    items: Array<{
      id: string
      content: string
      position: 'left' | 'right'
    }>
  }
  // 对话框状态
  dialogs: Record<string, {
    visible: boolean
    data?: any
  }>
  // 加载状态
  loading: {
    global: boolean
    components: Record<string, boolean>
  }
  // 通知
  notifications: Array<{
    id: string
    type: 'success' | 'warning' | 'error' | 'info'
    title: string
    message: string
    duration?: number
    timestamp: number
  }>
  // 布局设置
  layout: {
    headerHeight: number
    footerHeight: number
    contentPadding: number
  }
}

export const useUIStore = defineStore('ui', {
  state: (): UIStoreState => ({
    theme: {
      mode: 'light',
      primaryColor: '#1890ff',
      fontSize: 14
    },
    locale: 'zh-CN',
    panels: {},
    sidebar: {
      collapsed: false,
      width: 240
    },
    toolbar: {
      visible: true,
      items: ['save', 'undo', 'redo', 'format']
    },
    statusBar: {
      visible: true,
      items: []
    },
    dialogs: {},
    loading: {
      global: false,
      components: {}
    },
    notifications: [],
    layout: {
      headerHeight: 60,
      footerHeight: 30,
      contentPadding: 16
    }
  }),

  getters: {
    /**
     * 获取当前主题模式
     */
    currentTheme: (state) => {
      if (state.theme.mode === 'auto') {
        // 根据系统主题自动判断
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
      }
      return state.theme.mode
    },

    /**
     * 获取可见的面板
     */
    visiblePanels: (state) => {
      return Object.entries(state.panels)
        .filter(([, panel]) => panel.visible)
        .map(([id, panel]) => ({ id, ...panel }))
    },

    /**
     * 按位置分组的面板
     */
    panelsByPosition: (state) => {
      const groups: Record<PanelPosition, Array<{ id: string } & PanelState>> = {
        left: [],
        right: [],
        top: [],
        bottom: [],
        center: []
      }

      Object.entries(state.panels).forEach(([id, panel]) => {
        if (panel.visible) {
          // 这里需要从插件元数据中获取位置信息
          // 暂时默认为 right
          const position: PanelPosition = PanelPosition.RIGHT
          groups[position].push({ id, ...panel })
        }
      })

      // 按顺序排序
      Object.values(groups).forEach(group => {
        group.sort((a, b) => (a.order || 0) - (b.order || 0))
      })

      return groups
    },

    /**
     * 获取活动的通知
     */
    activeNotifications: (state) => {
      const now = Date.now()
      return state.notifications.filter(notification => {
        if (!notification.duration) return true
        return now - notification.timestamp < notification.duration
      })
    },

    /**
     * 检查是否有全局加载状态
     */
    isGlobalLoading: (state) => state.loading.global,

    /**
     * 检查组件是否在加载
     */
    isComponentLoading: (state) => (componentId: string) => {
      return state.loading.components[componentId] || false
    }
  },

  actions: {
    /**
     * 设置主题
     */
    setTheme(theme: Partial<UIStoreState['theme']>) {
      Object.assign(this.theme, theme)
      this.applyTheme()
    },

    /**
     * 切换主题模式
     */
    toggleThemeMode() {
      const modes: Array<'light' | 'dark' | 'auto'> = ['light', 'dark', 'auto']
      const currentIndex = modes.indexOf(this.theme.mode)
      this.theme.mode = modes[(currentIndex + 1) % modes.length]
      this.applyTheme()
    },

    /**
     * 应用主题到DOM
     */
    applyTheme() {
      const root = document.documentElement
      const theme = this.currentTheme
      
      root.setAttribute('data-theme', theme)
      root.style.setProperty('--primary-color', this.theme.primaryColor)
      root.style.setProperty('--font-size', `${this.theme.fontSize}px`)
    },

    /**
     * 设置语言
     */
    setLocale(locale: string) {
      this.locale = locale
    },

    /**
     * 注册面板
     */
    registerPanel(panelId: string, initialState: Partial<PanelState> = {}) {
      this.panels[panelId] = {
        id: panelId,
        visible: false,
        collapsed: false,
        order: 0,
        ...initialState
      }
    },

    /**
     * 注销面板
     */
    unregisterPanel(panelId: string) {
      delete this.panels[panelId]
    },

    /**
     * 显示面板
     */
    showPanel(panelId: string) {
      if (this.panels[panelId]) {
        this.panels[panelId].visible = true
      }
    },

    /**
     * 隐藏面板
     */
    hidePanel(panelId: string) {
      if (this.panels[panelId]) {
        this.panels[panelId].visible = false
      }
    },

    /**
     * 切换面板显示状态
     */
    togglePanel(panelId: string) {
      if (this.panels[panelId]) {
        this.panels[panelId].visible = !this.panels[panelId].visible
      }
    },

    /**
     * 折叠/展开面板
     */
    togglePanelCollapse(panelId: string) {
      if (this.panels[panelId]) {
        this.panels[panelId].collapsed = !this.panels[panelId].collapsed
      }
    },

    /**
     * 设置面板尺寸
     */
    setPanelSize(panelId: string, width?: number, height?: number) {
      if (this.panels[panelId]) {
        if (width !== undefined) this.panels[panelId].width = width
        if (height !== undefined) this.panels[panelId].height = height
      }
    },

    /**
     * 设置面板顺序
     */
    setPanelOrder(panelId: string, order: number) {
      if (this.panels[panelId]) {
        this.panels[panelId].order = order
      }
    },

    /**
     * 切换侧边栏
     */
    toggleSidebar() {
      this.sidebar.collapsed = !this.sidebar.collapsed
    },

    /**
     * 设置侧边栏宽度
     */
    setSidebarWidth(width: number) {
      this.sidebar.width = Math.max(200, Math.min(400, width))
    },

    /**
     * 显示对话框
     */
    showDialog(dialogId: string, data?: any) {
      this.dialogs[dialogId] = {
        visible: true,
        data
      }
    },

    /**
     * 隐藏对话框
     */
    hideDialog(dialogId: string) {
      if (this.dialogs[dialogId]) {
        this.dialogs[dialogId].visible = false
      }
    },

    /**
     * 设置全局加载状态
     */
    setGlobalLoading(loading: boolean) {
      this.loading.global = loading
    },

    /**
     * 设置组件加载状态
     */
    setComponentLoading(componentId: string, loading: boolean) {
      if (loading) {
        this.loading.components[componentId] = true
      } else {
        delete this.loading.components[componentId]
      }
    },

    /**
     * 添加通知
     */
    addNotification(notification: Omit<UIStoreState['notifications'][0], 'id' | 'timestamp'>) {
      const id = `notification_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
      this.notifications.push({
        id,
        timestamp: Date.now(),
        duration: 5000, // 默认5秒
        ...notification
      })

      // 自动清理过期通知
      this.cleanupNotifications()

      return id
    },

    /**
     * 移除通知
     */
    removeNotification(notificationId: string) {
      this.notifications = this.notifications.filter(n => n.id !== notificationId)
    },

    /**
     * 清理过期通知
     */
    cleanupNotifications() {
      const now = Date.now()
      this.notifications = this.notifications.filter(notification => {
        if (!notification.duration) return true
        return now - notification.timestamp < notification.duration
      })
    },

    /**
     * 清空所有通知
     */
    clearNotifications() {
      this.notifications = []
    },

    /**
     * 添加状态栏项目
     */
    addStatusBarItem(item: UIStoreState['statusBar']['items'][0]) {
      const existingIndex = this.statusBar.items.findIndex(i => i.id === item.id)
      if (existingIndex >= 0) {
        this.statusBar.items[existingIndex] = item
      } else {
        this.statusBar.items.push(item)
      }
    },

    /**
     * 移除状态栏项目
     */
    removeStatusBarItem(itemId: string) {
      this.statusBar.items = this.statusBar.items.filter(item => item.id !== itemId)
    },

    /**
     * 重置UI状态
     */
    reset() {
      // 重置为默认状态，但保留主题和语言设置
      const { theme, locale } = this
      Object.assign(this, {
        panels: {},
        sidebar: { collapsed: false, width: 240 },
        dialogs: {},
        loading: { global: false, components: {} },
        notifications: [],
        statusBar: { visible: true, items: [] }
      })
      this.theme = theme
      this.locale = locale
    }
  }
})