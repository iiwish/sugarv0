// 聊天相关类型定义

export interface ChatMessage {
  id: string
  content: string
  type: 'user' | 'ai'
  timestamp: Date
  context?: ContextInfo | null
  error?: boolean
  metadata?: {
    agentName?: string
    processingTime?: number
    tokens?: number
  }
}

export interface ContextInfo {
  fileName: string
  fileId: string
  sheetName?: string
  selectedRange?: string
  sheetData?: any[][]
}

export interface ChatSession {
  id: string
  title: string
  messages: ChatMessage[]
  createdAt: Date
  updatedAt: Date
  context?: ContextInfo
}

export interface AIResponse {
  content: string
  type: 'text' | 'formula' | 'analysis'
  metadata?: {
    agentName?: string
    processingTime?: number
    confidence?: number
  }
}

export interface FormulaGenerationRequest {
  description: string
  context: ContextInfo
  targetRange?: string
  dataType?: 'number' | 'text' | 'date' | 'boolean'
}

export interface DataAnalysisRequest {
  description: string
  context: ContextInfo
  analysisType?: 'summary' | 'trend' | 'correlation' | 'anomaly'
  dataRange?: string
}

export interface ChatConfig {
  maxMessages: number
  autoSave: boolean
  enableContext: boolean
  defaultAgent: string
}