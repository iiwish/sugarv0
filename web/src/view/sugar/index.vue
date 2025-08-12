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
    <main class="main-content" :class="{ 'main-content--sidebar-collapsed': sidebarCollapsed }">
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
        </div>
      </div>
      
      <!-- Univer容器 -->
      <div id="univer-sheet-container" class="univer-container"></div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, onActivated, onDeactivated } from 'vue'
import { ElMessage } from 'element-plus'
import { DocumentAdd, Refresh } from '@element-plus/icons-vue'
import { useApp } from '@/composables/useApp'
import { useWorkspace } from '@/composables/useWorkspace'
import Sidebar from '@/components/Sidebar.vue'
import type { WorkspaceTreeNode, ApiResponse } from '@/types/api'
import type { UniverCorePlugin } from '@/plugins/univer-core'
import type { CustomFormulasPlugin } from '@/plugins/custom-formulas'

// 定义组件名称，用于keep-alive
defineOptions({
  name: 'SugarApp'
})

const containerRef = ref<HTMLDivElement | null>(null)
const sidebarRef = ref()
const sidebarCollapsed = ref(false)
const isSaving = ref(false)
const isRefreshing = ref(false)

// 使用工作空间管理
const workspace = useWorkspace()

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

onMounted(() => {
  updateHeight()
  window.addEventListener('resize', updateHeight)
  // 添加键盘快捷键监听
  addKeyboardListeners()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateHeight)
  removeKeyboardListeners()
})

// 处理keep-alive的激活和停用
onActivated(() => {
  // 重新添加键盘监听器
  addKeyboardListeners()
})

onDeactivated(() => {
  // 移除键盘监听器
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
    const univerCorePlugin = app.pluginManager?.getPlugin('univer-core') as UniverCorePlugin
    if (!univerCorePlugin) {
      throw new Error('Univer核心插件未找到')
    }

    // 获取当前工作簿
    const workbook = univerCorePlugin.getCurrentWorkbook()
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
    }) as unknown as ApiResponse<any>

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
      const response = await getWorkbookContent({ id: data.id }) as unknown as ApiResponse<any>
      
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
      const response = await getWorkbookContent({ id: data.id }) as unknown as ApiResponse<any>
      
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
    const univerCorePlugin = app.pluginManager?.getPlugin('univer-core') as UniverCorePlugin
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
    const workbook = await univerCorePlugin.createWorkbook(workbookData)
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
  // 当侧边栏状态改变时，重新计算容器高度
  updateHeight()
}

// 处理团队切换
const handleTeamChange = (teamId: string) => {
  console.log('团队切换:', teamId)
  // 这里可以添加团队切换后的逻辑，比如清空当前工作区状态等
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
    const customFormulasPlugin = app.pluginManager?.getPlugin('custom-formulas') as CustomFormulasPlugin
    if (!customFormulasPlugin) {
      throw new Error('自定义公式插件未找到')
    }

    // 调用插件的刷新方法
    const result = await customFormulasPlugin.refreshDatabaseFormulas()
    
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

    // 获取统计信息并记录
    const stats = customFormulasPlugin.getDatabaseFormulaStats()
    console.log('数据库公式统计:', stats)
    
    // 如果有统计信息，在控制台显示详细信息
    if (stats.total > 0) {
      console.log(`当前工作簿包含 ${stats.total} 个数据库公式:`, stats.byType)
    }
    
  } catch (error) {
    console.error('刷新公式失败:', error)
    ElMessage.error('刷新公式失败: ' + (error as Error).message)
  } finally {
    isRefreshing.value = false
  }
}

// useApp 组合式函数封装了所有初始化和清理逻辑。
// 它会在 onMounted 时自动运行，并在 onBeforeUnmount 时自动关闭。
// 现在还支持 onActivated 和 onDeactivated 来处理keep-alive的状态切换。
const app = useApp()

// 组件挂载时初始化工作空间
onMounted(() => {
  workspace.initialize()
})

// 暴露应用状态供调试使用
defineExpose({
  app,
  sidebarRef,
  workspace
})
</script>

<style scoped>
.sugar-app-container {
  width: 100%;
  /* height is now set dynamically */
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
  transition: margin-left 0.3s ease;
}

.main-content--sidebar-collapsed {
  /* 当侧边栏折叠时的样式调整 */
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
}
</style>