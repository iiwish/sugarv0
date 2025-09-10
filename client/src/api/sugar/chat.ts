import service from '@/utils/request'
import type { ApiResponse, ChatMessage, FormulaQueryRequest, FormulaQueryResponse } from '@/types/api'

// 聊天相关 API

// @Tags Chat
// @Summary 发送聊天消息
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "聊天消息数据"
// @Success 200 {object} response.Response{data=object,msg=string} "发送成功"
// @Router /chat/sendMessage [post]
export const sendChatMessage = (data: {
  content: string
  sessionId?: string
  type?: 'text' | 'formula'
}): Promise<ApiResponse<ChatMessage>> => {
  return service.post('/chat/sendMessage', data)
}

// @Tags Chat
// @Summary 获取聊天历史
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param sessionId query string false "会话ID"
// @Success 200 {object} response.Response{data=array,msg=string} "获取成功"
// @Router /chat/getHistory [get]
export const getChatHistory = (params?: {
  sessionId?: string
  page?: number
  pageSize?: number
}): Promise<ApiResponse<ChatMessage[]>> => {
  return service.get('/chat/getHistory', params)
}

// @Tags Chat
// @Summary 创建新的聊天会话
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "会话数据"
// @Success 200 {object} response.Response{data=object,msg=string} "创建成功"
// @Router /chat/createSession [post]
export const createChatSession = (data: {
  title?: string
}): Promise<ApiResponse<{ sessionId: string }>> => {
  return service.post('/chat/createSession', data)
}

// @Tags Chat
// @Summary 删除聊天会话
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param sessionId query string true "会话ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /chat/deleteSession [delete]
export const deleteChatSession = (params: {
  sessionId: string
}): Promise<ApiResponse<any>> => {
  return service.delete('/chat/deleteSession', params)
}

// 公式查询相关 API

// @Tags Formula
// @Summary AI公式查询
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "公式查询数据"
// @Success 200 {object} response.Response{data=object,msg=string} "查询成功"
// @Router /formula/query [post]
export const queryFormula = (data: FormulaQueryRequest): Promise<ApiResponse<FormulaQueryResponse>> => {
  return service.post('/formula/query', data)
}

// @Tags Formula
// @Summary 获取公式建议
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param query query string true "查询关键词"
// @Success 200 {object} response.Response{data=array,msg=string} "获取成功"
// @Router /formula/suggestions [get]
export const getFormulaSuggestions = (params: {
  query: string
  limit?: number
}): Promise<ApiResponse<string[]>> => {
  return service.get('/formula/suggestions', params)
}

// @Tags Formula
// @Summary 验证公式语法
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "公式验证数据"
// @Success 200 {object} response.Response{data=object,msg=string} "验证成功"
// @Router /formula/validate [post]
export const validateFormula = (data: {
  formula: string
  context?: any
}): Promise<ApiResponse<{
  isValid: boolean
  errors?: string[]
  suggestions?: string[]
}>> => {
  return service.post('/formula/validate', data)
}

// 团队相关 API

// @Tags Team
// @Summary 获取团队列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /team/list [get]
export const getSugarTeamsList = (params?: {
  page?: number
  pageSize?: number
}): Promise<ApiResponse<{ list: any[] }>> => {
  return service.get('/team/list', params)
}

// 工作空间树相关 API

// @Tags Workspace
// @Summary 获取工作空间树形结构
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param teamId query string true "团队ID"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /workspace/tree [get]
export const getWorkspaceTree = (params: {
  teamId: string
}): Promise<ApiResponse<{ tree: any[] }>> => {
  return service.get('/workspace/tree', params)
}

// 文件夹操作相关 API

// @Tags Folder
// @Summary 创建文件夹
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "文件夹数据"
// @Success 200 {object} response.Response{data=object,msg=string} "创建成功"
// @Router /folder/create [post]
export const createFolder = (data: any): Promise<ApiResponse<any>> => {
  return service.post('/folder/create', data)
}

// @Tags Folder
// @Summary 重命名文件或文件夹
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "重命名数据"
// @Success 200 {object} response.Response{msg=string} "重命名成功"
// @Router /folder/rename [put]
export const renameItem = (data: any): Promise<ApiResponse<any>> => {
  return service.put('/folder/rename', data)
}

// @Tags Folder
// @Summary 删除文件或文件夹
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "文件或文件夹ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /folder/delete [delete]
export const deleteItem = (params: { id: string }): Promise<ApiResponse<any>> => {
  return service.delete('/folder/delete', params)
}