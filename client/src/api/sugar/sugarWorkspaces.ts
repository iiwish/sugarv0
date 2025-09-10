import service from '@/utils/request'
import type { ApiResponse, WorkspaceTreeNode, CreateFolderData, RenameItemData } from '@/types/api'

// @Tags SugarWorkspaces
// @Summary 创建Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarWorkspaces true "创建Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /sugarWorkspaces/createSugarWorkspaces [post]
export const createSugarWorkspaces = (data: any) => {
  return service.post('/sugarWorkspaces/createSugarWorkspaces', data)
}

// @Tags SugarWorkspaces
// @Summary 删除Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarWorkspaces true "删除Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarWorkspaces/deleteSugarWorkspaces [delete]
export const deleteSugarWorkspaces = (params: { id: string }) => {
  return service.delete('/sugarWorkspaces/deleteSugarWorkspaces', params)
}

// @Tags SugarWorkspaces
// @Summary 批量删除Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /sugarWorkspaces/deleteSugarWorkspaces [delete]
export const deleteSugarWorkspacesByIds = (params: { ids: string[] }) => {
  return service.delete('/sugarWorkspaces/deleteSugarWorkspacesByIds', params)
}

// @Tags SugarWorkspaces
// @Summary 更新Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.SugarWorkspaces true "更新Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /sugarWorkspaces/updateSugarWorkspaces [put]
export const updateSugarWorkspaces = (data: any) => {
  return service.put('/sugarWorkspaces/updateSugarWorkspaces', data)
}

// @Tags SugarWorkspaces
// @Summary 用id查询Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.SugarWorkspaces true "用id查询Sugar文件列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /sugarWorkspaces/findSugarWorkspaces [get]
export const findSugarWorkspaces = (params: { id: string }) => {
  return service.get('/sugarWorkspaces/findSugarWorkspaces', params)
}

// @Tags SugarWorkspaces
// @Summary 分页获取Sugar文件列表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取Sugar文件列表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /sugarWorkspaces/getSugarWorkspacesList [get]
export const getSugarWorkspacesList = (params: { page?: number; pageSize?: number }) => {
  return service.get('/sugarWorkspaces/getSugarWorkspacesList', params)
}

// @Tags SugarWorkspaces
// @Summary 不需要鉴权的Sugar文件列表接口
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarWorkspacesSearch true "分页获取Sugar文件列表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarWorkspaces/getSugarWorkspacesPublic [get]
export const getSugarWorkspacesPublic = () => {
  return service.get('/sugarWorkspaces/getSugarWorkspacesPublic')
}

// @Tags SugarWorkspaces
// @Summary 创建新的工作簿文件
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "创建工作簿文件数据"
// @Success 200 {object} response.Response{data=object,msg=string} "创建成功"
// @Router /sugarWorkspaces/createWorkbookFile [post]
export const createWorkbookFile = (data: {
  name: string
  parentId?: string
  teamId: string
}): Promise<ApiResponse<WorkspaceTreeNode>> => {
  return service.post('/sugarWorkspaces/createWorkbookFile', data)
}

// @Tags SugarWorkspaces
// @Summary 保存工作簿内容
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "保存工作簿内容数据"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /sugarWorkspaces/saveWorkbookContent [put]
export const saveWorkbookContent = (data: {
  id: string
  content: any
}): Promise<ApiResponse<any>> => {
  return service.put('/sugarWorkspaces/saveWorkbookContent', data)
}

// @Tags SugarWorkspaces
// @Summary 获取工作簿内容
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "文件ID"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarWorkspaces/getWorkbookContent [get]
export const getWorkbookContent = (params: { id: string }): Promise<ApiResponse<any>> => {
  return service.get('/sugarWorkspaces/getWorkbookContent', params)
}