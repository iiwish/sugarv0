package sugar

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/sugar"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"go.uber.org/zap"
)

// AIInteractionManager AI交互管理器 - 负责AI相关的交互逻辑
type AIInteractionManager struct {
	llmService system.SysLLMService
}

// NewAIInteractionManager 创建AI交互管理器
func NewAIInteractionManager() *AIInteractionManager {
	return &AIInteractionManager{
		llmService: system.SysLLMService{},
	}
}

// BuildSystemPrompt 构建智能系统提示词
func (aim *AIInteractionManager) BuildSystemPrompt(agent *sugar.SugarAgents, userId string) string {
	// 构建工具调用的系统提示词（不包含具体的分析提示词）
	enhancedPrompt := fmt.Sprintf(`你是一个专业的数据分析助手，专门负责调用数据分析工具。

📋 重要工作流程指导：
1. **使用智能匿名化分析工具**：对于贡献度分析需求，请使用 smart_anonymized_analyzer 工具，它会自动完成数据验证和匿名化分析的完整流程
2. **精确匹配原则**：生成的筛选条件必须与用户问题中的具体实体对应，避免过于宽泛或不存在的条件
3. **语义顺序原则**：调用工具时，groupByDimensions参数中的维度必须按照语义逻辑顺序排列（从大到小、从主要到次要），这样有利于后续匿名化还原时保持语句通顺性
4. **数据验证策略**：工具会自动验证数据可用性，如果数据不足会给出明确提示
5. **结果可信度评估**：基于实际数据的完整性和代表性评估结论的可信度

🔧 工具使用指南：
- **推荐工具**：smart_anonymized_analyzer - 完整的智能匿名化分析流程
- **备用工具**：data_scope_explorer（仅数据探索）、anonymized_data_analyzer（传统匿名化）
- 当前用户ID为 %s，调用工具时必须传递此用户ID
- 启用数据验证（enableDataValidation: true）以确保数据质量
- **维度排序示例**：货币资金分析时应按 ['银行名称', '账户类型', '币种'] 的顺序

💡 智能分析策略：
- 优先分析数据中贡献度最高的维度组合
- 对异常值和趋势变化提供深入洞察
- 结合业务常识给出可操作的建议
- 明确说明分析的局限性和数据范围`, userId)

	return enhancedPrompt
}

// BuildAnalysisSystemPrompt 构建数据分析的系统提示词（包含Agent配置的提示词）
func (aim *AIInteractionManager) BuildAnalysisSystemPrompt(agent *sugar.SugarAgents, userId string) string {
	// 基础提示词优先使用Agent中定义的Prompt字段
	basePrompt := ""
	if agent != nil && agent.Prompt != nil && *agent.Prompt != "" {
		basePrompt = *agent.Prompt
		global.GVA_LOG.Info("使用Agent配置的系统提示词",
			zap.String("agentName", aim.safeString(agent.Name)),
			zap.String("promptLength", fmt.Sprintf("%d", len(*agent.Prompt))))
	} else {
		// 如果Agent没有定义Prompt，则使用一个通用的、鼓励性的默认值
		basePrompt = "你是一个专业的财务数据分析师，请根据下文的匿名化数据和分析要求，给出深入、有洞察力的分析报告。"
		global.GVA_LOG.Info("使用默认系统提示词",
			zap.String("reason", "Agent未配置Prompt字段"))
	}

	// 提供一个通用的、关于如何处理匿名化数据的附加上下文
	anonymizationContext := `
---
**匿名化数据处理指南:**
- **数据已脱敏**: 你接收到的数据中，维度名称（如 D01）和维度值（如 D01_V01）都经过了匿名化处理。
- **关注相对关系**: 分析的重点应放在数据的模式、趋势和相对贡献度上，而不是具体的绝对值。
- **贡献度分析**: 数据中包含了"贡献度百分比"和"是否为正向驱动"，请利用这些信息来识别关键影响因素。
- **还原业务含义**: 在输出最终结论时，系统会自动将匿名代号解码回真实的业务术语，所以请在分析时大胆使用这些代号，并想象它们代表的真实业务含义。
`

	// 将基础提示词和附加上下文结合起来
	finalPrompt := fmt.Sprintf("%s\n%s", basePrompt, anonymizationContext)

	global.GVA_LOG.Debug("构建分析系统提示词完成",
		zap.Int("finalPromptLength", len(finalPrompt)))

	return finalPrompt
}

