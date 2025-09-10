import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import * as chatApi from '@/api/sugar/chat'
import type { ChatMessage, ChatSession, FormulaQueryRequest, FormulaQueryResponse } from '@/types/api'

/**
 * 聊天功能组合式函数
 * 提供AI聊天和公式查询功能
 */
export function useChat() {
  // 当前聊天会话
  const currentSession = ref<ChatSession | null>(null)
  
  // 聊天消息列表
  const messages = ref<ChatMessage[]>([])
  
  // 聊天会话列表
  const sessions = ref<ChatSession[]>([])
  
  // 加载状态
  const isLoading = ref(false)
  
  // 发送状态
  const isSending = ref(false)
  
  // 错误信息
  const error = ref<string | null>(null)
  
  // 输入框内容
  const inputText = ref('')
  
  // 聊天设置
  const chatSettings = reactive({
    autoScroll: true,
    showTimestamp: true,
    enableSound: false,
    maxMessages: 100
  })

  // 计算属性
  const hasMessages = computed(() => messages.value.length > 0)
  const canSend = computed(() => inputText.value.trim().length > 0 && !isSending.value)
  const messageCount = computed(() => messages.value.length)
  const sessionCount = computed(() => sessions.value.length)

  /**
   * 初始化聊天功能
   */
  const initialize = async () => {
    try {
      isLoading.value = true
      error.value = null
      
      // 加载聊天设置
      loadChatSettings()
      
      // 加载聊天会话列表
      await loadSessions()
      
      // 如果有会话，加载最近的会话
      if (sessions.value.length > 0) {
        await loadSession(sessions.value[0].id)
      }
      
    } catch (err) {
      error.value = err instanceof Error ? err.message : '聊天功能初始化失败'
      ElMessage.error(error.value)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 创建新的聊天会话
   */
  const createSession = async (title?: string) => {
    try {
      const sessionTitle = title || `会话 ${sessions.value.length + 1}`
      
      // 调用API创建会话
      const response = await chatApi.createChatSession({ title: sessionTitle })
      
      const session: ChatSession = {
        id: response.data.sessionId,
        title: sessionTitle,
        messages: [],
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString()
      }
      
      sessions.value.unshift(session)
      currentSession.value = session
      messages.value = []
      
      saveSessions()
      ElMessage.success('新会话已创建')
      
      return session
    } catch (err) {
      const message = err instanceof Error ? err.message : '创建会话失败'
      ElMessage.error(message)
      throw err
    }
  }

  /**
   * 加载聊天会话
   */
  const loadSession = async (sessionId: string) => {
    try {
      isLoading.value = true
      
      const session = sessions.value.find(s => s.id === sessionId)
      if (!session) {
        throw new Error('会话不存在')
      }
      
      currentSession.value = session
      
      // 从API加载消息历史
      try {
        const response = await chatApi.getChatHistory({ sessionId })
        messages.value = response.data || []
      } catch (apiError) {
        // 如果API失败，尝试从本地存储加载
        const savedMessages = localStorage.getItem(`chat_messages_${sessionId}`)
        if (savedMessages) {
          messages.value = JSON.parse(savedMessages)
        } else {
          messages.value = []
        }
      }
      
    } catch (err) {
      error.value = err instanceof Error ? err.message : '加载会话失败'
      ElMessage.error(error.value)
    } finally {
      isLoading.value = false
    }
  }

  /**
   * 发送消息
   */
  const sendMessage = async (content: string, type: 'text' | 'formula' = 'text') => {
    if (!canSend.value) return
    
    try {
      isSending.value = true
      
      // 如果没有当前会话，创建一个
      if (!currentSession.value) {
        await createSession()
      }
      
      // 创建用户消息
      const userMessage: ChatMessage = {
        id: Date.now().toString(),
        content,
        sender: 'user',
        timestamp: Date.now(),
        type: type === 'formula' ? 'formula' : 'text'
      }
      
      // 添加到消息列表
      messages.value.push(userMessage)
      
      // 清空输入框
      inputText.value = ''
      
      // 保存消息
      saveMessages()
      
      // 发送到AI服务
      let response: string
      
      if (type === 'formula') {
        // 公式查询
        const queryRequest: FormulaQueryRequest = {
          query: content,
          context: JSON.stringify({
            sessionId: currentSession.value!.id,
            previousMessages: messages.value.slice(-5) // 最近5条消息作为上下文
          }),
          sessionId: currentSession.value!.id
        }
        
        const result = await chatApi.queryFormula(queryRequest)
        response = formatFormulaResponse(result.data)
      } else {
        // 普通聊天
        const result = await chatApi.sendChatMessage({
          content,
          sessionId: currentSession.value!.id,
          type: 'text'
        })
        response = result.data.content
      }
      
      // 创建AI回复消息
      const aiMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        content: response,
        sender: 'assistant',
        timestamp: Date.now(),
        type: type === 'formula' ? 'formula' : 'text'
      }
      
      // 添加到消息列表
      messages.value.push(aiMessage)
      
      // 更新会话信息
      currentSession.value!.updatedAt = new Date().toISOString()
      currentSession.value!.messages = messages.value
      
      // 保存数据
      saveMessages()
      saveSessions()
      
    } catch (err) {
      const message = err instanceof Error ? err.message : '发送消息失败'
      ElMessage.error(message)
      
      // 创建错误消息
      const errorMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        content: `抱歉，发生了错误：${message}`,
        sender: 'assistant',
        timestamp: Date.now(),
        type: 'text'
      }
      
      messages.value.push(errorMessage)
      saveMessages()
    } finally {
      isSending.value = false
    }
  }

  /**
   * 格式化公式查询响应
   */
  const formatFormulaResponse = (response: FormulaQueryResponse): string => {
    let result = ''
    
    if (response.formula) {
      result += `**建议公式：**\n\`${response.formula}\`\n\n`
    }
    
    if (response.explanation) {
      result += `**说明：**\n${response.explanation}\n\n`
    }
    
    if (response.examples && response.examples.length > 0) {
      result += `**示例：**\n${response.examples.join('\n')}\n\n`
    }
    
    if (response.confidence) {
      result += `**置信度：**\n${Math.round(response.confidence * 100)}%`
    }
    
    return result || '抱歉，无法为您的查询提供合适的公式建议。'
  }

  /**
   * 删除消息
   */
  const deleteMessage = (messageId: string) => {
    const index = messages.value.findIndex(msg => msg.id === messageId)
    if (index > -1) {
      messages.value.splice(index, 1)
      saveMessages()
      ElMessage.success('消息已删除')
    }
  }

  /**
   * 清空当前会话消息
   */
  const clearMessages = () => {
    messages.value = []
    if (currentSession.value) {
      currentSession.value.messages = []
      currentSession.value.updatedAt = new Date().toISOString()
      saveSessions()
    }
    saveMessages()
    ElMessage.success('消息已清空')
  }

  /**
   * 删除会话
   */
  const deleteSession = async (sessionId: string) => {
    try {
      // 调用API删除会话
      await chatApi.deleteChatSession({ sessionId })
      
      const index = sessions.value.findIndex(s => s.id === sessionId)
      if (index > -1) {
        sessions.value.splice(index, 1)
        
        // 如果删除的是当前会话，切换到其他会话或创建新会话
        if (currentSession.value?.id === sessionId) {
          if (sessions.value.length > 0) {
            await loadSession(sessions.value[0].id)
          } else {
            currentSession.value = null
            messages.value = []
          }
        }
        
        // 删除会话的消息数据
        localStorage.removeItem(`chat_messages_${sessionId}`)
        
        saveSessions()
        ElMessage.success('会话已删除')
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '删除会话失败'
      ElMessage.error(message)
    }
  }

  /**
   * 重命名会话
   */
  const renameSession = (sessionId: string, newTitle: string) => {
    const session = sessions.value.find(s => s.id === sessionId)
    if (session) {
      session.title = newTitle
      session.updatedAt = new Date().toISOString()
      saveSessions()
      ElMessage.success('会话已重命名')
    }
  }

  /**
   * 搜索消息
   */
  const searchMessages = (keyword: string) => {
    return messages.value.filter(msg => 
      msg.content.toLowerCase().includes(keyword.toLowerCase())
    )
  }

  /**
   * 导出聊天记录
   */
  const exportChat = (sessionId?: string) => {
    try {
      const targetSessionId = sessionId || currentSession.value?.id
      if (!targetSessionId) {
        throw new Error('没有可导出的会话')
      }
      
      const session = sessions.value.find(s => s.id === targetSessionId)
      const sessionMessages = sessionId 
        ? JSON.parse(localStorage.getItem(`chat_messages_${sessionId}`) || '[]')
        : messages.value
      
      const exportData = {
        session,
        messages: sessionMessages,
        exportTime: new Date().toISOString()
      }
      
      const blob = new Blob([JSON.stringify(exportData, null, 2)], {
        type: 'application/json'
      })
      
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `chat_${session?.title || 'session'}_${Date.now()}.json`
      a.click()
      
      URL.revokeObjectURL(url)
      ElMessage.success('聊天记录已导出')
    } catch (err) {
      const message = err instanceof Error ? err.message : '导出失败'
      ElMessage.error(message)
    }
  }

  /**
   * 加载聊天会话列表
   */
  const loadSessions = async () => {
    try {
      const saved = localStorage.getItem('chat_sessions')
      if (saved) {
        sessions.value = JSON.parse(saved)
      }
    } catch (err) {
      console.warn('加载聊天会话失败:', err)
    }
  }

  /**
   * 保存聊天会话列表
   */
  const saveSessions = () => {
    try {
      localStorage.setItem('chat_sessions', JSON.stringify(sessions.value))
    } catch (err) {
      console.warn('保存聊天会话失败:', err)
    }
  }

  /**
   * 保存当前会话的消息
   */
  const saveMessages = () => {
    if (!currentSession.value) return
    
    try {
      localStorage.setItem(
        `chat_messages_${currentSession.value.id}`,
        JSON.stringify(messages.value)
      )
    } catch (err) {
      console.warn('保存聊天消息失败:', err)
    }
  }

  /**
   * 加载聊天设置
   */
  const loadChatSettings = () => {
    try {
      const saved = localStorage.getItem('chat_settings')
      if (saved) {
        Object.assign(chatSettings, JSON.parse(saved))
      }
    } catch (err) {
      console.warn('加载聊天设置失败:', err)
    }
  }

  /**
   * 保存聊天设置
   */
  const saveChatSettings = () => {
    try {
      localStorage.setItem('chat_settings', JSON.stringify(chatSettings))
      ElMessage.success('设置已保存')
    } catch (err) {
      console.warn('保存聊天设置失败:', err)
    }
  }

  return {
    // 状态
    currentSession,
    messages,
    sessions,
    isLoading,
    isSending,
    error,
    inputText,
    chatSettings,
    
    // 计算属性
    hasMessages,
    canSend,
    messageCount,
    sessionCount,
    
    // 方法
    initialize,
    createSession,
    loadSession,
    sendMessage,
    deleteMessage,
    clearMessages,
    deleteSession,
    renameSession,
    searchMessages,
    exportChat,
    saveChatSettings
  }
}