package sugar

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	sugarReq "github.com/flipped-aurora/gin-vue-admin/server/model/sugar/request"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// SugarExecutionLogService 执行日志服务
type SugarExecutionLogService struct{}

// ExecutionLogContext 执行日志上下文，用于跟踪单次AIFETCH执行的日志信息
type ExecutionLogContext struct {
	LogID                int64
	StartTime            time.Time
	InputPayload         datatypes.JSON
	AnonymizationEnabled bool
	AnonymizedInput      datatypes.JSON
	AnonymizedOutput     *string
	// AI交互相关字段
	SystemPrompt   *string
	UserMessage    *string
	LlmConfig      datatypes.JSON
	AiInteractions []AIInteractionRecord
	ToolCalls      []ToolCallRecord
	RawLlmResponse *string
	TokenUsage     *TokenUsageRecord
}

// AIInteractionRecord AI交互记录，包含单轮对话
type AIInteractionRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	Type         string    `json:"type"`             // "request" 或 "response"
	Role         string    `json:"role"`             // "system", "user", "assistant"
	Content      string    `json:"content"`          // 消息内容
	Tokens       *int      `json:"tokens,omitempty"` // Token数量
	Model        *string   `json:"model,omitempty"`  // 使用的模型
	IsAnonymized bool      `json:"is_anonymized"`    // 是否为匿名化内容
}

// ToolCallRecord 工具调用记录
type ToolCallRecord struct {
	Timestamp    time.Time              `json:"timestamp"`
	ToolName     string                 `json:"tool_name"`
	Arguments    map[string]interface{} `json:"arguments"`
	Result       interface{}            `json:"result,omitempty"`
	Error        *string                `json:"error,omitempty"`
	DurationMs   int                    `json:"duration_ms"`
	IsAnonymized bool                   `json:"is_anonymized"`
}

// TokenUsageRecord Token使用记录
type TokenUsageRecord struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// CreateExecutionLog 创建执行日志记录
func (s *SugarExecutionLogService) CreateExecutionLog(ctx context.Context, req *sugarReq.SugarFormulaAiFetchRequest, userId string, agentId *string) (*ExecutionLogContext, error) {
	// 构建输入负载
	inputPayload := map[string]interface{}{
		"agentName":   req.AgentName,
		"description": req.Description,
		"dataRange":   req.DataRange,
		"userId":      userId,
	}

	inputJSON, err := json.Marshal(inputPayload)
	if err != nil {
		global.GVA_LOG.Error("序列化输入负载失败", zap.Error(err))
		return nil, fmt.Errorf("序列化输入负载失败: %w", err)
	}

	// 创建日志记录
	logEntry := sugar.SugarExecutionLogs{
		LogType:              "ai_agent",
		UserId:               &userId,
		AgentId:              agentId,
		InputPayload:         datatypes.JSON(inputJSON),
		Status:               "pending",
		AnonymizationEnabled: false, // 初始状态为false，后续根据实际情况更新
		ExecutedAt:           time.Now(),
	}

	// 保存到数据库
	if err := global.GVA_DB.Create(&logEntry).Error; err != nil {
		global.GVA_LOG.Error("创建执行日志失败", zap.Error(err))
		return nil, fmt.Errorf("创建执行日志失败: %w", err)
	}

	global.GVA_LOG.Info("创建执行日志成功",
		zap.Int64("logId", logEntry.Id),
		zap.String("userId", userId),
		zap.String("agentName", req.AgentName))

	// 返回日志上下文
	return &ExecutionLogContext{
		LogID:                logEntry.Id,
		StartTime:            logEntry.ExecutedAt,
		InputPayload:         datatypes.JSON(inputJSON),
		AnonymizationEnabled: false,
		AiInteractions:       make([]AIInteractionRecord, 0),
		ToolCalls:            make([]ToolCallRecord, 0),
	}, nil
}