// BuildUserMessage 构建用户消息
func (aim *AIInteractionManager) BuildUserMessage(description string, semantic *string, dataRange string) string {
	message := description

	// 如果有语义模型标识，从数据库获取详细信息
	if semantic != nil && *semantic != "" {
		semanticInfo, err := aim.getSemanticModelInfo(*semantic)
		if err != nil {
			global.GVA_LOG.Warn("获取语义模型信息失败", zap.String("semantic", *semantic), zap.Error(err))
			message += fmt.Sprintf("\n\n可用的数据模型信息：\n%s", *semantic)
		} else {
			message += fmt.Sprintf("\n\n可用的数据模型信息：\n%s", semanticInfo)
		}
	}

	// 如果提供了DataRange数据，将其包含在提示词中
	if dataRange != "" {
		global.GVA_LOG.Info("包含DataRange数据到提示词中", zap.Int("dataRangeLength", len(dataRange)))
		message += fmt.Sprintf("\n\n相关数据范围：\n%s", dataRange)
	}

	return message
}

// CallLLMWithTools 调用LLM并传入工具定义
func (aim *AIInteractionManager) CallLLMWithTools(ctx context.Context, llmConfig *system.LLMConfig, systemPrompt, userMessage string) (string, error) {
	// 准备工具定义（智能匿名化数据分析工具）
	tools := []system.ToolDefinition{
		{
			Name:        "smart_anonymized_analyzer",
			Description: "智能匿名化数据分析工具，自动进行数据范围探索和匿名化分析的完整流程。该工具会先验证数据可用性，然后进行匿名化贡献度分析，确保数据安全和分析准确性。调用时请确保维度按语义逻辑顺序排列。",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"modelName": map[string]interface{}{
						"type":        "string",
						"description": "要分析的语义模型名称。",
					},
					"targetMetric": map[string]interface{}{
						"type":        "string",
						"description": "需要分析的核心指标列名，例如 '销售金额'、'利润' 等。",
					},
					"currentPeriodFilters": map[string]interface{}{
						"type":        "object",
						"description": "获取本期数据的筛选条件，格式为 {\"列名\": \"筛选值\"}。",
					},
					"basePeriodFilters": map[string]interface{}{
						"type":        "object",
						"description": "获取基期（如上期、预算）数据的筛选条件，格式为 {\"列名\": \"筛选值\"}。",
					},
					"groupByDimensions": map[string]interface{}{
						"type":        "array",
						"items":       map[string]string{"type": "string"},
						"description": "进行分组和归因分析的维度列名列表，如 ['区域', '产品类别']。**重要：请按照语义逻辑顺序排列维度，例如从大到小、从主要到次要的顺序，这样有利于后续匿名化还原时保持语句的通顺性。比如货币资金分析时应按 ['银行名称', '账户类型', '币种'] 的顺序排列。**",
					},
					"userId": map[string]interface{}{
						"type":        "string",
						"description": "发起请求的用户ID，工具内部需要此参数进行鉴权。",
					},
					"enableDataValidation": map[string]interface{}{
						"type":        "boolean",
						"description": "是否启用数据范围验证，默认为true。启用后会先验证筛选条件的有效性。",
					},
				},
				"required": []string{"modelName", "targetMetric", "currentPeriodFilters", "basePeriodFilters", "groupByDimensions", "userId"},
			},
		},
	}

	// 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 调用LLM，传入工具定义
	global.GVA_LOG.Info("开始调用LLM",
		zap.String("model", llmConfig.ModelName),
		zap.Int("toolsCount", len(tools)),
		zap.Int("messagesCount", len(messages)))

	llmResponse, err := aim.llmService.ChatWithTools(ctx, *llmConfig, messages, tools)
	if err != nil {
		global.GVA_LOG.Error("LLM调用失败", zap.Error(err))
		return "", fmt.Errorf("LLM调用失败: %w", err)
	}

	global.GVA_LOG.Info("LLM调用成功", zap.String("responseLength", fmt.Sprintf("%d", len(llmResponse))))
	global.GVA_LOG.Debug("LLM原始响应", zap.String("llmResponse", llmResponse))

	return llmResponse, nil
}

