package sugar

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SugarWorkspacesApi struct{}

// CreateSugarWorkspaces 创建Sugar文件列表
// @Tags SugarWorkspaces
// @Summary 创建Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarWorkspaces true "创建Sugar文件列表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarWorkspaces/createSugarWorkspaces [post]
func (sugarWorkspacesApi *SugarWorkspacesApi) CreateSugarWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	var sugarWorkspaces sugar.SugarWorkspaces
	err := c.ShouldBindJSON(&sugarWorkspaces)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	sugarWorkspaces.CreatedBy = &userIdStr

	err = sugarWorkspacesService.CreateSugarWorkspaces(ctx, &sugarWorkspaces)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteSugarWorkspaces 删除Sugar文件列表
// @Tags SugarWorkspaces
// @Summary 删除Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "文件ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarWorkspaces/deleteSugarWorkspaces [delete]
func (sugarWorkspacesApi *SugarWorkspacesApi) DeleteSugarWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Query("id")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	err := sugarWorkspacesService.DeleteSugarWorkspaces(ctx, id, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarWorkspacesByIds 批量删除Sugar文件列表
// @Tags SugarWorkspaces
// @Summary 批量删除Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param ids query []string true "文件ID列表"
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarWorkspaces/deleteSugarWorkspacesByIds [delete]
func (sugarWorkspacesApi *SugarWorkspacesApi) DeleteSugarWorkspacesByIds(c *gin.Context) {
	ctx := c.Request.Context()
	ids := c.QueryArray("ids[]")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	err := sugarWorkspacesService.DeleteSugarWorkspacesByIds(ctx, ids, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarWorkspaces 更新Sugar文件列表
// @Tags SugarWorkspaces
// @Summary 更新Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarWorkspaces true "更新Sugar文件列表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarWorkspaces/updateSugarWorkspaces [put]
func (sugarWorkspacesApi *SugarWorkspacesApi) UpdateSugarWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	var sugarWorkspaces sugar.SugarWorkspaces
	err := c.ShouldBindJSON(&sugarWorkspaces)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	sugarWorkspaces.UpdatedBy = &userIdStr

	err = sugarWorkspacesService.UpdateSugarWorkspaces(ctx, sugarWorkspaces, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarWorkspaces 用id查询Sugar文件列表
// @Tags SugarWorkspaces
// @Summary 用id查询Sugar文件列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "文件ID"
// @Success 200 {object} response.Response{data=sugar.SugarWorkspaces,msg=string} "查询成功"
// @Router /sugarWorkspaces/findSugarWorkspaces [get]
func (sugarWorkspacesApi *SugarWorkspacesApi) FindSugarWorkspaces(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Query("id")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	resugarWorkspaces, err := sugarWorkspacesService.GetSugarWorkspaces(ctx, id, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(resugarWorkspaces, c)
}

// GetSugarWorkspacesList 分页获取Sugar文件列表列表
// @Tags SugarWorkspaces
// @Summary 分页获取Sugar文件列表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarWorkspacesSearch true "分页获取Sugar文件列表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarWorkspaces/getSugarWorkspacesList [get]
func (sugarWorkspacesApi *SugarWorkspacesApi) GetSugarWorkspacesList(c *gin.Context) {
	ctx := c.Request.Context()
	var pageInfo sugarReq.SugarWorkspacesSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	list, total, err := sugarWorkspacesService.GetSugarWorkspacesInfoListByUser(ctx, pageInfo, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}
