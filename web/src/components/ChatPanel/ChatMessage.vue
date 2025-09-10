<template>
  <div class="chat-message" :class="messageClass">
    <div class="message-avatar" v-if="message.type === 'ai'">
      <el-icon class="avatar-icon"><Avatar /></el-icon>
    </div>
    
    <div class="message-content">
      <div class="message-bubble" :class="bubbleClass">
        <div class="message-header" v-if="showHeader">
          <span class="message-sender">{{ senderName }}</span>
          <span class="message-time">{{ formatTime(message.timestamp) }}</span>
        </div>
        
        <div class="message-text" v-html="formattedContent"></div>
        
        <!-- 上下文信息 -->
        <div class="message-context" v-if="message.context && showContext">
          <el-tag size="small" type="info">{{ message.context.fileName }}</el-tag>
          <el-tag size="small" type="success" v-if="message.context.sheetName">
            {{ message.context.sheetName }}
          </el-tag>
          <el-tag size="small" type="warning" v-if="message.context.selectedRange">
            {{ message.context.selectedRange }}
          </el-tag>
        </div>
        
        <!-- 元数据信息 -->
        <div class="message-metadata" v-if="message.metadata && showMetadata">
          <span v-if="message.metadata.agentName" class="metadata-item">
            <el-icon><User /></el-icon>
            {{ message.metadata.agentName }}
          </span>
          <span v-if="message.metadata.processingTime" class="metadata-item">
            <el-icon><Timer /></el-icon>
            {{ message.metadata.processingTime }}ms
          </span>
        </div>
      </div>
      
      <!-- 操作按钮 -->
      <div class="message-actions" v-if="showActions">
        <el-button
          type="text"
          size="small"
          @click="handleCopy"
          :icon="DocumentCopy"
          title="复制内容"
        />
        <el-button
          type="text"
          size="small"
          @click="handleRetry"
          :icon="RefreshRight"
          title="重新发送"
          v-if="message.type === 'user' || message.error"
        />
        <el-button
          type="text"
          size="small"
          @click="handleInsertFormula"
          :icon="EditPen"
          title="插入公式"
          v-if="isFormulaContent"
        />
      </div>
    </div>
    
    <div class="message-avatar" v-if="message.type === 'user'">
      <el-icon class="avatar-icon"><User /></el-icon>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Avatar,
  User,
  Timer,
  DocumentCopy,
  RefreshRight,
  EditPen
} from '@element-plus/icons-vue'
import type { ChatMessage } from '@/types/chat'

// 定义组件属性
interface Props {
  message: ChatMessage
  showHeader?: boolean
  showContext?: boolean
  showMetadata?: boolean
  showActions?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showHeader: true,
  showContext: true,
  showMetadata: true,
  showActions: true
})

// 定义事件
const emit = defineEmits<{
  retry: [messageId: string]
  copy: [content: string]
  insertFormula: [formula: string]
}>()

// 计算属性
const messageClass = computed(() => ({
  'chat-message--user': props.message.type === 'user',
  'chat-message--ai': props.message.type === 'ai',
  'chat-message--error': props.message.error
}))

const bubbleClass = computed(() => ({
  'message-bubble--user': props.message.type === 'user',
  'message-bubble--ai': props.message.type === 'ai',
  'message-bubble--error': props.message.error
}))

const senderName = computed(() => {
  if (props.message.type === 'user') {
    return '我'
  } else {
    return props.message.metadata?.agentName || 'AI助手'
  }
})

const formattedContent = computed(() => {
  let content = props.message.content
  
  // 处理换行
  content = content.replace(/\n/g, '<br>')
  
  // 处理代码块
  content = content.replace(/```([\s\S]*?)```/g, '<pre class="code-block">$1</pre>')
  
  // 处理行内代码
  content = content.replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
  
  // 处理公式（以=开头的内容）
  content = content.replace(/^(=.+)$/gm, '<code class="formula-code">$1</code>')
  
  // 处理粗体
  content = content.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
  
  // 处理斜体
  content = content.replace(/\*(.*?)\*/g, '<em>$1</em>')
  
  return content
})

