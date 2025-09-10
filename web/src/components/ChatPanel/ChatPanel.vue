<template>
  <div class="chat-panel" :class="{ 'chat-panel--collapsed': collapsed }">
    <!-- 聊天面板头部 -->
    <div class="chat-header">
      <div class="chat-title" v-if="!collapsed">
        <el-icon class="chat-icon"><ChatDotRound /></el-icon>
        <span>AI 助手</span>
      </div>
      <div class="chat-actions">
        <el-tooltip :content="collapsed ? '展开聊天面板' : '折叠聊天面板'" placement="left">
          <el-button
            type="text"
            @click="toggleCollapse"
            class="collapse-btn"
            :icon="collapsed ? Expand : Fold"
          />
        </el-tooltip>
        <el-tooltip content="清空对话历史" placement="left" v-if="!collapsed">
          <el-button
            type="text"
            @click="handleClearHistory"
            class="clear-btn"
            :icon="Delete"
          />
        </el-tooltip>
      </div>
    </div>

    <!-- 聊天消息区域 -->
    <div class="chat-content" v-show="!collapsed">
      <div class="chat-messages" ref="messagesContainer">
        <div v-if="messages.length === 0" class="empty-state">
          <el-icon class="empty-icon"><ChatDotRound /></el-icon>
          <p class="empty-text">开始与AI助手对话</p>
          <div class="quick-actions">
            <el-button size="small" @click="handleQuickAction('analyze')">
              分析当前数据
            </el-button>
            <el-button size="small" @click="handleQuickAction('formula')">
              生成公式
            </el-button>
          </div>
        </div>
        
        <ChatMessage
          v-for="message in messages"
          :key="message.id"
          :message="message"
          @retry="handleRetry"
          @copy="handleCopy"
        />
        
        <!-- 加载状态 -->
        <div v-if="isLoading" class="loading-message">
          <div class="message-bubble ai-message">
            <div class="typing-indicator">
              <span></span>
              <span></span>
              <span></span>
            </div>
            <p class="loading-text">AI正在思考中...</p>
          </div>
        </div>
      </div>

      <!-- 上下文信息显示 -->
      <div class="context-info" v-if="contextInfo">
        <el-tag size="small" type="info" class="context-tag">
          <el-icon><Document /></el-icon>
          {{ contextInfo.fileName }}
        </el-tag>
        <el-tag size="small" type="success" class="context-tag" v-if="contextInfo.sheetName">
          <el-icon><Grid /></el-icon>
          {{ contextInfo.sheetName }}
        </el-tag>
        <el-tag size="small" type="warning" class="context-tag" v-if="contextInfo.selectedRange">
          <el-icon><Select /></el-icon>
          {{ contextInfo.selectedRange }}
        </el-tag>
      </div>

      <!-- 消息输入区域 -->
      <ChatInput
        @send="handleSendMessage"
        @quick-action="handleQuickAction"
        :disabled="isLoading"
        :context="contextInfo"
      />
    </div>

    <!-- 折叠状态下的快捷操作 -->
    <div class="collapsed-actions" v-show="collapsed">
      <el-tooltip content="快速分析" placement="left">
        <el-button
          type="text"
          @click="handleQuickAction('analyze')"
          class="collapsed-action-btn"
          :icon="TrendCharts"
        />
      </el-tooltip>
      <el-tooltip content="生成公式" placement="left">
        <el-button
          type="text"
          @click="handleQuickAction('formula')"
          class="collapsed-action-btn"
          :icon="EditPen"
        />
      </el-tooltip>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ChatDotRound,
  Expand,
  Fold,
  Delete,
  Document,
  Grid,
  Select,
  TrendCharts,
  EditPen
} from '@element-plus/icons-vue'
import ChatMessage from './ChatMessage.vue'
import ChatInput from './ChatInput.vue'
import { useChatStore } from '@/stores/chatStore'
import { useWorkspace } from '@/composables/useWorkspace'
import { useApp } from '@/composables/useApp'
import type { ChatMessage as ChatMessageType, ContextInfo } from '@/types/chat'

