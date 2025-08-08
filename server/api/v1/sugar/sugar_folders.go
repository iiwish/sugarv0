package sugar

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SugarFoldersApi struct{}

// GetWorkspaceTree 获取工作空间文件夹树形结构
// @Tags SugarFolders
// @Summary 获取工作空间文件夹树形结构
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param teamId query string false "团队ID，不传则获取当前用户所有团队的工作空间"
// @Success 200 {object} response.Response{data=sugarRes.SugarFoldersGetWorkspaceTreeResponse,msg=string} "获取成功"
// @Router /sugarFolders/getWorkspaceTree [get]
func (s *SugarFoldersApi) GetWorkspaceTree(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFoldersGetWorkspaceTreeRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := sugarFoldersService.GetWorkspaceTree(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("获取工作空间树形结构失败!", zap.Error(err))
		response.FailWithMessage("获取工作空间树形结构失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// CreateFolder 创建文件夹
// @Tags SugarFolders
// @Summary 创建文件夹
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugarReq.SugarFoldersCreateFolderRequest true "创建文件夹数据"
// @Success 200 {object} response.Response{data=sugarRes.SugarFoldersCreateFolderResponse,msg=string} "创建成功"
// @Router /sugarFolders/createFolder [post]
func (s *SugarFoldersApi) CreateFolder(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFoldersCreateFolderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := sugarFoldersService.CreateFolder(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("创建文件夹失败!", zap.Error(err))
		response.FailWithMessage("创建文件夹失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// RenameItem 重命名文件夹或文件
// @Tags SugarFolders
// @Summary 重命名文件夹或文件
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugarReq.SugarFoldersRenameRequest true "重命名数据"
// @Success 200 {object} response.Response{data=sugarRes.SugarFoldersRenameResponse,msg=string} "重命名成功"
// @Router /sugarFolders/rename [put]
func (s *SugarFoldersApi) RenameItem(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFoldersRenameRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := sugarFoldersService.RenameItem(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("重命名失败!", zap.Error(err))
		response.FailWithMessage("重命名失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// MoveItem 移动文件夹或文件
// @Tags SugarFolders
// @Summary 移动文件夹或文件
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugarReq.SugarFoldersMoveRequest true "移动数据"
// @Success 200 {object} response.Response{data=sugarRes.SugarFoldersMoveResponse,msg=string} "移动成功"
// @Router /sugarFolders/move [put]
func (s *SugarFoldersApi) MoveItem(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFoldersMoveRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := sugarFoldersService.MoveItem(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("移动失败!", zap.Error(err))
		response.FailWithMessage("移动失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// DeleteItem 删除文件夹或文件
// @Tags SugarFolders
// @Summary 删除文件夹或文件
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "要删除的项目ID"
// @Success 200 {object} response.Response{data=sugarRes.SugarFoldersDeleteResponse,msg=string} "删除成功"
// @Router /sugarFolders/deleteItem [delete]
func (s *SugarFoldersApi) DeleteItem(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFoldersDeleteRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := sugarFoldersService.DeleteItem(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// GetFolderContent 获取文件夹内容
// @Tags SugarFolders
// @Summary 获取文件夹内容
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param folderId query string true "文件夹ID"
// @Param page query number false "页码"
// @Param pageSize query number false "每页数量"
// @Success 200 {object} response.Response{data=sugarRes.SugarFoldersGetFolderContentResponse,msg=string} "获取成功"
// @Router /sugarFolders/getFolderContent [get]
func (s *SugarFoldersApi) GetFolderContent(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFoldersGetFolderContentRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := sugarFoldersService.GetFolderContent(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("获取文件夹内容失败!", zap.Error(err))
		response.FailWithMessage("获取文件夹内容失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}
