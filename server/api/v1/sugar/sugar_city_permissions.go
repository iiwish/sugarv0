package sugar

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type SugarCityPermissionsApi struct {}



// CreateSugarCityPermissions 创建sugarCityPermissions表
// @Tags SugarCityPermissions
// @Summary 创建sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarCityPermissions true "创建sugarCityPermissions表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarCityPermissions/createSugarCityPermissions [post]
func (sugarCityPermissionsApi *SugarCityPermissionsApi) CreateSugarCityPermissions(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var sugarCityPermissions sugar.SugarCityPermissions
	err := c.ShouldBindJSON(&sugarCityPermissions)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarCityPermissionsService.CreateSugarCityPermissions(ctx,&sugarCityPermissions)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteSugarCityPermissions 删除sugarCityPermissions表
// @Tags SugarCityPermissions
// @Summary 删除sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarCityPermissions true "删除sugarCityPermissions表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarCityPermissions/deleteSugarCityPermissions [delete]
func (sugarCityPermissionsApi *SugarCityPermissionsApi) DeleteSugarCityPermissions(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := sugarCityPermissionsService.DeleteSugarCityPermissions(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarCityPermissionsByIds 批量删除sugarCityPermissions表
// @Tags SugarCityPermissions
// @Summary 批量删除sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarCityPermissions/deleteSugarCityPermissionsByIds [delete]
func (sugarCityPermissionsApi *SugarCityPermissionsApi) DeleteSugarCityPermissionsByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := sugarCityPermissionsService.DeleteSugarCityPermissionsByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarCityPermissions 更新sugarCityPermissions表
// @Tags SugarCityPermissions
// @Summary 更新sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarCityPermissions true "更新sugarCityPermissions表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarCityPermissions/updateSugarCityPermissions [put]
func (sugarCityPermissionsApi *SugarCityPermissionsApi) UpdateSugarCityPermissions(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var sugarCityPermissions sugar.SugarCityPermissions
	err := c.ShouldBindJSON(&sugarCityPermissions)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarCityPermissionsService.UpdateSugarCityPermissions(ctx,sugarCityPermissions)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarCityPermissions 用id查询sugarCityPermissions表
// @Tags SugarCityPermissions
// @Summary 用id查询sugarCityPermissions表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询sugarCityPermissions表"
// @Success 200 {object} response.Response{data=sugar.SugarCityPermissions,msg=string} "查询成功"
// @Router /sugarCityPermissions/findSugarCityPermissions [get]
func (sugarCityPermissionsApi *SugarCityPermissionsApi) FindSugarCityPermissions(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	resugarCityPermissions, err := sugarCityPermissionsService.GetSugarCityPermissions(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(resugarCityPermissions, c)
}
// GetSugarCityPermissionsList 分页获取sugarCityPermissions表列表
// @Tags SugarCityPermissions
// @Summary 分页获取sugarCityPermissions表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarCityPermissionsSearch true "分页获取sugarCityPermissions表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarCityPermissions/getSugarCityPermissionsList [get]
func (sugarCityPermissionsApi *SugarCityPermissionsApi) GetSugarCityPermissionsList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo sugarReq.SugarCityPermissionsSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := sugarCityPermissionsService.GetSugarCityPermissionsInfoList(ctx,pageInfo)
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

// GetSugarCityPermissionsPublic 不需要鉴权的sugarCityPermissions表接口
// @Tags SugarCityPermissions
// @Summary 不需要鉴权的sugarCityPermissions表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarCityPermissions/getSugarCityPermissionsPublic [get]
func (sugarCityPermissionsApi *SugarCityPermissionsApi) GetSugarCityPermissionsPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    sugarCityPermissionsService.GetSugarCityPermissionsPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的sugarCityPermissions表接口信息",
    }, "获取成功", c)
}
