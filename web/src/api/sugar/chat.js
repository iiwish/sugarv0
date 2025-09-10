import service from '@/utils/request'

// @Tags Chat
// @Summary 发送聊天消息
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "聊天消息数据"
// @Success 200 {object} response.Response{data=object,msg=string} "发送成功"
// @Router /chat/sendMessage [post]
export const sendChatMessage = (data) => {
  return service({
    url: '/api/sugarFormulaQuery/executeAiFetch',
    method: 'post',
    data
  })
}

// @Tags Chat
// @Summary 执行AI数据分析
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "数据分析请求"
// @Success 200 {object} response.Response{data=object,msg=string} "分析成功"
// @Router /chat/analyzeData [post]
export const analyzeData = (data) => {
  return service({
    url: '/api/sugarFormulaQuery/executeAiExplainRange',
    method: 'post',
    data
  })
}

// @Tags Chat
// @Summary 生成AI公式
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "公式生成请求"
// @Success 200 {object} response.Response{data=object,msg=string} "生成成功"
// @Router /chat/generateFormula [post]
export const generateFormula = (data) => {
  return service({
    url: '/api/sugarFormulaQuery/executeAiFetch',
    method: 'post',
    data
  })
}

// @Tags Chat
// @Summary 获取可用的AI代理列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=array,msg=string} "获取成功"
// @Router /chat/getAgents [get]
export const getChatAgents = () => {
  return service({
    url: '/sugarAgents/getSugarAgentsList',
    method: 'get',
    params: {
      page: 1,
      pageSize: 100
    }
  })
}

// @Tags Chat
// @Summary 获取聊天历史
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param sessionId query string false "会话ID"
// @Success 200 {object} response.Response{data=array,msg=string} "获取成功"
// @Router /chat/getHistory [get]
export const getChatHistory = (params) => {
  return service({
    url: '/chat/getHistory',
    method: 'get',
    params
  })
}

// @Tags Chat
// @Summary 保存聊天会话
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "会话数据"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /chat/saveSession [post]
export const saveChatSession = (data) => {
  return service({
    url: '/chat/saveSession',
    method: 'post',
    data
  })
}

// @Tags Chat
// @Summary 删除聊天会话
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param sessionId query string true "会话ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /chat/deleteSession [delete]
export const deleteChatSession = (params) => {
  return service({
    url: '/chat/deleteSession',
    method: 'delete',
    params
  })
}