// 定义组件属性
interface Props {
  defaultCollapsed?: boolean
  width?: string
  collapsedWidth?: string
}

const props = withDefaults(defineProps<Props>(), {
  defaultCollapsed: false,
  width: '320px',
  collapsedWidth: '48px'
})

// 定义事件
const emit = defineEmits<{
  collapseChange: [collapsed: boolean]
}>()

// 响应式数据
const collapsed = ref(props.defaultCollapsed)
const messagesContainer = ref<HTMLElement>()
const isLoading = ref(false)

// 使用stores和composables
const chatStore = useChatStore()
const workspace = useWorkspace()
const app = useApp()

// 计算属性
const messages = computed(() => chatStore.messages)

// 获取当前上下文信息
const contextInfo = computed((): ContextInfo | null => {
  const currentNode = workspace.currentNode.value
  if (!currentNode || currentNode.type !== 'file') {
    return null
  }

  // 获取Univer核心插件来获取当前sheet信息
  const univerCorePlugin = app.pluginManager?.getPlugin('univer-core')
  let sheetName = ''
  let selectedRange = ''

  if (univerCorePlugin) {
    try {
      const workbook = univerCorePlugin.getCurrentWorkbook()
      if (workbook) {
        const activeSheet = workbook.getActiveSheet()
        if (activeSheet) {
          sheetName = activeSheet.getName()
          
          // 获取选中区域
          const selection = activeSheet.getSelection()
          if (selection) {
            const ranges = selection.getActiveRanges()
            if (ranges && ranges.length > 0) {
              selectedRange = ranges[0].toString()
            }
          }
        }
      }
    } catch (error) {
      console.warn('获取sheet信息失败:', error)
    }
  }

  return {
    fileName: currentNode.name,
    fileId: currentNode.id,
    sheetName,
    selectedRange
  }
})

// 方法
const toggleCollapse = () => {
  collapsed.value = !collapsed.value
  emit('collapseChange', collapsed.value)
}

const handleSendMessage = async (content: string, type: 'text' | 'analyze' | 'formula' = 'text') => {
  if (!content.trim()) return

  try {
    isLoading.value = true
    
    // 添加用户消息
    const userMessage: ChatMessageType = {
      id: Date.now().toString(),
      content,
      type: 'user',
      timestamp: new Date(),
      context: contextInfo.value
    }
    
    chatStore.addMessage(userMessage)
    await scrollToBottom()

    // 根据消息类型调用不同的AI服务
    let aiResponse: string
    
    if (type === 'analyze') {
      aiResponse = await chatStore.analyzeData(content, contextInfo.value)
    } else if (type === 'formula') {
      aiResponse = await chatStore.generateFormula(content, contextInfo.value)
    } else {
      aiResponse = await chatStore.sendMessage(content, contextInfo.value)
    }

    // 添加AI回复
    const aiMessage: ChatMessageType = {
      id: (Date.now() + 1).toString(),
      content: aiResponse,
      type: 'ai',
      timestamp: new Date(),
      context: contextInfo.value
    }
    
    chatStore.addMessage(aiMessage)
    await scrollToBottom()

  } catch (error) {
    console.error('发送消息失败:', error)
    ElMessage.error('发送消息失败: ' + (error as Error).message)
    
    // 添加错误消息
    const errorMessage: ChatMessageType = {
      id: (Date.now() + 1).toString(),
      content: '抱歉，我遇到了一些问题，请稍后重试。',
      type: 'ai',
      timestamp: new Date(),
      error: true,
      context: contextInfo.value
    }
    
    chatStore.addMessage(errorMessage)
    await scrollToBottom()
  } finally {
    isLoading.value = false
  }
}

