package sugar

import (
	"encoding/json"
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

// CreateWorkbookFile 创建新的工作簿文件
// @Tags SugarWorkspaces
// @Summary 创建新的工作簿文件
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "创建工作簿文件数据"
// @Success 200 {object} response.Response{data=sugar.SugarWorkspaces,msg=string} "创建成功"
// @Router /sugarWorkspaces/createWorkbookFile [post]
func (sugarWorkspacesApi *SugarWorkspacesApi) CreateWorkbookFile(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		Name     string  `json:"name" binding:"required"`
		ParentId *string `json:"parentId"`
		TeamId   string  `json:"teamId" binding:"required"`
		Content  any     `json:"content"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	// 如果没有提供内容，使用默认的工作簿数据
	var defaultContent any
	if req.Content != nil {
		defaultContent = req.Content
	} else {
		defaultContent = map[string]interface{}{
			"id":         "",
			"sheetOrder": []string{"sheet-001"},
			"name":       req.Name,
			"appVersion": "0.1.0",
			"locale":     "zh-CN",
			"styles":     map[string]interface{}{},
			"sheets": map[string]interface{}{
				"sheet-001": map[string]interface{}{
					"id":   "sheet-001",
					"name": "Sheet1",
					"cellData": map[string]interface{}{
						"0": map[string]interface{}{
							"0": map[string]interface{}{"v": "新建表格"},
						},
					},
				},
			},
		}
	}

	// 转换为JSON
	contentBytes, err := json.Marshal(defaultContent)
	if err != nil {
		response.FailWithMessage("内容格式错误", c)
		return
	}

	workspace, err := sugarWorkspacesService.CreateWorkbookFile(ctx, req.Name, req.ParentId, req.TeamId, userIdStr, contentBytes)
	if err != nil {
		global.GVA_LOG.Error("创建工作簿文件失败!", zap.Error(err))
		response.FailWithMessage("创建失败:"+err.Error(), c)
		return
	}

	response.OkWithData(workspace, c)
}

// SaveWorkbookContent 保存工作簿内容
// @Tags SugarWorkspaces
// @Summary 保存工作簿内容
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body object true "保存工作簿内容数据"
// @Success 200 {object} response.Response{msg=string} "保存成功"
// @Router /sugarWorkspaces/saveWorkbookContent [put]
func (sugarWorkspacesApi *SugarWorkspacesApi) SaveWorkbookContent(c *gin.Context) {
	ctx := c.Request.Context()
	var req struct {
		Id      string `json:"id" binding:"required"`
		Content any    `json:"content" binding:"required"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	// 转换为JSON
	contentBytes, err := json.Marshal(req.Content)
	if err != nil {
		response.FailWithMessage("内容格式错误", c)
		return
	}

	err = sugarWorkspacesService.SaveWorkbookContent(ctx, req.Id, contentBytes, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("保存工作簿内容失败!", zap.Error(err))
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("保存成功", c)
}

// GetWorkbookContent 获取工作簿内容
// @Tags SugarWorkspaces
// @Summary 获取工作簿内容
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param id query string true "文件ID"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /sugarWorkspaces/getWorkbookContent [get]
func (sugarWorkspacesApi *SugarWorkspacesApi) GetWorkbookContent(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Query("id")
	if id == "" {
		response.FailWithMessage("文件ID不能为空", c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	content, err := sugarWorkspacesService.GetWorkbookContent(ctx, id, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("获取工作簿内容失败!", zap.Error(err))
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}

	response.OkWithData(content, c)
}
