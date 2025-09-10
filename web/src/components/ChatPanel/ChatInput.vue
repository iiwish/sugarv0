<template>
  <div class="chat-input">
    <!-- 快捷操作按钮 -->
    <div class="quick-actions" v-if="showQuickActions && !inputFocused">
      <el-button
        size="small"
        type="primary"
        plain
        @click="handleQuickAction('analyze')"
        :icon="TrendCharts"
        :disabled="disabled || !hasContext"
      >
        分析数据
      </el-button>
      <el-button
        size="small"
        type="success"
        plain
        @click="handleQuickAction('formula')"
        :icon="EditPen"
        :disabled="disabled || !hasContext"
      >
        生成公式
      </el-button>
    </div>

    <!-- 输入区域 -->
    <div class="input-container">
      <el-input
        ref="inputRef"
        v-model="inputText"
        type="textarea"
        :placeholder="placeholder"
        :disabled="disabled"
        :autosize="{ minRows: 1, maxRows: 4 }"
        @keydown="handleKeyDown"
        @focus="handleFocus"
        @blur="handleBlur"
        class="message-input"
      />
      
      <!-- 输入工具栏 -->
      <div class="input-toolbar">
        <div class="toolbar-left">
          <!-- 文件上传 -->
          <el-tooltip content="上传文件" placement="top">
            <el-button
              type="text"
              size="small"
              @click="handleFileUpload"
              :icon="Paperclip"
              :disabled="disabled"
            />
          </el-tooltip>
          
          <!-- 表情 -->
          <el-tooltip content="插入表情" placement="top">
            <el-button
              type="text"
              size="small"
              @click="handleEmojiPicker"
              :icon="ChatDotRound"
              :disabled="disabled"
            />
          </el-tooltip>
          
          <!-- 字数统计 -->
          <span class="char-count" v-if="inputText.length > 0">
            {{ inputText.length }}/{{ maxLength }}
          </span>
        </div>
        
        <div class="toolbar-right">
          <!-- 清空输入 -->
          <el-tooltip content="清空输入" placement="top" v-if="inputText.length > 0">
            <el-button
              type="text"
              size="small"
              @click="handleClear"
              :icon="Delete"
              :disabled="disabled"
            />
          </el-tooltip>
          
          <!-- 发送按钮 -->
          <el-button
            type="primary"
            size="small"
            @click="handleSend"
            :disabled="disabled || !canSend"
            :loading="disabled"
            :icon="Position"
          >
            发送
          </el-button>
        </div>
      </div>
    </div>

    <!-- 输入提示 -->
    <div class="input-hints" v-if="showHints">
      <div class="hint-item" v-for="hint in hints" :key="hint.text" @click="handleHintClick(hint)">
        <el-icon><Lightning /></el-icon>
        <span>{{ hint.text }}</span>
      </div>
    </div>

    <!-- 隐藏的文件输入 -->
    <input
      ref="fileInputRef"
      type="file"
      style="display: none"
      @change="handleFileChange"
      accept=".xlsx,.xls,.csv,.txt"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  TrendCharts,
  EditPen,
  Paperclip,
  ChatDotRound,
  Delete,
  Position,
  Lightning
} from '@element-plus/icons-vue'
import type { ContextInfo } from '@/types/chat'

// 定义组件属性
interface Props {
  disabled?: boolean
  placeholder?: string
  maxLength?: number
  showQuickActions?: boolean
  context?: ContextInfo | null
}

const props = withDefaults(defineProps<Props>(), {
  disabled: false,
  placeholder: '输入消息...',
  maxLength: 2000,
  showQuickActions: true,
  context: null
})

// 定义事件
const emit = defineEmits<{
  send: [content: string, type?: 'text' | 'analyze' | 'formula']
  quickAction: [action: string]
  fileUpload: [file: File]
}>()

// 响应式数据
const inputText = ref('')
const inputFocused = ref(false)
const inputRef = ref()
const fileInputRef = ref()

// 计算属性
const hasContext = computed(() => {
  return props.context && props.context.fileName
})

const canSend = computed(() => {
  return inputText.value.trim().length > 0 && inputText.value.length <= props.maxLength
})

const showHints = computed(() => {
  return inputFocused.value && inputText.value.length === 0 && hasContext.value
})

const hints = computed(() => {
  if (!hasContext.value) return []
  
  const contextHints = [
    { text: '分析当前工作表的数据趋势', action: 'analyze' },
    { text: '生成求和公式', action: 'formula' },
    { text: '解释选中区域的数据', action: 'explain' },
    { text: '查找数据中的异常值', action: 'anomaly' }
  ]
  
  if (props.context?.selectedRange) {
    contextHints.unshift({
      text: `分析选中区域 ${props.context.selectedRange}`,
      action: 'analyze-range'
    })
  }
  
  return contextHints
})

