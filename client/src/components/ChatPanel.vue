<template>
  <div class="chat-panel" :class="{ 'chat-panel--collapsed': collapsed }">
    <!-- 聊天面板头部 -->
    <div class="chat-header">
      <div class="chat-title" v-if="!collapsed">
        <el-icon><ChatDotRound /></el-icon>
        <span>AI助手</span>
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

    <!-- 聊天内容区域 -->
    <div class="chat-content" v-if="!collapsed">
      <!-- 会话列表 -->
      <div class="session-list" v-if="showSessionList">
        <div class="session-header">
          <span>聊天会话</span>
          <el-button
            type="primary"
            size="small"
            @click="createNewSession"
            :icon="Plus"
          >
            新建
          </el-button>
        </div>
        <div class="sessions">
          <div
            v-for="session in chat.sessions.value"
            :key="session.id"
            class="session-item"
            :class="{ active: chat.currentSession.value?.id === session.id }"
            @click="loadSession(session.id)"
          >
            <div class="session-info">
              <span class="session-title">{{ session.title }}</span>
              <span class="session-time">{{ formatTime(session.updatedAt) }}</span>
            </div>
            <el-dropdown @command="handleSessionCommand" trigger="click">
              <el-button type="text" size="small">
                <el-icon><MoreFilled /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="{ action: 'rename', sessionId: session.id }">
                    重命名
                  </el-dropdown-item>
                  <el-dropdown-item :command="{ action: 'delete', sessionId: session.id }" divided>
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>

      <!-- 聊天消息区域 -->
      <div class="chat-messages" v-else>
        <div class="messages-header">
          <el-button
            type="text"
            size="small"
            @click="showSessionList = true"
            :icon="ArrowLeft"
          >
            返回会话列表
          </el-button>
          <span v-if="chat.currentSession.value">{{ chat.currentSession.value.title }}</span>
          <el-button
            type="text"
            size="small"
            @click="clearMessages"
            :icon="Delete"
          >
            清空
          </el-button>
        </div>

        <!-- 消息列表 -->
        <div class="message-list" ref="messageListRef">
          <div
            v-for="message in chat.messages.value"
            :key="message.id"
            class="message-item"
            :class="`message-${message.sender}`"
          >
            <div class="message-avatar">
              <el-icon v-if="message.sender === 'user'">
                <User />
              </el-icon>
              <el-icon v-else>
                <Service />
              </el-icon>
            </div>
            <div class="message-content">
              <div class="message-text" v-html="formatMessage(message.content)"></div>
              <div class="message-time">{{ formatTime(message.timestamp) }}</div>
            </div>
          </div>
          
          <!-- 加载状态 -->
          <div v-if="chat.isSending.value" class="message-item message-assistant">
            <div class="message-avatar">
              <el-icon><Service /></el-icon>
            </div>
            <div class="message-content">
              <div class="typing-indicator">
                <span></span>
                <span></span>
                <span></span>
              </div>
            </div>
          </div>
        </div>

        <!-- 输入区域 -->
        <div class="chat-input">
          <div class="input-toolbar">
            <el-button-group size="small">
              <el-button
                :type="inputMode === 'text' ? 'primary' : 'default'"
                @click="inputMode = 'text'"
              >
                普通聊天
              </el-button>
              <el-button
                :type="inputMode === 'formula' ? 'primary' : 'default'"
                @click="inputMode = 'formula'"
              >
                公式查询
              </el-button>
            </el-button-group>
          </div>
          <div class="input-area">
            <el-input
              v-model="chat.inputText.value"
              type="textarea"
              :rows="3"
              :placeholder="inputMode === 'formula' ? '请描述您需要的公式功能...' : '请输入您的问题...'"
              @keydown.ctrl.enter="sendMessage"
              @keydown.meta.enter="sendMessage"
            />
            <el-button
              type="primary"
              @click="sendMessage"
              :loading="chat.isSending.value"
              :disabled="!chat.canSend.value"
              class="send-btn"
            >
              发送
            </el-button>
          </div>
        </div>
      </div>
    </div>

    <!-- 折叠状态下的快捷按钮 -->
    <div class="collapsed-actions" v-if="collapsed">
      <el-tooltip content="AI助手" placement="left">
        <el-button
          type="text"
          size="small"
          @click="toggleCollapse"
          class="collapsed-btn"
        >
          <el-icon><ChatDotRound /></el-icon>
        </el-button>
      </el-tooltip>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ChatDotRound,
  Expand,
  Fold,
  Plus,
  MoreFilled,
  ArrowLeft,
  Delete,
  User,
  Service
} from '@element-plus/icons-vue'
import { useChat } from '@/composables'

// 组件属性
interface Props {
  defaultCollapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  defaultCollapsed: false
})

// 组件事件
const emit = defineEmits<{
  'collapse-change': [collapsed: boolean]
}>()

// 响应式数据
const collapsed = ref(props.defaultCollapsed)
const showSessionList = ref(true)
const inputMode = ref<'text' | 'formula'>('text')
const messageListRef = ref<HTMLElement>()

// 使用聊天功能
const chat = useChat()

// 计算属性
const panelWidth = computed(() => collapsed.value ? '60px' : '320px')

// 初始化
onMounted(async () => {
  await chat.initialize()
  
  // 如果有会话，显示消息列表
  if (chat.currentSession.value) {
    showSessionList.value = false
  }
})

// 切换折叠状态
const toggleCollapse = () => {
  collapsed.value = !collapsed.value
  emit('collapse-change', collapsed.value)
}

// 创建新会话
const createNewSession = async () => {
  try {
    await chat.createSession()
    showSessionList.value = false
    scrollToBottom()
  } catch (error) {
    console.error('创建会话失败:', error)
  }
}