// PerformDataAnalysis 对获取的数据进行AI分析（上下文感知版本）
func (aim *AIInteractionManager) PerformDataAnalysis(ctx context.Context, dataText string, userDescription string, agent *sugar.SugarAgents, llmConfig *system.LLMConfig) (string, error) {
	global.GVA_LOG.Info("开始执行上下文感知数据分析",
		zap.String("userDescription", userDescription),
		zap.String("dataLength", fmt.Sprintf("%d", len(dataText))))

	// 使用专门的分析提示词（包含Agent配置的Prompt字段）
	systemPrompt := aim.BuildAnalysisSystemPrompt(agent, "")
	global.GVA_LOG.Debug("构建分析系统提示词", zap.String("systemPrompt", systemPrompt))

	// 构建完整的用户消息，包含用户的原始提示词和匿名化数据
	userMessage := aim.buildCompleteAnalysisMessage(dataText, userDescription, agent)
	global.GVA_LOG.Debug("构建完整分析用户消息", zap.String("userMessage", userMessage))

	// 构建消息列表
	messages := []system.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userMessage},
	}

	// 调用LLM进行分析
	global.GVA_LOG.Info("开始调用LLM进行上下文感知数据分析", zap.String("model", llmConfig.ModelName))
	response, err := aim.llmService.ChatSimple(ctx, *llmConfig, messages)
	if err != nil {
		global.GVA_LOG.Error("AI数据分析调用失败", zap.Error(err))
		return "", fmt.Errorf("AI分析失败: %w", err)
	}

	// 对分析结果进行质量评估
	qualityScore := aim.evaluateAnalysisQuality(response, dataText, userDescription)
	global.GVA_LOG.Info("数据分析完成",
		zap.String("responseLength", fmt.Sprintf("%d", len(response))),
		zap.Float64("qualityScore", qualityScore))

	global.GVA_LOG.Debug("数据分析响应", zap.String("response", response))
	return response, nil
}

// ParseToolCallResponse 解析工具调用响应
func (aim *AIInteractionManager) ParseToolCallResponse(llmResponse string) (*ToolCallResponse, error) {
	var toolCallResp ToolCallResponse
	err := json.Unmarshal([]byte(llmResponse), &toolCallResp)

	// 如果解析失败或不是工具调用，则认为是普通文本响应
	if err != nil || toolCallResp.Type != "tool_call" {
		if err != nil {
			global.GVA_LOG.Debug("响应不是JSON格式，作为普通文本处理", zap.Error(err))
		} else {
			global.GVA_LOG.Debug("响应类型不是工具调用，作为普通文本处理", zap.String("type", toolCallResp.Type))
		}
		return nil, fmt.Errorf("不是工具调用响应")
	}

	return &toolCallResp, nil
}

