<template>
  <div class="workspace-tree">
    <!-- 搜索框 -->
    <div class="search-container">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索文件和文件夹..."
        :prefix-icon="Search"
        clearable
        @input="handleSearch"
        class="search-input"
      />
    </div>

    <!-- 操作按钮 -->
    <div class="tree-actions">
      <el-button
        type="primary"
        size="small"
        @click="handleCreateFolder"
        :icon="FolderAdd"
      >
        新建文件夹
      </el-button>
      <el-button
        type="success"
        size="small"
        @click="handleCreateFile"
        :icon="DocumentAdd"
      >
        新建文件
      </el-button>
    </div>

    <!-- 快捷标签页 -->
    <div class="quick-tabs">
      <el-tabs v-model="activeTab" size="small" @tab-click="handleTabClick">
        <el-tab-pane label="文件树" name="tree" />
        <el-tab-pane label="最近" name="recent" />
        <el-tab-pane label="收藏" name="favorites" />
      </el-tabs>
    </div>

    <!-- 内容区域 -->
    <div class="tree-container">
      <!-- 文件树视图 -->
      <div v-show="activeTab === 'tree'" class="tab-content">
        <!-- 搜索结果 -->
        <div v-if="searchKeyword && searchResults.length > 0" class="search-results">
          <div class="search-results-header">
            <span>搜索结果 ({{ searchResults.length }})</span>
            <el-button type="text" size="small" @click="clearSearch">清除</el-button>
          </div>
          <div
            v-for="item in searchResults"
            :key="item.id"
            class="search-result-item"
            @click="handleNodeClick(item)"
          >
            <el-icon class="node-icon">
              <Folder v-if="item.type === 'folder'" />
              <Document v-else />
            </el-icon>
            <span class="node-label">{{ item.name }}</span>
            <el-button
              type="text"
              size="small"
              @click.stop="toggleFavorite(item)"
              :icon="isFavorite(item.id) ? StarFilled : Star"
              :class="{ 'is-favorite': isFavorite(item.id) }"
            />
          </div>
        </div>

        <!-- 树形结构 -->
        <el-tree
          v-show="!searchKeyword"
          ref="treeRef"
          :data="treeData"
          :props="treeProps"
          :expand-on-click-node="false"
          :default-expand-all="false"
          node-key="id"
          @node-click="handleNodeClick"
          @node-dblclick="handleNodeDoubleClick"
          @node-contextmenu="handleContextMenu"
          class="workspace-tree-view"
        >
          <template #default="{ node, data }">
            <div 
              class="tree-node"
              @mouseenter="hoveredNodeId = data.id"
              @mouseleave="hoveredNodeId = ''"
            >
              <el-icon class="node-icon">
                <Folder v-if="data.type === 'folder'" />
                <Document v-else />
              </el-icon>
              <span class="node-label">{{ node.label }}</span>
              <div class="node-actions" v-if="hoveredNodeId === data.id">
                <el-button
                  type="text"
                  size="small"
                  @click.stop="toggleFavorite(data)"
                  :icon="isFavorite(data.id) ? StarFilled : Star"
                  :class="{ 'is-favorite': isFavorite(data.id) }"
                />
                <el-button
                  type="text"
                  size="small"
                  @click.stop="handleRename(data)"
                  :icon="Edit"
                />
                <el-button
                  type="text"
                  size="small"
                  @click.stop="handleDelete(data)"
                  :icon="Delete"
                />
              </div>
            </div>
          </template>
        </el-tree>
      </div>

      <!-- 最近访问视图 -->
      <div v-show="activeTab === 'recent'" class="tab-content">
        <div class="recent-files-header">
          <span>最近访问</span>
          <el-button type="text" size="small" @click="clearRecentFiles">清空</el-button>
        </div>
        <div v-if="recentFiles.length === 0" class="empty-state">
          <el-icon><Document /></el-icon>
          <span>暂无最近访问的文件</span>
        </div>
        <div
          v-for="item in recentFiles"
          :key="item.id"
          class="recent-file-item"
          @click="handleNodeClick(item)"
        >
          <el-icon class="node-icon">
            <Document />
          </el-icon>
          <div class="file-info">
            <span class="file-name">{{ item.name }}</span>
            <span class="file-path">{{ getFilePath(item) }}</span>
          </div>
          <el-button
            type="text"
            size="small"
            @click.stop="toggleFavorite(item)"
            :icon="isFavorite(item.id) ? StarFilled : Star"
            :class="{ 'is-favorite': isFavorite(item.id) }"
          />
        </div>
      </div>

      <!-- 收藏视图 -->
      <div v-show="activeTab === 'favorites'" class="tab-content">
        <div class="favorites-header">
          <span>收藏</span>
        </div>
        <div v-if="favorites.length === 0" class="empty-state">
          <el-icon><Star /></el-icon>
          <span>暂无收藏的文件或文件夹</span>
        </div>
        <div
          v-for="item in favorites"
          :key="item.id"
          class="favorite-item"
          @click="handleNodeClick(item)"
        >
          <el-icon class="node-icon">
            <Folder v-if="item.type === 'folder'" />
            <Document v-else />
          </el-icon>
          <div class="file-info">
            <span class="file-name">{{ item.name }}</span>
            <span class="file-path">{{ getFilePath(item) }}</span>
          </div>
          <el-button
            type="text"
            size="small"
            @click.stop="toggleFavorite(item)"
            :icon="StarFilled"
            class="is-favorite"
          />
        </div>
      </div>
    </div>

    <!-- 右键菜单 -->
    <el-dropdown
      ref="contextMenuRef"
      trigger="contextmenu"
      :teleported="false"
      class="context-menu"
    >
      <span></span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item @click="handleCreateFolder" :icon="FolderAdd">
            新建文件夹
          </el-dropdown-item>
          <el-dropdown-item @click="handleCreateFile" :icon="DocumentAdd">
            新建文件
          </el-dropdown-item>
          <el-dropdown-item
            v-if="contextMenuData"
            @click="handleRename(contextMenuData)"
            :icon="Edit"
          >
            重命名
          </el-dropdown-item>
          <el-dropdown-item
            v-if="contextMenuData"
            @click="handleDelete(contextMenuData)"
            :icon="Delete"
            divided
          >
            删除
          </el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>

    <!-- 重命名对话框 -->
    <el-dialog
      v-model="renameDialogVisible"
      title="重命名"
      width="400px"
    >
      <el-form @submit.prevent="confirmRename">
        <el-form-item label="名称">
          <el-input
            v-model="newName"
            placeholder="请输入新名称"
            @keyup.enter="confirmRename"
            ref="renameInputRef"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="renameDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmRename">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Folder,
  Document,
  FolderAdd,
  DocumentAdd,
  Edit,
  Delete,
  Search,
  Star,
  StarFilled
} from '@element-plus/icons-vue'
import {
  getWorkspaceTree,
  createFolder,
  renameItem,
  deleteItem
} from '@/api/sugar/sugarFolders'
import type {
  ApiResponse,
  WorkspaceTreeNode,
  TeamInfo,
  CreateFolderData,
  RenameItemData
} from '@/types/api'
import { useWorkspace } from '@/composables/useWorkspace'

