import service from '@/utils/request'
// @Tags SugarWorkspaces
// @Summary 创建Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarWorkspaces true "创建Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarWorkspaces/createSugarWorkspaces [post]
export const createSugarWorkspaces = (data) => {
  return service({
    url: '/sugarWorkspaces/createSugarWorkspaces',
    method: 'post',
    data
  })
}

// @Tags SugarWorkspaces
// @Summary 删除Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarWorkspaces true "删除Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarWorkspaces/deleteSugarWorkspaces [delete]
export const deleteSugarWorkspaces = (params) => {
  return service({
    url: '/sugarWorkspaces/deleteSugarWorkspaces',
    method: 'delete',
    params
  })
}

// @Tags SugarWorkspaces
// @Summary 批量删除Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarWorkspaces/deleteSugarWorkspaces [delete]
export const deleteSugarWorkspacesByIds = (params) => {
  return service({
    url: '/sugarWorkspaces/deleteSugarWorkspacesByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarWorkspaces
// @Summary 更新Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarWorkspaces true "更新Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarWorkspaces/updateSugarWorkspaces [put]
export const updateSugarWorkspaces = (data) => {
  return service({
    url: '/sugarWorkspaces/updateSugarWorkspaces',
    method: 'put',
    data
  })
}

// @Tags SugarWorkspaces
// @Summary 用id查询Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarWorkspaces true "用id查询Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarWorkspaces/findSugarWorkspaces [get]
export const findSugarWorkspaces = (params) => {
  return service({
    url: '/sugarWorkspaces/findSugarWorkspaces',
    method: 'get',
    params
  })
}

// @Tags SugarWorkspaces
// @Summary 分页获取Sugar文件列表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取Sugar文件列表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarWorkspaces/getSugarWorkspacesList [get]
export const getSugarWorkspacesList = (params) => {
  return service({
    url: '/sugarWorkspaces/getSugarWorkspacesList',
    method: 'get',
    params
  })
}

// @Tags SugarWorkspaces
// @Summary 不需要鉴权的Sugar文件列表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarWorkspacesSearch true "分页获取Sugar文件列表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarWorkspaces/getSugarWorkspacesPublic [get]
export const getSugarWorkspacesPublic = () => {
  return service({
    url: '/sugarWorkspaces/getSugarWorkspacesPublic',
    method: 'get',
  })
}
