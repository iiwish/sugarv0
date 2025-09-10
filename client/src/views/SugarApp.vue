<template>
  <div ref="containerRef" class="sugar-app-container">
    <!-- 侧边栏 -->
    <Sidebar
      ref="sidebarRef"
      @node-click="handleNodeClick"
      @file-create="handleFileCreate"
      @collapse-change="handleSidebarCollapseChange"
      @team-change="handleTeamChange"
      class="app-sidebar"
    />
    
    <!-- 主内容区域 -->
    <main class="main-content" :class="{
      'main-content--sidebar-collapsed': sidebarCollapsed,
      'main-content--chat-collapsed': chatCollapsed
    }">
      <!-- 工具栏 -->
      <div class="toolbar" v-if="workspace.currentNode.value && workspace.currentNode.value.type === 'file'">
        <div class="toolbar-left">
          <span class="current-file-name">{{ workspace.currentNode.value.name }}</span>
        </div>
        <div class="toolbar-right">
          <el-button
            type="success"
            size="small"
            @click="handleRefreshFormulas"
            :icon="Refresh"
            :loading="isRefreshing"
            title="刷新所有数据库公式 (Ctrl+R 或 F5)"
          >
            {{ isRefreshing ? '刷新中...' : '刷新数据' }}
          </el-button>
          <el-button
            type="primary"
            size="small"
            @click="handleSaveFile"
            :icon="DocumentAdd"
            :loading="isSaving"
          >
            保存 (Ctrl+S)
          </el-button>
          <el-button
            type="info"
            size="small"
            @click="handleToggleChat"
            :icon="ChatDotRound"
            title="切换AI助手面板"
          >
            AI助手
          </el-button>
        </div>
      </div>
      
      <!-- Univer容器 -->
      <div id="univer-sheet-container" class="univer-container"></div>
      
      <!-- 欢迎页面 -->
      <div v-if="!workspace.currentNode.value" class="welcome-container">
        <div class="welcome-content">
          <div class="welcome-icon">
            <el-icon size="64" color="#409eff">
              <DocumentAdd />
            </el-icon>
          </div>
          <h2>欢迎使用 Sugar Analytics</h2>
          <p>智能数据分析平台，基于 Univer 表格引擎</p>
          <div class="welcome-actions">
            <el-button type="primary" @click="handleCreateNewFile">
              <el-icon><DocumentAdd /></el-icon>
              创建新文件
            </el-button>
            <el-button @click="handleOpenRecentFile" :disabled="!hasRecentFiles">
              <el-icon><FolderOpened /></el-icon>
              打开最近文件
            </el-button>
          </div>
          <div class="recent-files" v-if="hasRecentFiles">
            <h3>最近访问</h3>
            <div class="recent-file-list">
              <div
                v-for="file in workspace.recentFiles.value.slice(0, 5)"
                :key="file.id"
                class="recent-file-item"
                @click="handleNodeClick(file)"
              >
                <el-icon><Document /></el-icon>
                <span>{{ file.name }}</span>
                <span class="file-path">{{ getFilePath(file) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- 聊天面板 -->
    <ChatPanel
      ref="chatPanelRef"
      @collapse-change="handleChatCollapseChange"
      class="app-chat-panel"
      :default-collapsed="chatCollapsed"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, onActivated, onDeactivated, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { DocumentAdd, Refresh, ChatDotRound, FolderOpened, Document } from '@element-plus/icons-vue'
import { useApp, useWorkspace, usePluginManager } from '@/composables'
import Sidebar from '@/components/Sidebar.vue'
import ChatPanel from '@/components/ChatPanel.vue'
import type { WorkspaceTreeNode } from '@/types/api'

// 定义组件名称，用于keep-alive
defineOptions({
  name: 'SugarApp'
})

const containerRef = ref<HTMLDivElement | null>(null)
const sidebarRef = ref()
const chatPanelRef = ref()
const sidebarCollapsed = ref(false)
const chatCollapsed = ref(false)
const isSaving = ref(false)
const isRefreshing = ref(false)

// 使用组合式函数
const app = useApp()
const workspace = useWorkspace()
const pluginManager = usePluginManager()

// 计算属性
const hasRecentFiles = computed(() => workspace.recentFiles.value.length > 0)

const updateHeight = () => {
  if (containerRef.value) {
    const top = containerRef.value.offsetTop
    containerRef.value.style.height = `calc(100vh - ${top}px - 60px)`
  }
}

// 添加和移除键盘事件监听器的函数
const addKeyboardListeners = () => {
  window.addEventListener('keydown', handleKeyDown)
}

const removeKeyboardListeners = () => {
  window.removeEventListener('keydown', handleKeyDown)
}

onMounted(async () => {
  updateHeight()
  window.addEventListener('resize', updateHeight)
  addKeyboardListeners()
  
  // 初始化应用
  await app.initialize()
  
  // 初始化插件管理器
  await pluginManager.initialize()
  
  // 初始化工作空间
  workspace.initialize()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateHeight)
  removeKeyboardListeners()
  
  // 清理资源
  pluginManager.cleanup()
})

// 处理keep-alive的激活和停用
onActivated(() => {
  addKeyboardListeners()
})

onDeactivated(() => {
  removeKeyboardListeners()
})

// 处理键盘快捷键
const handleKeyDown = (event: KeyboardEvent) => {
  // Ctrl+S 保存文件
  if (event.ctrlKey && event.key === 's') {
    event.preventDefault()
    handleSaveFile()
  }
  
  // Ctrl+R 或 F5 刷新数据库公式
  if ((event.ctrlKey && event.key === 'r') || event.key === 'F5') {
    event.preventDefault()
    handleRefreshFormulas()
  }
}

// 保存当前文件
const handleSaveFile = async () => {
  const currentNode = workspace.currentNode.value
  if (!currentNode || currentNode.type !== 'file') {
    ElMessage.warning('请先选择要保存的文件')
    return
  }

  if (isSaving.value) {
    return // 防止重复保存
  }

  try {
    isSaving.value = true
    
    // 获取Univer核心插件
    const univerCorePlugin = pluginManager.registeredPlugins.value.get('univer-core')
    if (!univerCorePlugin) {
      throw new Error('Univer核心插件未找到')
    }

    // 获取当前工作簿
    const workbook = univerCorePlugin.getCurrentWorkbook?.()
    if (!workbook) {
      ElMessage.warning('没有打开的工作簿')
      return
    }

    // 获取工作簿数据
    const workbookData = workbook.getSnapshot()
    
    // 调用保存API
    const { saveWorkbookContent } = await import('@/api/sugar/sugarWorkspaces')
    const response = await saveWorkbookContent({
      id: currentNode.id,
      content: workbookData
    })

    if (response?.code === 0) {
      ElMessage.success('文件保存成功')
    } else {
      ElMessage.error(response?.msg || '保存失败')
    }
  } catch (error) {
    console.error('保存文件失败:', error)
    ElMessage.error('保存文件失败')
  } finally {
    isSaving.value = false
  }
}

// 处理工作空间节点点击
const handleNodeClick = async (data: WorkspaceTreeNode) => {
  console.log('节点点击:', data)
  workspace.setCurrentNode(data)
  
  if (data.type === 'file') {
    try {
      // 获取文件内容并在Univer中打开
      const { getWorkbookContent } = await import('@/api/sugar/sugarWorkspaces')
      const response = await getWorkbookContent({ id: data.id })
      
      if (response?.code === 0) {
        await loadWorkbookInUniver(data, response.data)
        ElMessage.success(`文件 "${data.name}" 已打开`)
      } else {
        ElMessage.error('获取文件内容失败')
      }
    } catch (error) {
      console.error('打开文件失败:', error)
      ElMessage.error('打开文件失败')
    }
  } else {
    // 文件夹点击，可以展开/折叠
    ElMessage.info(`选中文件夹: ${data.name}`)
  }
}

// 处理文件创建
const handleFileCreate = async (data: WorkspaceTreeNode | string | undefined) => {
  console.log('文件创建事件:', data)
  
  // 如果传入的是WorkspaceTreeNode对象，说明是新创建的文件
  if (data && typeof data === 'object' && 'id' in data) {
    try {
      // 获取文件内容
      const { getWorkbookContent } = await import('@/api/sugar/sugarWorkspaces')
      const response = await getWorkbookContent({ id: data.id })
      
      if (response?.code === 0) {
        // 在Univer中创建并打开工作簿
        await loadWorkbookInUniver(data, response.data)
        ElMessage.success(`文件 "${data.name}" 已打开`)
      } else {
        ElMessage.error('获取文件内容失败')
      }
    } catch (error) {
      console.error('打开新建文件失败:', error)
      ElMessage.error('打开文件失败')
    }
  }
}

// 在Univer中加载工作簿
const loadWorkbookInUniver = async (fileData: WorkspaceTreeNode, content: any) => {
  try {
    // 获取Univer核心插件
    const univerCorePlugin = pluginManager.registeredPlugins.value.get('univer-core')
    if (!univerCorePlugin) {
      throw new Error('Univer核心插件未找到')
    }

    // 准备工作簿数据
    const workbookData = {
      id: fileData.id,
      name: fileData.name,
      ...content
    }

    // 创建工作簿
    const workbook = await univerCorePlugin.createWorkbook?.(workbookData)
    console.log('工作簿已在Univer中创建:', workbook)
    
    // 设置当前文件节点
    workspace.setCurrentNode(fileData)
    
  } catch (error) {
    console.error('在Univer中加载工作簿失败:', error)
    throw error
  }
}

// 处理侧边栏折叠状态变化
const handleSidebarCollapseChange = (collapsed: boolean) => {
  sidebarCollapsed.value = collapsed
  updateHeight()
}

// 处理聊天面板折叠状态变化
const handleChatCollapseChange = (collapsed: boolean) => {
  chatCollapsed.value = collapsed
  updateHeight()
}

// 切换聊天面板
const handleToggleChat = () => {
  if (chatPanelRef.value) {
    chatPanelRef.value.toggleCollapse()
  }
}

// 处理团队切换
const handleTeamChange = (teamId: string) => {
  console.log('团队切换:', teamId)
  workspace.setCurrentNode(null)
  ElMessage.success('团队切换成功')
}

// 刷新所有数据库公式
const handleRefreshFormulas = async () => {
  const currentNode = workspace.currentNode.value
  if (!currentNode || currentNode.type !== 'file') {
    ElMessage.warning('请先打开一个文件')
    return
  }

  if (isRefreshing.value) {
    return // 防止重复刷新
  }

  try {
    isRefreshing.value = true
    
    // 获取自定义公式插件
    const customFormulasPlugin = pluginManager.registeredPlugins.value.get('custom-formulas')
    if (!customFormulasPlugin) {
      throw new Error('自定义公式插件未找到')
    }

    // 调用插件的刷新方法
    const result = await customFormulasPlugin.refreshDatabaseFormulas?.()
    
    if (!result) {
      ElMessage.info('插件不支持公式刷新功能')
      return
    }
    
    // 根据结果显示相应的消息
    if (result.success > 0 && result.failed === 0) {
      ElMessage.success({
        message: `数据库公式刷新成功！共处理 ${result.success} 项`,
        duration: 3000,
        showClose: true
      })
    } else if (result.success > 0 && result.failed > 0) {
      ElMessage.warning({
        message: `数据库公式部分刷新成功！成功 ${result.success} 项，失败 ${result.failed} 项`,
        duration: 5000,
        showClose: true
      })
      console.warn('刷新失败的详情:', result.errors)
    } else if (result.failed > 0) {
      ElMessage.error({
        message: `数据库公式刷新失败！失败 ${result.failed} 项`,
        duration: 5000,
        showClose: true
      })
      console.error('刷新失败的详情:', result.errors)
    } else {
      ElMessage.info({
        message: '没有找到需要刷新的数据库公式',
        duration: 3000,
        showClose: true
      })
    }
    
  } catch (error) {
    console.error('刷新公式失败:', error)
    ElMessage.error('刷新公式失败: ' + (error as Error).message)
  } finally {
    isRefreshing.value = false
  }
}

// 创建新文件
const handleCreateNewFile = async () => {
  try {
    const { value: fileName } = await ElMessageBox.prompt('请输入文件名', '创建新文件', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputPattern: /^[^\\/:*?"<>|]+$/,
      inputErrorMessage: '文件名不能包含特殊字符'
    })
    
    if (fileName) {
      // 这里应该调用创建文件的API
      ElMessage.success(`文件 "${fileName}" 创建成功`)
    }
  } catch {
    // 用户取消
  }
}

// 打开最近文件
const handleOpenRecentFile = () => {
  if (workspace.recentFiles.value.length > 0) {
    handleNodeClick(workspace.recentFiles.value[0])
  }
}

// 获取文件路径
const getFilePath = (file: WorkspaceTreeNode) => {
  // 这里应该根据文件的层级关系构建路径
  return file.parentId ? `/${file.name}` : file.name
}

// 暴露应用状态供调试使用
defineExpose({
  app,
  workspace,
  pluginManager,
  sidebarRef,
  chatPanelRef
})
</script>

<style scoped>
.sugar-app-container {
  width: 100%;
  display: flex;
  flex-direction: row;
}

.app-sidebar {
  flex-shrink: 0;
}

.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding: 8px;
  transition: all 0.3s ease;
}

