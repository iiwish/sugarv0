import service from '@/utils/request'
// @Tags SugarTeamMembers
// @Summary 创建sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarTeamMembers true "创建sugarTeamMembers表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarTeamMembers/createSugarTeamMembers [post]
export const createSugarTeamMembers = (data) => {
  return service({
    url: '/sugarTeamMembers/createSugarTeamMembers',
    method: 'post',
    data
  })
}

// @Tags SugarTeamMembers
// @Summary 删除sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarTeamMembers true "删除sugarTeamMembers表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarTeamMembers/deleteSugarTeamMembers [delete]
export const deleteSugarTeamMembers = (params) => {
  return service({
    url: '/sugarTeamMembers/deleteSugarTeamMembers',
    method: 'delete',
    params
  })
}

// @Tags SugarTeamMembers
// @Summary 批量删除sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除sugarTeamMembers表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarTeamMembers/deleteSugarTeamMembers [delete]
export const deleteSugarTeamMembersByIds = (params) => {
  return service({
    url: '/sugarTeamMembers/deleteSugarTeamMembersByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarTeamMembers
// @Summary 更新sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarTeamMembers true "更新sugarTeamMembers表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarTeamMembers/updateSugarTeamMembers [put]
export const updateSugarTeamMembers = (data) => {
  return service({
    url: '/sugarTeamMembers/updateSugarTeamMembers',
    method: 'put',
    data
  })
}

// @Tags SugarTeamMembers
// @Summary 用id查询sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarTeamMembers true "用id查询sugarTeamMembers表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarTeamMembers/findSugarTeamMembers [get]
export const findSugarTeamMembers = (params) => {
  return service({
    url: '/sugarTeamMembers/findSugarTeamMembers',
    method: 'get',
    params
  })
}

// @Tags SugarTeamMembers
// @Summary 分页获取sugarTeamMembers表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取sugarTeamMembers表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarTeamMembers/getSugarTeamMembersList [get]
export const getSugarTeamMembersList = (params) => {
  return service({
    url: '/sugarTeamMembers/getSugarTeamMembersList',
    method: 'get',
    params
  })
}

// @Tags SugarTeamMembers
// @Summary 不需要鉴权的sugarTeamMembers表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarTeamMembersSearch true "分页获取sugarTeamMembers表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarTeamMembers/getSugarTeamMembersPublic [get]
export const getSugarTeamMembersPublic = () => {
  return service({
    url: '/sugarTeamMembers/getSugarTeamMembersPublic',
    method: 'get',
  })
}