// 定义组件属性
interface Props {
  height?: string
  selectedTeamId?: string
  teams?: TeamInfo[]
}

const props = withDefaults(defineProps<Props>(), {
  height: '100%',
  selectedTeamId: '',
  teams: () => []
})

// 定义事件
const emit = defineEmits<{
  nodeClick: [data: WorkspaceTreeNode]
  fileCreate: [parentId?: string]
  teamChange: [teamId: string]
}>()

// 使用工作空间管理
const workspace = useWorkspace()

// 响应式数据
const treeRef = ref()
const contextMenuRef = ref()
const renameInputRef = ref()
const treeData = ref<WorkspaceTreeNode[]>([])
const hoveredNodeId = ref('')
const contextMenuData = ref<WorkspaceTreeNode | null>(null)
const renameDialogVisible = ref(false)
const newName = ref('')
const currentRenameItem = ref<WorkspaceTreeNode | null>(null)
const selectedNode = ref<WorkspaceTreeNode | null>(null)

// 搜索相关
const searchKeyword = ref('')
const searchResults = ref<WorkspaceTreeNode[]>([])

// 标签页
const activeTab = ref('tree')

// 从工作空间管理中获取状态
const { recentFiles, favorites, isFavorite, toggleFavorite, clearRecentFiles } = workspace

// 树形组件配置
const treeProps = {
  children: 'children',
  label: 'name',
  isLeaf: (data: WorkspaceTreeNode) => data.type !== 'folder'
}

