<template>
  <div class="sidebar" :class="{ 'sidebar--collapsed': collapsed }">
    <!-- 侧边栏头部 -->
    <div class="sidebar-header">
      <div class="sidebar-title" v-if="!collapsed">
        <el-icon><FolderOpened /></el-icon>
        <span>工作空间</span>
      </div>
      <el-button
        type="text"
        size="small"
        @click="toggleCollapse"
        class="collapse-btn"
      >
        <el-icon>
          <component :is="collapsed ? 'Expand' : 'Fold'" />
        </el-icon>
      </el-button>
    </div>

    <!-- 团队选择 -->
    <div class="team-selector" v-if="!collapsed">
      <el-select
        v-model="selectedTeam"
        placeholder="选择团队"
        @change="handleTeamChange"
        style="width: 100%"
      >
        <el-option
          v-for="team in teams"
          :key="team.id"
          :label="team.teamName"
          :value="team.id"
        />
      </el-select>
    </div>

    <!-- 工具栏 -->
    <div class="sidebar-toolbar" v-if="!collapsed">
      <el-button
        type="primary"
        size="small"
        @click="handleCreateFile"
        :icon="DocumentAdd"
      >
        新建文件
      </el-button>
      <el-button
        type="default"
        size="small"
        @click="handleCreateFolder"
        :icon="FolderAdd"
      >
        新建文件夹
      </el-button>
    </div>

    <!-- 文件树 -->
    <div class="file-tree" v-if="!collapsed">
      <el-tree
        ref="treeRef"
        :data="treeData"
        :props="treeProps"
        node-key="id"
        :expand-on-click-node="false"
        :default-expand-all="false"
        @node-click="handleNodeClick"
        @node-contextmenu="handleContextMenu"
        class="workspace-tree"
      >
        <template #default="{ node, data }">
          <div class="tree-node">
            <el-icon class="node-icon">
              <component :is="data.type === 'folder' ? 'Folder' : 'Document'" />
            </el-icon>
            <span class="node-label">{{ data.name }}</span>
          </div>
        </template>
      </el-tree>
    </div>

    <!-- 折叠状态下的快捷按钮 -->
    <div class="collapsed-actions" v-if="collapsed">
      <el-tooltip content="新建文件" placement="right">
        <el-button
          type="text"
          size="small"
          @click="handleCreateFile"
          class="collapsed-btn"
        >
          <el-icon><DocumentAdd /></el-icon>
        </el-button>
      </el-tooltip>
      <el-tooltip content="新建文件夹" placement="right">
        <el-button
          type="text"
          size="small"
          @click="handleCreateFolder"
          class="collapsed-btn"
        >
          <el-icon><FolderAdd /></el-icon>
        </el-button>
      </el-tooltip>
    </div>

    <!-- 右键菜单 -->
    <el-dropdown
      ref="contextMenuRef"
      trigger="contextmenu"
      @command="handleContextMenuCommand"
    >
      <span></span>
      <template #dropdown>
        <el-dropdown-menu>
          <el-dropdown-item command="rename">重命名</el-dropdown-item>
          <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
        </el-dropdown-menu>
      </template>
    </el-dropdown>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  FolderOpened,
  Expand,
  Fold,
  DocumentAdd,
  FolderAdd,
  Folder,
  Document
} from '@element-plus/icons-vue'
import * as chatApi from '@/api/sugar/chat'
import type { WorkspaceTreeNode, TeamInfo } from '@/types/api'

// 组件事件
const emit = defineEmits<{
  'node-click': [node: WorkspaceTreeNode]
  'file-create': [node: WorkspaceTreeNode]
  'collapse-change': [collapsed: boolean]
  'team-change': [teamId: string]
}>()

// 响应式数据
const collapsed = ref(false)
const selectedTeam = ref('')
const teams = ref<TeamInfo[]>([])
const treeData = ref<WorkspaceTreeNode[]>([])
const contextMenuNode = ref<WorkspaceTreeNode | null>(null)
const treeRef = ref()
const contextMenuRef = ref()

// 树形组件配置
const treeProps = {
  children: 'children',
  label: 'name'
}

// 计算属性
const sidebarWidth = computed(() => collapsed.value ? '60px' : '280px')

// 初始化
onMounted(async () => {
  await loadTeams()
  if (teams.value.length > 0) {
    selectedTeam.value = teams.value[0].id
    await loadWorkspaceTree()
  }
})

// 加载团队列表
const loadTeams = async () => {
  try {
    const response = await chatApi.getSugarTeamsList()
    if (response.code === 0) {
      teams.value = response.data.list || []
    }
  } catch (error) {
    console.error('加载团队列表失败:', error)
    ElMessage.error('加载团队列表失败')
  }
}

// 加载工作空间树
const loadWorkspaceTree = async () => {
  if (!selectedTeam.value) return
  
  try {
    const response = await chatApi.getWorkspaceTree({ teamId: selectedTeam.value })
    if (response.code === 0) {
      treeData.value = response.data.tree || []
    }
  } catch (error) {
    console.error('加载工作空间树失败:', error)
    ElMessage.error('加载工作空间树失败')
  }
}

// 切换折叠状态
const toggleCollapse = () => {
  collapsed.value = !collapsed.value
  emit('collapse-change', collapsed.value)
}