// UpdateExecutionLogWithAnonymization 更新执行日志的匿名化信息
func (s *SugarExecutionLogService) UpdateExecutionLogWithAnonymization(ctx context.Context, logCtx *ExecutionLogContext, anonymizedInput interface{}, anonymizedOutput *string) error {
	// 序列化匿名化输入
	var anonymizedInputJSON datatypes.JSON
	if anonymizedInput != nil {
		inputBytes, err := json.Marshal(anonymizedInput)
		if err != nil {
			global.GVA_LOG.Error("序列化匿名化输入失败", zap.Error(err))
			return fmt.Errorf("序列化匿名化输入失败: %w", err)
		}
		anonymizedInputJSON = datatypes.JSON(inputBytes)
	}

	// 首先从数据库读取当前记录，确保不会覆盖已存在的匿名化数据
	var currentLog sugar.SugarExecutionLogs
	if err := global.GVA_DB.Where("id = ?", logCtx.LogID).First(&currentLog).Error; err != nil {
		global.GVA_LOG.Error("读取当前执行日志失败", zap.Error(err), zap.Int64("logId", logCtx.LogID))
		return fmt.Errorf("读取当前执行日志失败: %w", err)
	}

	// 构建更新数据 - 只更新匿名化启用状态
	updates := map[string]interface{}{
		"anonymization_enabled": true,
	}

	// 保护匿名化输入：只有当传入新数据时才更新，否则保持数据库中的值
	if anonymizedInput != nil && len(anonymizedInputJSON) > 0 {
		updates["anonymized_input"] = anonymizedInputJSON
		// 更新上下文
		logCtx.AnonymizedInput = anonymizedInputJSON
	} else if len(currentLog.AnonymizedInput) > 0 {
		// 保持数据库中现有的匿名化输入数据
		updates["anonymized_input"] = currentLog.AnonymizedInput
		// 同步更新上下文，确保上下文与数据库一致
		logCtx.AnonymizedInput = currentLog.AnonymizedInput
	}

	// 保护匿名化输出：只有当传入新数据时才更新，否则保持数据库中的值
	if anonymizedOutput != nil && *anonymizedOutput != "" {
		updates["anonymized_output"] = *anonymizedOutput
		// 更新上下文
		logCtx.AnonymizedOutput = anonymizedOutput
	} else if currentLog.AnonymizedOutput != nil && *currentLog.AnonymizedOutput != "" {
		// 保持数据库中现有的匿名化输出数据
		updates["anonymized_output"] = *currentLog.AnonymizedOutput
		// 同步更新上下文，确保上下文与数据库一致
		logCtx.AnonymizedOutput = currentLog.AnonymizedOutput
	}

	if err := global.GVA_DB.Model(&sugar.SugarExecutionLogs{}).Where("id = ?", logCtx.LogID).Updates(updates).Error; err != nil {
		global.GVA_LOG.Error("更新执行日志匿名化信息失败", zap.Error(err), zap.Int64("logId", logCtx.LogID))
		return fmt.Errorf("更新执行日志匿名化信息失败: %w", err)
	}

	global.GVA_LOG.Info("更新执行日志匿名化信息成功", zap.Int64("logId", logCtx.LogID))

	// 确保匿名化启用状态已更新
	logCtx.AnonymizationEnabled = true

	return nil
}

