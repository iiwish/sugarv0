<template>
  <div class="dashboard-container">
    <!-- 顶部导航栏 -->
    <header class="dashboard-header">
      <div class="header-left">
        <h1 class="app-title">Sugar 表格应用</h1>
      </div>
      <div class="header-right">
        <el-dropdown @command="handleUserCommand">
          <span class="user-info">
            <el-avatar :size="32" :src="userStore.userInfo?.avatar">
              {{ userStore.userInfo?.nickname?.charAt(0) || 'U' }}
            </el-avatar>
            <span class="username">{{ userStore.userInfo?.nickname || userStore.userInfo?.username }}</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item command="settings">设置</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <!-- 主内容区域 -->
    <main class="dashboard-main">
      <!-- 这里将放置 Sugar 表格应用的主要内容 -->
      <div class="welcome-section">
        <el-card class="welcome-card">
          <template #header>
            <div class="card-header">
              <span>欢迎使用 Sugar 表格应用</span>
            </div>
          </template>
          <div class="welcome-content">
            <p>这是一个基于 Univer 的智能数据分析平台</p>
            <p>支持自定义公式、AI 助手和实时协作功能</p>
            <el-button type="primary" @click="startDemo">开始体验</el-button>
          </div>
        </el-card>
      </div>

      <!-- 功能预览区域 -->
      <div class="features-section">
        <el-row :gutter="20">
          <el-col :span="8">
            <el-card class="feature-card">
              <template #header>
                <div class="card-header">
                  <el-icon><Document /></el-icon>
                  <span>智能表格</span>
                </div>
              </template>
              <p>基于 Univer 的强大表格引擎，支持复杂的数据处理和分析</p>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card class="feature-card">
              <template #header>
                <div class="card-header">
                  <el-icon><ChatDotRound /></el-icon>
                  <span>AI 助手</span>
                </div>
              </template>
              <p>集成 AI 助手，提供智能公式建议和数据分析支持</p>
            </el-card>
          </el-col>
          <el-col :span="8">
            <el-card class="feature-card">
              <template #header>
                <div class="card-header">
                  <el-icon><Share /></el-icon>
                  <span>实时协作</span>
                </div>
              </template>
              <p>支持多人实时协作编辑，团队工作更高效</p>
            </el-card>
          </el-col>
        </el-row>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowDown, Document, ChatDotRound, Share } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

// 处理用户下拉菜单命令
const handleUserCommand = async (command: string) => {
  switch (command) {
    case 'profile':
      ElMessage.info('个人资料功能开发中')
      break
    case 'settings':
      ElMessage.info('设置功能开发中')
      break
    case 'logout':
      try {
        await ElMessageBox.confirm(
          '确定要退出登录吗？',
          '确认退出',
          {
            confirmButtonText: '确定',
            cancelButtonText: '取消',
            type: 'warning'
          }
        )
        await userStore.logout()
        ElMessage.success('已退出登录')
        router.push('/login')
      } catch (error) {
        // 用户取消退出
      }
      break
  }
}

// 开始演示
const startDemo = () => {
  router.push('/sugar')
}
</script>

<style scoped>
.dashboard-container {
  min-height: 100vh;
  background-color: #f5f7fa;
  display: flex;
  flex-direction: column;
}

.dashboard-header {
  background: white;
  padding: 0 24px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  position: sticky;
  top: 0;
  z-index: 100;
}

.header-left {
  display: flex;
  align-items: center;
}

.app-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #2c3e50;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.user-info:hover {
  background-color: #f5f7fa;
}

.username {
  font-size: 14px;
  color: #606266;
}

.dashboard-main {
  flex: 1;
  padding: 24px;
  overflow: auto;
}

.welcome-section {
  margin-bottom: 24px;
}

.welcome-card {
  max-width: 800px;
  margin: 0 auto;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.welcome-content {
  text-align: center;
  padding: 20px 0;
}

.welcome-content p {
  margin: 8px 0;
  color: #606266;
  font-size: 16px;
}

.features-section {
  max-width: 1200px;
  margin: 0 auto;
}

.feature-card {
  height: 160px;
  transition: transform 0.2s, box-shadow 0.2s;
}

.feature-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.feature-card .card-header {
  color: #409eff;
}

.feature-card p {
  color: #606266;
  line-height: 1.6;
  margin: 16px 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .dashboard-header {
    padding: 0 16px;
  }
  
  .app-title {
    font-size: 18px;
  }
  
  .dashboard-main {
    padding: 16px;
  }
  
  .features-section .el-col {
    margin-bottom: 16px;
  }
}
</style>