// 处理团队切换
const handleTeamChange = async (teamId: string) => {
  selectedTeam.value = teamId
  await loadWorkspaceTree()
  emit('team-change', teamId)
}

// 处理节点点击
const handleNodeClick = (data: WorkspaceTreeNode) => {
  emit('node-click', data)
}

// 处理右键菜单
const handleContextMenu = (event: MouseEvent, data: WorkspaceTreeNode) => {
  event.preventDefault()
  contextMenuNode.value = data
  // 这里可以显示自定义右键菜单
}

// 处理右键菜单命令
const handleContextMenuCommand = async (command: string) => {
  if (!contextMenuNode.value) return

  switch (command) {
    case 'rename':
      await handleRename(contextMenuNode.value)
      break
    case 'delete':
      await handleDelete(contextMenuNode.value)
      break
  }
}

// 创建文件
const handleCreateFile = async () => {
  try {
    const { value: fileName } = await ElMessageBox.prompt('请输入文件名', '创建文件', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputPattern: /^[^\\/:*?"<>|]+$/,
      inputErrorMessage: '文件名不能包含特殊字符'
    })

    if (fileName && selectedTeam.value) {
      const response = await chatApi.createFolder({
        name: fileName,
        type: 'file',
        teamId: selectedTeam.value,
        content: {
          // 默认的空工作簿结构
          sheets: [{
            id: 'sheet1',
            name: 'Sheet1',
            cellData: {}
          }]
        }
      })

      if (response.code === 0) {
        await loadWorkspaceTree()
        ElMessage.success('文件创建成功')
        
        // 触发文件创建事件
        const newFile: WorkspaceTreeNode = {
          id: response.data.id,
          name: fileName,
          type: 'file',
          teamId: selectedTeam.value
        }
        emit('file-create', newFile)
      } else {
        ElMessage.error(response.msg || '创建文件失败')
      }
    }
  } catch {
    // 用户取消
  }
}

// 创建文件夹
const handleCreateFolder = async () => {
  try {
    const { value: folderName } = await ElMessageBox.prompt('请输入文件夹名', '创建文件夹', {
      confirmButtonText: '创建',
      cancelButtonText: '取消',
      inputPattern: /^[^\\/:*?"<>|]+$/,
      inputErrorMessage: '文件夹名不能包含特殊字符'
    })

    if (folderName && selectedTeam.value) {
      const response = await chatApi.createFolder({
        name: folderName,
        type: 'folder',
        teamId: selectedTeam.value
      })

      if (response.code === 0) {
        await loadWorkspaceTree()
        ElMessage.success('文件夹创建成功')
      } else {
        ElMessage.error(response.msg || '创建文件夹失败')
      }
    }
  } catch {
    // 用户取消
  }
}

// 重命名
const handleRename = async (node: WorkspaceTreeNode) => {
  try {
    const { value: newName } = await ElMessageBox.prompt('请输入新名称', '重命名', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValue: node.name,
      inputPattern: /^[^\\/:*?"<>|]+$/,
      inputErrorMessage: '名称不能包含特殊字符'
    })

    if (newName && newName !== node.name) {
      const response = await chatApi.renameItem({
        id: node.id,
        name: newName
      })

      if (response.code === 0) {
        await loadWorkspaceTree()
        ElMessage.success('重命名成功')
      } else {
        ElMessage.error(response.msg || '重命名失败')
      }
    }
  } catch {
    // 用户取消
  }
}

// 删除
const handleDelete = async (node: WorkspaceTreeNode) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 "${node.name}" 吗？此操作不可恢复。`,
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const response = await chatApi.deleteItem({ id: node.id })
    if (response.code === 0) {
      await loadWorkspaceTree()
      ElMessage.success('删除成功')
    } else {
      ElMessage.error(response.msg || '删除失败')
    }
  } catch {
    // 用户取消
  }
}

// 暴露方法
defineExpose({
  loadWorkspaceTree,
  toggleCollapse
})
</script>

<style scoped>
.sidebar {
  width: v-bind(sidebarWidth);
  height: 100%;
  background: #f8f9fa;
  border-right: 1px solid #e9ecef;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
}

.sidebar--collapsed {
  width: 60px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-bottom: 1px solid #e9ecef;
  background: white;
}

.sidebar-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #2c3e50;
}

.collapse-btn {
  padding: 4px;
  min-height: auto;
}

.team-selector {
  padding: 12px;
  border-bottom: 1px solid #e9ecef;
}

.sidebar-toolbar {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  border-bottom: 1px solid #e9ecef;
}

.file-tree {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.workspace-tree {
  background: transparent;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.node-icon {
  font-size: 16px;
  color: #6c757d;
}

.node-label {
  flex: 1;
  font-size: 14px;
  color: #495057;
}

.collapsed-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px 8px;
}

.collapsed-btn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.collapsed-btn:hover {
  background: #e9ecef;
}

/* 滚动条样式 */
.file-tree::-webkit-scrollbar {
  width: 6px;
}

.file-tree::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.file-tree::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.file-tree::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sidebar {
    position: absolute;
    left: 0;
    top: 0;
    z-index: 1000;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  }
  
  .sidebar--collapsed {
    width: 0;
    border-right: none;
  }
}
</style>