// FinishExecutionLog 完成执行日志记录
func (s *SugarExecutionLogService) FinishExecutionLog(ctx context.Context, logCtx *ExecutionLogContext, finalResult string, status string, errorMessage *string) error {
	// 计算执行耗时
	durationMs := int(time.Since(logCtx.StartTime).Milliseconds())

	// 构建更新数据
	updates := map[string]interface{}{
		"status":       status,
		"final_result": finalResult,
		"duration_ms":  durationMs,
	}

	if errorMessage != nil {
		updates["error_message"] = *errorMessage
	}

	// 首先从数据库读取当前记录，确保不会覆盖已存在的匿名化数据
	var currentLog sugar.SugarExecutionLogs
	if err := global.GVA_DB.Where("id = ?", logCtx.LogID).First(&currentLog).Error; err != nil {
		global.GVA_LOG.Error("读取当前执行日志失败", zap.Error(err), zap.Int64("logId", logCtx.LogID))
		return fmt.Errorf("读取当前执行日志失败: %w", err)
	}

	// 保护匿名化相关字段：只有当上下文中有值或数据库中已有值时才更新
	if logCtx.AnonymizationEnabled || currentLog.AnonymizationEnabled {
		updates["anonymization_enabled"] = logCtx.AnonymizationEnabled || currentLog.AnonymizationEnabled
	}

	// 保护匿名化输入：优先使用上下文中的值，如果上下文中没有则保持数据库中的值
	if logCtx.AnonymizedInput != nil && len(logCtx.AnonymizedInput) > 0 {
		updates["anonymized_input"] = logCtx.AnonymizedInput
	} else if len(currentLog.AnonymizedInput) > 0 {
		updates["anonymized_input"] = currentLog.AnonymizedInput
	}

	// 保护匿名化输出：优先使用上下文中的值，如果上下文中没有则保持数据库中的值
	if logCtx.AnonymizedOutput != nil && *logCtx.AnonymizedOutput != "" {
		updates["anonymized_output"] = *logCtx.AnonymizedOutput
	} else if currentLog.AnonymizedOutput != nil && *currentLog.AnonymizedOutput != "" {
		updates["anonymized_output"] = *currentLog.AnonymizedOutput
	}

	// 更新数据库记录
	if err := global.GVA_DB.Model(&sugar.SugarExecutionLogs{}).Where("id = ?", logCtx.LogID).Updates(updates).Error; err != nil {
		global.GVA_LOG.Error("完成执行日志失败", zap.Error(err), zap.Int64("logId", logCtx.LogID))
		return fmt.Errorf("完成执行日志失败: %w", err)
	}

	global.GVA_LOG.Info("完成执行日志成功",
		zap.Int64("logId", logCtx.LogID),
		zap.String("status", status),
		zap.Int("durationMs", durationMs))

	return nil
}

// RecordAISystemPrompt 记录AI系统提示词
func (s *SugarExecutionLogService) RecordAISystemPrompt(ctx context.Context, logCtx *ExecutionLogContext, systemPrompt string) {
	if logCtx == nil {
		return
	}

	logCtx.SystemPrompt = &systemPrompt

	// 记录为AI交互
	interaction := AIInteractionRecord{
		Timestamp:    time.Now(),
		Type:         "request",
		Role:         "system",
		Content:      systemPrompt,
		IsAnonymized: false,
	}
	logCtx.AiInteractions = append(logCtx.AiInteractions, interaction)

	global.GVA_LOG.Debug("记录系统提示词",
		zap.Int64("logId", logCtx.LogID),
		zap.Int("promptLength", len(systemPrompt)))
}

// RecordAIUserMessage 记录用户消息
func (s *SugarExecutionLogService) RecordAIUserMessage(ctx context.Context, logCtx *ExecutionLogContext, userMessage string) {
	if logCtx == nil {
		return
	}

	logCtx.UserMessage = &userMessage

	// 记录为AI交互
	interaction := AIInteractionRecord{
		Timestamp:    time.Now(),
		Type:         "request",
		Role:         "user",
		Content:      userMessage,
		IsAnonymized: false,
	}
	logCtx.AiInteractions = append(logCtx.AiInteractions, interaction)

	global.GVA_LOG.Debug("记录用户消息",
		zap.Int64("logId", logCtx.LogID),
		zap.Int("messageLength", len(userMessage)))
}

