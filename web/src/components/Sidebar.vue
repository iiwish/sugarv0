<template>
  <div class="sidebar" :class="{ 'sidebar--collapsed': collapsed }">
    <!-- 侧边栏头部 - 团队选择器 -->
    <div class="sidebar-header">
      <div class="team-selector" v-if="!collapsed">
        <el-select
          v-model="selectedTeamId"
          placeholder="选择团队"
          @change="handleTeamChange"
          :loading="teamsLoading"
          class="team-select"
          size="small"
        >
          <el-option
            v-for="team in teams"
            :key="team.id"
            :label="team.teamName"
            :value="team.id"
          />
        </el-select>
      </div>
      <el-button
        type="text"
        @click="toggleCollapse"
        class="collapse-btn"
        :icon="collapsed ? Expand : Fold"
      />
    </div>

    <!-- 工作空间树 -->
    <div class="sidebar-content" v-show="!collapsed">
      <WorkspaceTree
        ref="workspaceTreeRef"
        :selected-team-id="selectedTeamId"
        :teams="teams"
        @node-click="handleNodeClick"
        @file-create="handleFileCreate"
        @team-change="handleTeamChange"
      />
    </div>

    <!-- 折叠状态下的快捷操作 -->
    <div class="sidebar-collapsed-actions" v-show="collapsed">
      <el-tooltip content="展开工作空间" placement="right">
        <el-button
          type="text"
          @click="toggleCollapse"
          class="collapsed-action-btn"
          :icon="Folder"
        />
      </el-tooltip>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Folder, Expand, Fold } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import WorkspaceTree from './WorkspaceTree.vue'
import { getSugarTeamsList } from '@/api/sugar/sugarTeams'
import type { WorkspaceTreeNode, TeamInfo, ApiResponse } from '@/types/api'

// 定义组件属性
interface Props {
  defaultCollapsed?: boolean
  width?: string
  collapsedWidth?: string
}

const props = withDefaults(defineProps<Props>(), {
  defaultCollapsed: false,
  width: '230px',
  collapsedWidth: '48px'
})

// 定义事件
const emit = defineEmits<{
  nodeClick: [data: WorkspaceTreeNode]
  fileCreate: [parentId?: string]
  collapseChange: [collapsed: boolean]
  teamChange: [teamId: string]
}>()

// 响应式数据
const collapsed = ref(props.defaultCollapsed)
const workspaceTreeRef = ref()
const selectedTeamId = ref('')
const teams = ref<TeamInfo[]>([])
const teamsLoading = ref(false)

// 加载团队列表
const loadTeams = async () => {
  try {
    teamsLoading.value = true
    const response = await getSugarTeamsList({}) as unknown as ApiResponse<{ list: TeamInfo[] }>
    if (response?.code === 0) {
      teams.value = response.data?.list || []
      // 如果没有选中的团队且有团队数据，默认选择第一个
      if (!selectedTeamId.value && teams.value.length > 0) {
        selectedTeamId.value = teams.value[0].id
        emit('teamChange', selectedTeamId.value)
      }
    } else {
      ElMessage.error(response?.msg || '加载团队列表失败')
    }
  } catch (error) {
    console.error('加载团队列表失败:', error)
    ElMessage.error('加载团队列表失败')
  } finally {
    teamsLoading.value = false
  }
}

// 处理团队切换
const handleTeamChange = () => {
  emit('teamChange', selectedTeamId.value)
}

// 切换折叠状态
const toggleCollapse = () => {
  collapsed.value = !collapsed.value
  emit('collapseChange', collapsed.value)
}

// 处理节点点击
const handleNodeClick = (data: WorkspaceTreeNode) => {
  emit('nodeClick', data)
}

// 处理文件创建
const handleFileCreate = (parentId?: string) => {
  emit('fileCreate', parentId)
}

// 刷新工作空间树
const refreshTree = () => {
  workspaceTreeRef.value?.refresh()
}

// 组件挂载时初始化
onMounted(async () => {
  await loadTeams()
})

// 暴露方法
defineExpose({
  refreshTree,
  toggleCollapse,
  loadTeams
})
</script>

<style scoped>
.sidebar {
  height: 100%;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  width: v-bind('props.width');
  min-width: v-bind('props.width');
}

.sidebar--collapsed {
  width: v-bind('props.collapsedWidth');
  min-width: v-bind('props.collapsedWidth');
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
}

.team-selector {
  flex: 1;
  margin-right: 8px;
}

.team-select {
  width: 100%;
}

.team-select :deep(.el-input__inner) {
  font-size: 14px;
  font-weight: 500;
}

.collapse-btn {
  padding: 4px;
  color: #606266;
  flex-shrink: 0;
}

.collapse-btn:hover {
  color: #409eff;
}

.sidebar-content {
  flex: 1;
  overflow: hidden;
}

.sidebar-collapsed-actions {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 0;
  gap: 8px;
}

.collapsed-action-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #606266;
  border-radius: 4px;
}

.collapsed-action-btn:hover {
  background: #f5f7fa;
  color: #409eff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    z-index: 1000;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.1);
  }
  
  .sidebar--collapsed {
    transform: translateX(-100%);
  }
}
</style>