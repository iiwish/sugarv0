import service from '@/utils/request'
// @Tags SugarTeams
// @Summary 创建团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarTeams true "创建团队信息表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarTeams/createSugarTeams [post]
export const createSugarTeams = (data) => {
  return service({
    url: '/sugarTeams/createSugarTeams',
    method: 'post',
    data
  })
}

// @Tags SugarTeams
// @Summary 删除团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarTeams true "删除团队信息表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarTeams/deleteSugarTeams [delete]
export const deleteSugarTeams = (params) => {
  return service({
    url: '/sugarTeams/deleteSugarTeams',
    method: 'delete',
    params
  })
}

// @Tags SugarTeams
// @Summary 批量删除团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除团队信息表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarTeams/deleteSugarTeams [delete]
export const deleteSugarTeamsByIds = (params) => {
  return service({
    url: '/sugarTeams/deleteSugarTeamsByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarTeams
// @Summary 更新团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarTeams true "更新团队信息表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarTeams/updateSugarTeams [put]
export const updateSugarTeams = (data) => {
  return service({
    url: '/sugarTeams/updateSugarTeams',
    method: 'put',
    data
  })
}

// @Tags SugarTeams
// @Summary 用id查询团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarTeams true "用id查询团队信息表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarTeams/findSugarTeams [get]
export const findSugarTeams = (params) => {
  return service({
    url: '/sugarTeams/findSugarTeams',
    method: 'get',
    params
  })
}

// @Tags SugarTeams
// @Summary 分页获取团队信息表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取团队信息表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarTeams/getSugarTeamsList [get]
export const getSugarTeamsList = (params) => {
  return service({
    url: '/sugarTeams/getSugarTeamsList',
    method: 'get',
    params
  })
}

// @Tags SugarTeams
// @Summary 不需要鉴权的团队信息表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarTeamsSearch true "分页获取团队信息表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarTeams/getSugarTeamsPublic [get]
export const getSugarTeamsPublic = () => {
  return service({
    url: '/sugarTeams/getSugarTeamsPublic',
    method: 'get',
  })
}