// ParseSmartAnalyzerParams 解析智能匿名化分析工具参数
func (aim *AIInteractionManager) ParseSmartAnalyzerParams(arguments string) (*SmartAnalyzerParams, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("解析工具调用参数失败: %w", err)
	}

	// 提取参数
	modelName, _ := args["modelName"].(string)
	targetMetric, _ := args["targetMetric"].(string)
	currentPeriodFilters, _ := args["currentPeriodFilters"].(map[string]interface{})
	basePeriodFilters, _ := args["basePeriodFilters"].(map[string]interface{})
	groupByDimensionsInterface, _ := args["groupByDimensions"].([]interface{})
	userId, _ := args["userId"].(string)
	enableDataValidation, _ := args["enableDataValidation"].(bool)

	// 默认启用数据验证
	if !enableDataValidation {
		enableDataValidation = true
	}

	var groupByDimensions []string
	for _, item := range groupByDimensionsInterface {
		if str, ok := item.(string); ok {
			groupByDimensions = append(groupByDimensions, str)
		}
	}

	return &SmartAnalyzerParams{
		ModelName:            modelName,
		TargetMetric:         targetMetric,
		CurrentPeriodFilters: currentPeriodFilters,
		BasePeriodFilters:    basePeriodFilters,
		GroupByDimensions:    groupByDimensions,
		UserId:               userId,
		EnableDataValidation: enableDataValidation,
	}, nil
}

// ParseDataScopeParams 解析数据范围探索工具参数
func (aim *AIInteractionManager) ParseDataScopeParams(arguments string) (*DataScopeParams, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("解析工具调用参数失败: %w", err)
	}

	// 提取参数
	modelName, _ := args["modelName"].(string)
	exploreDimensionsInterface, _ := args["exploreDimensions"].([]interface{})
	sampleFilters, _ := args["sampleFilters"].(map[string]interface{})
	userId, _ := args["userId"].(string)

	var exploreDimensions []string
	for _, item := range exploreDimensionsInterface {
		if str, ok := item.(string); ok {
			exploreDimensions = append(exploreDimensions, str)
		}
	}

	return &DataScopeParams{
		ModelName:         modelName,
		ExploreDimensions: exploreDimensions,
		SampleFilters:     sampleFilters,
		UserId:            userId,
	}, nil
}

// ParseAnonymizedAnalyzerParams 解析匿名化数据分析工具参数
func (aim *AIInteractionManager) ParseAnonymizedAnalyzerParams(arguments string) (*AnonymizedAnalyzerParams, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return nil, fmt.Errorf("解析工具调用参数失败: %w", err)
	}

	// 提取参数
	modelName, _ := args["modelName"].(string)
	targetMetric, _ := args["targetMetric"].(string)
	currentPeriodFilters, _ := args["currentPeriodFilters"].(map[string]interface{})
	basePeriodFilters, _ := args["basePeriodFilters"].(map[string]interface{})
	groupByDimensionsInterface, _ := args["groupByDimensions"].([]interface{})
	userId, _ := args["userId"].(string)

	var groupByDimensions []string
	for _, item := range groupByDimensionsInterface {
		if str, ok := item.(string); ok {
			groupByDimensions = append(groupByDimensions, str)
		}
	}

	return &AnonymizedAnalyzerParams{
		ModelName:            modelName,
		TargetMetric:         targetMetric,
		CurrentPeriodFilters: currentPeriodFilters,
		BasePeriodFilters:    basePeriodFilters,
		GroupByDimensions:    groupByDimensions,
		UserId:               userId,
	}, nil
}

// 私有方法