// 加载树形数据
const loadTreeData = async () => {
  if (!props.selectedTeamId) return
  
  try {
    const params = { teamId: props.selectedTeamId }
    const response = await getWorkspaceTree(params) as unknown as ApiResponse<{ tree: WorkspaceTreeNode[] }>
    if (response?.code === 0) {
      // 修复：使用 response.data.tree 而不是 response.data
      treeData.value = response.data?.tree || []
    } else {
      ElMessage.error(response?.msg || '加载工作空间树形结构失败')
    }
  } catch (error) {
    console.error('加载工作空间树形结构失败:', error)
    ElMessage.error('加载工作空间树形结构失败')
  }
}

// 处理团队切换
const handleTeamChange = async (teamId: string) => {
  treeData.value = [] // 清空树形数据
  emit('teamChange', teamId)
  await loadTreeData() // 重新加载树形数据
}

// 处理搜索
const handleSearch = () => {
  if (searchKeyword.value.trim()) {
    searchResults.value = workspace.searchNodes(searchKeyword.value, treeData.value)
  } else {
    searchResults.value = []
  }
}

// 清除搜索
const clearSearch = () => {
  searchKeyword.value = ''
  searchResults.value = []
}

// 处理标签页点击
const handleTabClick = (tab: any) => {
  activeTab.value = tab.name
}

// 处理节点点击
const handleNodeClick = (data: WorkspaceTreeNode) => {
  selectedNode.value = data
  workspace.setCurrentNode(data)
  emit('nodeClick', data)
}

// 处理节点双击
const handleNodeDoubleClick = (data: WorkspaceTreeNode) => {
  if (data.type === 'folder') {
    // 双击文件夹：展开/折叠
    const node = treeRef.value?.getNode(data.id)
    if (node) {
      node.expanded = !node.expanded
    }
  } else if (data.type === 'file') {
    // 双击文件：加载数据到右侧表格
    workspace.setCurrentNode(data)
    emit('nodeClick', data)
    ElMessage.success(`正在加载文件: ${data.name}`)
  }
}

// 获取文件路径（用于显示）
const getFilePath = (node: WorkspaceTreeNode): string => {
  // 这里可以根据实际需求实现路径获取逻辑
  return node.parentId ? `/${node.name}` : node.name
}

// 处理右键菜单
const handleContextMenu = (event: MouseEvent, data: WorkspaceTreeNode) => {
  event.preventDefault()
  contextMenuData.value = data
  // 这里可以添加右键菜单的显示逻辑
}

// 创建文件夹
const handleCreateFolder = async () => {
  try {
    // 根据当前选中的节点确定父节点ID
    let parentId: string | undefined = undefined
    
    if (selectedNode.value) {
      if (selectedNode.value.type === 'folder') {
        // 如果选中的是文件夹，则在该文件夹下创建
        parentId = selectedNode.value.id
      } else if (selectedNode.value.type === 'file') {
        // 如果选中的是文件，则在该文件的父文件夹下创建
        parentId = selectedNode.value.parentId
      }
    } else if (contextMenuData.value) {
      // 如果是右键菜单触发，使用右键菜单的节点
      if (contextMenuData.value.type === 'folder') {
        parentId = contextMenuData.value.id
      } else {
        parentId = contextMenuData.value.parentId
      }
    }
    
    const folderData: CreateFolderData = {
      name: '新建文件夹',
      type: 'folder',
      parentId,
      teamId: props.selectedTeamId || ''
    }
    
    const response = await createFolder(folderData) as unknown as ApiResponse
    if (response?.code === 0) {
      ElMessage.success('文件夹创建成功')
      loadTreeData()
    } else {
      ElMessage.error(response?.msg || '创建文件夹失败')
    }
  } catch (error) {
    console.error('创建文件夹失败:', error)
    ElMessage.error('创建文件夹失败')
  }
  contextMenuData.value = null
}

// 创建文件
const handleCreateFile = () => {
  // 根据当前选中的节点确定父节点ID
  let parentId: string | undefined = undefined
  
  if (selectedNode.value) {
    if (selectedNode.value.type === 'folder') {
      // 如果选中的是文件夹，则在该文件夹下创建
      parentId = selectedNode.value.id
    } else if (selectedNode.value.type === 'file') {
      // 如果选中的是文件，则在该文件的父文件夹下创建
      parentId = selectedNode.value.parentId
    }
  } else if (contextMenuData.value) {
    // 如果是右键菜单触发，使用右键菜单的节点
    if (contextMenuData.value.type === 'folder') {
      parentId = contextMenuData.value.id
    } else {
      parentId = contextMenuData.value.parentId
    }
  }
  
  emit('fileCreate', parentId)
  contextMenuData.value = null
}