// RecordLLMConfig 记录LLM配置信息
func (s *SugarExecutionLogService) RecordLLMConfig(ctx context.Context, logCtx *ExecutionLogContext, llmConfig interface{}) error {
	if logCtx == nil {
		return nil
	}

	configJSON, err := json.Marshal(llmConfig)
	if err != nil {
		global.GVA_LOG.Error("序列化LLM配置失败", zap.Error(err))
		return fmt.Errorf("序列化LLM配置失败: %w", err)
	}

	logCtx.LlmConfig = datatypes.JSON(configJSON)

	global.GVA_LOG.Debug("记录LLM配置",
		zap.Int64("logId", logCtx.LogID),
		zap.String("config", string(configJSON)))

	return nil
}

// RecordLLMResponse 记录LLM响应
func (s *SugarExecutionLogService) RecordLLMResponse(ctx context.Context, logCtx *ExecutionLogContext, response string, model *string, tokens *int) {
	if logCtx == nil {
		return
	}

	logCtx.RawLlmResponse = &response

	// 记录为AI交互
	interaction := AIInteractionRecord{
		Timestamp:    time.Now(),
		Type:         "response",
		Role:         "assistant",
		Content:      response,
		Model:        model,
		Tokens:       tokens,
		IsAnonymized: false,
	}
	logCtx.AiInteractions = append(logCtx.AiInteractions, interaction)

	global.GVA_LOG.Debug("记录LLM响应",
		zap.Int64("logId", logCtx.LogID),
		zap.Int("responseLength", len(response)))
}

// RecordToolCall 记录工具调用
func (s *SugarExecutionLogService) RecordToolCall(ctx context.Context, logCtx *ExecutionLogContext, toolName string, arguments map[string]interface{}, result interface{}, error *string, durationMs int, isAnonymized bool) {
	if logCtx == nil {
		return
	}

	toolCall := ToolCallRecord{
		Timestamp:    time.Now(),
		ToolName:     toolName,
		Arguments:    arguments,
		Result:       result,
		Error:        error,
		DurationMs:   durationMs,
		IsAnonymized: isAnonymized,
	}
	logCtx.ToolCalls = append(logCtx.ToolCalls, toolCall)

	global.GVA_LOG.Debug("记录工具调用",
		zap.Int64("logId", logCtx.LogID),
		zap.String("toolName", toolName),
		zap.Int("durationMs", durationMs),
		zap.Bool("isAnonymized", isAnonymized))
}

// RecordTokenUsage 记录Token使用情况
func (s *SugarExecutionLogService) RecordTokenUsage(ctx context.Context, logCtx *ExecutionLogContext, promptTokens, completionTokens, totalTokens int) {
	if logCtx == nil {
		return
	}

	tokenUsage := &TokenUsageRecord{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
	}
	logCtx.TokenUsage = tokenUsage

	global.GVA_LOG.Debug("记录Token使用",
		zap.Int64("logId", logCtx.LogID),
		zap.Int("promptTokens", promptTokens),
		zap.Int("completionTokens", completionTokens),
		zap.Int("totalTokens", totalTokens))
}

