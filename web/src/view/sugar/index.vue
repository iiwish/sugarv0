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
      <div id="univer-sheet-container" class="univer-container"></div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage } from 'element-plus'
import { useApp } from '@/composables/useApp'
import { useWorkspace } from '@/composables/useWorkspace'
import Sidebar from '@/components/Sidebar.vue'
import type { WorkspaceTreeNode } from '@/types/api'

// 定义组件名称，用于keep-alive
defineOptions({
  name: 'SugarApp'
})

const containerRef = ref<HTMLDivElement | null>(null)
const sidebarRef = ref()
const sidebarCollapsed = ref(false)

// 使用工作空间管理
const workspace = useWorkspace()

const updateHeight = () => {
  if (containerRef.value) {
    const top = containerRef.value.offsetTop
    containerRef.value.style.height = `calc(100vh - ${top}px - 60px)`
  }
}

onMounted(() => {
  updateHeight()
  window.addEventListener('resize', updateHeight)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateHeight)
})

// 处理工作空间节点点击
const handleNodeClick = (data: WorkspaceTreeNode) => {
  console.log('节点点击:', data)
  workspace.setCurrentNode(data)
  
  if (data.type === 'file') {
    // 这里可以添加打开文件的逻辑
    ElMessage.info(`打开文件: ${data.name}`)
    // 可以在这里集成到Univer中打开文件
  } else {
    // 文件夹点击，可以展开/折叠
    ElMessage.info(`选中文件夹: ${data.name}`)
  }
}

// 处理文件创建
const handleFileCreate = (parentId?: string) => {
  console.log('创建文件，父级ID:', parentId)
  // 这里可以添加创建新文件的逻辑
  ElMessage.info('创建新文件功能待实现')
  // 可以在这里创建新的工作簿并在Univer中打开
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
  overflow: hidden;
  padding: 8px;
  transition: margin-left 0.3s ease;
}

.main-content--sidebar-collapsed {
  /* 当侧边栏折叠时的样式调整 */
}

.univer-container {
  flex: 1;
  border-radius: 8px;
  background-color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sugar-app-container {
    position: relative;
  }
  
  .main-content {
    padding: 4px;
  }
}
</style>