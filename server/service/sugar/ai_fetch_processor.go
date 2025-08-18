package sugar

import (
	"context"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	sugarRes "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/sugar/advanced_contribution_analyzer"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// AiFetchProcessor AI获取处理器 - 负责协调整个AIFETCH流程
type AiFetchProcessor struct {
	dataProcessor          *DataProcessor
	contributionAnalyzer   *ContributionAnalyzer
	anonymizationProcessor *AnonymizationProcessor
	aiInteractionManager   *AIInteractionManager
	executionLogger        *ExecutionLogger
}

// NewAiFetchProcessor 创建AI获取处理器
func NewAiFetchProcessor(advancedAnalyzer *advanced_contribution_analyzer.AdvancedContributionService) *AiFetchProcessor {
	return &AiFetchProcessor{
		dataProcessor:          NewDataProcessor(),
		contributionAnalyzer:   NewContributionAnalyzer(advancedAnalyzer),
		anonymizationProcessor: NewAnonymizationProcessor(),
		aiInteractionManager:   NewAIInteractionManager(),
		executionLogger:        NewExecutionLogger(),
	}
}

// ProcessRequest 处理AIFETCH请求的主流程
func (p *AiFetchProcessor) ProcessRequest(ctx context.Context, req *sugarReq.SugarFormulaAiFetchRequest, userId string) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始处理AIFETCH请求",
		zap.String("agentName", req.AgentName),
		zap.String("description", req.Description),
		zap.String("userId", userId))

	// 1. 获取Agent信息和LLM配置
	agent, llmConfig, err := p.dataProcessor.GetAgentAndLLMConfig(ctx, req.AgentName, userId)
	if err != nil {
		return sugarRes.NewAiErrorResponse(err.Error()), nil
	}

	// 2. 创建执行日志
	logCtx, err := p.executionLogger.CreateLog(ctx, req, userId, agent.Id)
	if err != nil {
		global.GVA_LOG.Error("创建执行日志失败", zap.Error(err))
	}

	// 3. 构建系统提示词和用户消息
	systemPrompt := p.aiInteractionManager.BuildSystemPrompt(agent, userId)
	userMessage := p.aiInteractionManager.BuildUserMessage(req.Description, agent.Semantic, req.DataRange)

	// 记录到日志
	if logCtx != nil {
		p.executionLogger.RecordAIPrompts(ctx, logCtx, systemPrompt, userMessage, llmConfig)
	}

	// 4. 调用LLM获取工具调用指令
	llmResponse, err := p.aiInteractionManager.CallLLMWithTools(ctx, llmConfig, systemPrompt, userMessage)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.FinishWithError(ctx, logCtx, "AI分析失败: "+err.Error())
		}
		return sugarRes.NewAiErrorResponse("AI分析失败: " + err.Error()), nil
	}

	// 记录LLM响应
	if logCtx != nil {
		p.executionLogger.RecordLLMResponse(ctx, logCtx, llmResponse, llmConfig.ModelName)
	}

	// 5. 处理LLM响应和工具调用
	result, err := p.processLLMResponse(ctx, llmResponse, userId, req, agent, llmConfig, logCtx)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.FinishWithError(ctx, logCtx, err.Error())
		}
		return result, err
	}

	// 6. 记录成功日志
	if logCtx != nil && result != nil {
		p.executionLogger.FinishWithSuccess(ctx, logCtx, result.Text)
	}

	return result, nil
}

// processLLMResponse 处理LLM响应，可能包含工具调用
func (p *AiFetchProcessor) processLLMResponse(ctx context.Context, llmResponse string, userId string, req *sugarReq.SugarFormulaAiFetchRequest, agent *sugar.SugarAgents, llmConfig *system.LLMConfig, logCtx *ExecutionLogContext) (*sugarRes.SugarFormulaAiResponse, error) {
	global.GVA_LOG.Info("开始处理LLM响应", zap.String("userId", userId))

	// 解析工具调用
	toolCallResp, err := p.aiInteractionManager.ParseToolCallResponse(llmResponse)
	if err != nil {
		// 不是工具调用，直接返回文本响应
		return sugarRes.NewAiSuccessResponseWithText(llmResponse), nil
	}

	if len(toolCallResp.Content) == 0 {
		return sugarRes.NewAiSuccessResponseWithText(llmResponse), nil
	}

	// 处理工具调用
	toolCall := toolCallResp.Content[0]
	global.GVA_LOG.Info("处理工具调用",
		zap.String("functionName", toolCall.Function.Name),
		zap.String("arguments", toolCall.Function.Arguments))

	switch toolCall.Function.Name {
	case "smart_anonymized_analyzer":
		return p.handleSmartAnonymizedAnalyzer(ctx, toolCall, logCtx, req, agent, llmConfig)
	case "data_scope_explorer":
		return p.handleDataScopeExplorer(ctx, toolCall, logCtx)
	case "anonymized_data_analyzer":
		return p.handleAnonymizedDataAnalyzer(ctx, toolCall, logCtx, req, agent, llmConfig)
	default:
		return sugarRes.NewAiSuccessResponseWithText(llmResponse), nil
	}
}

