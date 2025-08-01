package sugar

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type SugarRowLevelOverridesApi struct {}



// CreateSugarRowLevelOverrides 创建Sugar行级权限豁免表
// @Tags SugarRowLevelOverrides
// @Summary 创建Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarRowLevelOverrides true "创建Sugar行级权限豁免表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarRowLevelOverrides/createSugarRowLevelOverrides [post]
func (sugarRowLevelOverridesApi *SugarRowLevelOverridesApi) CreateSugarRowLevelOverrides(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var sugarRowLevelOverrides sugar.SugarRowLevelOverrides
	err := c.ShouldBindJSON(&sugarRowLevelOverrides)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarRowLevelOverridesService.CreateSugarRowLevelOverrides(ctx,&sugarRowLevelOverrides)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteSugarRowLevelOverrides 删除Sugar行级权限豁免表
// @Tags SugarRowLevelOverrides
// @Summary 删除Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarRowLevelOverrides true "删除Sugar行级权限豁免表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarRowLevelOverrides/deleteSugarRowLevelOverrides [delete]
func (sugarRowLevelOverridesApi *SugarRowLevelOverridesApi) DeleteSugarRowLevelOverrides(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := sugarRowLevelOverridesService.DeleteSugarRowLevelOverrides(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarRowLevelOverridesByIds 批量删除Sugar行级权限豁免表
// @Tags SugarRowLevelOverrides
// @Summary 批量删除Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarRowLevelOverrides/deleteSugarRowLevelOverridesByIds [delete]
func (sugarRowLevelOverridesApi *SugarRowLevelOverridesApi) DeleteSugarRowLevelOverridesByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := sugarRowLevelOverridesService.DeleteSugarRowLevelOverridesByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarRowLevelOverrides 更新Sugar行级权限豁免表
// @Tags SugarRowLevelOverrides
// @Summary 更新Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarRowLevelOverrides true "更新Sugar行级权限豁免表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarRowLevelOverrides/updateSugarRowLevelOverrides [put]
func (sugarRowLevelOverridesApi *SugarRowLevelOverridesApi) UpdateSugarRowLevelOverrides(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var sugarRowLevelOverrides sugar.SugarRowLevelOverrides
	err := c.ShouldBindJSON(&sugarRowLevelOverrides)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarRowLevelOverridesService.UpdateSugarRowLevelOverrides(ctx,sugarRowLevelOverrides)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarRowLevelOverrides 用id查询Sugar行级权限豁免表
// @Tags SugarRowLevelOverrides
// @Summary 用id查询Sugar行级权限豁免表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询Sugar行级权限豁免表"
// @Success 200 {object} response.Response{data=sugar.SugarRowLevelOverrides,msg=string} "查询成功"
// @Router /sugarRowLevelOverrides/findSugarRowLevelOverrides [get]
func (sugarRowLevelOverridesApi *SugarRowLevelOverridesApi) FindSugarRowLevelOverrides(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	resugarRowLevelOverrides, err := sugarRowLevelOverridesService.GetSugarRowLevelOverrides(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(resugarRowLevelOverrides, c)
}
// GetSugarRowLevelOverridesList 分页获取Sugar行级权限豁免表列表
// @Tags SugarRowLevelOverrides
// @Summary 分页获取Sugar行级权限豁免表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarRowLevelOverridesSearch true "分页获取Sugar行级权限豁免表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarRowLevelOverrides/getSugarRowLevelOverridesList [get]
func (sugarRowLevelOverridesApi *SugarRowLevelOverridesApi) GetSugarRowLevelOverridesList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo sugarReq.SugarRowLevelOverridesSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := sugarRowLevelOverridesService.GetSugarRowLevelOverridesInfoList(ctx,pageInfo)
	if err != nil {
	    global.GVA_LOG.Error("获取失败!", zap.Error(err))
        response.FailWithMessage("获取失败:" + err.Error(), c)
        return
    }
    response.OkWithDetailed(response.PageResult{
        List:     list,
        Total:    total,
        Page:     pageInfo.Page,
        PageSize: pageInfo.PageSize,
    }, "获取成功", c)
}

// GetSugarRowLevelOverridesPublic 不需要鉴权的Sugar行级权限豁免表接口
// @Tags SugarRowLevelOverrides
// @Summary 不需要鉴权的Sugar行级权限豁免表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarRowLevelOverrides/getSugarRowLevelOverridesPublic [get]
func (sugarRowLevelOverridesApi *SugarRowLevelOverridesApi) GetSugarRowLevelOverridesPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    sugarRowLevelOverridesService.GetSugarRowLevelOverridesPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的Sugar行级权限豁免表接口信息",
    }, "获取成功", c)
}