.main-content--sidebar-collapsed {
  /* 当侧边栏折叠时的样式调整 */
}

.main-content--chat-collapsed {
  /* 当聊天面板折叠时的样式调整 */
}

.app-chat-panel {
  flex-shrink: 0;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 16px;
  background: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
  border-radius: 8px 8px 0 0;
  margin-bottom: 8px;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.current-file-name {
  font-size: 14px;
  font-weight: 500;
  color: #495057;
  margin-right: 16px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.univer-container {
  flex: 1;
  border-radius: 8px;
  background-color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 当有工具栏时，调整容器样式 */
.main-content:has(.toolbar) .univer-container {
  border-radius: 0 0 8px 8px;
}

/* 欢迎页面样式 */
.welcome-container {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-radius: 8px;
}

.welcome-content {
  text-align: center;
  max-width: 600px;
  padding: 40px;
}

.welcome-icon {
  margin-bottom: 24px;
}

.welcome-content h2 {
  font-size: 28px;
  color: #2c3e50;
  margin-bottom: 12px;
  font-weight: 600;
}

.welcome-content p {
  font-size: 16px;
  color: #7f8c8d;
  margin-bottom: 32px;
  line-height: 1.6;
}

.welcome-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
  margin-bottom: 40px;
}

.recent-files {
  text-align: left;
}

.recent-files h3 {
  font-size: 18px;
  color: #2c3e50;
  margin-bottom: 16px;
  border-bottom: 2px solid #3498db;
  padding-bottom: 8px;
}

.recent-file-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.recent-file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: white;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.recent-file-item:hover {
  background: #f8f9fa;
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.recent-file-item .el-icon {
  color: #3498db;
  font-size: 18px;
}

.recent-file-item span:first-of-type {
  font-weight: 500;
  color: #2c3e50;
}

.file-path {
  font-size: 12px;
  color: #95a5a6;
  margin-left: auto;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sugar-app-container {
    position: relative;
  }
  
  .main-content {
    padding: 4px;
  }
  
  .toolbar {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  
  .toolbar-right {
    width: 100%;
    justify-content: flex-end;
  }
  
  .current-file-name {
    margin-right: 0;
    margin-bottom: 4px;
  }
  
  .welcome-content {
    padding: 20px;
  }
  
  .welcome-actions {
    flex-direction: column;
    align-items: center;
  }
  
  .recent-file-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }
  
  .file-path {
    margin-left: 0;
  }
}
</style>