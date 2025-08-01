import service from '@/utils/request'
// @Tags SugarAgents
// @Summary 创建sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarAgents true "创建sugar智能体表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarAgents/createSugarAgents [post]
export const createSugarAgents = (data) => {
  return service({
    url: '/sugarAgents/createSugarAgents',
    method: 'post',
    data
  })
}

// @Tags SugarAgents
// @Summary 删除sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarAgents true "删除sugar智能体表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarAgents/deleteSugarAgents [delete]
export const deleteSugarAgents = (params) => {
  return service({
    url: '/sugarAgents/deleteSugarAgents',
    method: 'delete',
    params
  })
}

// @Tags SugarAgents
// @Summary 批量删除sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除sugar智能体表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarAgents/deleteSugarAgents [delete]
export const deleteSugarAgentsByIds = (params) => {
  return service({
    url: '/sugarAgents/deleteSugarAgentsByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarAgents
// @Summary 更新sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarAgents true "更新sugar智能体表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarAgents/updateSugarAgents [put]
export const updateSugarAgents = (data) => {
  return service({
    url: '/sugarAgents/updateSugarAgents',
    method: 'put',
    data
  })
}

// @Tags SugarAgents
// @Summary 用id查询sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarAgents true "用id查询sugar智能体表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarAgents/findSugarAgents [get]
export const findSugarAgents = (params) => {
  return service({
    url: '/sugarAgents/findSugarAgents',
    method: 'get',
    params
  })
}

// @Tags SugarAgents
// @Summary 分页获取sugar智能体表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取sugar智能体表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarAgents/getSugarAgentsList [get]
export const getSugarAgentsList = (params) => {
  return service({
    url: '/sugarAgents/getSugarAgentsList',
    method: 'get',
    params
  })
}

// @Tags SugarAgents
// @Summary 不需要鉴权的sugar智能体表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarAgentsSearch true "分页获取sugar智能体表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarAgents/getSugarAgentsPublic [get]
export const getSugarAgentsPublic = () => {
  return service({
    url: '/sugarAgents/getSugarAgentsPublic',
    method: 'get',
  })
}
