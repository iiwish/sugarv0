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

type SugarTeamsApi struct{}

// CreateSugarTeams 创建团队信息表
// @Tags SugarTeams
// @Summary 创建团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.SugarTeamsCreateRequest true "创建团队信息表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarTeams/createSugarTeams [post]
func (sugarTeamsApi *SugarTeamsApi) CreateSugarTeams(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var req sugarReq.SugarTeamsCreateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	// 转换为模型实体
	sugarTeams := sugar.SugarTeams{
		TeamName:   &req.TeamName,
		IsPersonal: &req.IsPersonal,
		OwnerId:    &userIdStr,
		CreatedBy:  &userIdStr,
	}

	err = sugarTeamsService.CreateSugarTeams(ctx, &sugarTeams)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteSugarTeams 删除团队信息表
// @Tags SugarTeams
// @Summary 删除团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarTeams true "删除团队信息表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarTeams/deleteSugarTeams [delete]
func (sugarTeamsApi *SugarTeamsApi) DeleteSugarTeams(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	err := sugarTeamsService.DeleteSugarTeams(ctx, id, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarTeamsByIds 批量删除团队信息表
// @Tags SugarTeams
// @Summary 批量删除团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarTeams/deleteSugarTeamsByIds [delete]
func (sugarTeamsApi *SugarTeamsApi) DeleteSugarTeamsByIds(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	err := sugarTeamsService.DeleteSugarTeamsByIds(ctx, ids, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarTeams 更新团队信息表
// @Tags SugarTeams
// @Summary 更新团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.SugarTeamsUpdateRequest true "更新团队信息表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarTeams/updateSugarTeams [put]
func (sugarTeamsApi *SugarTeamsApi) UpdateSugarTeams(c *gin.Context) {
	// 从ctx获取标准context进行业务行为
	ctx := c.Request.Context()

	var req sugarReq.SugarTeamsUpdateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	// 转换为模型实体
	sugarTeams := sugar.SugarTeams{
		Id:        &req.Id,
		TeamName:  req.TeamName,
		UpdatedBy: &userIdStr,
	}

	err = sugarTeamsService.UpdateSugarTeams(ctx, sugarTeams)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarTeams 用id查询团队信息表
// @Tags SugarTeams
// @Summary 用id查询团队信息表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "用id查询团队信息表"
// @Success 200 {object} response.Response{data=sugar.SugarTeams,msg=string} "查询成功"
// @Router /sugarTeams/findSugarTeams [get]
func (sugarTeamsApi *SugarTeamsApi) FindSugarTeams(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	id := c.Query("id")
	resugarTeams, err := sugarTeamsService.GetSugarTeams(ctx, id)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(resugarTeams, c)
}

// GetSugarTeamsList 分页获取团队信息表列表
// @Tags SugarTeams
// @Summary 分页获取团队信息表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarTeamsSearch true "分页获取团队信息表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarTeams/getSugarTeamsList [get]
func (sugarTeamsApi *SugarTeamsApi) GetSugarTeamsList(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	var pageInfo sugarReq.SugarTeamsSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// userId := utils.GetUserID(c)
	// userIdStr := strconv.Itoa(int(userId))

	list, total, err := sugarTeamsService.GetSugarTeamsInfoList(ctx, pageInfo)
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

// GetSugarTeamsPublic 不需要鉴权的团队信息表接口
// @Tags SugarTeams
// @Summary 不需要鉴权的团队信息表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarTeams/getSugarTeamsPublic [get]
func (sugarTeamsApi *SugarTeamsApi) GetSugarTeamsPublic(c *gin.Context) {
	// 创建业务用Context
	ctx := c.Request.Context()

	// 此接口不需要鉴权
	// 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
	sugarTeamsService.GetSugarTeamsPublic(ctx)
	response.OkWithDetailed(gin.H{
		"info": "不需要鉴权的团队信息表接口信息",
	}, "获取成功", c)
}
