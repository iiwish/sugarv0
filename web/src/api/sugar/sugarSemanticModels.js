import service from '@/utils/request'
// @Tags SugarSemanticModels
// @Summary 创建Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarSemanticModels true "创建Sugar指标语义表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarSemanticModels/createSugarSemanticModels [post]
export const createSugarSemanticModels = (data) => {
  return service({
    url: '/sugarSemanticModels/createSugarSemanticModels',
    method: 'post',
    data
  })
}

// @Tags SugarSemanticModels
// @Summary 删除Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarSemanticModels true "删除Sugar指标语义表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarSemanticModels/deleteSugarSemanticModels [delete]
export const deleteSugarSemanticModels = (params) => {
  return service({
    url: '/sugarSemanticModels/deleteSugarSemanticModels',
    method: 'delete',
    params
  })
}

// @Tags SugarSemanticModels
// @Summary 批量删除Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除Sugar指标语义表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarSemanticModels/deleteSugarSemanticModels [delete]
export const deleteSugarSemanticModelsByIds = (params) => {
  return service({
    url: '/sugarSemanticModels/deleteSugarSemanticModelsByIds',
    method: 'delete',
    params
  })
}

// @Tags SugarSemanticModels
// @Summary 更新Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarSemanticModels true "更新Sugar指标语义表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarSemanticModels/updateSugarSemanticModels [put]
export const updateSugarSemanticModels = (data) => {
  return service({
    url: '/sugarSemanticModels/updateSugarSemanticModels',
    method: 'put',
    data
  })
}

// @Tags SugarSemanticModels
// @Summary 用id查询Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarSemanticModels true "用id查询Sugar指标语义表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarSemanticModels/findSugarSemanticModels [get]
export const findSugarSemanticModels = (params) => {
  return service({
    url: '/sugarSemanticModels/findSugarSemanticModels',
    method: 'get',
    params
  })
}

// @Tags SugarSemanticModels
// @Summary 分页获取Sugar指标语义表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取Sugar指标语义表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarSemanticModels/getSugarSemanticModelsList [get]
export const getSugarSemanticModelsList = (params) => {
  return service({
    url: '/sugarSemanticModels/getSugarSemanticModelsList',
    method: 'get',
    params
  })
}

// @Tags SugarSemanticModels
// @Summary 不需要鉴权的Sugar指标语义表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarSemanticModelsSearch true "分页获取Sugar指标语义表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarSemanticModels/getSugarSemanticModelsPublic [get]
export const getSugarSemanticModelsPublic = () => {
  return service({
    url: '/sugarSemanticModels/getSugarSemanticModelsPublic',
    method: 'get',
  })
}