// getSemanticModelInfo 根据语义模型名称或ID获取详细信息
func (aim *AIInteractionManager) getSemanticModelInfo(semantic string) (string, error) {
	var models []sugar.SugarSemanticModels

	err := global.GVA_DB.Where("(name = ? OR id = ?) AND deleted_at IS NULL", semantic, semantic).Find(&models).Error
	if err != nil {
		return "", fmt.Errorf("查询语义模型失败: %w", err)
	}
	if len(models) == 0 {
		return "", fmt.Errorf("未找到语义模型: %s", semantic)
	}
	var builder strings.Builder
	for i, model := range models {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(fmt.Sprintf("模型名称: %s\n", aim.safeString(model.Name)))
		if model.Description != nil && *model.Description != "" {
			builder.WriteString(fmt.Sprintf("模型描述: %s\n", *model.Description))
		}
		if model.SourceTableName != nil && *model.SourceTableName != "" {
			builder.WriteString(fmt.Sprintf("数据表: %s\n", *model.SourceTableName))
		}
		if len(model.ParameterConfig) > 0 {
			paramInfo, err := aim.parseParameterConfig(model.ParameterConfig)
			if err != nil {
				global.GVA_LOG.Warn("解析参数配置失败", zap.Error(err))
			} else {
				builder.WriteString(fmt.Sprintf("可用筛选条件:\n%s\n", paramInfo))
			}
		}
		if len(model.ReturnableColumnsConfig) > 0 {
			columnInfo, err := aim.parseReturnableColumnsConfig(model.ReturnableColumnsConfig)
			if err != nil {
				global.GVA_LOG.Warn("解析返回字段配置失败", zap.Error(err))
			} else {
				builder.WriteString(fmt.Sprintf("可返回字段:\n%s", columnInfo))
			}
		}
	}
	return builder.String(), nil
}

// parseParameterConfig 解析参数配置JSON
func (aim *AIInteractionManager) parseParameterConfig(configJSON []byte) (string, error) {
	var config map[string]map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return "", fmt.Errorf("解析参数配置JSON失败: %w", err)
	}
	var builder strings.Builder
	for paramName, paramConfig := range config {
		builder.WriteString(fmt.Sprintf("  - %s: ", paramName))
		if desc, ok := paramConfig["description"].(string); ok {
			builder.WriteString(desc)
		}
		if paramType, ok := paramConfig["type"].(string); ok {
			builder.WriteString(fmt.Sprintf(" (类型: %s)", paramType))
		}
		if operator, ok := paramConfig["operator"].(string); ok {
			builder.WriteString(fmt.Sprintf(" [操作符: %s]", operator))
		}
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

// parseReturnableColumnsConfig 解析返回字段配置JSON
func (aim *AIInteractionManager) parseReturnableColumnsConfig(configJSON []byte) (string, error) {
	var config map[string]map[string]interface{}
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return "", fmt.Errorf("解析返回字段配置JSON失败: %w", err)
	}
	var builder strings.Builder
	for columnName, columnConfig := range config {
		builder.WriteString(fmt.Sprintf("  - %s: ", columnName))
		if desc, ok := columnConfig["description"].(string); ok {
			builder.WriteString(desc)
		}
		if columnType, ok := columnConfig["type"].(string); ok {
			builder.WriteString(fmt.Sprintf(" (类型: %s)", columnType))
		}
		builder.WriteString("\n")
	}
	return builder.String(), nil
}

// buildCompleteAnalysisMessage 构建完整的分析消息，包含用户提示词和系统提示词上下文
func (aim *AIInteractionManager) buildCompleteAnalysisMessage(dataText string, userDescription string, agent *sugar.SugarAgents) string {
	var builder strings.Builder

	// 首先包含用户的原始提示词作为分析背景
	builder.WriteString("用户分析需求：")
	builder.WriteString(userDescription)
	builder.WriteString("\n\n")

	// 如果Agent有特定的分析指导，也包含进来
	if agent.Prompt != nil && *agent.Prompt != "" {
		builder.WriteString("分析指导原则：")
		builder.WriteString(*agent.Prompt)
		builder.WriteString("\n\n")
	}

	// 然后提供匿名化数据
	builder.WriteString("请基于以下匿名化数据进行分析：\n\n")
	builder.WriteString("--- 匿名化数据 ---\n")
	builder.WriteString(dataText)
	builder.WriteString("\n--- 结束 ---\n\n")

	builder.WriteString("请结合用户需求和分析指导原则，对上述匿名化数据进行深入分析。")

	return builder.String()
}

