package sugar

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type SugarDbConnectionsApi struct {}



// CreateSugarDbConnections 创建Sugar数据库配置表
// @Tags SugarDbConnections
// @Summary 创建Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarDbConnections true "创建Sugar数据库配置表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarDbConnections/createSugarDbConnections [post]
func (sugarDbConnectionsApi *SugarDbConnectionsApi) CreateSugarDbConnections(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var sugarDbConnections sugar.SugarDbConnections
	err := c.ShouldBindJSON(&sugarDbConnections)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarDbConnectionsService.CreateSugarDbConnections(ctx,&sugarDbConnections)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteSugarDbConnections 删除Sugar数据库配置表
// @Tags SugarDbConnections
// @Summary 删除Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarDbConnections true "删除Sugar数据库配置表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarDbConnections/deleteSugarDbConnections [delete]
func (sugarDbConnectionsApi *SugarDbConnectionsApi) DeleteSugarDbConnections(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := sugarDbConnectionsService.DeleteSugarDbConnections(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarDbConnectionsByIds 批量删除Sugar数据库配置表
// @Tags SugarDbConnections
// @Summary 批量删除Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarDbConnections/deleteSugarDbConnectionsByIds [delete]
func (sugarDbConnectionsApi *SugarDbConnectionsApi) DeleteSugarDbConnectionsByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := sugarDbConnectionsService.DeleteSugarDbConnectionsByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarDbConnections 更新Sugar数据库配置表
// @Tags SugarDbConnections
// @Summary 更新Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarDbConnections true "更新Sugar数据库配置表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarDbConnections/updateSugarDbConnections [put]
func (sugarDbConnectionsApi *SugarDbConnectionsApi) UpdateSugarDbConnections(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var sugarDbConnections sugar.SugarDbConnections
	err := c.ShouldBindJSON(&sugarDbConnections)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarDbConnectionsService.UpdateSugarDbConnections(ctx,sugarDbConnections)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarDbConnections 用id查询Sugar数据库配置表
// @Tags SugarDbConnections
// @Summary 用id查询Sugar数据库配置表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "用id查询Sugar数据库配置表"
// @Success 200 {object} response.Response{data=sugar.SugarDbConnections,msg=string} "查询成功"
// @Router /sugarDbConnections/findSugarDbConnections [get]
func (sugarDbConnectionsApi *SugarDbConnectionsApi) FindSugarDbConnections(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	resugarDbConnections, err := sugarDbConnectionsService.GetSugarDbConnections(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(resugarDbConnections, c)
}
// GetSugarDbConnectionsList 分页获取Sugar数据库配置表列表
// @Tags SugarDbConnections
// @Summary 分页获取Sugar数据库配置表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarDbConnectionsSearch true "分页获取Sugar数据库配置表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarDbConnections/getSugarDbConnectionsList [get]
func (sugarDbConnectionsApi *SugarDbConnectionsApi) GetSugarDbConnectionsList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo sugarReq.SugarDbConnectionsSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := sugarDbConnectionsService.GetSugarDbConnectionsInfoList(ctx,pageInfo)
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

// GetSugarDbConnectionsPublic 不需要鉴权的Sugar数据库配置表接口
// @Tags SugarDbConnections
// @Summary 不需要鉴权的Sugar数据库配置表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarDbConnections/getSugarDbConnectionsPublic [get]
func (sugarDbConnectionsApi *SugarDbConnectionsApi) GetSugarDbConnectionsPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    sugarDbConnectionsService.GetSugarDbConnectionsPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的Sugar数据库配置表接口信息",
    }, "获取成功", c)
}