// 加载会话
const loadSession = async (sessionId: string) => {
  try {
    await chat.loadSession(sessionId)
    showSessionList.value = false
    await nextTick()
    scrollToBottom()
  } catch (error) {
    console.error('加载会话失败:', error)
  }
}

// 处理会话命令
const handleSessionCommand = async (command: { action: string, sessionId: string }) => {
  const { action, sessionId } = command
  
  switch (action) {
    case 'rename':
      await renameSession(sessionId)
      break
    case 'delete':
      await deleteSession(sessionId)
      break
  }
}

// 重命名会话
const renameSession = async (sessionId: string) => {
  try {
    const session = chat.sessions.value.find(s => s.id === sessionId)
    if (!session) return
    
    const { value: newTitle } = await ElMessageBox.prompt('请输入新的会话名称', '重命名会话', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputValue: session.title
    })
    
    if (newTitle && newTitle !== session.title) {
      chat.renameSession(sessionId, newTitle)
    }
  } catch {
    // 用户取消
  }
}

// 删除会话
const deleteSession = async (sessionId: string) => {
  try {
    await ElMessageBox.confirm('确定要删除这个会话吗？', '确认删除', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await chat.deleteSession(sessionId)
    
    // 如果删除的是当前会话，返回会话列表
    if (chat.currentSession.value?.id === sessionId) {
      showSessionList.value = true
    }
  } catch {
    // 用户取消
  }
}

// 发送消息
const sendMessage = async () => {
  if (!chat.canSend.value) return
  
  try {
    await chat.sendMessage(chat.inputText.value, inputMode.value)
    await nextTick()
    scrollToBottom()
  } catch (error) {
    console.error('发送消息失败:', error)
  }
}

// 清空消息
const clearMessages = async () => {
  try {
    await ElMessageBox.confirm('确定要清空当前会话的所有消息吗？', '确认清空', {
      confirmButtonText: '清空',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    chat.clearMessages()
  } catch {
    // 用户取消
  }
}

// 滚动到底部
const scrollToBottom = () => {
  if (messageListRef.value) {
    messageListRef.value.scrollTop = messageListRef.value.scrollHeight
  }
}

// 格式化时间
const formatTime = (timestamp: string | number) => {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  if (diff < 60000) { // 1分钟内
    return '刚刚'
  } else if (diff < 3600000) { // 1小时内
    return `${Math.floor(diff / 60000)}分钟前`
  } else if (diff < 86400000) { // 1天内
    return `${Math.floor(diff / 3600000)}小时前`
  } else {
    return date.toLocaleDateString()
  }
}

// 格式化消息内容
const formatMessage = (content: string) => {
  // 简单的 Markdown 渲染
  return content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/`(.*?)`/g, '<code>$1</code>')
    .replace(/\n/g, '<br>')
}

// 暴露方法
defineExpose({
  toggleCollapse,
  createNewSession,
  scrollToBottom
})
</script>

<style scoped>
.chat-panel {
  width: v-bind(panelWidth);
  height: 100%;
  background: #f8f9fa;
  border-left: 1px solid #e9ecef;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  overflow: hidden;
}

.chat-panel--collapsed {
  width: 60px;
}

.chat-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-bottom: 1px solid #e9ecef;
  background: white;
}

.chat-title {
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

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 会话列表样式 */
.session-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.session-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-bottom: 1px solid #e9ecef;
  background: white;
}

.sessions {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  margin-bottom: 8px;
  background: white;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.session-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 4px rgba(64, 158, 255, 0.1);
}

.session-item.active {
  border-color: #409eff;
  background: #f0f8ff;
}

.session-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.session-title {
  font-weight: 500;
  color: #2c3e50;
  font-size: 14px;
}

.session-time {
  font-size: 12px;
  color: #6c757d;
}

/* 聊天消息样式 */
.chat-messages {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.messages-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  border-bottom: 1px solid #e9ecef;
  background: white;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-item {
  display: flex;
  gap: 12px;
}

.message-user {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #409eff;
  color: white;
  flex-shrink: 0;
}

.message-user .message-avatar {
  background: #67c23a;
}

.message-content {
  flex: 1;
  max-width: calc(100% - 44px);
}

.message-text {
  background: white;
  padding: 12px 16px;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  word-wrap: break-word;
  line-height: 1.5;
}

.message-user .message-text {
  background: #409eff;
  color: white;
}

.message-time {
  font-size: 12px;
  color: #6c757d;
  margin-top: 4px;
  text-align: right;
}

.message-user .message-time {
  text-align: left;
}

/* 输入打字指示器 */
.typing-indicator {
  display: flex;
  gap: 4px;
  padding: 12px 16px;
}

.typing-indicator span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #6c757d;
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

/* 输入区域样式 */
.chat-input {
  border-top: 1px solid #e9ecef;
  background: white;
}

.input-toolbar {
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
}

.input-area {
  padding: 12px;
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.input-area .el-textarea {
  flex: 1;
}

.send-btn {
  height: 40px;
}

/* 折叠状态样式 */
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
.sessions::-webkit-scrollbar,
.message-list::-webkit-scrollbar {
  width: 6px;
}

.sessions::-webkit-scrollbar-track,
.message-list::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.sessions::-webkit-scrollbar-thumb,
.message-list::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.sessions::-webkit-scrollbar-thumb:hover,
.message-list::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .chat-panel {
    position: absolute;
    right: 0;
    top: 0;
    z-index: 1000;
    box-shadow: -2px 0 8px rgba(0, 0, 0, 0.1);
  }
  
  .chat-panel--collapsed {
    width: 0;
    border-left: none;
  }
}
</style>