// evaluateAnalysisQuality 评估分析结果质量
func (aim *AIInteractionManager) evaluateAnalysisQuality(response string, dataText string, userDescription string) float64 {
	qualityScore := 0.0
	maxScore := 100.0

	// 1. 结构完整性 (30分)
	structureScore := 0.0
	if strings.Contains(response, "整体趋势") || strings.Contains(response, "总体") {
		structureScore += 10.0
	}
	if strings.Contains(response, "驱动因子") || strings.Contains(response, "关键") {
		structureScore += 10.0
	}
	if strings.Contains(response, "建议") || strings.Contains(response, "洞察") {
		structureScore += 10.0
	}
	qualityScore += structureScore

	// 2. 数据引用度 (25分)
	dataReferenceScore := 0.0
	// 检查是否引用了具体的代号
	if strings.Contains(response, "D01") || strings.Contains(response, "D02") {
		dataReferenceScore += 10.0
	}
	// 检查是否引用了具体的百分比
	if strings.Contains(response, "%") {
		dataReferenceScore += 10.0
	}
	// 检查是否有量化描述
	if strings.Contains(response, "贡献度") {
		dataReferenceScore += 5.0
	}
	qualityScore += dataReferenceScore

	// 3. 逻辑连贯性 (20分)
	logicalScore := 0.0
	responseLen := len(response)
	if responseLen > 200 {
		logicalScore += 10.0
	}
	if responseLen > 500 {
		logicalScore += 10.0
	}
	qualityScore += logicalScore

	// 4. 问题相关性 (25分)
	relevanceScore := 0.0
	userWords := strings.Fields(strings.ToLower(userDescription))
	responseLower := strings.ToLower(response)

	matchedWords := 0
	for _, word := range userWords {
		if len(word) > 2 && strings.Contains(responseLower, word) {
			matchedWords++
		}
	}

	if len(userWords) > 0 {
		relevanceRatio := float64(matchedWords) / float64(len(userWords))
		relevanceScore = relevanceRatio * 25.0
	}
	qualityScore += relevanceScore

	// 计算最终得分
	finalScore := (qualityScore / maxScore) * 100.0
	if finalScore > 100.0 {
		finalScore = 100.0
	}

	global.GVA_LOG.Debug("分析质量评估详情",
		zap.Float64("structureScore", structureScore),
		zap.Float64("dataReferenceScore", dataReferenceScore),
		zap.Float64("logicalScore", logicalScore),
		zap.Float64("relevanceScore", relevanceScore),
		zap.Float64("finalScore", finalScore))

	return finalScore
}

// safeString 安全地获取字符串指针的值
func (aim *AIInteractionManager) safeString(str *string) string {
	if str == nil {
		return ""
	}
	return *str
}

// 参数结构体定义

// SmartAnalyzerParams 智能匿名化分析工具参数
type SmartAnalyzerParams struct {
	ModelName            string                 `json:"modelName"`
	TargetMetric         string                 `json:"targetMetric"`
	CurrentPeriodFilters map[string]interface{} `json:"currentPeriodFilters"`
	BasePeriodFilters    map[string]interface{} `json:"basePeriodFilters"`
	GroupByDimensions    []string               `json:"groupByDimensions"`
	UserId               string                 `json:"userId"`
	EnableDataValidation bool                   `json:"enableDataValidation"`
}

// DataScopeParams 数据范围探索工具参数
type DataScopeParams struct {
	ModelName         string                 `json:"modelName"`
	ExploreDimensions []string               `json:"exploreDimensions"`
	SampleFilters     map[string]interface{} `json:"sampleFilters"`
	UserId            string                 `json:"userId"`
}

// AnonymizedAnalyzerParams 匿名化数据分析工具参数
type AnonymizedAnalyzerParams struct {
	ModelName            string                 `json:"modelName"`
	TargetMetric         string                 `json:"targetMetric"`
	CurrentPeriodFilters map[string]interface{} `json:"currentPeriodFilters"`
	BasePeriodFilters    map[string]interface{} `json:"basePeriodFilters"`
	GroupByDimensions    []string               `json:"groupByDimensions"`
	UserId               string                 `json:"userId"`
}
