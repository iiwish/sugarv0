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

type SugarAgentsApi struct{}

// CreateSugarAgents 创建sugar智能体表
// @Tags SugarAgents
// @Summary 创建sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarAgents true "创建sugar智能体表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarAgents/createSugarAgents [post]
func (sugarAgentsApi *SugarAgentsApi) CreateSugarAgents(c *gin.Context) {
	ctx := c.Request.Context()
	var sugarAgents sugar.SugarAgents
	err := c.ShouldBindJSON(&sugarAgents)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	sugarAgents.CreatedBy = &userIdStr

	err = sugarAgentsService.CreateSugarAgents(ctx, &sugarAgents)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteSugarAgents 删除sugar智能体表
// @Tags SugarAgents
// @Summary 删除sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarAgents true "删除sugar智能体表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarAgents/deleteSugarAgents [delete]
func (sugarAgentsApi *SugarAgentsApi) DeleteSugarAgents(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Query("id")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	err := sugarAgentsService.DeleteSugarAgents(ctx, id, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarAgentsByIds 批量删除sugar智能体表
// @Tags SugarAgents
// @Summary 批量删除sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param ids query []string true "批量删除"
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarAgents/deleteSugarAgentsByIds [delete]
func (sugarAgentsApi *SugarAgentsApi) DeleteSugarAgentsByIds(c *gin.Context) {
	ctx := c.Request.Context()
	ids := c.QueryArray("ids[]")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	err := sugarAgentsService.DeleteSugarAgentsByIds(ctx, ids, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarAgents 更新sugar智能体表
// @Tags SugarAgents
// @Summary 更新sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarAgents true "更新sugar智能体表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarAgents/updateSugarAgents [put]
func (sugarAgentsApi *SugarAgentsApi) UpdateSugarAgents(c *gin.Context) {
	ctx := c.Request.Context()
	var sugarAgents sugar.SugarAgents
	err := c.ShouldBindJSON(&sugarAgents)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	sugarAgents.UpdatedBy = &userIdStr

	err = sugarAgentsService.UpdateSugarAgents(ctx, sugarAgents, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarAgents 用id查询sugar智能体表
// @Tags SugarAgents
// @Summary 用id查询sugar智能体表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "ID"
// @Success 200 {object} response.Response{data=sugar.SugarAgents,msg=string} "查询成功"
// @Router /sugarAgents/findSugarAgents [get]
func (sugarAgentsApi *SugarAgentsApi) FindSugarAgents(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Query("id")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	resugarAgents, err := sugarAgentsService.GetSugarAgents(ctx, id, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(resugarAgents, c)
}

// GetSugarAgentsList 分页获取sugar智能体表列表
// @Tags SugarAgents
// @Summary 分页获取sugar智能体表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarAgentsSearch true "分页获取sugar智能体表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarAgents/getSugarAgentsList [get]
func (sugarAgentsApi *SugarAgentsApi) GetSugarAgentsList(c *gin.Context) {
	ctx := c.Request.Context()
	var pageInfo sugarReq.SugarAgentsSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	list, total, err := sugarAgentsService.GetSugarAgentsListByUser(ctx, pageInfo, userIdStr)
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