const handleQuickAction = (action: string) => {
  const context = contextInfo.value
  
  if (!context) {
    ElMessage.warning('请先打开一个工作簿文件')
    return
  }

  let prompt = ''
  let type: 'text' | 'analyze' | 'formula' = 'text'

  switch (action) {
    case 'analyze':
      prompt = `请分析当前工作表"${context.sheetName}"中的数据`
      if (context.selectedRange) {
        prompt += `，重点关注选中区域 ${context.selectedRange}`
      }
      type = 'analyze'
      break
    case 'formula':
      prompt = '请根据当前数据生成一个有用的公式'
      if (context.selectedRange) {
        prompt += `，应用到选中区域 ${context.selectedRange}`
      }
      type = 'formula'
      break
  }

  if (prompt) {
    handleSendMessage(prompt, type)
  }
}

const handleRetry = (messageId: string) => {
  const message = messages.value.find(m => m.id === messageId)
  if (message && message.type === 'user') {
    handleSendMessage(message.content)
  }
}

const handleCopy = (content: string) => {
  navigator.clipboard.writeText(content).then(() => {
    ElMessage.success('已复制到剪贴板')
  }).catch(() => {
    ElMessage.error('复制失败')
  })
}

const handleClearHistory = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有对话历史吗？此操作不可撤销。',
      '确认清空',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    chatStore.clearMessages()
    ElMessage.success('对话历史已清空')
  } catch {
    // 用户取消操作
  }
}

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

// 监听消息变化，自动滚动到底部
watch(
  () => messages.value.length,
  () => {
    scrollToBottom()
  }
)

// 组件挂载时初始化
onMounted(() => {
  chatStore.initialize()
})

// 暴露方法
defineExpose({
  toggleCollapse,
  sendMessage: handleSendMessage,
  clearHistory: handleClearHistory
})
</script>

<style scoped>
.chat-panel {
  height: 100%;
  background: #fff;
  border-left: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  width: v-bind('props.width');
  min-width: v-bind('props.width');
}

.chat-panel--collapsed {
  width: v-bind('props.collapsedWidth');
  min-width: v-bind('props.collapsedWidth');
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  background: #fafafa;
  min-height: 48px;
}

.chat-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.chat-icon {
  color: #409eff;
}

.chat-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.collapse-btn,
.clear-btn {
  padding: 4px;
  color: #606266;
}

.collapse-btn:hover,
.clear-btn:hover {
  color: #409eff;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #909399;
  text-align: center;
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  color: #c0c4cc;
}

.empty-text {
  margin: 0 0 16px 0;
  font-size: 14px;
}

.quick-actions {
  display: flex;
  gap: 8px;
}

.loading-message {
  display: flex;
  justify-content: flex-start;
}

.message-bubble {
  max-width: 80%;
  padding: 12px 16px;
  border-radius: 12px;
  background: #f0f2f5;
  position: relative;
}

.ai-message {
  background: #f0f2f5;
}

.typing-indicator {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
}

.typing-indicator span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #409eff;
  animation: typing 1.4s infinite ease-in-out;
}

.typing-indicator span:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-indicator span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes typing {
  0%, 80%, 100% {
    transform: scale(0);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

.loading-text {
  margin: 0;
  font-size: 12px;
  color: #909399;
}

.context-info {
  padding: 8px 16px;
  border-top: 1px solid #e4e7ed;
  background: #f8f9fa;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.context-tag {
  display: flex;
  align-items: center;
  gap: 4px;
}

.collapsed-actions {
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

/* 滚动条样式 */
.chat-messages::-webkit-scrollbar {
  width: 4px;
}

.chat-messages::-webkit-scrollbar-track {
  background: transparent;
}

.chat-messages::-webkit-scrollbar-thumb {
  background: #c0c4cc;
  border-radius: 2px;
}

.chat-messages::-webkit-scrollbar-thumb:hover {
  background: #909399;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .chat-panel {
    position: fixed;
    right: 0;
    top: 0;
    z-index: 1000;
    box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1);
  }
  
  .chat-panel--collapsed {
    transform: translateX(100%);
  }
}
</style>