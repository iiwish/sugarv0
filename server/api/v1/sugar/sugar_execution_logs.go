package sugar

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type SugarExecutionLogsApi struct {}



// CreateSugarExecutionLogs 创建sugar操作日志表
// @Tags SugarExecutionLogs
// @Summary 创建sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarExecutionLogs true "创建sugar操作日志表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarExecutionLogs/createSugarExecutionLogs [post]
func (sugarExecutionLogsApi *SugarExecutionLogsApi) CreateSugarExecutionLogs(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var sugarExecutionLogs sugar.SugarExecutionLogs
	err := c.ShouldBindJSON(&sugarExecutionLogs)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarExecutionLogsService.CreateSugarExecutionLogs(ctx,&sugarExecutionLogs)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteSugarExecutionLogs 删除sugar操作日志表
// @Tags SugarExecutionLogs
// @Summary 删除sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarExecutionLogs true "删除sugar操作日志表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarExecutionLogs/deleteSugarExecutionLogs [delete]
func (sugarExecutionLogsApi *SugarExecutionLogsApi) DeleteSugarExecutionLogs(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := sugarExecutionLogsService.DeleteSugarExecutionLogs(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarExecutionLogsByIds 批量删除sugar操作日志表
// @Tags SugarExecutionLogs
// @Summary 批量删除sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarExecutionLogs/deleteSugarExecutionLogsByIds [delete]
func (sugarExecutionLogsApi *SugarExecutionLogsApi) DeleteSugarExecutionLogsByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := sugarExecutionLogsService.DeleteSugarExecutionLogsByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarExecutionLogs 更新sugar操作日志表
// @Tags SugarExecutionLogs
// @Summary 更新sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarExecutionLogs true "更新sugar操作日志表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarExecutionLogs/updateSugarExecutionLogs [put]
func (sugarExecutionLogsApi *SugarExecutionLogsApi) UpdateSugarExecutionLogs(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var sugarExecutionLogs sugar.SugarExecutionLogs
	err := c.ShouldBindJSON(&sugarExecutionLogs)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarExecutionLogsService.UpdateSugarExecutionLogs(ctx,sugarExecutionLogs)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarExecutionLogs 用id查询sugar操作日志表
// @Tags SugarExecutionLogs
// @Summary 用id查询sugar操作日志表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询sugar操作日志表"
// @Success 200 {object} response.Response{data=sugar.SugarExecutionLogs,msg=string} "查询成功"
// @Router /sugarExecutionLogs/findSugarExecutionLogs [get]
func (sugarExecutionLogsApi *SugarExecutionLogsApi) FindSugarExecutionLogs(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	resugarExecutionLogs, err := sugarExecutionLogsService.GetSugarExecutionLogs(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(resugarExecutionLogs, c)
}
// GetSugarExecutionLogsList 分页获取sugar操作日志表列表
// @Tags SugarExecutionLogs
// @Summary 分页获取sugar操作日志表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarExecutionLogsSearch true "分页获取sugar操作日志表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarExecutionLogs/getSugarExecutionLogsList [get]
func (sugarExecutionLogsApi *SugarExecutionLogsApi) GetSugarExecutionLogsList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo sugarReq.SugarExecutionLogsSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := sugarExecutionLogsService.GetSugarExecutionLogsInfoList(ctx,pageInfo)
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

// GetSugarExecutionLogsPublic 不需要鉴权的sugar操作日志表接口
// @Tags SugarExecutionLogs
// @Summary 不需要鉴权的sugar操作日志表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarExecutionLogs/getSugarExecutionLogsPublic [get]
func (sugarExecutionLogsApi *SugarExecutionLogsApi) GetSugarExecutionLogsPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    sugarExecutionLogsService.GetSugarExecutionLogsPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的sugar操作日志表接口信息",
    }, "获取成功", c)
}
