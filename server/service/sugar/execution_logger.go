package sugar

import (
	"context"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// ExecutionLogger 执行日志管理器 - 负责记录执行过程中的各种日志
type ExecutionLogger struct {
	executionLogService SugarExecutionLogService
}

// NewExecutionLogger 创建执行日志管理器
func NewExecutionLogger() *ExecutionLogger {
	return &ExecutionLogger{
		executionLogService: SugarExecutionLogService{},
	}
}

// CreateLog 创建执行日志
func (el *ExecutionLogger) CreateLog(ctx context.Context, req *sugarReq.SugarFormulaAiFetchRequest, userId string, agentId *string) (*ExecutionLogContext, error) {
	logCtx, err := el.executionLogService.CreateExecutionLog(ctx, req, userId, agentId)
	if err != nil {
		global.GVA_LOG.Error("创建执行日志失败", zap.Error(err))
		return nil, err
	}
	return logCtx, nil
}

// RecordAIPrompts 记录AI提示词信息
func (el *ExecutionLogger) RecordAIPrompts(ctx context.Context, logCtx *ExecutionLogContext, systemPrompt, userMessage string, llmConfig *system.LLMConfig) {
	if logCtx == nil {
		return
	}

	el.executionLogService.RecordAISystemPrompt(ctx, logCtx, systemPrompt)
	el.executionLogService.RecordAIUserMessage(ctx, logCtx, userMessage)
	_ = el.executionLogService.RecordLLMConfig(ctx, logCtx, llmConfig)
}

// RecordLLMResponse 记录LLM响应
func (el *ExecutionLogger) RecordLLMResponse(ctx context.Context, logCtx *ExecutionLogContext, response, modelName string) {
	if logCtx == nil {
		return
	}

	el.executionLogService.RecordLLMResponse(ctx, logCtx, response, &modelName, nil)
}

// RecordAnonymization 记录匿名化信息
func (el *ExecutionLogger) RecordAnonymization(ctx context.Context, logCtx *ExecutionLogContext, aiDataText, toolCallArgs string, usedAdvanced bool) {
	if logCtx == nil {
		return
	}

	anonymizedInputData := map[string]interface{}{
		"aiDataText":             aiDataText,
		"toolCall":               toolCallArgs,
		"isEncrypted":            true,
		"used_advanced_analyzer": usedAdvanced,
	}
	_ = el.executionLogService.UpdateExecutionLogWithAnonymization(ctx, logCtx, anonymizedInputData, nil)
}

// RecordAnonymizationOutput 记录匿名化输出
func (el *ExecutionLogger) RecordAnonymizationOutput(ctx context.Context, logCtx *ExecutionLogContext, analysisResult string) {
	if logCtx == nil {
		return
	}

	_ = el.executionLogService.UpdateExecutionLogWithAnonymization(ctx, logCtx, nil, &analysisResult)
}

// RecordToolCallError 记录工具调用错误
func (el *ExecutionLogger) RecordToolCallError(ctx context.Context, logCtx *ExecutionLogContext, toolName string, args interface{}, errorMsg string, startTime time.Time) {
	if logCtx == nil {
		return
	}

	durationMs := int(time.Since(startTime).Milliseconds())

	// 类型断言处理
	var argsMap map[string]interface{}
	if args != nil {
		if m, ok := args.(map[string]interface{}); ok {
			argsMap = m
		} else {
			argsMap = map[string]interface{}{"raw_args": args}
		}
	}

	el.executionLogService.RecordToolCall(ctx, logCtx, toolName, argsMap, nil, &errorMsg, durationMs, false)
}

// RecordToolCallSuccess 记录工具调用成功（智能匿名化分析工具）
func (el *ExecutionLogger) RecordToolCallSuccess(ctx context.Context, logCtx *ExecutionLogContext, toolName string, args interface{}, decodedResult string, dataCount int, usedAdvanced bool, startTime time.Time) {
	if logCtx == nil {
		return
	}

	durationMs := int(time.Since(startTime).Milliseconds())
	toolResult := map[string]interface{}{
		"decoded_result":         decodedResult,
		"anonymized_data_count":  dataCount,
		"used_advanced_analyzer": usedAdvanced,
	}

	// 类型断言处理
	var argsMap map[string]interface{}
	if args != nil {
		if m, ok := args.(map[string]interface{}); ok {
			argsMap = m
		} else {
			argsMap = map[string]interface{}{"raw_args": args}
		}
	}

	el.executionLogService.RecordToolCall(ctx, logCtx, toolName, argsMap, toolResult, nil, durationMs, true)
}

// RecordToolCallSuccessWithResult 记录工具调用成功（通用版本）
func (el *ExecutionLogger) RecordToolCallSuccessWithResult(ctx context.Context, logCtx *ExecutionLogContext, toolName string, args interface{}, result map[string]interface{}, startTime time.Time) {
	if logCtx == nil {
		return
	}

	durationMs := int(time.Since(startTime).Milliseconds())

	// 类型断言处理
	var argsMap map[string]interface{}
	if args != nil {
		if m, ok := args.(map[string]interface{}); ok {
			argsMap = m
		} else {
			argsMap = map[string]interface{}{"raw_args": args}
		}
	}

	el.executionLogService.RecordToolCall(ctx, logCtx, toolName, argsMap, result, nil, durationMs, true)
}

// FinishWithSuccess 完成执行并记录成功
func (el *ExecutionLogger) FinishWithSuccess(ctx context.Context, logCtx *ExecutionLogContext, result string) {
	if logCtx == nil {
		return
	}

	// 更新AI交互信息到数据库
	_ = el.executionLogService.UpdateExecutionLogWithAIInteractionsToDatabase(ctx, logCtx)

	// 记录成功日志
	_ = el.executionLogService.FinishExecutionLog(ctx, logCtx, result, "success", nil)
}

// FinishWithError 完成执行并记录错误
func (el *ExecutionLogger) FinishWithError(ctx context.Context, logCtx *ExecutionLogContext, errorMsg string) {
	if logCtx == nil {
		return
	}

	_ = el.executionLogService.FinishExecutionLog(ctx, logCtx, "", "failed", &errorMsg)
}
