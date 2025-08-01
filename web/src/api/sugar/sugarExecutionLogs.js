import service from '@/utils/request'
// @Tags SugarExecutionLogs
// @Summary 创建sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarExecutionLogs true "创建sugar操作日志表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarExecutionLogs/createSugarExecutionLogs [post]
export const createSugarExecutionLogs = (data) => {
  return service({
    url: '/sugarExecutionLogs/createSugarExecutionLogs',
    method: 'post',
    data
  })
}

// @Tags SugarExecutionLogs
// @Summary 删除sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarExecutionLogs true "删除sugar操作日志表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarExecutionLogs/deleteSugarExecutionLogs [delete]
export const deleteSugarExecutionLogs = (params) => {
  return service({
    url: '/sugarExecutionLogs/deleteSugarExecutionLogs',
    method: 'delete',
    params
  })
}

// @Tags SugarExecutionLogs
// @Summary 批量删除sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除sugar操作日志表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarExecutionLogs/deleteSugarExecutionLogs [delete]
export const deleteSugarExecutionLogsByIds = (params) => {
  return service({
    url: '/sugarExecutionLogs/deleteSugarExecutionLogsByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarExecutionLogs
// @Summary 更新sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarExecutionLogs true "更新sugar操作日志表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarExecutionLogs/updateSugarExecutionLogs [put]
export const updateSugarExecutionLogs = (data) => {
  return service({
    url: '/sugarExecutionLogs/updateSugarExecutionLogs',
    method: 'put',
    data
  })
}

// @Tags SugarExecutionLogs
// @Summary 用id查询sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarExecutionLogs true "用id查询sugar操作日志表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarExecutionLogs/findSugarExecutionLogs [get]
export const findSugarExecutionLogs = (params) => {
  return service({
    url: '/sugarExecutionLogs/findSugarExecutionLogs',
    method: 'get',
    params
  })
}

// @Tags SugarExecutionLogs
// @Summary 分页获取sugar操作日志表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取sugar操作日志表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarExecutionLogs/getSugarExecutionLogsList [get]
export const getSugarExecutionLogsList = (params) => {
  return service({
    url: '/sugarExecutionLogs/getSugarExecutionLogsList',
    method: 'get',
    params
  })
}

// @Tags SugarExecutionLogs
// @Summary 不需要鉴权的sugar操作日志表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarExecutionLogsSearch true "分页获取sugar操作日志表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarExecutionLogs/getSugarExecutionLogsPublic [get]
export const getSugarExecutionLogsPublic = () => {
  return service({
    url: '/sugarExecutionLogs/getSugarExecutionLogsPublic',
    method: 'get',
  })
}
