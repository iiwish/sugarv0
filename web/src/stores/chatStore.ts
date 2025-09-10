import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { ChatMessage, ChatSession, ContextInfo, AIResponse } from '@/types/chat'
import { aiFormulaManager } from '@/plugins/custom-formulas/formulas/ai'

export const useChatStore = defineStore('chat', () => {
  // 状态
  const messages = ref<ChatMessage[]>([])
  const sessions = ref<ChatSession[]>([])
  const currentSessionId = ref<string | null>(null)
  const isLoading = ref(false)
  const config = ref({
    maxMessages: 100,
    autoSave: true,
    enableContext: true,
    defaultAgent: 'AI助手'
  })

  // 计算属性
  const currentSession = computed(() => {
    return sessions.value.find(s => s.id === currentSessionId.value) || null
  })

  const hasMessages = computed(() => messages.value.length > 0)

  const lastMessage = computed(() => {
    return messages.value[messages.value.length - 1] || null
  })

  // 方法
  const initialize = () => {
    loadFromStorage()
    
    // 如果没有当前会话，创建一个新的
    if (!currentSessionId.value && sessions.value.length === 0) {
      createNewSession()
    }
  }

  const createNewSession = (title?: string): string => {
    const sessionId = Date.now().toString()
    const session: ChatSession = {
      id: sessionId,
      title: title || '新对话',
      messages: [],
      createdAt: new Date(),
      updatedAt: new Date()
    }
    
    sessions.value.push(session)
    currentSessionId.value = sessionId
    messages.value = []
    
    saveToStorage()
    return sessionId
  }

  const switchSession = (sessionId: string) => {
    const session = sessions.value.find(s => s.id === sessionId)
    if (session) {
      currentSessionId.value = sessionId
      messages.value = [...session.messages]
    }
  }

  const deleteSession = (sessionId: string) => {
    const index = sessions.value.findIndex(s => s.id === sessionId)
    if (index > -1) {
      sessions.value.splice(index, 1)
      
      if (currentSessionId.value === sessionId) {
        if (sessions.value.length > 0) {
          switchSession(sessions.value[0].id)
        } else {
          createNewSession()
        }
      }
      
      saveToStorage()
    }
  }

  const addMessage = (message: ChatMessage) => {
    messages.value.push(message)
    
    // 更新当前会话
    if (currentSessionId.value) {
      const session = sessions.value.find(s => s.id === currentSessionId.value)
      if (session) {
        session.messages = [...messages.value]
        session.updatedAt = new Date()
        
        // 自动更新会话标题（使用第一条用户消息）
        if (session.title === '新对话' && message.type === 'user') {
          session.title = message.content.slice(0, 20) + (message.content.length > 20 ? '...' : '')
        }
      }
    }
    
    // 限制消息数量
    if (messages.value.length > config.value.maxMessages) {
      messages.value = messages.value.slice(-config.value.maxMessages)
    }
    
    if (config.value.autoSave) {
      saveToStorage()
    }
  }

  const clearMessages = () => {
    messages.value = []
    
    if (currentSessionId.value) {
      const session = sessions.value.find(s => s.id === currentSessionId.value)
      if (session) {
        session.messages = []
        session.updatedAt = new Date()
      }
    }
    
    saveToStorage()
  }

  const removeMessage = (messageId: string) => {
    const index = messages.value.findIndex(m => m.id === messageId)
    if (index > -1) {
      messages.value.splice(index, 1)
      
      // 更新会话
      if (currentSessionId.value) {
        const session = sessions.value.find(s => s.id === currentSessionId.value)
        if (session) {
          session.messages = [...messages.value]
          session.updatedAt = new Date()
        }
      }
      
      saveToStorage()
    }
  }

  // AI交互方法
  const sendMessage = async (content: string, context?: ContextInfo | null): Promise<string> => {
    try {
      isLoading.value = true
      
      // 构建请求数据
      const requestData = {
        agentName: config.value.defaultAgent,
        description: content,
        dataRange: context?.selectedRange || ''
      }
      
      // 调用AI服务
      const result = await aiFormulaManager.executeAIRequest(
        '/api/sugarFormulaQuery/executeAiFetch',
        requestData
      )
      
      if (result && result.data && result.data.text) {
        return result.data.text
      } else if (result && result.msg) {
        return result.msg
      } else {
        throw new Error('AI响应格式错误')
      }
    } catch (error) {
      console.error('发送消息失败:', error)
      throw new Error('发送消息失败: ' + (error as Error).message)
    } finally {
      isLoading.value = false
    }
  }

  const analyzeData = async (description: string, context?: ContextInfo | null): Promise<string> => {
    try {
      isLoading.value = true
      
      if (!context || !context.fileName) {
        throw new Error('需要打开工作簿文件才能进行数据分析')
      }
      
      // 获取当前工作表数据
      let dataRange = ''
      if (context.selectedRange) {
        dataRange = `选中区域: ${context.selectedRange}`
      } else if (context.sheetName) {
        dataRange = `工作表: ${context.sheetName}`
      }
      
      const analysisPrompt = `请分析以下数据：
文件: ${context.fileName}
${dataRange}

分析要求: ${description}

请提供详细的数据分析报告，包括趋势、模式和洞察。`
      
      return await sendMessage(analysisPrompt, context)
    } catch (error) {
      console.error('数据分析失败:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  const generateFormula = async (description: string, context?: ContextInfo | null): Promise<string> => {
    try {
      isLoading.value = true
      
      if (!context || !context.fileName) {
        throw new Error('需要打开工作簿文件才能生成公式')
      }
      
      const formulaPrompt = `请根据以下要求生成Excel公式：
文件: ${context.fileName}
工作表: ${context.sheetName || '当前工作表'}
${context.selectedRange ? `目标区域: ${context.selectedRange}` : ''}

公式要求: ${description}

请提供完整的公式代码，并解释公式的作用和使用方法。`
      
      return await sendMessage(formulaPrompt, context)
    } catch (error) {
      console.error('公式生成失败:', error)
      throw error
    } finally {
      isLoading.value = false
    }
  }

  // 存储相关方法
  const saveToStorage = () => {
    try {
      const data = {
        sessions: sessions.value,
        currentSessionId: currentSessionId.value,
        config: config.value
      }
      localStorage.setItem('chat_store', JSON.stringify(data))
    } catch (error) {
      console.warn('保存聊天数据失败:', error)
    }
  }

  const loadFromStorage = () => {
    try {
      const saved = localStorage.getItem('chat_store')
      if (saved) {
        const data = JSON.parse(saved)
        
        // 恢复会话数据
        if (data.sessions && Array.isArray(data.sessions)) {
          sessions.value = data.sessions.map((session: any) => ({
            ...session,
            createdAt: new Date(session.createdAt),
            updatedAt: new Date(session.updatedAt),
            messages: session.messages.map((msg: any) => ({
              ...msg,
              timestamp: new Date(msg.timestamp)
            }))
          }))
        }
        
        // 恢复当前会话
        if (data.currentSessionId) {
          currentSessionId.value = data.currentSessionId
          const currentSession = sessions.value.find(s => s.id === data.currentSessionId)
          if (currentSession) {
            messages.value = [...currentSession.messages]
          }
        }
        
        // 恢复配置
        if (data.config) {
          config.value = { ...config.value, ...data.config }
        }
      }
    } catch (error) {
      console.warn('加载聊天数据失败:', error)
    }
  }

  const exportChat = (sessionId?: string): string => {
    const session = sessionId 
      ? sessions.value.find(s => s.id === sessionId)
      : currentSession.value
    
    if (!session) {
      throw new Error('未找到要导出的会话')
    }
    
    const exportData = {
      title: session.title,
      createdAt: session.createdAt,
      messages: session.messages.map(msg => ({
        type: msg.type,
        content: msg.content,
        timestamp: msg.timestamp,
        context: msg.context
      }))
    }
    
    return JSON.stringify(exportData, null, 2)
  }

  const importChat = (data: string): string => {
    try {
      const importData = JSON.parse(data)
      
      const sessionId = Date.now().toString()
      const session: ChatSession = {
        id: sessionId,
        title: importData.title || '导入的对话',
        messages: importData.messages.map((msg: any, index: number) => ({
          id: `${sessionId}_${index}`,
          type: msg.type,
          content: msg.content,
          timestamp: new Date(msg.timestamp),
          context: msg.context
        })),
        createdAt: new Date(importData.createdAt),
        updatedAt: new Date()
      }
      
      sessions.value.push(session)
      saveToStorage()
      
      return sessionId
    } catch (error) {
      throw new Error('导入数据格式错误')
    }
  }

  // 返回store接口
  return {
    // 状态
    messages,
    sessions,
    currentSessionId,
    isLoading,
    config,
    
    // 计算属性
    currentSession,
    hasMessages,
    lastMessage,
    
    // 方法
    initialize,
    createNewSession,
    switchSession,
    deleteSession,
    addMessage,
    clearMessages,
    removeMessage,
    sendMessage,
    analyzeData,
    generateFormula,
    saveToStorage,
    loadFromStorage,
    exportChat,
    importChat
  }
})