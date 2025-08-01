import service from '@/utils/request'
// @Tags SugarDbConnections
// @Summary 创建Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarDbConnections true "创建Sugar数据库配置表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarDbConnections/createSugarDbConnections [post]
export const createSugarDbConnections = (data) => {
  return service({
    url: '/sugarDbConnections/createSugarDbConnections',
    method: 'post',
    data
  })
}

// @Tags SugarDbConnections
// @Summary 删除Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarDbConnections true "删除Sugar数据库配置表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarDbConnections/deleteSugarDbConnections [delete]
export const deleteSugarDbConnections = (params) => {
  return service({
    url: '/sugarDbConnections/deleteSugarDbConnections',
    method: 'delete',
    params
  })
}

// @Tags SugarDbConnections
// @Summary 批量删除Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除Sugar数据库配置表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarDbConnections/deleteSugarDbConnections [delete]
export const deleteSugarDbConnectionsByIds = (params) => {
  return service({
    url: '/sugarDbConnections/deleteSugarDbConnectionsByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarDbConnections
// @Summary 更新Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarDbConnections true "更新Sugar数据库配置表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarDbConnections/updateSugarDbConnections [put]
export const updateSugarDbConnections = (data) => {
  return service({
    url: '/sugarDbConnections/updateSugarDbConnections',
    method: 'put',
    data
  })
}

// @Tags SugarDbConnections
// @Summary 用id查询Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarDbConnections true "用id查询Sugar数据库配置表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarDbConnections/findSugarDbConnections [get]
export const findSugarDbConnections = (params) => {
  return service({
    url: '/sugarDbConnections/findSugarDbConnections',
    method: 'get',
    params
  })
}

// @Tags SugarDbConnections
// @Summary 分页获取Sugar数据库配置表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取Sugar数据库配置表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarDbConnections/getSugarDbConnectionsList [get]
export const getSugarDbConnectionsList = (params) => {
  return service({
    url: '/sugarDbConnections/getSugarDbConnectionsList',
    method: 'get',
    params
  })
}

// @Tags SugarDbConnections
// @Summary 不需要鉴权的Sugar数据库配置表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarDbConnectionsSearch true "分页获取Sugar数据库配置表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarDbConnections/getSugarDbConnectionsPublic [get]
export const getSugarDbConnectionsPublic = () => {
  return service({
    url: '/sugarDbConnections/getSugarDbConnectionsPublic',
    method: 'get',
  })
}