// handleSmartAnonymizedAnalyzer 处理智能匿名化分析工具调用
func (p *AiFetchProcessor) handleSmartAnonymizedAnalyzer(ctx context.Context, toolCall system.OpenAIToolCall, logCtx *ExecutionLogContext, req *sugarReq.SugarFormulaAiFetchRequest, agent *sugar.SugarAgents, llmConfig *system.LLMConfig) (*sugarRes.SugarFormulaAiResponse, error) {
	toolCallStartTime := time.Now()

	// 解析工具参数
	params, err := p.aiInteractionManager.ParseSmartAnalyzerParams(toolCall.Function.Arguments)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, nil, err.Error(), toolCallStartTime)
		}
		return sugarRes.NewAiErrorResponse("解析工具调用参数失败: " + err.Error()), nil
	}

	global.GVA_LOG.Info("提取智能匿名化工具参数",
		zap.String("modelName", params.ModelName),
		zap.String("targetMetric", params.TargetMetric),
		zap.Strings("groupByDimensions", params.GroupByDimensions),
		zap.Bool("enableDataValidation", params.EnableDataValidation))

	// 数据验证
	var validationMessage string
	if params.EnableDataValidation {
		validationResult, err := p.dataProcessor.ValidateDataAvailability(ctx, params.ModelName, params.GroupByDimensions, params.CurrentPeriodFilters, params.BasePeriodFilters, params.UserId)
		if err != nil {
			if logCtx != nil {
				p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, params, "数据可用性验证失败: "+err.Error(), toolCallStartTime)
			}
			return sugarRes.NewAiErrorResponse("数据可用性验证失败: " + err.Error()), nil
		}

		if !validationResult.IsDataAvailable {
			validationMessage = fmt.Sprintf("⚠️%s\n\n", validationResult.ValidationMessage)
		} else {
			// validationMessage = "✅数据验证通过\n\n"
			validationMessage = ""
		}
	}

	// 执行贡献度分析和匿名化处理
	aiDataText, session, usedAdvanced, err := p.contributionAnalyzer.PerformAnalysis(ctx, params.ModelName, params.TargetMetric, params.CurrentPeriodFilters, params.BasePeriodFilters, params.GroupByDimensions, params.UserId)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, params, "分析处理失败: "+err.Error(), toolCallStartTime)
		}
		return sugarRes.NewAiErrorResponse("分析处理失败: " + err.Error()), nil
	}

	// 记录匿名化信息到日志
	if logCtx != nil {
		p.executionLogger.RecordAnonymization(ctx, logCtx, aiDataText, toolCall.Function.Arguments, usedAdvanced)
	}

	// AI分析
	analysisResult, err := p.aiInteractionManager.PerformDataAnalysis(ctx, aiDataText, req.Description, agent, llmConfig)
	if err != nil {
		return sugarRes.NewAiErrorResponse("AI数据分析失败: " + err.Error()), nil
	}

	// 记录匿名化输出
	if logCtx != nil {
		p.executionLogger.RecordAnonymizationOutput(ctx, logCtx, analysisResult)
	}

	// 解密AI分析结果
	decodedResult, err := p.anonymizationProcessor.DecodeAIResponse(session, analysisResult)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, params, "AI结果解密失败: "+err.Error(), toolCallStartTime)
		}
		return sugarRes.NewAiErrorResponse("AI结果解密失败: " + err.Error()), nil
	}

	// 组合最终结果
	finalResult := validationMessage + decodedResult

	// 记录工具调用成功
	if logCtx != nil {
		p.executionLogger.RecordToolCallSuccess(ctx, logCtx, toolCall.Function.Name, params, decodedResult, len(session.AIReadyData), usedAdvanced, toolCallStartTime)
	}

	return sugarRes.NewAiSuccessResponseWithText(finalResult), nil
}

