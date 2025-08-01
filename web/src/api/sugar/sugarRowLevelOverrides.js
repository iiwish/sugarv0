import service from '@/utils/request'
// @Tags SugarRowLevelOverrides
// @Summary 创建Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarRowLevelOverrides true "创建Sugar行级权限豁免表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarRowLevelOverrides/createSugarRowLevelOverrides [post]
export const createSugarRowLevelOverrides = (data) => {
  return service({
    url: '/sugarRowLevelOverrides/createSugarRowLevelOverrides',
    method: 'post',
    data
  })
}

// @Tags SugarRowLevelOverrides
// @Summary 删除Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarRowLevelOverrides true "删除Sugar行级权限豁免表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarRowLevelOverrides/deleteSugarRowLevelOverrides [delete]
export const deleteSugarRowLevelOverrides = (params) => {
  return service({
    url: '/sugarRowLevelOverrides/deleteSugarRowLevelOverrides',
    method: 'delete',
    params
  })
}

// @Tags SugarRowLevelOverrides
// @Summary 批量删除Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除Sugar行级权限豁免表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarRowLevelOverrides/deleteSugarRowLevelOverrides [delete]
export const deleteSugarRowLevelOverridesByIds = (params) => {
  return service({
    url: '/sugarRowLevelOverrides/deleteSugarRowLevelOverridesByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarRowLevelOverrides
// @Summary 更新Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarRowLevelOverrides true "更新Sugar行级权限豁免表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarRowLevelOverrides/updateSugarRowLevelOverrides [put]
export const updateSugarRowLevelOverrides = (data) => {
  return service({
    url: '/sugarRowLevelOverrides/updateSugarRowLevelOverrides',
    method: 'put',
    data
  })
}

// @Tags SugarRowLevelOverrides
// @Summary 用id查询Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarRowLevelOverrides true "用id查询Sugar行级权限豁免表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarRowLevelOverrides/findSugarRowLevelOverrides [get]
export const findSugarRowLevelOverrides = (params) => {
  return service({
    url: '/sugarRowLevelOverrides/findSugarRowLevelOverrides',
    method: 'get',
    params
  })
}

// @Tags SugarRowLevelOverrides
// @Summary 分页获取Sugar行级权限豁免表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取Sugar行级权限豁免表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarRowLevelOverrides/getSugarRowLevelOverridesList [get]
export const getSugarRowLevelOverridesList = (params) => {
  return service({
    url: '/sugarRowLevelOverrides/getSugarRowLevelOverridesList',
    method: 'get',
    params
  })
}

// @Tags SugarRowLevelOverrides
// @Summary 不需要鉴权的Sugar行级权限豁免表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarRowLevelOverridesSearch true "分页获取Sugar行级权限豁免表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarRowLevelOverrides/getSugarRowLevelOverridesPublic [get]
export const getSugarRowLevelOverridesPublic = () => {
  return service({
    url: '/sugarRowLevelOverrides/getSugarRowLevelOverridesPublic',
    method: 'get',
  })
}