const isFormulaContent = computed(() => {
  return props.message.content.includes('=') && 
         (props.message.content.includes('SUM') || 
          props.message.content.includes('AVERAGE') || 
          props.message.content.includes('COUNT') ||
          props.message.content.includes('AI.'))
})

// 方法
const formatTime = (timestamp: Date): string => {
  const now = new Date()
  const diff = now.getTime() - timestamp.getTime()
  
  if (diff < 60000) { // 1分钟内
    return '刚刚'
  } else if (diff < 3600000) { // 1小时内
    return `${Math.floor(diff / 60000)}分钟前`
  } else if (diff < 86400000) { // 24小时内
    return `${Math.floor(diff / 3600000)}小时前`
  } else {
    return timestamp.toLocaleDateString('zh-CN', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }
}

const handleCopy = () => {
  emit('copy', props.message.content)
}

const handleRetry = () => {
  emit('retry', props.message.id)
}

const handleInsertFormula = () => {
  // 提取公式内容
  const formulaMatch = props.message.content.match(/=(.*?)(?:\n|$)/g)
  if (formulaMatch && formulaMatch.length > 0) {
    const formula = formulaMatch[0].trim()
    emit('insertFormula', formula)
  } else {
    ElMessage.warning('未找到有效的公式内容')
  }
}
</script>

<style scoped>
.chat-message {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;
}

.chat-message--user {
  flex-direction: row-reverse;
}

.chat-message--ai {
  flex-direction: row;
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  background: #f0f2f5;
}

.chat-message--user .message-avatar {
  background: #409eff;
  color: white;
}

.chat-message--ai .message-avatar {
  background: #67c23a;
  color: white;
}

.avatar-icon {
  font-size: 16px;
}

.message-content {
  flex: 1;
  max-width: calc(100% - 44px);
}

.message-bubble {
  padding: 12px 16px;
  border-radius: 12px;
  position: relative;
  word-wrap: break-word;
  word-break: break-word;
}

.message-bubble--user {
  background: #409eff;
  color: white;
  margin-left: 20%;
}

.message-bubble--ai {
  background: #f0f2f5;
  color: #303133;
  margin-right: 20%;
}

.message-bubble--error {
  background: #fef0f0;
  border: 1px solid #fbc4c4;
  color: #f56c6c;
}

.message-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 12px;
  opacity: 0.8;
}

.message-sender {
  font-weight: 500;
}

.message-time {
  font-size: 11px;
}

.message-text {
  line-height: 1.5;
  font-size: 14px;
}

.message-text :deep(.code-block) {
  background: rgba(0, 0, 0, 0.1);
  padding: 8px 12px;
  border-radius: 4px;
  margin: 8px 0;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  overflow-x: auto;
}

.message-text :deep(.inline-code) {
  background: rgba(0, 0, 0, 0.1);
  padding: 2px 4px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.message-text :deep(.formula-code) {
  background: #e6f7ff;
  color: #1890ff;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  font-weight: 500;
}

.message-bubble--user .message-text :deep(.code-block),
.message-bubble--user .message-text :deep(.inline-code),
.message-bubble--user .message-text :deep(.formula-code) {
  background: rgba(255, 255, 255, 0.2);
  color: inherit;
}

.message-context {
  margin-top: 8px;
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.message-metadata {
  margin-top: 8px;
  display: flex;
  gap: 12px;
  font-size: 11px;
  opacity: 0.7;
}

.metadata-item {
  display: flex;
  align-items: center;
  gap: 2px;
}

.message-actions {
  margin-top: 4px;
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.chat-message:hover .message-actions {
  opacity: 1;
}

.chat-message--user .message-actions {
  justify-content: flex-end;
}

.chat-message--ai .message-actions {
  justify-content: flex-start;
}

.message-actions .el-button {
  padding: 4px;
  color: #909399;
}

.message-actions .el-button:hover {
  color: #409eff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .message-bubble--user,
  .message-bubble--ai {
    margin-left: 0;
    margin-right: 0;
  }
  
  .message-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
  }
}
</style>