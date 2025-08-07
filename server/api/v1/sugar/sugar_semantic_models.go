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

type SugarSemanticModelsApi struct{}

// CreateSugarSemanticModels 创建Sugar指标语义表
// @Tags SugarSemanticModels
// @Summary 创建Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarSemanticModels true "创建Sugar指标语义表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarSemanticModels/createSugarSemanticModels [post]
func (sugarSemanticModelsApi *SugarSemanticModelsApi) CreateSugarSemanticModels(c *gin.Context) {
	ctx := c.Request.Context()
	var sugarSemanticModels sugar.SugarSemanticModels
	err := c.ShouldBindJSON(&sugarSemanticModels)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	sugarSemanticModels.CreatedBy = &userIdStr

	err = sugarSemanticModelsService.CreateSugarSemanticModels(ctx, &sugarSemanticModels)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

// DeleteSugarSemanticModels 删除Sugar指标语义表
// @Tags SugarSemanticModels
// @Summary 删除Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "指标ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarSemanticModels/deleteSugarSemanticModels [delete]
func (sugarSemanticModelsApi *SugarSemanticModelsApi) DeleteSugarSemanticModels(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Query("id")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	err := sugarSemanticModelsService.DeleteSugarSemanticModels(ctx, id, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarSemanticModelsByIds 批量删除Sugar指标语义表
// @Tags SugarSemanticModels
// @Summary 批量删除Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param ids query []string true "指标ID列表"
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarSemanticModels/deleteSugarSemanticModelsByIds [delete]
func (sugarSemanticModelsApi *SugarSemanticModelsApi) DeleteSugarSemanticModelsByIds(c *gin.Context) {
	ctx := c.Request.Context()
	ids := c.QueryArray("ids[]")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	err := sugarSemanticModelsService.DeleteSugarSemanticModelsByIds(ctx, ids, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarSemanticModels 更新Sugar指标语义表
// @Tags SugarSemanticModels
// @Summary 更新Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarSemanticModels true "更新Sugar指标语义表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarSemanticModels/updateSugarSemanticModels [put]
func (sugarSemanticModelsApi *SugarSemanticModelsApi) UpdateSugarSemanticModels(c *gin.Context) {
	ctx := c.Request.Context()
	var sugarSemanticModels sugar.SugarSemanticModels
	err := c.ShouldBindJSON(&sugarSemanticModels)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))
	sugarSemanticModels.UpdatedBy = &userIdStr

	err = sugarSemanticModelsService.UpdateSugarSemanticModels(ctx, sugarSemanticModels, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarSemanticModels 用id查询Sugar指标语义表
// @Tags SugarSemanticModels
// @Summary 用id查询Sugar指标语义表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "指标ID"
// @Success 200 {object} response.Response{data=sugar.SugarSemanticModels,msg=string} "查询成功"
// @Router /sugarSemanticModels/findSugarSemanticModels [get]
func (sugarSemanticModelsApi *SugarSemanticModelsApi) FindSugarSemanticModels(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Query("id")
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	resugarSemanticModels, err := sugarSemanticModelsService.GetSugarSemanticModels(ctx, id, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:"+err.Error(), c)
		return
	}
	response.OkWithData(resugarSemanticModels, c)
}

// GetSugarSemanticModelsList 分页获取Sugar指标语义表列表
// @Tags SugarSemanticModels
// @Summary 分页获取Sugar指标语义表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarSemanticModelsSearch true "分页获取Sugar指标语义表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarSemanticModels/getSugarSemanticModelsList [get]
func (sugarSemanticModelsApi *SugarSemanticModelsApi) GetSugarSemanticModelsList(c *gin.Context) {
	ctx := c.Request.Context()
	var pageInfo sugarReq.SugarSemanticModelsSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	list, total, err := sugarSemanticModelsService.GetSugarSemanticModelsListByUser(ctx, pageInfo, userIdStr)
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