// UpdateExecutionLogWithAIInteractions 更新执行日志的AI交互信息
func (s *SugarExecutionLogService) UpdateExecutionLogWithAIInteractions(ctx context.Context, logCtx *ExecutionLogContext) error {
	if logCtx == nil {
		return fmt.Errorf("日志上下文为空")
	}

	// 记录AI交互信息到日志，但不更新数据库（因为字段可能不存在）
	global.GVA_LOG.Info("记录AI交互信息",
		zap.Int64("logId", logCtx.LogID),
		zap.Int("interactionCount", len(logCtx.AiInteractions)),
		zap.Int("toolCallCount", len(logCtx.ToolCalls)),
		zap.Bool("hasSystemPrompt", logCtx.SystemPrompt != nil),
		zap.Bool("hasUserMessage", logCtx.UserMessage != nil),
		zap.Bool("hasLlmResponse", logCtx.RawLlmResponse != nil),
		zap.Bool("hasTokenUsage", logCtx.TokenUsage != nil))

	// 可选：记录详细的AI交互到日志文件
	if len(logCtx.AiInteractions) > 0 {
		for i, interaction := range logCtx.AiInteractions {
			global.GVA_LOG.Debug("AI交互详情",
				zap.Int64("logId", logCtx.LogID),
				zap.Int("interactionIndex", i),
				zap.String("type", interaction.Type),
				zap.String("role", interaction.Role),
				zap.Int("contentLength", len(interaction.Content)),
				zap.Bool("isAnonymized", interaction.IsAnonymized),
				zap.Time("timestamp", interaction.Timestamp))
		}
	}

	if len(logCtx.ToolCalls) > 0 {
		for i, toolCall := range logCtx.ToolCalls {
			global.GVA_LOG.Debug("工具调用详情",
				zap.Int64("logId", logCtx.LogID),
				zap.Int("toolCallIndex", i),
				zap.String("toolName", toolCall.ToolName),
				zap.Int("durationMs", toolCall.DurationMs),
				zap.Bool("isAnonymized", toolCall.IsAnonymized),
				zap.Bool("hasError", toolCall.Error != nil),
				zap.Time("timestamp", toolCall.Timestamp))
		}
	}

	return nil
}

// UpdateExecutionLogWithAIInteractionsToDatabase 更新执行日志的AI交互信息到数据库（需要数据库字段支持）
func (s *SugarExecutionLogService) UpdateExecutionLogWithAIInteractionsToDatabase(ctx context.Context, logCtx *ExecutionLogContext) error {
	if logCtx == nil {
		return fmt.Errorf("日志上下文为空")
	}

	// 序列化AI交互记录
	var aiInteractionsJSON datatypes.JSON
	if len(logCtx.AiInteractions) > 0 {
		interactionsBytes, err := json.Marshal(logCtx.AiInteractions)
		if err != nil {
			global.GVA_LOG.Error("序列化AI交互记录失败", zap.Error(err))
			return fmt.Errorf("序列化AI交互记录失败: %w", err)
		}
		aiInteractionsJSON = datatypes.JSON(interactionsBytes)
	}

	// 序列化工具调用记录
	var toolCallsJSON datatypes.JSON
	if len(logCtx.ToolCalls) > 0 {
		toolCallsBytes, err := json.Marshal(logCtx.ToolCalls)
		if err != nil {
			global.GVA_LOG.Error("序列化工具调用记录失败", zap.Error(err))
			return fmt.Errorf("序列化工具调用记录失败: %w", err)
		}
		toolCallsJSON = datatypes.JSON(toolCallsBytes)
	}

	// 序列化Token使用记录
	var tokenUsageJSON datatypes.JSON
	if logCtx.TokenUsage != nil {
		tokenUsageBytes, err := json.Marshal(logCtx.TokenUsage)
		if err != nil {
			global.GVA_LOG.Error("序列化Token使用记录失败", zap.Error(err))
			return fmt.Errorf("序列化Token使用记录失败: %w", err)
		}
		tokenUsageJSON = datatypes.JSON(tokenUsageBytes)
	}

	// 构建更新数据
	updates := map[string]interface{}{}

	if logCtx.SystemPrompt != nil {
		updates["system_prompt"] = *logCtx.SystemPrompt
	}
	if logCtx.UserMessage != nil {
		updates["user_message"] = *logCtx.UserMessage
	}
	if len(logCtx.LlmConfig) > 0 {
		updates["llm_config"] = logCtx.LlmConfig
	}
	if len(aiInteractionsJSON) > 0 {
		updates["ai_interactions"] = aiInteractionsJSON
	}
	if len(toolCallsJSON) > 0 {
		updates["tool_calls"] = toolCallsJSON
	}
	if logCtx.RawLlmResponse != nil {
		updates["raw_llm_response"] = *logCtx.RawLlmResponse
	}
	if len(tokenUsageJSON) > 0 {
		updates["token_usage"] = tokenUsageJSON
	}

	// 更新数据库记录
	if len(updates) > 0 {
		// 首先从数据库读取当前记录，确保不会覆盖已存在的匿名化数据
		var currentLog sugar.SugarExecutionLogs
		if err := global.GVA_DB.Where("id = ?", logCtx.LogID).First(&currentLog).Error; err != nil {
			global.GVA_LOG.Error("读取当前执行日志失败", zap.Error(err), zap.Int64("logId", logCtx.LogID))
			return fmt.Errorf("读取当前执行日志失败: %w", err)
		}

		// 保护匿名化相关字段：只有当上下文中有值或数据库中已有值时才更新
		if logCtx.AnonymizationEnabled || currentLog.AnonymizationEnabled {
			updates["anonymization_enabled"] = logCtx.AnonymizationEnabled || currentLog.AnonymizationEnabled
		}

		// 保护匿名化输入：优先使用上下文中的值，如果上下文中没有则保持数据库中的值
		if logCtx.AnonymizedInput != nil && len(logCtx.AnonymizedInput) > 0 {
			updates["anonymized_input"] = logCtx.AnonymizedInput
		} else if len(currentLog.AnonymizedInput) > 0 {
			updates["anonymized_input"] = currentLog.AnonymizedInput
		}

		// 保护匿名化输出：优先使用上下文中的值，如果上下文中没有则保持数据库中的值
		if logCtx.AnonymizedOutput != nil && *logCtx.AnonymizedOutput != "" {
			updates["anonymized_output"] = *logCtx.AnonymizedOutput
		} else if currentLog.AnonymizedOutput != nil && *currentLog.AnonymizedOutput != "" {
			updates["anonymized_output"] = *currentLog.AnonymizedOutput
		}

		if err := global.GVA_DB.Model(&sugar.SugarExecutionLogs{}).Where("id = ?", logCtx.LogID).Updates(updates).Error; err != nil {
			global.GVA_LOG.Error("更新执行日志AI交互信息失败", zap.Error(err), zap.Int64("logId", logCtx.LogID))
			return fmt.Errorf("更新执行日志AI交互信息失败: %w", err)
		}
	}

	global.GVA_LOG.Info("更新执行日志AI交互信息成功",
		zap.Int64("logId", logCtx.LogID),
		zap.Int("interactionCount", len(logCtx.AiInteractions)),
		zap.Int("toolCallCount", len(logCtx.ToolCalls)))

	return nil
}

