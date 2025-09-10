import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import type { WorkspaceTreeNode, TeamInfo } from '@/types/api'

/**
 * 工作空间管理组合式函数
 * 提供工作空间的状态管理和操作方法
 */
export function useWorkspace() {
  // 当前选中的工作空间节点
  const currentNode = ref<WorkspaceTreeNode | null>(null)
  
  // 当前选中的团队
  const currentTeam = ref<TeamInfo | null>(null)
  
  // 最近访问的文件列表
  const recentFiles = ref<WorkspaceTreeNode[]>([])
  
  // 收藏的文件/文件夹列表
  const favorites = ref<WorkspaceTreeNode[]>([])
  
  // 工作空间设置
  const workspaceSettings = reactive({
    autoSave: true,
    showHiddenFiles: false,
    sortBy: 'name' as 'name' | 'date' | 'type',
    sortOrder: 'asc' as 'asc' | 'desc'
  })

  /**
   * 设置当前选中的节点
   */
  const setCurrentNode = (node: WorkspaceTreeNode | null) => {
    currentNode.value = node
    
    // 如果是文件，添加到最近访问列表
    if (node && node.type === 'file') {
      addToRecentFiles(node)
    }
  }

  /**
   * 设置当前团队
   */
  const setCurrentTeam = (team: TeamInfo | null) => {
    currentTeam.value = team
  }

  /**
   * 添加到最近访问文件列表
   */
  const addToRecentFiles = (file: WorkspaceTreeNode) => {
    // 移除已存在的相同文件
    const index = recentFiles.value.findIndex(f => f.id === file.id)
    if (index > -1) {
      recentFiles.value.splice(index, 1)
    }
    
    // 添加到列表开头
    recentFiles.value.unshift(file)
    
    // 限制最近文件数量
    if (recentFiles.value.length > 10) {
      recentFiles.value = recentFiles.value.slice(0, 10)
    }
    
    // 保存到本地存储
    saveRecentFiles()
  }

  /**
   * 切换收藏状态
   */
  const toggleFavorite = (node: WorkspaceTreeNode) => {
    const index = favorites.value.findIndex(f => f.id === node.id)
    
    if (index > -1) {
      // 移除收藏
      favorites.value.splice(index, 1)
      ElMessage.success('已取消收藏')
    } else {
      // 添加收藏
      favorites.value.push(node)
      ElMessage.success('已添加到收藏')
    }
    
    // 保存到本地存储
    saveFavorites()
  }

  /**
   * 检查是否已收藏
   */
  const isFavorite = (nodeId: string): boolean => {
    return favorites.value.some(f => f.id === nodeId)
  }

  /**
   * 清空最近访问文件
   */
  const clearRecentFiles = () => {
    recentFiles.value = []
    saveRecentFiles()
    ElMessage.success('已清空最近访问文件')
  }

  /**
   * 获取节点路径（面包屑导航用）
   */
  const getNodePath = (node: WorkspaceTreeNode, allNodes: WorkspaceTreeNode[]): WorkspaceTreeNode[] => {
    const path: WorkspaceTreeNode[] = [node]
    let currentParentId = node.parentId
    
    while (currentParentId) {
      const parent = findNodeById(currentParentId, allNodes)
      if (parent) {
        path.unshift(parent)
        currentParentId = parent.parentId
      } else {
        break
      }
    }
    
    return path
  }

  /**
   * 根据ID查找节点
   */
  const findNodeById = (id: string, nodes: WorkspaceTreeNode[]): WorkspaceTreeNode | null => {
    for (const node of nodes) {
      if (node.id === id) {
        return node
      }
      if (node.children) {
        const found = findNodeById(id, node.children)
        if (found) return found
      }
    }
    return null
  }

  /**
   * 搜索节点
   */
  const searchNodes = (keyword: string, nodes: WorkspaceTreeNode[]): WorkspaceTreeNode[] => {
    const results: WorkspaceTreeNode[] = []
    
    const search = (nodeList: WorkspaceTreeNode[]) => {
      for (const node of nodeList) {
        if (node.name.toLowerCase().includes(keyword.toLowerCase())) {
          results.push(node)
        }
        if (node.children) {
          search(node.children)
        }
      }
    }
    
    search(nodes)
    return results
  }

  /**
   * 保存最近访问文件到本地存储
   */
  const saveRecentFiles = () => {
    try {
      localStorage.setItem('workspace_recent_files', JSON.stringify(recentFiles.value))
    } catch (error) {
      console.warn('保存最近访问文件失败:', error)
    }
  }

  /**
   * 从本地存储加载最近访问文件
   */
  const loadRecentFiles = () => {
    try {
      const saved = localStorage.getItem('workspace_recent_files')
      if (saved) {
        recentFiles.value = JSON.parse(saved)
      }
    } catch (error) {
      console.warn('加载最近访问文件失败:', error)
    }
  }

  /**
   * 保存收藏到本地存储
   */
  const saveFavorites = () => {
    try {
      localStorage.setItem('workspace_favorites', JSON.stringify(favorites.value))
    } catch (error) {
      console.warn('保存收藏失败:', error)
    }
  }

  /**
   * 从本地存储加载收藏
   */
  const loadFavorites = () => {
    try {
      const saved = localStorage.getItem('workspace_favorites')
      if (saved) {
        favorites.value = JSON.parse(saved)
      }
    } catch (error) {
      console.warn('加载收藏失败:', error)
    }
  }

  /**
   * 保存工作空间设置
   */
  const saveSettings = () => {
    try {
      localStorage.setItem('workspace_settings', JSON.stringify(workspaceSettings))
    } catch (error) {
      console.warn('保存工作空间设置失败:', error)
    }
  }

  /**
   * 加载工作空间设置
   */
  const loadSettings = () => {
    try {
      const saved = localStorage.getItem('workspace_settings')
      if (saved) {
        Object.assign(workspaceSettings, JSON.parse(saved))
      }
    } catch (error) {
      console.warn('加载工作空间设置失败:', error)
    }
  }

  /**
   * 初始化工作空间
   */
  const initialize = () => {
    loadRecentFiles()
    loadFavorites()
    loadSettings()
  }

  return {
    // 状态
    currentNode,
    currentTeam,
    recentFiles,
    favorites,
    workspaceSettings,
    
    // 方法
    setCurrentNode,
    setCurrentTeam,
    addToRecentFiles,
    toggleFavorite,
    isFavorite,
    clearRecentFiles,
    getNodePath,
    findNodeById,
    searchNodes,
    saveSettings,
    initialize
  }
}