// handleDataScopeExplorer 处理数据范围探索工具调用
func (p *AiFetchProcessor) handleDataScopeExplorer(ctx context.Context, toolCall system.OpenAIToolCall, logCtx *ExecutionLogContext) (*sugarRes.SugarFormulaAiResponse, error) {
	toolCallStartTime := time.Now()

	// 解析参数
	params, err := p.aiInteractionManager.ParseDataScopeParams(toolCall.Function.Arguments)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, nil, "解析工具调用参数失败: "+err.Error(), toolCallStartTime)
		}
		return sugarRes.NewAiErrorResponse("解析工具调用参数失败: " + err.Error()), nil
	}

	// 执行数据范围探索
	scopeInfo, err := p.dataProcessor.ExploreDataScope(ctx, params.ModelName, params.ExploreDimensions, params.SampleFilters, params.UserId)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, params, "数据范围探索失败: "+err.Error(), toolCallStartTime)
		}
		return sugarRes.NewAiErrorResponse("数据范围探索失败: " + err.Error()), nil
	}

	// 格式化结果
	resultText := p.dataProcessor.FormatDataScopeResult(scopeInfo)

	// 记录工具调用成功
	if logCtx != nil {
		toolResult := map[string]interface{}{
			"scope_info":          scopeInfo,
			"explored_dimensions": len(params.ExploreDimensions),
			"total_records":       scopeInfo.TotalRecords,
			"dimension_coverage":  scopeInfo.DimensionCoverage,
		}
		p.executionLogger.RecordToolCallSuccessWithResult(ctx, logCtx, toolCall.Function.Name, params, toolResult, toolCallStartTime)
	}

	return sugarRes.NewAiSuccessResponseWithText(resultText), nil
}

// handleAnonymizedDataAnalyzer 处理匿名化数据分析工具调用（向后兼容）
func (p *AiFetchProcessor) handleAnonymizedDataAnalyzer(ctx context.Context, toolCall system.OpenAIToolCall, logCtx *ExecutionLogContext, req *sugarReq.SugarFormulaAiFetchRequest, agent *sugar.SugarAgents, llmConfig *system.LLMConfig) (*sugarRes.SugarFormulaAiResponse, error) {
	toolCallStartTime := time.Now()

	// 解析参数
	params, err := p.aiInteractionManager.ParseAnonymizedAnalyzerParams(toolCall.Function.Arguments)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, nil, "解析工具调用参数失败: "+err.Error(), toolCallStartTime)
		}
		return sugarRes.NewAiErrorResponse("解析工具调用参数失败: " + err.Error()), nil
	}

	// 执行匿名化数据处理
	anonymizedResult, err := p.anonymizationProcessor.ProcessAnonymizedDataAnalysis(ctx, params.ModelName, params.TargetMetric, params.CurrentPeriodFilters, params.BasePeriodFilters, params.GroupByDimensions, params.UserId)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, params, "匿名化数据处理失败: "+err.Error(), toolCallStartTime)
		}
		return sugarRes.NewAiErrorResponse("匿名化数据处理失败: " + err.Error()), nil
	}

	// 序列化匿名化数据
	aiDataText, err := p.anonymizationProcessor.SerializeAnonymizedDataToText(anonymizedResult.AIReadyData)
	if err != nil {
		return sugarRes.NewAiErrorResponse("匿名化数据序列化失败: " + err.Error()), nil
	}

	// 记录匿名化信息
	if logCtx != nil {
		p.executionLogger.RecordAnonymization(ctx, logCtx, aiDataText, toolCall.Function.Arguments, false)
	}

	// 进行AI分析
	analysisResult, err := p.aiInteractionManager.PerformDataAnalysis(ctx, aiDataText, req.Description, agent, llmConfig)
	if err != nil {
		return sugarRes.NewAiErrorResponse("AI数据分析失败: " + err.Error()), nil
	}

	// 记录匿名化输出
	if logCtx != nil {
		p.executionLogger.RecordAnonymizationOutput(ctx, logCtx, analysisResult)
	}

	// 解密AI分析结果
	decodedResult, err := p.anonymizationProcessor.DecodeAIResponseLegacy(anonymizedResult, analysisResult)
	if err != nil {
		if logCtx != nil {
			p.executionLogger.RecordToolCallError(ctx, logCtx, toolCall.Function.Name, params, "AI结果解密失败: "+err.Error(), toolCallStartTime)
		}
		return sugarRes.NewAiErrorResponse("AI结果解密失败: " + err.Error()), nil
	}

	// 记录工具调用成功
	if logCtx != nil {
		p.executionLogger.RecordToolCallSuccess(ctx, logCtx, toolCall.Function.Name, params, decodedResult, len(anonymizedResult.AIReadyData), false, toolCallStartTime)
	}

	return sugarRes.NewAiSuccessResponseWithText(decodedResult), nil
}
