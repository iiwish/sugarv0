import service from '@/utils/request'

// @Tags SugarFolders
// @Summary 获取工作空间文件夹树形结构
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param teamId query string false "团队ID，不传则获取当前用户所有团队的工作空间"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarFolders/getWorkspaceTree [get]
export const getWorkspaceTree = (params) => {
  return service({
    url: '/sugarFolders/getWorkspaceTree',
    method: 'get',
    params
  })
}

// @Tags SugarFolders
// @Summary 创建文件夹
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "创建文件夹数据"
// @Success 200 {object} response.Response{data=object,msg=string} "创建成功"
// @Router /sugarFolders/createFolder [post]
export const createFolder = (data) => {
  return service({
    url: '/sugarFolders/createFolder',
    method: 'post',
    data
  })
}

// @Tags SugarFolders
// @Summary 重命名文件夹或文件
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "重命名数据"
// @Success 200 {object} response.Response{data=object,msg=string} "重命名成功"
// @Router /sugarFolders/rename [put]
export const renameItem = (data) => {
  return service({
    url: '/sugarFolders/rename',
    method: 'put',
    data
  })
}

// @Tags SugarFolders
// @Summary 移动文件夹或文件
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "移动数据"
// @Success 200 {object} response.Response{data=object,msg=string} "移动成功"
// @Router /sugarFolders/move [put]
export const moveItem = (data) => {
  return service({
    url: '/sugarFolders/move',
    method: 'put',
    data
  })
}

// @Tags SugarFolders
// @Summary 删除文件夹或文件
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "要删除的项目ID"
// @Success 200 {object} response.Response{data=object,msg=string} "删除成功"
// @Router /sugarFolders/deleteItem [delete]
export const deleteItem = (params) => {
  return service({
    url: '/sugarFolders/deleteItem',
    method: 'delete',
    params
  })
}

// @Tags SugarFolders
// @Summary 获取文件夹内容
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param folderId query string true "文件夹ID"
// @Param page query number false "页码"
// @Param pageSize query number false "每页数量"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarFolders/getFolderContent [get]
export const getFolderContent = (params) => {
  return service({
    url: '/sugarFolders/getFolderContent',
    method: 'get',
    params
  })
}