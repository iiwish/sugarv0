package sugar

import (
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
    "github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
    "github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
    sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

type SugarTeamMembersApi struct {}



// CreateSugarTeamMembers 创建sugarTeamMembers表
// @Tags SugarTeamMembers
// @Summary 创建sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarTeamMembers true "创建sugarTeamMembers表"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /sugarTeamMembers/createSugarTeamMembers [post]
func (sugarTeamMembersApi *SugarTeamMembersApi) CreateSugarTeamMembers(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var sugarTeamMembers sugar.SugarTeamMembers
	err := c.ShouldBindJSON(&sugarTeamMembers)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarTeamMembersService.CreateSugarTeamMembers(ctx,&sugarTeamMembers)
	if err != nil {
        global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败:" + err.Error(), c)
		return
	}
    response.OkWithMessage("创建成功", c)
}

// DeleteSugarTeamMembers 删除sugarTeamMembers表
// @Tags SugarTeamMembers
// @Summary 删除sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarTeamMembers true "删除sugarTeamMembers表"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /sugarTeamMembers/deleteSugarTeamMembers [delete]
func (sugarTeamMembersApi *SugarTeamMembersApi) DeleteSugarTeamMembers(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	err := sugarTeamMembersService.DeleteSugarTeamMembers(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// DeleteSugarTeamMembersByIds 批量删除sugarTeamMembers表
// @Tags SugarTeamMembers
// @Summary 批量删除sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /sugarTeamMembers/deleteSugarTeamMembersByIds [delete]
func (sugarTeamMembersApi *SugarTeamMembersApi) DeleteSugarTeamMembersByIds(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	ids := c.QueryArray("ids[]")
	err := sugarTeamMembersService.DeleteSugarTeamMembersByIds(ctx,ids)
	if err != nil {
        global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("批量删除成功", c)
}

// UpdateSugarTeamMembers 更新sugarTeamMembers表
// @Tags SugarTeamMembers
// @Summary 更新sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugar.SugarTeamMembers true "更新sugarTeamMembers表"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /sugarTeamMembers/updateSugarTeamMembers [put]
func (sugarTeamMembersApi *SugarTeamMembersApi) UpdateSugarTeamMembers(c *gin.Context) {
    // 从ctx获取标准context进行业务行为
    ctx := c.Request.Context()

	var sugarTeamMembers sugar.SugarTeamMembers
	err := c.ShouldBindJSON(&sugarTeamMembers)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	err = sugarTeamMembersService.UpdateSugarTeamMembers(ctx,sugarTeamMembers)
	if err != nil {
        global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败:" + err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

// FindSugarTeamMembers 用id查询sugarTeamMembers表
// @Tags SugarTeamMembers
// @Summary 用id查询sugarTeamMembers表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query int true "用id查询sugarTeamMembers表"
// @Success 200 {object} response.Response{data=sugar.SugarTeamMembers,msg=string} "查询成功"
// @Router /sugarTeamMembers/findSugarTeamMembers [get]
func (sugarTeamMembersApi *SugarTeamMembersApi) FindSugarTeamMembers(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	id := c.Query("id")
	resugarTeamMembers, err := sugarTeamMembersService.GetSugarTeamMembers(ctx,id)
	if err != nil {
        global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败:" + err.Error(), c)
		return
	}
	response.OkWithData(resugarTeamMembers, c)
}
// GetSugarTeamMembersList 分页获取sugarTeamMembers表列表
// @Tags SugarTeamMembers
// @Summary 分页获取sugarTeamMembers表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query sugarReq.SugarTeamMembersSearch true "分页获取sugarTeamMembers表列表"
// @Success 200 {object} response.Response{data=response.PageResult,msg=string} "获取成功"
// @Router /sugarTeamMembers/getSugarTeamMembersList [get]
func (sugarTeamMembersApi *SugarTeamMembersApi) GetSugarTeamMembersList(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

	var pageInfo sugarReq.SugarTeamMembersSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	list, total, err := sugarTeamMembersService.GetSugarTeamMembersInfoList(ctx,pageInfo)
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

// GetSugarTeamMembersPublic 不需要鉴权的sugarTeamMembers表接口
// @Tags SugarTeamMembers
// @Summary 不需要鉴权的sugarTeamMembers表接口
// @Accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarTeamMembers/getSugarTeamMembersPublic [get]
func (sugarTeamMembersApi *SugarTeamMembersApi) GetSugarTeamMembersPublic(c *gin.Context) {
    // 创建业务用Context
    ctx := c.Request.Context()

    // 此接口不需要鉴权
    // 示例为返回了一个固定的消息接口，一般本接口用于C端服务，需要自己实现业务逻辑
    sugarTeamMembersService.GetSugarTeamMembersPublic(ctx)
    response.OkWithDetailed(gin.H{
       "info": "不需要鉴权的sugarTeamMembers表接口信息",
    }, "获取成功", c)
}
