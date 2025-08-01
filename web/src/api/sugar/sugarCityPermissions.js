import service from '@/utils/request'
// @Tags SugarCityPermissions
// @Summary 创建sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarCityPermissions true "创建sugarCityPermissions表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarCityPermissions/createSugarCityPermissions [post]
export const createSugarCityPermissions = (data) => {
  return service({
    url: '/sugarCityPermissions/createSugarCityPermissions',
    method: 'post',
    data
  })
}

// @Tags SugarCityPermissions
// @Summary 删除sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarCityPermissions true "删除sugarCityPermissions表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarCityPermissions/deleteSugarCityPermissions [delete]
export const deleteSugarCityPermissions = (params) => {
  return service({
    url: '/sugarCityPermissions/deleteSugarCityPermissions',
    method: 'delete',
    params
  })
}

// @Tags SugarCityPermissions
// @Summary 批量删除sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除sugarCityPermissions表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarCityPermissions/deleteSugarCityPermissions [delete]
export const deleteSugarCityPermissionsByIds = (params) => {
  return service({
    url: '/sugarCityPermissions/deleteSugarCityPermissionsByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarCityPermissions
// @Summary 更新sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarCityPermissions true "更新sugarCityPermissions表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarCityPermissions/updateSugarCityPermissions [put]
export const updateSugarCityPermissions = (data) => {
  return service({
    url: '/sugarCityPermissions/updateSugarCityPermissions',
    method: 'put',
    data
  })
}

// @Tags SugarCityPermissions
// @Summary 用id查询sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarCityPermissions true "用id查询sugarCityPermissions表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarCityPermissions/findSugarCityPermissions [get]
export const findSugarCityPermissions = (params) => {
  return service({
    url: '/sugarCityPermissions/findSugarCityPermissions',
    method: 'get',
    params
  })
}

// @Tags SugarCityPermissions
// @Summary 分页获取sugarCityPermissions表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取sugarCityPermissions表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarCityPermissions/getSugarCityPermissionsList [get]
export const getSugarCityPermissionsList = (params) => {
  return service({
    url: '/sugarCityPermissions/getSugarCityPermissionsList',
    method: 'get',
    params
  })
}

// @Tags SugarCityPermissions
// @Summary 不需要鉴权的sugarCityPermissions表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarCityPermissionsSearch true "分页获取sugarCityPermissions表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarCityPermissions/getSugarCityPermissionsPublic [get]
export const getSugarCityPermissionsPublic = () => {
  return service({
    url: '/sugarCityPermissions/getSugarCityPermissionsPublic',
    method: 'get',
  })
}