// CreateExecutionLogForDbQuery 创建数据库查询执行日志记录
func (s *SugarExecutionLogService) CreateExecutionLogForDbQuery(ctx context.Context, userId string, connectionId *string, inputPayload interface{}) (*ExecutionLogContext, error) {
	inputJSON, err := json.Marshal(inputPayload)
	if err != nil {
		global.GVA_LOG.Error("序列化数据库查询输入负载失败", zap.Error(err))
		return nil, fmt.Errorf("序列化输入负载失败: %w", err)
	}

	// 创建日志记录
	logEntry := sugar.SugarExecutionLogs{
		LogType:              "db_query",
		UserId:               &userId,
		ConnectionId:         connectionId,
		InputPayload:         datatypes.JSON(inputJSON),
		Status:               "pending",
		AnonymizationEnabled: false,
		ExecutedAt:           time.Now(),
	}

	// 保存到数据库
	if err := global.GVA_DB.Create(&logEntry).Error; err != nil {
		global.GVA_LOG.Error("创建数据库查询执行日志失败", zap.Error(err))
		return nil, fmt.Errorf("创建执行日志失败: %w", err)
	}

	global.GVA_LOG.Info("创建数据库查询执行日志成功",
		zap.Int64("logId", logEntry.Id),
		zap.String("userId", userId))

	// 返回日志上下文
	return &ExecutionLogContext{
		LogID:                logEntry.Id,
		StartTime:            logEntry.ExecutedAt,
		InputPayload:         datatypes.JSON(inputJSON),
		AnonymizationEnabled: false,
	}, nil
}