// 方法
const handleSend = () => {
  if (!canSend.value) return
  
  const content = inputText.value.trim()
  emit('send', content)
  inputText.value = ''
  
  // 重新聚焦输入框
  nextTick(() => {
    inputRef.value?.focus()
  })
}

const handleQuickAction = (action: string) => {
  emit('quickAction', action)
}

const handleKeyDown = (event: KeyboardEvent) => {
  // Ctrl/Cmd + Enter 发送消息
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    handleSend()
  }
  
  // Shift + Enter 换行（默认行为）
  if (event.shiftKey && event.key === 'Enter') {
    return
  }
  
  // Enter 发送消息（可配置）
  if (event.key === 'Enter' && !event.shiftKey && !event.ctrlKey && !event.metaKey) {
    event.preventDefault()
    handleSend()
  }
}

const handleFocus = () => {
  inputFocused.value = true
}

const handleBlur = () => {
  // 延迟设置，避免点击提示时立即隐藏
  setTimeout(() => {
    inputFocused.value = false
  }, 200)
}

const handleClear = () => {
  inputText.value = ''
  inputRef.value?.focus()
}

const handleFileUpload = () => {
  fileInputRef.value?.click()
}

const handleFileChange = (event: Event) => {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  
  if (file) {
    // 检查文件大小（限制为10MB）
    if (file.size > 10 * 1024 * 1024) {
      ElMessage.error('文件大小不能超过10MB')
      return
    }
    
    emit('fileUpload', file)
    
    // 清空文件输入
    target.value = ''
  }
}

const handleEmojiPicker = () => {
  // 简单的表情插入
  const emojis = ['😊', '👍', '❤️', '🎉', '💡', '🔥', '✨', '📊', '📈', '💯']
  const randomEmoji = emojis[Math.floor(Math.random() * emojis.length)]
  inputText.value += randomEmoji
  inputRef.value?.focus()
}

const handleHintClick = (hint: { text: string; action: string }) => {
  inputText.value = hint.text
  
  // 根据提示类型发送不同类型的消息
  let messageType: 'text' | 'analyze' | 'formula' = 'text'
  if (hint.action.includes('analyze')) {
    messageType = 'analyze'
  } else if (hint.action.includes('formula')) {
    messageType = 'formula'
  }
  
  emit('send', hint.text, messageType)
  inputText.value = ''
}

// 监听上下文变化，更新占位符
watch(
  () => props.context,
  (newContext) => {
    if (newContext) {
      // 可以根据上下文动态更新占位符
    }
  },
  { deep: true }
)

// 暴露方法
defineExpose({
  focus: () => inputRef.value?.focus(),
  clear: () => inputText.value = '',
  setText: (text: string) => inputText.value = text
})
</script>

<style scoped>
.chat-input {
  border-top: 1px solid #e4e7ed;
  background: #fff;
}

.quick-actions {
  padding: 8px 16px;
  display: flex;
  gap: 8px;
  border-bottom: 1px solid #f0f0f0;
}

.input-container {
  padding: 12px 16px;
}

.message-input {
  margin-bottom: 8px;
}

.message-input :deep(.el-textarea__inner) {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 8px 12px;
  font-size: 14px;
  line-height: 1.5;
  resize: none;
  transition: border-color 0.2s;
}

.message-input :deep(.el-textarea__inner):focus {
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.1);
}

.input-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.char-count {
  font-size: 12px;
  color: #909399;
  margin-left: 8px;
}

.input-hints {
  padding: 8px 16px;
  border-top: 1px solid #f0f0f0;
  background: #fafafa;
}

.hint-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
  color: #606266;
  transition: all 0.2s;
  margin-bottom: 4px;
}

.hint-item:last-child {
  margin-bottom: 0;
}

.hint-item:hover {
  background: #e6f7ff;
  color: #409eff;
}

.hint-item .el-icon {
  font-size: 12px;
  color: #409eff;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .quick-actions {
    flex-wrap: wrap;
    gap: 4px;
  }
  
  .quick-actions .el-button {
    flex: 1;
    min-width: 0;
  }
  
  .input-toolbar {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .toolbar-left,
  .toolbar-right {
    justify-content: space-between;
  }
}

/* 动画效果 */
.quick-actions {
  animation: slideDown 0.3s ease-out;
}

.input-hints {
  animation: slideDown 0.2s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>