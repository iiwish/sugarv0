package sugar

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SugarFormulaQueryApi struct{}

// ExecuteSugarCalc 执行 SUGAR.CALC 公式
// @Tags SugarFormulaQuery
// @Summary 执行 SUGAR.CALC 公式
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugarReq.SugarFormulaCalcRequest true "SUGAR.CALC 公式请求"
// @Success 200 {object} response.Response{data=sugarRes.SugarFormulaCalcResponse,msg=string} "执行成功"
// @Router /sugarFormulaQuery/executeCalc [post]
func (s *SugarFormulaQueryApi) ExecuteSugarCalc(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFormulaCalcRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := sugarFormulaQueryService.ExecuteCalcFormula(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("SUGAR.CALC 执行失败!", zap.Error(err))
		response.FailWithMessage("SUGAR.CALC 执行失败: "+err.Error(), c)
		return
	}

	// 检查业务层返回的错误
	if result.Error != "" {
		response.FailWithMessage(result.Error, c)
		return
	}

	response.OkWithData(result, c)
}

// ExecuteSugarGet 执行 SUGAR.GET 公式
// @Tags SugarFormulaQuery
// @Summary 执行 SUGAR.GET 公式
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugarReq.SugarFormulaGetRequest true "SUGAR.GET 公式请求"
// @Success 200 {object} response.Response{data=sugarRes.SugarFormulaGetResponse,msg=string} "执行成功"
// @Router /sugarFormulaQuery/executeGet [post]
func (s *SugarFormulaQueryApi) ExecuteSugarGet(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFormulaGetRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := sugarFormulaQueryService.ExecuteGetFormula(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("SUGAR.GET 执行失败!", zap.Error(err))
		response.FailWithMessage("SUGAR.GET 执行失败: "+err.Error(), c)
		return
	}

	// 检查业务层返回的错误
	if result.Error != "" {
		response.FailWithMessage(result.Error, c)
		return
	}

	response.OkWithData(result, c)
}

// ExecuteAiFetch 执行 AIFETCH 公式
// @Tags SugarFormulaQuery
// @Summary 执行 AIFETCH 公式
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugarReq.SugarFormulaAiFetchRequest true "AIFETCH 公式请求"
// @Success 200 {object} response.Response{data=sugarRes.SugarFormulaAiResponse,msg=string} "执行成功"
// @Router /sugarFormulaQuery/executeAiFetch [post]
func (s *SugarFormulaQueryApi) ExecuteAiFetch(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFormulaAiFetchRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := service.ServiceGroupApp.SugarServiceGroup.GetSugarFormulaAiService().ExecuteAiFetchFormula(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("AIFETCH 执行失败!", zap.Error(err))
		response.FailWithMessage("AIFETCH 执行失败: "+err.Error(), c)
		return
	}

	// 检查业务层返回的错误
	if result.Error != "" {
		response.FailWithMessage(result.Error, c)
		return
	}

	response.OkWithData(result, c)
}

// ExecuteAiExplain 执行 AIEXPLAINRange 公式
// @Tags SugarFormulaQuery
// @Summary 执行 AIEXPLAINRange 公式
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body sugarReq.SugarFormulaAiExplainRequest true "AIEXPLAINRange 公式请求"
// @Success 200 {object} response.Response{data=sugarRes.SugarFormulaAiResponse,msg=string} "执行成功"
// @Router /sugarFormulaQuery/executeAiExplainRange [post]
func (s *SugarFormulaQueryApi) ExecuteAiExplainRange(c *gin.Context) {
	ctx := c.Request.Context()
	var req sugarReq.SugarFormulaAiExplainRangeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userId := utils.GetUserID(c)
	userIdStr := strconv.Itoa(int(userId))

	result, err := service.ServiceGroupApp.SugarServiceGroup.GetSugarFormulaAiService().ExecuteAiExplainFormula(ctx, &req, userIdStr)
	if err != nil {
		global.GVA_LOG.Error("AIEXPLAINRange 执行失败!", zap.Error(err))
		response.FailWithMessage("AIEXPLAINRange 执行失败: "+err.Error(), c)
		return
	}

	// 检查业务层返回的错误
	if result.Error != "" {
		response.FailWithMessage(result.Error, c)
		return
	}

	response.OkWithData(result, c)
}