// 重命名
const handleRename = (data: WorkspaceTreeNode) => {
  currentRenameItem.value = data
  newName.value = data.name
  renameDialogVisible.value = true
  nextTick(() => {
    renameInputRef.value?.focus()
  })
}

// 确认重命名
const confirmRename = async () => {
  if (!newName.value.trim()) {
    ElMessage.warning('名称不能为空')
    return
  }

  if (!currentRenameItem.value) {
    return
  }

  try {
    const renameData: RenameItemData = {
      id: currentRenameItem.value.id,
      name: newName.value.trim()
    }
    
    const response = await renameItem(renameData) as unknown as ApiResponse
    if (response?.code === 0) {
      ElMessage.success('重命名成功')
      renameDialogVisible.value = false
      loadTreeData()
    } else {
      ElMessage.error(response?.msg || '重命名失败')
    }
  } catch (error) {
    console.error('重命名失败:', error)
    ElMessage.error('重命名失败')
  }
}

// 删除
const handleDelete = async (data: WorkspaceTreeNode) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 "${data.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await deleteItem({ id: data.id }) as unknown as ApiResponse
    if (response?.code === 0) {
      ElMessage.success('删除成功')
      loadTreeData()
    } else {
      ElMessage.error(response?.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 监听选中的团队变化
watch(() => props.selectedTeamId, async () => {
  if (props.selectedTeamId) {
    await loadTreeData()
  }
})

// 组件挂载时初始化
onMounted(async () => {
  workspace.initialize()
  if (props.selectedTeamId) {
    await loadTreeData()
  }
})

// 暴露方法
defineExpose({
  refresh: loadTreeData
})
</script>

<style scoped>
.workspace-tree {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.search-container {
  padding: 8px 12px;
  border-bottom: 1px solid #e4e7ed;
}

.search-input {
  width: 100%;
}

.tree-actions {
  padding: 8px 12px;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  gap: 8px;
}

.quick-tabs {
  border-bottom: 1px solid #e4e7ed;
}

.tree-container {
  flex: 1;
  overflow: auto;
}

.tab-content {
  padding: 8px 0;
}

.workspace-tree-view {
  background: transparent;
}

.tree-node {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 2px 0;
}

.node-icon {
  margin-right: 6px;
  color: #606266;
}

.node-label {
  flex: 1;
  font-size: 14px;
  color: #303133;
}

.node-actions {
  display: flex;
  gap: 0;
  opacity: 0.8;
  flex-shrink: 0;
  max-width: 48px;
  align-items: center;
}

.tree-node:hover .node-actions {
  opacity: 1;
}

.node-actions .el-button {
  padding: 1px 2px !important;
  min-width: 20px !important;
  width: 20px !important;
  height: 20px !important;
  border: none !important;
  background: transparent !important;
  margin: 0 !important;
}

.node-actions .el-button .el-icon {
  font-size: 12px !important;
}

.node-actions .el-button:hover {
  background: rgba(64, 158, 255, 0.1) !important;
}

.node-actions .el-button.is-favorite {
  color: #f56c6c !important;
}

.node-actions .el-button.is-favorite:hover {
  background: rgba(245, 108, 108, 0.1) !important;
}

.search-results {
  padding: 0 12px;
}

.search-results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  font-size: 12px;
  color: #909399;
  border-bottom: 1px solid #e4e7ed;
}

.search-result-item,
.recent-file-item,
.favorite-item {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  border-radius: 4px;
  margin: 2px 0;
}

.search-result-item:hover,
.recent-file-item:hover,
.favorite-item:hover {
  background: #f5f7fa;
}

.recent-files-header,
.favorites-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  font-size: 12px;
  color: #909399;
  border-bottom: 1px solid #e4e7ed;
}

.file-info {
  flex: 1;
  margin-left: 8px;
  display: flex;
  flex-direction: column;
}

.file-name {
  font-size: 14px;
  color: #303133;
}

.file-path {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  color: #909399;
  font-size: 14px;
}

.empty-state .el-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.is-favorite {
  color: #f56c6c !important;
}

.context-menu {
  position: fixed;
  z-index: 9999;
}

:deep(.el-tree-node__content) {
  height: 32px;
  padding: 0 12px;
}

:deep(.el-tree-node__content:hover) {
  background-color: #f5f7fa;
}

:deep(.el-tree-node__expand-icon) {
  color: #c0c4cc;
}

:deep(.el-tree-node__expand-icon.expanded) {
  transform: rotate(90deg);
}

:deep(.el-tabs__header) {
  margin: 0;
}

:deep(.el-tabs__nav-wrap) {
  padding: 0 12px;
}
</style>